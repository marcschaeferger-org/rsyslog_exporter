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

type action struct {
	Name              string `json:"name"`
	Processed         int64  `json:"processed"`
	Failed            int64  `json:"failed"`
	Suspended         int64  `json:"suspended"`
	SuspendedDuration int64  `json:"suspended.duration"`
	Resumed           int64  `json:"resumed"`
}

func NewActionFromJSON(b []byte) (*action, error) {
	var pstat action
	err := json.Unmarshal(b, &pstat)
	if err != nil {
		return nil, fmt.Errorf("failed to decode action stat `%v`: %v", string(b), err)
	}
	return &pstat, nil
}

func (a *action) ToPoints() []*model.Point {
	points := make([]*model.Point, 5)

	points[0] = &model.Point{
		Name:        "action_processed",
		Type:        model.Counter,
		Value:       a.Processed,
		Description: "messages processed",
		LabelName:   "action",
		LabelValue:  a.Name,
	}

	points[1] = &model.Point{
		Name:        "action_failed",
		Type:        model.Counter,
		Value:       a.Failed,
		Description: "messages failed",
		LabelName:   "action",
		LabelValue:  a.Name,
	}

	points[2] = &model.Point{
		Name:        "action_suspended",
		Type:        model.Counter,
		Value:       a.Suspended,
		Description: "times suspended",
		LabelName:   "action",
		LabelValue:  a.Name,
	}

	points[3] = &model.Point{
		Name:        "action_suspended_duration",
		Type:        model.Counter,
		Value:       a.SuspendedDuration,
		Description: "time spent suspended",
		LabelName:   "action",
		LabelValue:  a.Name,
	}

	points[4] = &model.Point{
		Name:        "action_resumed",
		Type:        model.Counter,
		Value:       a.Resumed,
		Description: "times resumed",
		LabelName:   "action",
		LabelValue:  a.Name,
	}

	return points
}
