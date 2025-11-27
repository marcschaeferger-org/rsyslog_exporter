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
	"flag"
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

func main() {
	setupSyslog()

	flag.Parse()
	re := exporter.New()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Print("interrupt received, exiting")
		os.Exit(0)
	}()

	go func() {
		if err := re.Run(*silent); err != nil {
			log.Printf("exporter run ended with error: %v", err)
			os.Exit(1)
		}
		log.Print("exporter run ended normally, exiting")
		os.Exit(0)
	}()

	prometheus.MustRegister(re)
	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		// nolint:errcheck
		w.Write([]byte(`<html>
<head><title>Rsyslog exporter</title></head>
<body>
<h1>Rsyslog exporter</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
`))
	})

	// Configure server with sensible timeouts to mitigate slowloris and
	// similar DoS attacks. Use DefaultServeMux by leaving Handler nil.
	srv := &http.Server{
		Addr:              *listenAddress,
		Handler:           nil,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if *certPath == "" && *keyPath == "" {
		log.Printf("Listening on %s", *listenAddress)
		log.Fatal(srv.ListenAndServe())
	}
	if *certPath == "" || *keyPath == "" {
		log.Fatal("Both tls.server-crt and tls.server-key must be specified")
	}
	log.Printf("Listening for TLS on %s", *listenAddress)
	log.Fatal(srv.ListenAndServeTLS(*certPath, *keyPath))
}

func setupSyslog() {
	logwriter, e := syslog.New(syslog.LOG_NOTICE|syslog.LOG_SYSLOG, "rsyslog_exporter")
	if e == nil {
		log.SetOutput(logwriter)
	}
}
