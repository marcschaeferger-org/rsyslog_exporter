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
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type PointType int

const (
	Counter PointType = iota
	Gauge
)

type Point struct {
	Name        string
	Description string
	Type        PointType
	Value       int64
	LabelName   string
	LabelValue  string
}

func (p *Point) PromDescription() *prometheus.Desc {
	var variableLabels []string
	if p.PromLabelName() != "" {
		variableLabels = []string{p.PromLabelName()}
	}
	return prometheus.NewDesc(
		prometheus.BuildFQName("", "rsyslog", p.Name),
		p.Description,
		variableLabels,
		nil,
	)
}

func (p *Point) PromType() prometheus.ValueType {
	if p.Type == Counter {
		return prometheus.CounterValue
	}
	return prometheus.GaugeValue
}

func (p *Point) PromValue() float64 {
	return float64(p.Value)
}

func (p *Point) PromLabelValue() string {
	return p.LabelValue
}

func (p *Point) PromLabelName() string {
	return p.LabelName
}

func (p *Point) Key() string {
	if p.LabelValue == "" {
		return p.Name
	}
	return fmt.Sprintf("%s.%s", p.Name, p.LabelValue)
}
