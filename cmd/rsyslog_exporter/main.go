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
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	exporter "github.com/prometheus-community/rsyslog_exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress = flag.String("web.listen-address", ":9104", "Address to listen on for web interface and telemetry.")
	metricPath    = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	certPath      = flag.String("tls.server-crt", "", "Path to PEM encoded file containing TLS server cert.")
	keyPath       = flag.String("tls.server-key", "", "Path to PEM encoded file containing TLS server key (unencrypted).")
	silent        = flag.Bool("silent", false, "Disable logging of errors in handling stats lines")
)

// test hooks
var (
	// newSyslog remains injectable for tests.
	newSyslog = func(priority syslog.Priority, tag string) (io.Writer, error) { return syslog.New(priority, tag) }
	// exitOnErr logs and exits with code 1 by default; tests can override.
	exitOnErr = func(err error) { log.Printf("fatal: %v", err); osExit(1) }
	// osExit allows tests to intercept os.Exit calls.
	osExit = os.Exit
	// makeRootContext allows tests to control the root context used by main.
	makeRootContext = func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) }
)

func setupSyslog() io.Writer {
	w, err := newSyslog(syslog.LOG_NOTICE|syslog.LOG_SYSLOG, "rsyslog_exporter")
	if err == nil && w != nil {
		log.SetOutput(w)
		return w
	}
	return nil
}

func main() {
	_ = setupSyslog()
	flag.Parse()
	re := exporter.New()

	// root context for the application; cancel on shutdown to allow
	// future components to observe cancellation.
	ctx, cancel := makeRootContext()
	defer cancel()

	// start exporter loop (reads stdin until EOF). Pass root context so
	// it can be canceled on shutdown.
	go func() {
		if err := re.Run(ctx, *silent); err != nil {
			log.Printf("exporter run ended with error: %v", err)
		} else {
			log.Print("exporter run ended normally")
		}
	}()

	mux := http.NewServeMux()
	// use a fresh registry to avoid double registration during tests,
	// but register the standard collectors so runtime/process metrics
	// are exposed in production.
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registerHandlers(mux, *metricPath, re, reg)

	srv := buildServer(*listenAddress, mux)

	// start the HTTP server asynchronously and get an error channel.
	serverErrC := startServerAsync(srv, *listenAddress, *certPath, *keyPath)

	// listen for SIGINT and SIGTERM and trigger graceful shutdown.
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigC:
		log.Printf("signal received: %v, shutting down", sig)
		// give the server up to 5s to shutdown cleanly
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := shutdownServer(srv, shutdownCtx); err != nil {
			log.Printf("error during server shutdown: %v", err)
		} else {
			log.Print("server shutdown complete")
		}
		// cancel root context so other components can stop if wired up
		cancel()
		// ensure the shutdown timeout context is cancelled before exiting
		shutdownCancel()
		osExit(0)
	case err := <-serverErrC:
		// server terminated on its own; if it's a real error, report it.
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			exitOnErr(err)
		}
	case <-ctx.Done():
		// defensive: if root context is canceled, attempt shutdown as above
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := shutdownServer(srv, shutdownCtx); err != nil {
			log.Printf("error during server shutdown: %v", err)
		}
		// ensure the shutdown timeout context is cancelled before exiting
		shutdownCancel()
		osExit(0)
	}
}

// registerHandlers wires endpoints onto mux using provided registry.
func registerHandlers(mux *http.ServeMux, metricPath string, re *exporter.Exporter, reg *prometheus.Registry) {
	// safe register: ignore AlreadyRegistered
	_ = reg.Register(re)
	mux.Handle(metricPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		// nolint:errcheck
		w.Write([]byte(`<html>
<head><title>Rsyslog exporter</title></head>
<body>
<h1>Rsyslog exporter</h1>
<p><a href='` + metricPath + `'>Metrics</a></p>
</body>
</html>
`))
	})
}

func buildServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

var startServerAsync = func(srv *http.Server, listenAddr, certPath, keyPath string) <-chan error {
	errC := make(chan error, 1)

	go func() {
		if certPath == "" && keyPath == "" {
			log.Printf("Listening on %s", listenAddr)
			errC <- srv.ListenAndServe()
			return
		}
		if certPath == "" || keyPath == "" {
			errC <- errors.New("both tls.server-crt and tls.server-key must be specified")
			return
		}
		log.Printf("Listening for TLS on %s", listenAddr)
		errC <- srv.ListenAndServeTLS(certPath, keyPath)
	}()

	return errC
}

// (old setupSyslog removed; use the injectable setupSyslog above)

// shutdownServer is injectable for tests to simulate server shutdown behavior.
var shutdownServer = func(srv *http.Server, ctx context.Context) error {
	return srv.Shutdown(ctx)
}
