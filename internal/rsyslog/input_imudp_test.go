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
	inputIMUDPLog = []byte(`{ "name": "test_input_imudp", "origin": "imudp", "called.recvmmsg":1000, "called.recvmsg":2000, "msgs.received":500}`)
)

func TestGetInputIMUDP(t *testing.T) {
	if got := GetStatType(inputIMUDPLog); got != TypeInputIMDUP {
		t.Errorf(th.DetectedTypeFmt, TypeInputIMDUP, got)
	}
	pstat, err := NewInputIMUDPFromJSON(inputIMUDPLog)
	if err != nil {
		t.Fatalf("parse input imudp stat failed: %v", err)
	}
	th.AssertEqString(t, "name", "test_input_imudp", pstat.Name)
	th.AssertEqInt(t, "recvmmsg", 1000, pstat.Recvmmsg)
	th.AssertEqInt(t, "recvmsg", 2000, pstat.Recvmsg)
	th.AssertEqInt(t, "received", 500, pstat.Received)
}

func TestInputIMUDPtoPoints(t *testing.T) {
	pstat, err := NewInputIMUDPFromJSON(inputIMUDPLog)
	if err != nil {
		t.Fatalf("parse input imudp stat failed: %v", err)
	}
	points := pstat.ToPoints()
	expected := []struct {
		name  string
		value int64
	}{
		{"input_called_recvmmsg", 1000},
		{"input_called_recvmsg", 2000},
		{"input_received", 500},
	}
	if len(points) != len(expected) {
		t.Fatalf(th.ExpectedPointsFmt, len(expected), len(points))
	}
	for i, exp := range expected {
		th.AssertEqString(t, exp.name+" name", exp.name, points[i].Name)
		th.AssertEqInt(t, exp.name+" value", exp.value, points[i].Value)
		th.AssertEqString(t, exp.name+" label", "test_input_imudp", points[i].LabelValue)
	}
}
