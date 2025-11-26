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
	"testing"

	"github.com/prometheus-community/rsyslog_exporter/internal/model"
)

var (
	queueStat = []byte(`{"name":"main Q","size":10,"enqueued":20,"full":30,"discarded.full":40,"discarded.nf":50,"maxqsize":60}`)
)

func TestNewQueueFromJSON(t *testing.T) {
	logType := GetStatType(queueStat)
	if logType != TypeQueue {
		t.Errorf("detected pstat type should be %d but is %d", TypeQueue, logType)
	}

	pstat, err := NewQueueFromJSON([]byte(queueStat))
	if err != nil {
		t.Fatalf("expected parsing queue stat not to fail, got: %v", err)
	}

	if want, got := "main Q", pstat.Name; want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := int64(10), pstat.Size; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	if want, got := int64(20), pstat.Enqueued; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	if want, got := int64(30), pstat.Full; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	if want, got := int64(40), pstat.DiscardedFull; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	if want, got := int64(50), pstat.DiscardedNf; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	if want, got := int64(60), pstat.MaxQsize; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}
}

func TestQueueToPoints(t *testing.T) {
	pstat, err := NewQueueFromJSON([]byte(queueStat))
	if err != nil {
		t.Fatalf("expected parsing queue stat not to fail, got: %v", err)
	}
	points := pstat.ToPoints()

	type expectation struct {
		idx        int
		name       string
		value      int64
		metricType model.PointType
		labelValue string
	}
	expected := []expectation{
		{0, "queue_size", 10, model.Gauge, "main Q"},
		{1, "queue_enqueued", 20, model.Counter, "main Q"},
		{2, "queue_full", 30, model.Counter, "main Q"},
		{3, "queue_discarded_full", 40, model.Counter, "main Q"},
		{4, "queue_discarded_not_full", 50, model.Counter, "main Q"},
		{5, "queue_max_size", 60, model.Gauge, "main Q"},
	}

	for _, exp := range expected {
		if exp.idx >= len(points) {
			t.Fatalf("expected point index %d to exist", exp.idx)
		}
		pt := points[exp.idx]
		if pt.Name != exp.name {
			t.Errorf("idx %d: want name %s got %s", exp.idx, exp.name, pt.Name)
		}
		if pt.Value != exp.value {
			t.Errorf("%s: want value %d got %d", exp.name, exp.value, pt.Value)
		}
		if pt.Type != exp.metricType {
			t.Errorf("%s: want type %d got %d", exp.name, exp.metricType, pt.Type)
		}
		if pt.LabelValue != exp.labelValue {
			t.Errorf("%s: want label %s got %s", exp.name, exp.labelValue, pt.LabelValue)
		}
	}
}
