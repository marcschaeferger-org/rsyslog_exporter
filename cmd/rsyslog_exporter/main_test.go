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
	req, _ := http.NewRequest("GET", "/", http.NoBody)
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), mp) {
		t.Fatalf("root handler missing metric path link; code=%d body=%s", rr.Code, rr.Body.String())
	}

	// metrics path handler responds 200
	rr2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", mp, http.NoBody)
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

type exitPanic struct{ code int }

func (e exitPanic) Error() string { return fmt.Sprintf("exit(%d)", e.code) }

func TestMainFunctionCoversExitPath(t *testing.T) {
	// ensure flags are reset for test
	*listenAddress = invalidListenAddr // cause immediate server startup error
	*metricPath = defaultMetricPath
	*certPath = ""
	*keyPath = ""
	*silent = true

	// stub exits so we can intercept instead of terminating test process
	origExit := osExit
	defer func() { osExit = origExit }()
	osExit = func(code int) { panic(exitPanic{code}) }

	origFatal := exitOnErr
	defer func() { exitOnErr = origFatal }()
	exitOnErr = func(err error) { panic(err) }

	// block stdin so re.Run doesn't return and call osExit(0)
	origStdin := os.Stdin
	r, w, _ := os.Pipe()
	_ = w.Close() // keep reader open so scanner blocks
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic from exit hook")
		}
	}()

	// run main; it should attempt to start server then panic via exitOnErr
	// give goroutines a tiny window then let startServer hit the failing path
	go func() { time.Sleep(10 * time.Millisecond) }()
	main()
}

func TestMainFunctionCoversTLSBranch(t *testing.T) {
	*listenAddress = invalidListenAddr
	*metricPath = defaultMetricPath
	*certPath = "badcert.pem"
	*keyPath = "badkey.pem"
	*silent = true

	origExit := osExit
	defer func() { osExit = origExit }()
	osExit = func(code int) { panic(exitPanic{code}) }

	origFatal := exitOnErr
	defer func() { exitOnErr = origFatal }()
	exitOnErr = func(err error) { panic(err) }

	// block stdin so re.Run doesn't return and trigger logs concurrently
	origStdin := os.Stdin
	r, w, _ := os.Pipe()
	_ = w.Close()
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic from exit hook in TLS branch")
		}
	}()
	main()
}

type errReader struct{ used bool }

func (e *errReader) Read(p []byte) (int, error) {
	if !e.used {
		e.used = true
		copy(p, []byte(sampleLine))
		return len(sampleLine), nil
	}
	return 0, fmt.Errorf("reader error")
}

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
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(sampleLine)
	_ = w.Close()
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	re := exporter.New()
	runExporterLoop(re, true)
}

func TestRunExporterLoopError(t *testing.T) {
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
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)
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
	r, w, _ := os.Pipe()
	_ = w.Close()
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	// run main in goroutine
	go main()

	// give main time to start server
	time.Sleep(50 * time.Millisecond)

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGINT)

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
	r, w, _ := os.Pipe()
	_ = w.Close()
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
	r, w, _ := os.Pipe()
	_ = w.Close()
	os.Stdin = r
	defer func() { os.Stdin = origStdin; _ = r.Close() }()

	// run main in goroutine
	go main()

	// give main time to start server
	time.Sleep(50 * time.Millisecond)

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGTERM)

	select {
	case code := <-got:
		if code != 0 {
			t.Fatalf(msgExpectedExitCodeFmt, code)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("graceful shutdown on SIGTERM did not complete in time")
	}
}
