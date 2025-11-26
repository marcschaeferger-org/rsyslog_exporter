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

// decoder turns a raw impstats JSON buffer into points.
type decoder func([]byte) ([]*model.Point, error)

// statDecoders maps stat Types to their respective decoding functions.
var statDecoders = map[rsyslog.Type]decoder{
	rsyslog.TypeAction: func(b []byte) ([]*model.Point, error) {
		a, err := rsyslog.NewActionFromJSON(b)
		if err != nil {
			return nil, err
		}
		return a.ToPoints(), nil
	},
	rsyslog.TypeInput: func(b []byte) ([]*model.Point, error) {
		i, err := rsyslog.NewInputFromJSON(b)
		if err != nil {
			return nil, err
		}
		return i.ToPoints(), nil
	},
	rsyslog.TypeInputIMDUP: func(b []byte) ([]*model.Point, error) {
		u, err := rsyslog.NewInputIMUDPFromJSON(b)
		if err != nil {
			return nil, err
		}
		return u.ToPoints(), nil
	},
	rsyslog.TypeQueue: func(b []byte) ([]*model.Point, error) {
		q, err := rsyslog.NewQueueFromJSON(b)
		if err != nil {
			return nil, err
		}
		return q.ToPoints(), nil
	},
	rsyslog.TypeResource: func(b []byte) ([]*model.Point, error) {
		r, err := rsyslog.NewResourceFromJSON(b)
		if err != nil {
			return nil, err
		}
		return r.ToPoints(), nil
	},
	rsyslog.TypeDynStat: func(b []byte) ([]*model.Point, error) {
		s, err := rsyslog.NewDynStatFromJSON(b)
		if err != nil {
			return nil, err
		}
		return s.ToPoints(), nil
	},
	rsyslog.TypeDynafileCache: func(b []byte) ([]*model.Point, error) {
		d, err := rsyslog.NewDynafileCacheFromJSON(b)
		if err != nil {
			return nil, err
		}
		return d.ToPoints(), nil
	},
	rsyslog.TypeForward: func(b []byte) ([]*model.Point, error) {
		f, err := rsyslog.NewForwardFromJSON(b)
		if err != nil {
			return nil, err
		}
		return f.ToPoints(), nil
	},
	rsyslog.TypeKubernetes: func(b []byte) ([]*model.Point, error) {
		k, err := rsyslog.NewKubernetesFromJSON(b)
		if err != nil {
			return nil, err
		}
		return k.ToPoints(), nil
	},
	rsyslog.TypeOmkafka: func(b []byte) ([]*model.Point, error) {
		o, err := rsyslog.NewOmkafkaFromJSON(b)
		if err != nil {
			return nil, err
		}
		return o.ToPoints(), nil
	},
}

func (re *Exporter) handleStatLine(rawbuf []byte) error {
	s := bytes.SplitN(rawbuf, []byte(" "), 4)
	if len(s) != 4 {
		return fmt.Errorf("failed to split log line, expected 4 columns, got: %v", len(s))
	}
	buf := s[3]
	pstatType := rsyslog.GetStatType(buf)
	dec, ok := statDecoders[pstatType]
	if !ok {
		return fmt.Errorf("unknown pstat type: %v", pstatType)
	}
	points, err := dec(buf)
	if err != nil {
		return err
	}
	for _, p := range points {
		if err := re.Set(p); err != nil {
			return err
		}
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

func (re *Exporter) run(silent bool) error {
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
		return err
	}
	log.Print("input ended, returning from run")
	return nil
}

// Run starts the exporter loop. Exported for use by the cmd package.
// It returns when stdin scanning ends; callers (e.g. main) should decide whether to exit the process.
func (re *Exporter) Run(silent bool) error {
	return re.run(silent)
}
