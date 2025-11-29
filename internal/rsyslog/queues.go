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

// Queue represents rsyslog queue statistics.
type Queue struct {
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	Enqueued      int64  `json:"enqueued"`
	Full          int64  `json:"full"`
	DiscardedFull int64  `json:"discarded.full"`
	DiscardedNf   int64  `json:"discarded.nf"`
	MaxQsize      int64  `json:"maxqsize"`
}

func NewQueueFromJSON(b []byte) (*Queue, error) {
	var pstat Queue
	err := json.Unmarshal(b, &pstat)
	if err != nil {
		return nil, fmt.Errorf("failed to decode queue stat `%v`: %w", string(b), err)
	}
	return &pstat, nil
}

func (q *Queue) ToPoints() []*model.Point {
	points := make([]*model.Point, 6)

	points[0] = &model.Point{
		Name:        "queue_size",
		Type:        model.Gauge,
		Value:       q.Size,
		Description: "messages currently in queue",
		LabelName:   "queue",
		LabelValue:  q.Name,
	}

	points[1] = &model.Point{
		Name:        "queue_enqueued",
		Type:        model.Counter,
		Value:       q.Enqueued,
		Description: "total messages enqueued",
		LabelName:   "queue",
		LabelValue:  q.Name,
	}

	points[2] = &model.Point{
		Name:        "queue_full",
		Type:        model.Counter,
		Value:       q.Full,
		Description: "times queue was full",
		LabelName:   "queue",
		LabelValue:  q.Name,
	}

	points[3] = &model.Point{
		Name:        "queue_discarded_full",
		Type:        model.Counter,
		Value:       q.DiscardedFull,
		Description: "messages discarded due to queue being full",
		LabelName:   "queue",
		LabelValue:  q.Name,
	}

	points[4] = &model.Point{
		Name:        "queue_discarded_not_full",
		Type:        model.Counter,
		Value:       q.DiscardedNf,
		Description: "messages discarded when queue not full",
		LabelName:   "queue",
		LabelValue:  q.Name,
	}

	points[5] = &model.Point{
		Name:        "queue_max_size",
		Type:        model.Gauge,
		Value:       q.MaxQsize,
		Description: "maximum size queue has reached",
		LabelName:   "queue",
		LabelValue:  q.Name,
	}

	return points
}
