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

package exporter

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"

	// sync not needed here; store provides locking

	"github.com/prometheus-community/rsyslog_exporter/internal/model"
	"github.com/prometheus-community/rsyslog_exporter/internal/rsyslog"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter collects and exposes rsyslog impstats metrics.
type Exporter struct {
	scanner *bufio.Scanner
	*model.Store
}

func newExporter() *Exporter {
	e := &Exporter{
		scanner: bufio.NewScanner(os.Stdin),
		Store:   model.NewStore(),
	}
	return e
}

// New returns an initialized Exporter.
func New() *Exporter { // exported for tests
	return newExporter()
}

func (re *Exporter) handleStatLine(rawbuf []byte) error {
	s := bytes.SplitN(rawbuf, []byte(" "), 4)
	if len(s) != 4 {
		return fmt.Errorf("failed to split log line, expected 4 columns, got: %v", len(s))
	}
	buf := s[3]

	pstatType := rsyslog.GetStatType(buf)

	switch pstatType {
	case rsyslog.TypeAction:
		a, err := rsyslog.NewActionFromJSON(buf)
		if err != nil {
			return err
		}
		for _, p := range a.ToPoints() {
			if err := re.Set(p); err != nil {
				return err
			}
		}

	case rsyslog.TypeInput:
		i, err := rsyslog.NewInputFromJSON(buf)
		if err != nil {
			return err
		}
		for _, p := range i.ToPoints() {
			if err := re.Set(p); err != nil {
				return err
			}
		}

	case rsyslog.TypeInputIMDUP:
		u, err := rsyslog.NewInputIMUDPFromJSON(buf)
		if err != nil {
			return err
		}
		for _, p := range u.ToPoints() {
			if err := re.Set(p); err != nil {
				return err
			}
		}

	case rsyslog.TypeQueue:
		q, err := rsyslog.NewQueueFromJSON(buf)
		if err != nil {
			return err
		}
		for _, p := range q.ToPoints() {
			if err := re.Set(p); err != nil {
				return err
			}
		}

	case rsyslog.TypeResource:
		r, err := rsyslog.NewResourceFromJSON(buf)
		if err != nil {
			return err
		}
		for _, p := range r.ToPoints() {
			if err := re.Set(p); err != nil {
				return err
			}
		}
	case rsyslog.TypeDynStat:
		s, err := rsyslog.NewDynStatFromJSON(buf)
		if err != nil {
			return err
		}
		for _, p := range s.ToPoints() {
			if err := re.Set(p); err != nil {
				return err
			}
		}
	case rsyslog.TypeDynafileCache:
		d, err := rsyslog.NewDynafileCacheFromJSON(buf)
		if err != nil {
			return err
		}
		for _, p := range d.ToPoints() {
			if err := re.Set(p); err != nil {
				return err
			}
		}
	case rsyslog.TypeForward:
		f, err := rsyslog.NewForwardFromJSON(buf)
		if err != nil {
			return err
		}
		for _, p := range f.ToPoints() {
			if err := re.Set(p); err != nil {
				return err
			}
		}
	case rsyslog.TypeKubernetes:
		k, err := rsyslog.NewKubernetesFromJSON(buf)
		if err != nil {
			return err
		}
		for _, p := range k.ToPoints() {
			if err := re.Set(p); err != nil {
				return err
			}
		}
	case rsyslog.TypeOmkafka:
		o, err := rsyslog.NewOmkafkaFromJSON(buf)
		if err != nil {
			return err
		}
		for _, p := range o.ToPoints() {
			if err := re.Set(p); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unknown pstat type: %v", pstatType)
	}
	return nil
}

// Describe sends the description of currently known metrics collected
// by this Collector to the provided channel. Note that this implementation
// does not necessarily send the "super-set of all possible descriptors" as
// defined by the Collector interface spec, depending on the timing of when
// it is called. The rsyslog exporter does not know all possible metrics
// it will export until the first full batch of rsyslog impstats messages
// are received via stdin. This is ok for now.
func (re *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc(
		prometheus.BuildFQName("", "rsyslog", "scrapes"),
		"times exporter has been scraped",
		nil, nil,
	)

	keys := re.Keys()

	for _, k := range keys {
		p, err := re.Get(k)
		if err == nil {
			ch <- p.PromDescription()
		} else {
			log.Printf("describe: failed to get point for key %s: %v", k, err)
		}
	}
}

// Collect is called by Prometheus when collecting metrics.
func (re *Exporter) Collect(ch chan<- prometheus.Metric) {
	keys := re.Keys()

	for _, k := range keys {
		p, err := re.Get(k)
		if err != nil {
			continue
		}

		labelValues := []string{}
		if p.PromLabelValue() != "" {
			labelValues = []string{p.PromLabelValue()}
		}
		metric := prometheus.MustNewConstMetric(
			p.PromDescription(),
			p.PromType(),
			p.PromValue(),
			labelValues...,
		)

		ch <- metric
	}
}

func (re *Exporter) run(silent bool) {
	errorPoint := &model.Point{
		Name:        "stats_line_errors",
		Type:        model.Counter,
		Description: "Counts errors during stats line handling",
	}
	// nolint:errcheck
	re.Set(errorPoint)
	for re.scanner.Scan() {
		err := re.handleStatLine(re.scanner.Bytes())
		if err != nil {
			errorPoint.Value += 1
			if !silent {
				log.Printf("error handling stats line: %v, line was: %s", err, re.scanner.Bytes())
			}
		}
	}
	if err := re.scanner.Err(); err != nil {
		log.Printf("error reading input: %v", err)
	}
	log.Print("input ended, exiting normally")
	os.Exit(0)
}

// Run starts the exporter loop. Exported for use by the cmd package.
func (re *Exporter) Run(silent bool) {
	re.run(silent)
}
