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

type resource struct {
	Name     string `json:"name"`
	Utime    int64  `json:"utime"`
	Stime    int64  `json:"stime"`
	Maxrss   int64  `json:"maxrss"`
	Minflt   int64  `json:"minflt"`
	Majflt   int64  `json:"majflt"`
	Inblock  int64  `json:"inblock"`
	Outblock int64  `json:"oublock"`
	Nvcsw    int64  `json:"nvcsw"`
	Nivcsw   int64  `json:"nivcsw"`
}

func NewResourceFromJSON(b []byte) (*resource, error) {
	var pstat resource
	err := json.Unmarshal(b, &pstat)
	if err != nil {
		return nil, fmt.Errorf("failed to decode resource stat `%v`: %v", string(b), err)
	}
	return &pstat, nil
}

func (r *resource) ToPoints() []*model.Point {
	points := make([]*model.Point, 9)

	points[0] = &model.Point{
		Name:        "resource_utime",
		Type:        model.Counter,
		Value:       r.Utime,
		Description: "user time used in microseconds",
		LabelName:   "resource",
		LabelValue:  r.Name,
	}

	points[1] = &model.Point{
		Name:        "resource_stime",
		Type:        model.Counter,
		Value:       r.Stime,
		Description: "system time used in microsends",
		LabelName:   "resource",
		LabelValue:  r.Name,
	}

	points[2] = &model.Point{
		Name:        "resource_maxrss",
		Type:        model.Gauge,
		Value:       r.Maxrss,
		Description: "maximum resident set size",
		LabelName:   "resource",
		LabelValue:  r.Name,
	}

	points[3] = &model.Point{
		Name:        "resource_minflt",
		Type:        model.Counter,
		Value:       r.Minflt,
		Description: "total minor faults",
		LabelName:   "resource",
		LabelValue:  r.Name,
	}

	points[4] = &model.Point{
		Name:        "resource_majflt",
		Type:        model.Counter,
		Value:       r.Majflt,
		Description: "total major faults",
		LabelName:   "resource",
		LabelValue:  r.Name,
	}

	points[5] = &model.Point{
		Name:        "resource_inblock",
		Type:        model.Counter,
		Value:       r.Inblock,
		Description: "filesystem input operations",
		LabelName:   "resource",
		LabelValue:  r.Name,
	}

	points[6] = &model.Point{
		Name:        "resource_oublock",
		Type:        model.Counter,
		Value:       r.Outblock,
		Description: "filesystem output operations",
		LabelName:   "resource",
		LabelValue:  r.Name,
	}

	points[7] = &model.Point{
		Name:        "resource_nvcsw",
		Type:        model.Counter,
		Value:       r.Nvcsw,
		Description: "voluntary context switches",
		LabelName:   "resource",
		LabelValue:  r.Name,
	}

	points[8] = &model.Point{
		Name:        "resource_nivcsw",
		Type:        model.Counter,
		Value:       r.Nivcsw,
		Description: "involuntary context switches",
		LabelName:   "resource",
		LabelValue:  r.Name,
	}

	return points
}
