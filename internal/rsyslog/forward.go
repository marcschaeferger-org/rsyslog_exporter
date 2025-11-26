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

// Forward represents rsyslog forwarding statistics.
type Forward struct {
	Name      string `json:"name"`
	BytesSent int64  `json:"bytes.sent"`
}

func NewForwardFromJSON(b []byte) (*Forward, error) {
	var pstat Forward
	err := json.Unmarshal(b, &pstat)
	if err != nil {
		return nil, fmt.Errorf("failed to decode forward stat `%v`: %v", string(b), err)
	}
	return &pstat, nil
}

func (f *Forward) ToPoints() []*model.Point {
	points := make([]*model.Point, 1)

	points[0] = &model.Point{
		Name:        "forward_bytes_total",
		Type:        model.Counter,
		Value:       f.BytesSent,
		Description: "bytes forwarded to destination",
		LabelName:   "destination",
		LabelValue:  f.Name,
	}

	return points
}
