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

// InputIMUDP represents rsyslog imudp input worker statistics.
type InputIMUDP struct {
	Name     string `json:"name"`
	Recvmmsg int64  `json:"called.recvmmsg"`
	Recvmsg  int64  `json:"called.recvmsg"`
	Received int64  `json:"msgs.received"`
}

func NewInputIMUDPFromJSON(b []byte) (*InputIMUDP, error) {
	var pstat InputIMUDP
	err := json.Unmarshal(b, &pstat)
	if err != nil {
		return nil, fmt.Errorf("error decoding input stat `%v`: %v", string(b), err)
	}
	return &pstat, nil
}

func (i *InputIMUDP) ToPoints() []*model.Point {
	points := make([]*model.Point, 3)

	points[0] = &model.Point{
		Name:        "input_called_recvmmsg",
		Type:        model.Counter,
		Value:       i.Recvmmsg,
		Description: "Number of recvmmsg called",
		LabelName:   "worker",
		LabelValue:  i.Name,
	}
	points[1] = &model.Point{
		Name:        "input_called_recvmsg",
		Type:        model.Counter,
		Value:       i.Recvmsg,
		Description: "Number of recvmsg called",
		LabelName:   "worker",
		LabelValue:  i.Name,
	}

	points[2] = &model.Point{
		Name:        "input_received",
		Type:        model.Counter,
		Value:       i.Received,
		Description: "messages received",
		LabelName:   "worker",
		LabelValue:  i.Name,
	}

	return points
}
