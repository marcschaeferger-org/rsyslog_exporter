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
	"strings"

	"github.com/prometheus-community/rsyslog_exporter/internal/model"
)

// DfcStat represents rsyslog dynafile cache statistics.
type DfcStat struct {
	Name          string `json:"name"`
	Origin        string `json:"origin"`
	Requests      int64  `json:"requests"`
	Level0        int64  `json:"level0"`
	Missed        int64  `json:"missed"`
	Evicted       int64  `json:"evicted"`
	MaxUsed       int64  `json:"maxused"`
	CloseTimeouts int64  `json:"closetimeouts"`
}

func NewDynafileCacheFromJSON(b []byte) (*DfcStat, error) {
	var pstat DfcStat
	err := json.Unmarshal(b, &pstat)
	if err != nil {
		return nil, fmt.Errorf("error decoding dynafile cache stat `%v`: %w", string(b), err)
	}
	pstat.Name = strings.TrimPrefix(pstat.Name, "dynafile cache ")
	return &pstat, nil
}

func (d *DfcStat) ToPoints() []*model.Point {
	points := make([]*model.Point, 6)

	points[0] = &model.Point{
		Name:        "dynafile_cache_requests",
		Type:        model.Counter,
		Value:       d.Requests,
		Description: "number of requests made to obtain a dynafile",
		LabelName:   "cache",
		LabelValue:  d.Name,
	}
	points[1] = &model.Point{
		Name:        "dynafile_cache_level0",
		Type:        model.Counter,
		Value:       d.Level0,
		Description: "number of requests for the current active file",
		LabelName:   "cache",
		LabelValue:  d.Name,
	}
	points[2] = &model.Point{
		Name:        "dynafile_cache_missed",
		Type:        model.Counter,
		Value:       d.Missed,
		Description: "number of cache misses",
		LabelName:   "cache",
		LabelValue:  d.Name,
	}
	points[3] = &model.Point{
		Name:        "dynafile_cache_evicted",
		Type:        model.Counter,
		Value:       d.Evicted,
		Description: "number of times a file needed to be evicted from cache",
		LabelName:   "cache",
		LabelValue:  d.Name,
	}
	points[4] = &model.Point{
		Name:        "dynafile_cache_maxused",
		Type:        model.Counter,
		Value:       d.MaxUsed,
		Description: "maximum number of cache entries ever used",
		LabelName:   "cache",
		LabelValue:  d.Name,
	}
	points[5] = &model.Point{
		Name:        "dynafile_cache_closetimeouts",
		Type:        model.Counter,
		Value:       d.CloseTimeouts,
		Description: "number of times a file was closed due to timeout settings",
		LabelName:   "cache",
		LabelValue:  d.Name,
	}

	return points
}
