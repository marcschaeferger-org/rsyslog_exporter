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

// Input represents generic rsyslog input statistics.
type Input struct {
	Name      string `json:"name"`
	Submitted int64  `json:"submitted"`
}

func NewInputFromJSON(b []byte) (*Input, error) {
	var pstat Input
	err := json.Unmarshal(b, &pstat)
	if err != nil {
		return nil, fmt.Errorf("error decoding input stat `%v`: %v", string(b), err)
	}
	return &pstat, nil
}

func (i *Input) ToPoints() []*model.Point {
	points := make([]*model.Point, 1)

	points[0] = &model.Point{
		Name:        "input_submitted",
		Type:        model.Counter,
		Value:       i.Submitted,
		Description: "messages submitted",
		LabelName:   "input",
		LabelValue:  i.Name,
	}

	return points
}
