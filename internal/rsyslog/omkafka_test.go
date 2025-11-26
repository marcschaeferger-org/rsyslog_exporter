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

package rsyslog

import (
	"fmt"
	"testing"

	"github.com/prometheus-community/rsyslog_exporter/internal/model"
)

var (
	omkafkaLog = []byte(`{ "name": "omkafka", "origin": "omkafka", "submitted": 59, "maxoutqsize": 9, "failures": 0, "topicdynacache.skipped": 57, "topicdynacache.miss": 2, "topicdynacache.evicted": 0, "acked": 55, "failures_msg_too_large": 0, "failures_unknown_topic": 0, "failures_queue_full": 0, "failures_unknown_partition": 0, "failures_other": 0, "errors_timed_out": 0, "errors_transport": 0, "errors_broker_down": 0, "errors_auth": 0, "errors_ssl": 0, "errors_other": 0, "rtt_avg_usec": 0, "throttle_avg_msec": 0, "int_latency_avg_usec": 0 }`)
)

func TestNewOmkafkaFromJSON(t *testing.T) {
	logType := GetStatType(omkafkaLog)
	if logType != TypeOmkafka {
		t.Errorf("detected pstat type should be %d but is %d", TypeOmkafka, logType)
	}

	_, err := NewOmkafkaFromJSON([]byte(omkafkaLog))
	if err != nil {
		t.Fatalf("expected parsing action not to fail, got: %v", err)
	}
}

func TestOmkafkaToPoints(t *testing.T) {
	pstat, err := NewOmkafkaFromJSON([]byte(omkafkaLog))
	if err != nil {
		t.Fatalf("expected parsing action not to fail, got: %v", err)
	}
	points := pstat.ToPoints()

	testCases := []*model.Point{
		{
			Name:       "input_submitted",
			Type:       model.Counter,
			Value:      59,
			LabelValue: "omkafka",
		},
		{
			Name:       "omkafka_messages",
			Type:       model.Counter,
			Value:      59,
			LabelValue: "submitted",
		},
		{
			Name:  "omkafka_maxoutqsize",
			Type:  model.Counter,
			Value: 9,
		},
		{
			Name:       "omkafka_messages",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "failures",
		},
		{
			Name:       "omkafka_topicdynacache",
			Type:       model.Counter,
			Value:      57,
			LabelValue: "skipped",
		},
		{
			Name:       "omkafka_topicdynacache",
			Type:       model.Counter,
			Value:      2,
			LabelValue: "miss",
		},
		{
			Name:       "omkafka_topicdynacache",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "evicted",
		},
		{
			Name:       "omkafka_messages",
			Type:       model.Counter,
			Value:      55,
			LabelValue: "acked",
		},
		{
			Name:       "omkafka_failures",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "msg_too_large",
		},

		{
			Name:       "omkafka_failures",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "unknown_topic",
		},
		{
			Name:       "omkafka_failures",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "queue_full",
		},
		{
			Name:       "omkafka_failures",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "unknown_partition",
		},
		{
			Name:       "omkafka_failures",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "other",
		},
		{
			Name:       "omkafka_errors",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "timed_out",
		},
		{
			Name:       "omkafka_errors",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "transport",
		},
		{
			Name:       "omkafka_errors",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "broker_down",
		},
		{
			Name:       "omkafka_errors",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "auth",
		},
		{
			Name:       "omkafka_errors",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "ssl",
		},
		{
			Name:       "omkafka_errors",
			Type:       model.Counter,
			Value:      0,
			LabelValue: "other",
		},
		{
			Name:  "omkafka_rtt_avg_usec_avg",
			Type:  model.Gauge,
			Value: 0,
		},
		{
			Name:  "omkafka_throttle_avg_msec_avg",
			Type:  model.Gauge,
			Value: 0,
		},
		{
			Name:  "omkafka_int_latency_avg_usec_avg",
			Type:  model.Gauge,
			Value: 0,
		},
	}

	for idx, tc := range testCases {
		t.Run(fmt.Sprintf("point idx %d", idx), func(t *testing.T) {
			p := points[idx]
			if p.Name != tc.Name {
				t.Errorf("got name %s; want %s", p.Name, tc.Name)
			}
			if p.Type != tc.Type {
				t.Errorf("got type %d; want %d", p.Type, tc.Type)
			}
			if p.Value != tc.Value {
				t.Errorf("got value %d; want %d", p.Value, tc.Value)
			}
			if p.LabelValue != tc.LabelValue {
				t.Errorf("got label value %s; want %s", p.LabelValue, tc.LabelValue)
			}
		})
	}

}
