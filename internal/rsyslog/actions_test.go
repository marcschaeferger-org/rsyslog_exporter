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
	th "github.com/prometheus-community/rsyslog_exporter/internal/testhelpers"
)

var (
	actionLog = []byte(`{"name":"test_action","processed":100000,"failed":2,"suspended":1,"suspended.duration":1000,"resumed":1}`)
)

func TestNewActionFromJSON(t *testing.T) {
	logType := GetStatType(actionLog)
	if logType != TypeAction {
		t.Errorf(th.DetectedTypeFmt, TypeAction, logType)
	}

	pstat, err := NewActionFromJSON([]byte(actionLog))
	if err != nil {
		t.Fatalf("expected parsing action not to fail, got: %v", err)
	}

	if want, got := "test_action", pstat.Name; want != got {
		t.Errorf(th.WantStringFmt, want, got)
	}

	if want, got := int64(100000), pstat.Processed; want != got {
		t.Errorf(th.WantedIntFmt, want, got)
	}

	if want, got := int64(2), pstat.Failed; want != got {
		t.Errorf(th.WantedIntFmt, want, got)
	}

	if want, got := int64(1), pstat.Suspended; want != got {
		t.Errorf(th.WantedIntFmt, want, got)
	}

	if want, got := int64(1000), pstat.SuspendedDuration; want != got {
		t.Errorf(th.WantedIntFmt, want, got)
	}

	if want, got := int64(1), pstat.Resumed; want != got {
		t.Errorf(th.WantedIntFmt, want, got)
	}
}

func TestActionToPoints(t *testing.T) {
	pstat, err := NewActionFromJSON([]byte(actionLog))
	if err != nil {
		t.Fatalf("expected parsing action not to fail, got: %v", err)
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
		{0, "action_processed", 100000, model.Counter, "test_action"},
		{1, "action_failed", 2, model.Counter, "test_action"},
		{2, "action_suspended", 1, model.Counter, "test_action"},
		{3, "action_suspended_duration", 1000, model.Counter, "test_action"},
		{4, "action_resumed", 1, model.Counter, "test_action"},
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
