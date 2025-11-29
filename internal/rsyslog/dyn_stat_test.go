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
	"reflect"
	"testing"

	"github.com/prometheus-community/rsyslog_exporter/internal/model"
	th "github.com/prometheus-community/rsyslog_exporter/internal/testhelpers"
)

func TestGetDynStat(t *testing.T) {
	log := []byte(`{ "name": "global", "origin": "dynstats", "values": { "` + th.MsgPerHostOpsOverflow + `": 1, "` + th.MsgPerHostNewMetricAdd + `": 3, "` + th.MsgPerHostNoMetric + `": 0, "` + th.MsgPerHostMetricsPurged + `": 0, "` + th.MsgPerHostOpsIgnored + `": 0 } }`)
	values := map[string]int64{
		th.MsgPerHostOpsOverflow:   1,
		th.MsgPerHostNewMetricAdd:  3,
		th.MsgPerHostNoMetric:      0,
		th.MsgPerHostMetricsPurged: 0,
		th.MsgPerHostOpsIgnored:    0,
	}

	if got := StatType(log); got != TypeDynStat {
		t.Errorf(th.DetectedTypeFmt, TypeDynStat, got)
	}

	pstat, err := NewDynStatFromJSON(log)
	if err != nil {
		t.Fatalf("expected parsing dynamic stat not to fail, got: %v", err)
	}

	th.AssertEqString(t, "dynstat name", "global", pstat.Name)

	if want, got := values, pstat.Values; !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected values, want: %+v got: %+v", want, got)
	}
}

func TestDynStatToPoints(t *testing.T) {
	log := []byte(`{ "name": "global", "origin": "dynstats", "values": { "msg_per_host.ops_overflow": 1, "msg_per_host.new_metric_add": 3, "msg_per_host.no_metric": 0, "msg_per_host.metrics_purged": 0, "msg_per_host.ops_ignored": 0 } }`)
	wants := map[string]model.Point{
		"msg_per_host.ops_overflow": {
			Name:        "dynstat_global",
			Type:        model.Counter,
			Value:       1,
			Description: th.DynamicStatisticBucketDescription,
			LabelName:   "counter",
			LabelValue:  th.MsgPerHostOpsOverflow,
		},
		"msg_per_host.new_metric_add": {
			Name:        "dynstat_global",
			Type:        model.Counter,
			Value:       3,
			Description: th.DynamicStatisticBucketDescription,
			LabelName:   "counter",
			LabelValue:  th.MsgPerHostNewMetricAdd,
		},
		"msg_per_host.no_metric": {
			Name:        "dynstat_global",
			Type:        model.Counter,
			Value:       0,
			Description: th.DynamicStatisticBucketDescription,
			LabelName:   "counter",
			LabelValue:  th.MsgPerHostNoMetric,
		},
		"msg_per_host.metrics_purged": {
			Name:        "dynstat_global",
			Type:        model.Counter,
			Value:       0,
			Description: th.DynamicStatisticBucketDescription,
			LabelName:   "counter",
			LabelValue:  th.MsgPerHostMetricsPurged,
		},
		"msg_per_host.ops_ignored": {
			Name:        "dynstat_global",
			Type:        model.Counter,
			Value:       0,
			Description: th.DynamicStatisticBucketDescription,
			LabelName:   "counter",
			LabelValue:  th.MsgPerHostOpsIgnored,
		},
	}

	seen := map[string]bool{}
	for name := range wants {
		seen[name] = false
	}

	pstat, err := NewDynStatFromJSON(log)
	if err != nil {
		t.Fatalf("expected parsing dyn stat not to fail, got: %v", err)
	}

	points := pstat.ToPoints()
	for _, got := range points {
		key := got.LabelValue
		want, ok := wants[key]
		if !ok {
			t.Errorf("unexpected point, got: %+v", got)
			continue
		}

		if !reflect.DeepEqual(want, *got) {
			t.Errorf("expected point to be %+v, got %+v", want, got)
		}

		if seen[key] {
			t.Errorf("point seen multiple times: %+v", got)
		}
		seen[key] = true
	}

	for name, ok := range seen {
		if !ok {
			t.Errorf("expected to see point with key %s, but did not", name)
		}
	}
}
