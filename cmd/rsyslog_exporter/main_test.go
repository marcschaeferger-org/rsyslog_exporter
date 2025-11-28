// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	exporter "github.com/prometheus-community/rsyslog_exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	invalidListenAddr         = "*********:-1"
	anyListenZero             = ":0"
	defaultMetricPath         = "/metrics"
	sampleLine                = "col1 col2 col3 {\"submitted\":1}\n"
	msgUnexpectedExitOnErrFmt = "unexpected exitOnErr: %v"
	msgExpectedExitCodeFmt    = "expected exit code 0, got %d"
	msgPipeFailedFmt          = "os.Pipe failed: %v"
	msgPipeCloseFailedFmt     = "failed to close pipe writer: %v"
	msgFindProcessFailedFmt   = "FindProcess failed: %v"
)

func TestSetupSyslogFallback(t *testing.T) {
	// Save original and restore at the end
	orig := newSyslog
	defer func() { newSyslog = orig }()

	// make newSyslog return an error to simulate unavailability
	newSyslog = func(_ syslog.Priority, _ string) (io.Writer, error) {
		return nil, errors.New("syslog unavailable")
	}

	w := setupSyslog()
	if w != nil {
		t.Fatalf("expected nil writer when syslog unavailable, got %v", w)
	}
}

func TestSetupSyslogSuccess(t *testing.T) {
	orig := newSyslog
	defer func() { newSyslog = orig }()
	buf := &bytes.Buffer{}
	newSyslog = func(_ syslog.Priority, _ string) (io.Writer, error) {
		return buf, nil
	}
	w := setupSyslog()
	if w == nil {
		t.Fatalf("expected writer, got nil")
	}
	log.Print("hello syslog")
	out := buf.String()
	if !strings.Contains(out, "hello syslog") {
		t.Fatalf("expected log output to be written to buffer, got: %s", out)
	}
}

func TestRegisterHandlersWithCustomMetricPath(t *testing.T) {
	// custom metric path
	mp := "/custommetrics"
	re := exporter.New()
	mux := http.NewServeMux()
	reg := prometheus.NewRegistry()
	registerHandlers(mux, mp, re, reg)

	// root handler returns HTML with link
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", http.NoBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), mp) {
		t.Fatalf("root handler missing metric path link; code=%d body=%s", rr.Code, rr.Body.String())
	}

	// metrics path handler responds 200
	rr2 := httptest.NewRecorder()
	req2, err := http.NewRequest("GET", mp, http.NoBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != 200 {
		t.Fatalf("metrics handler returned %d", rr2.Code)
	}
}

func TestBuildServerConfig(t *testing.T) {
	mux := http.NewServeMux()
	srv := buildServer(":0", mux)
	if srv.Addr == "" || srv.Handler == nil {
		t.Fatalf("server not configured")
	}
	if srv.ReadHeaderTimeout == 0 || srv.ReadTimeout == 0 || srv.WriteTimeout == 0 || srv.IdleTimeout == 0 {
		t.Fatalf("timeouts not set")
	}
}

func TestStartServerNoTLSImmediateError(t *testing.T) {
	origExit := exitOnErr
	defer func() { exitOnErr = origExit }()
	var gotErr error
	exitOnErr = func(err error) { gotErr = err }

	mux := http.NewServeMux()
	srv := buildServer(invalidListenAddr, mux) // invalid port forces immediate error
	startServer(srv, srv.Addr, "", "")
	if gotErr == nil {
		t.Fatalf("expected error from ListenAndServe")
	}
}

func TestStartServerTLSMissingOneFlag(t *testing.T) {
	origExit := exitOnErr
	defer func() { exitOnErr = origExit }()
	var got error
	exitOnErr = func(err error) { got = err }
	mux := http.NewServeMux()
	srv := buildServer(":0", mux)
	startServer(srv, srv.Addr, "cert.pem", "")
	if got == nil || got.Error() != "both tls.server-crt and tls.server-key must be specified" {
		t.Fatalf("unexpected error: %v", got)
	}
}

func TestStartServerTLSBothProvided(t *testing.T) {
	origExit := exitOnErr
	defer func() { exitOnErr = origExit }()
	var got error
	exitOnErr = func(err error) { got = err }
	mux := http.NewServeMux()
	srv := buildServer(":0", mux)
	startServer(srv, srv.Addr, "no-such-cert.pem", "no-such-key.pem")
	if got == nil {
		t.Fatalf("expected TLS serve error")
	}
}

// exitPanic removed: tests use channel-based interception for osExit now.

func TestMainFunctionCoversExitPath(t *testing.T) {
	// ensure flags are reset for test
	*listenAddress = invalidListenAddr // cause immediate server startup error
	*metricPath = defaultMetricPath
	*certPath = ""
	*keyPath = ""
	*silent = true

	// intercept osExit and exitOnErr so we can observe their invocation
	origExit := osExit
	defer func() { osExit = origExit }()
	gotExit := make(chan int, 1)
	osExit = func(code int) { gotExit <- code }

	origFatal := exitOnErr
	defer func() { exitOnErr = origFatal }()
	gotErr := make(chan error, 1)
	exitOnErr = func(err error) { gotErr <- err }

	// block stdin so re.Run doesn't return and call osExit(0)
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(msgPipeFailedFmt, err)
	}
	if err := w.Close(); err != nil { // keep reader open so scanner blocks
		t.Fatalf(msgPipeCloseFailedFmt, err)
	}
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	// run main and wait for either exitOnErr or osExit to be invoked
	go main()
	select {
	case e := <-gotErr:
		if e == nil {
			t.Fatalf("expected error from exitOnErr, got nil")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("exit hook was not invoked in time")
	}
}

func TestMainFunctionCoversTLSBranch(t *testing.T) {
	*listenAddress = invalidListenAddr
	*metricPath = defaultMetricPath
	*certPath = "badcert.pem"
	*keyPath = "badkey.pem"
	*silent = true

	origExit := osExit
	defer func() { osExit = origExit }()
	gotExit := make(chan int, 1)
	osExit = func(code int) { gotExit <- code }

	origFatal := exitOnErr
	defer func() { exitOnErr = origFatal }()
	gotErr := make(chan error, 1)
	exitOnErr = func(err error) { gotErr <- err }

	// block stdin so re.Run doesn't return and trigger logs concurrently
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(msgPipeFailedFmt, err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf(msgPipeCloseFailedFmt, err)
	}
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	go main()
	select {
	case e := <-gotErr:
		if e == nil {
			t.Fatalf("expected error from exitOnErr in TLS branch, got nil")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("exit hook was not invoked in time for TLS branch")
	}
}

// (previous errReader helper removed as unused)

// Test-only helpers (moved from testhooks_test.go) to avoid staticcheck warnings
var exporterRunHook = func(re *exporter.Exporter, silent bool) error {
	return re.Run(context.Background(), silent)
}

func runExporterLoop(re *exporter.Exporter, silent bool) {
	if err := exporterRunHook(re, silent); err != nil {
		log.Printf("exporter run ended with error: %v", err)
		return
	}
	log.Print("exporter run ended normally")
}

// startServer is a blocking wrapper used by tests; it starts server async and
// waits for the first error then forwards it to exitOnErr (keeps previous
// test behavior).
func startServer(srv *http.Server, listenAddr, certPath, keyPath string) {
	errC := startServerAsync(srv, listenAddr, certPath, keyPath)
	err := <-errC
	exitOnErr(err)
}

func runInterruptWatcher() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Print("interrupt received")
}

func TestRunExporterLoopNormal(t *testing.T) {
	t.Helper()
	// Prepare stdin for exporter.New() so its scanner reads one line and EOF
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(msgPipeFailedFmt, err)
	}
	if _, err := w.WriteString(sampleLine); err != nil {
		t.Fatalf("failed to write to pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf(msgPipeCloseFailedFmt, err)
	}
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	re := exporter.New()
	runExporterLoop(re, true)
}

func TestRunExporterLoopError(_ *testing.T) {
	origHook := exporterRunHook
	defer func() { exporterRunHook = origHook }()
	exporterRunHook = func(_ *exporter.Exporter, _ bool) error { return fmt.Errorf("boom") }
	re := exporter.New()
	runExporterLoop(re, true) // should hit error branch and return
}

func TestRunInterruptWatcher(t *testing.T) {
	// Run watcher and send an interrupt
	done := make(chan struct{})
	go func() { runInterruptWatcher(); close(done) }()
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf(msgFindProcessFailedFmt, err)
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatalf("failed to send interrupt: %v", err)
	}
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("interrupt watcher did not return in time")
	}
}

func TestGracefulShutdownSIGINT(t *testing.T) {
	// configure flags
	*listenAddress = anyListenZero
	*metricPath = defaultMetricPath
	*certPath = ""
	*keyPath = ""
	*silent = true

	// intercept osExit
	origExit := osExit
	defer func() { osExit = origExit }()
	got := make(chan int, 1)
	osExit = func(code int) { got <- code }

	// intercept exitOnErr to fail the test if invoked
	origFatal := exitOnErr
	defer func() { exitOnErr = origFatal }()
	exitOnErr = func(err error) { t.Fatalf(msgUnexpectedExitOnErrFmt, err) }

	// block stdin so exporter.Run doesn't return immediately
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(msgPipeFailedFmt, err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf(msgPipeCloseFailedFmt, err)
	}
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	// run main in goroutine
	go main()

	// give main time to start server
	time.Sleep(50 * time.Millisecond)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf(msgFindProcessFailedFmt, err)
	}
	if err := p.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	select {
	case code := <-got:
		if code != 0 {
			t.Fatalf(msgExpectedExitCodeFmt, code)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("graceful shutdown on SIGINT did not complete in time")
	}
}

func TestContextDonePath(t *testing.T) {
	*listenAddress = ":0"
	*metricPath = defaultMetricPath
	*certPath = ""
	*keyPath = ""
	*silent = true

	// capture osExit
	origExit := osExit
	defer func() { osExit = origExit }()
	got := make(chan int, 1)
	osExit = func(code int) { got <- code }

	// never call exitOnErr
	origFatal := exitOnErr
	defer func() { exitOnErr = origFatal }()
	exitOnErr = func(err error) { t.Fatalf(msgUnexpectedExitOnErrFmt, err) }

	// block stdin so exporter loop doesn't finish on its own
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(msgPipeFailedFmt, err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf(msgPipeCloseFailedFmt, err)
	}
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	// override root context to auto-cancel shortly after start
	origMk := makeRootContext
	defer func() { makeRootContext = origMk }()
	makeRootContext = func() (context.Context, context.CancelFunc) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(20 * time.Millisecond)
			cancel()
		}()
		return ctx, cancel
	}

	go main()
	select {
	case code := <-got:
		if code != 0 {
			t.Fatalf(msgExpectedExitCodeFmt, code)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("context-done path did not exit in time")
	}
}

func TestGracefulShutdownSIGTERM(t *testing.T) {
	// configure flags
	*listenAddress = anyListenZero
	*metricPath = defaultMetricPath
	*certPath = ""
	*keyPath = ""
	*silent = true

	// intercept osExit
	origExit := osExit
	defer func() { osExit = origExit }()
	got := make(chan int, 1)
	osExit = func(code int) { got <- code }

	// intercept exitOnErr to fail the test if invoked
	origFatal := exitOnErr
	defer func() { exitOnErr = origFatal }()
	exitOnErr = func(err error) { t.Fatalf(msgUnexpectedExitOnErrFmt, err) }

	// block stdin so exporter.Run doesn't return immediately
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(msgPipeFailedFmt, err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf(msgPipeCloseFailedFmt, err)
	}
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	// run main in goroutine
	go main()

	// give main time to start server
	time.Sleep(50 * time.Millisecond)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf(msgFindProcessFailedFmt, err)
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	select {
	case code := <-got:
		if code != 0 {
			t.Fatalf(msgExpectedExitCodeFmt, code)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("graceful shutdown on SIGTERM did not complete in time")
	}
}

func TestMainServerErrorPath(t *testing.T) {
	// configure flags to not interfere with startup
	*listenAddress = ":0"
	*metricPath = defaultMetricPath
	*certPath = ""
	*keyPath = ""
	*silent = true

	// stub startServerAsync to return a channel that yields an error
	origStart := startServerAsync
	defer func() { startServerAsync = origStart }()
	errC := make(chan error, 1)
	startServerAsync = func(_ *http.Server, _ string, _ string, _ string) <-chan error {
		return errC
	}

	// capture exitOnErr invocation
	origExitOnErr := exitOnErr
	defer func() { exitOnErr = origExitOnErr }()
	gotErr := make(chan error, 1)
	exitOnErr = func(err error) { gotErr <- err }

	// block stdin to keep exporter.Run from returning immediately
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(msgPipeFailedFmt, err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf(msgPipeCloseFailedFmt, err)
	}
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	// run main in goroutine
	go main()

	// send an error on the server channel
	testErr := errors.New("server failed")
	errC <- testErr

	select {
	case e := <-gotErr:
		if e == nil || e.Error() != testErr.Error() {
			t.Fatalf("unexpected error forwarded to exitOnErr: %v", e)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("exitOnErr was not called in time")
	}
}

func TestMainShutdownError(t *testing.T) {
	*listenAddress = anyListenZero
	*metricPath = defaultMetricPath
	*certPath = ""
	*keyPath = ""
	*silent = true

	// stub shutdownServer to return an error
	origShutdown := shutdownServer
	defer func() { shutdownServer = origShutdown }()
	shutdownServer = func(_ *http.Server, _ context.Context) error { return errors.New("shutdown failed") }

	// intercept osExit to prevent exiting and capture code
	origExit := osExit
	defer func() { osExit = origExit }()
	got := make(chan int, 1)
	osExit = func(code int) { got <- code }

	// intercept exitOnErr to fail the test if invoked
	origFatal := exitOnErr
	defer func() { exitOnErr = origFatal }()
	exitOnErr = func(err error) { t.Fatalf(msgUnexpectedExitOnErrFmt, err) }

	// block stdin
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf(msgPipeFailedFmt, err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf(msgPipeCloseFailedFmt, err)
	}
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	go main()
	time.Sleep(50 * time.Millisecond)
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf(msgFindProcessFailedFmt, err)
	}
	if err := p.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	select {
	case code := <-got:
		if code != 0 {
			t.Fatalf(msgExpectedExitCodeFmt, code)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("shutdown error path did not complete in time")
	}
}
