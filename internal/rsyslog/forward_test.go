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

	th "github.com/prometheus-community/rsyslog_exporter/internal/testhelpers"

	"github.com/prometheus-community/rsyslog_exporter/internal/model"
)

var (
	forwardLog = []byte(`{ "name": "TCP-FQDN-6514", "origin": "omfwd", "bytes.sent": 666 }`)
)

func TestNewForwardFromJSON(t *testing.T) {
	if got := GetStatType(forwardLog); got != TypeForward {
		t.Errorf(th.DetectedTypeFmt, TypeForward, got)
	}
	pstat, err := NewForwardFromJSON(forwardLog)
	if err != nil {
		t.Fatalf("parse forward stat failed: %v", err)
	}
	th.AssertEqString(t, "name", "TCP-FQDN-6514", pstat.Name)
	th.AssertEqInt(t, "bytes_sent", 666, pstat.BytesSent)
}

func TestForwardToPoints(t *testing.T) {
	pstat, err := NewForwardFromJSON(forwardLog)
	if err != nil {
		t.Fatalf("parse forward stat failed: %v", err)
	}
	points := pstat.ToPoints()
	if len(points) != 1 {
		t.Fatalf(th.ExpectedPointsFmt, 1, len(points))
	}
	p := points[0]
	th.AssertEqString(t, "point name", "forward_bytes_total", p.Name)
	th.AssertEqInt(t, "point value", 666, p.Value)
	th.AssertEqString(t, "point label", "TCP-FQDN-6514", p.LabelValue)
	if p.Type != model.Counter {
		t.Errorf("point type: expected %v, got %v", model.Counter, p.Type)
	}
}
