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
	"encoding/json"
	"fmt"

	"github.com/prometheus-community/rsyslog_exporter/internal/model"
)

// DynStat represents rsyslog dynamic statistics buckets.
type DynStat struct {
	Name   string           `json:"name"`
	Origin string           `json:"origin"`
	Values map[string]int64 `json:"values"`
}

func NewDynStatFromJSON(b []byte) (*DynStat, error) {
	var pstat DynStat
	err := json.Unmarshal(b, &pstat)
	if err != nil {
		return nil, fmt.Errorf("error decoding values stat `%v`: %w", string(b), err)
	}
	return &pstat, nil
}

func (i *DynStat) ToPoints() []*model.Point {
	// Preallocate a slice with capacity to avoid multiple allocations from append.
	points := make([]*model.Point, 0, len(i.Values))
	for name, value := range i.Values {
		points = append(points, &model.Point{
			Name:        fmt.Sprintf("dynstat_%s", i.Name),
			Type:        model.Counter,
			Value:       value,
			Description: fmt.Sprintf("dynamic statistic bucket %s", i.Name),
			LabelName:   "counter",
			LabelValue:  name,
		})
	}

	return points
}
