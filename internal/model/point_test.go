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

package model

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCounter(t *testing.T) {
	p1 := &Point{
		Name:  "my_counter",
		Type:  Counter,
		Value: int64(10),
	}

	if want, got := float64(10), p1.PromValue(); want != got {
		t.Errorf("want '%f', got '%f'", want, got)
	}

	if want, got := prometheus.ValueType(1), p1.PromType(); want != got {
		t.Errorf("want '%v', got '%v'", want, got)
	}

	wanted := `Desc{fqName: "rsyslog_my_counter", help: "", constLabels: {}, variableLabels: {}}`
	if want, got := wanted, p1.PromDescription().String(); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestGauge(t *testing.T) {
	p1 := &Point{
		Name:  "my_gauge",
		Type:  Gauge,
		Value: int64(10),
	}

	if want, got := float64(10), p1.PromValue(); want != got {
		t.Errorf("want '%f', got '%f'", want, got)
	}

	if want, got := prometheus.ValueType(2), p1.PromType(); want != got {
		t.Errorf("want '%v', got '%v'", want, got)
	}

	wanted := `Desc{fqName: "rsyslog_my_gauge", help: "", constLabels: {}, variableLabels: {}}`
	if want, got := wanted, p1.PromDescription().String(); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

}

func TestPromLabelValueAndKey(t *testing.T) {
	p := &Point{
		Name:       "foo",
		Type:       Gauge,
		Value:      7,
		LabelName:  "lbl",
		LabelValue: "v1",
	}

	if want, got := "v1", p.PromLabelValue(); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := "lbl", p.PromLabelName(); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := "foo.v1", p.Key(); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	// when LabelValue is empty, Key should be just the name
	p.LabelValue = ""
	if want, got := "foo", p.Key(); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestPromDescriptionWithLabel(t *testing.T) {
	p := &Point{Name: "foo", Description: "bar", LabelName: "lbl", LabelValue: "v"}
	d := p.PromDescription().String()
	if want := "variableLabels: {lbl}"; !strings.Contains(d, want) {
		t.Fatalf("expected %q in description: %s", want, d)
	}
}
