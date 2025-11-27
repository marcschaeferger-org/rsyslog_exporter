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
)

var (
	inputLog = []byte(`{"name":"test_input", "origin":"imuxsock", "submitted":1000}`)
)

func TestGetInput(t *testing.T) {
	logType := GetStatType(inputLog)
	if logType != TypeInput {
		t.Errorf(th.DetectedTypeFmt, TypeInput, logType)
	}

	pstat, err := NewInputFromJSON([]byte(inputLog))
	if err != nil {
		t.Fatalf("expected parsing input stat not to fail, got: %v", err)
	}

	if want, got := "test_input", pstat.Name; want != got {
		t.Errorf(th.WantStringFmt, want, got)
	}

	if want, got := int64(1000), pstat.Submitted; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}
}

func TestInputtoPoints(t *testing.T) {
	pstat, err := NewInputFromJSON([]byte(inputLog))
	if err != nil {
		t.Fatalf("expected parsing input stat not to fail, got: %v", err)
	}

	points := pstat.ToPoints()

	point := points[0]
	if want, got := "input_submitted", point.Name; want != got {
		t.Errorf(th.WantStringFmt, want, got)
	}

	if want, got := int64(1000), point.Value; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	if want, got := "test_input", point.LabelValue; want != got {
		t.Errorf(th.WantStringFmt, want, got)
	}
}
