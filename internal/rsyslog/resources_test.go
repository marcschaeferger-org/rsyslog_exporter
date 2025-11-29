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
	resourceLog = []byte(`{"name":"` + th.ResourceUsage + `","utime":10,"stime":20,"maxrss":30,"minflt":40,"majflt":50,"inblock":60,"outblock":70,"nvcsw":80,"nivcsw":90}`)
)

func TestNewResourceFromJSON(t *testing.T) {
	logType := StatType(resourceLog)
	if logType != TypeResource {
		t.Errorf(th.DetectedTypeFmt, TypeResource, logType)
	}

	pstat, err := NewResourceFromJSON([]byte(resourceLog))
	if err != nil {
		t.Fatalf("expected parsing resource stat not to fail, got: %v", err)
	}

	if want, got := th.ResourceUsage, pstat.Name; want != got {
		t.Errorf(th.WantStringFmt, want, got)
	}

	if want, got := int64(10), pstat.Utime; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	if want, got := int64(20), pstat.Stime; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	if want, got := int64(30), pstat.Maxrss; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	if want, got := int64(40), pstat.Minflt; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	if want, got := int64(50), pstat.Majflt; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	if want, got := int64(60), pstat.Inblock; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	if want, got := int64(70), pstat.Outblock; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	if want, got := int64(80), pstat.Nvcsw; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	if want, got := int64(90), pstat.Nivcsw; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}
}

func TestResourceToPoints(t *testing.T) {
	pstat, err := NewResourceFromJSON([]byte(resourceLog))
	if err != nil {
		t.Fatalf("expected parsing resource stat not to fail, got: %v", err)
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
		{0, "resource_utime", 10, model.Counter, th.ResourceUsage},
		{1, "resource_stime", 20, model.Counter, th.ResourceUsage},
		{2, "resource_maxrss", 30, model.Gauge, th.ResourceUsage},
		{3, "resource_minflt", 40, model.Counter, th.ResourceUsage},
		{4, "resource_majflt", 50, model.Counter, th.ResourceUsage},
		{5, "resource_inblock", 60, model.Counter, th.ResourceUsage},
		{6, "resource_oublock", 70, model.Counter, th.ResourceUsage},
		{7, "resource_nvcsw", 80, model.Counter, th.ResourceUsage},
		{8, "resource_nivcsw", 90, model.Counter, th.ResourceUsage},
	}

	for _, exp := range expected {
		if exp.idx >= len(points) {
			t.Fatalf(th.ExpectedIndexFmt, exp.idx)
		}
		pt := points[exp.idx]
		want := th.PointExpectation{Name: exp.name, Type: int(exp.metricType), Value: exp.value, Label: exp.labelValue}
		got := th.PointExpectation{Name: pt.Name, Type: int(pt.Type), Value: pt.Value, Label: pt.LabelValue}
		th.AssertPointFields(t, exp.idx, want, got)
	}
}
