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
	"errors"
	"flag"
	"io"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"time"

	exporter "github.com/prometheus-community/rsyslog_exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
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
	// exitOnErr allows tests to intercept fatal exits without os.Exit.
	exitOnErr = func(err error) { log.Fatal(err) }
	// osExit allows tests to intercept os.Exit calls.
	osExit = os.Exit
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

	go runInterruptWatcher()
	go runExporterLoop(re, *silent)

	mux := http.NewServeMux()
	registerHandlers(mux, *metricPath, re, prometheus.NewRegistry()) // use a fresh registry to avoid double registration during tests

	srv := buildServer(*listenAddress, mux)
	startServer(srv, *listenAddress, *certPath, *keyPath)
}

func runInterruptWatcher() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Print("interrupt received")
}

func runExporterLoop(re *exporter.Exporter, silent bool) {
	if err := re.Run(silent); err != nil {
		log.Printf("exporter run ended with error: %v", err)
		return
	}
	log.Print("exporter run ended normally")
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

func startServer(srv *http.Server, listenAddr, certPath, keyPath string) {
	if certPath == "" && keyPath == "" {
		log.Printf("Listening on %s", listenAddr)
		exitOnErr(srv.ListenAndServe())
		return
	}
	if certPath == "" || keyPath == "" {
		exitOnErr(errors.New("Both tls.server-crt and tls.server-key must be specified"))
		return
	}
	log.Printf("Listening for TLS on %s", listenAddr)
	exitOnErr(srv.ListenAndServeTLS(certPath, keyPath))
}

// (old setupSyslog removed; use the injectable setupSyslog above)
