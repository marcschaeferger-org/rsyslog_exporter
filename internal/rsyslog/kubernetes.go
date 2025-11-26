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
	"regexp"

	"github.com/prometheus-community/rsyslog_exporter/internal/model"
)

var (
	apiNameRegexp = regexp.MustCompile(`mmkubernetes\((\S+)\)`)
)

type kubernetes struct {
	Name                  string `json:"name"`
	Url                   string
	RecordSeen            int64 `json:"recordseen"`
	NamespaceMetaSuccess  int64 `json:"namespacemetadatasuccess"`
	NamespaceMetaNotFound int64 `json:"namespacemetadatanotfound"`
	NamespaceMetaBusy     int64 `json:"namespacemetadatabusy"`
	NamespaceMetaError    int64 `json:"namespacemetadataerror"`
	PodMetaSuccess        int64 `json:"podmetadatasuccess"`
	PodMetaNotFound       int64 `json:"podmetadatanotfound"`
	PodMetaBusy           int64 `json:"podmetadatabusy"`
	PodMetaError          int64 `json:"podmetadataerror"`
}

func NewKubernetesFromJSON(b []byte) (*kubernetes, error) {
	var pstat kubernetes
	err := json.Unmarshal(b, &pstat)
	if err != nil {
		return nil, fmt.Errorf("failed to decode kubernetes stat `%v`: %v", string(b), err)
	}
	matches := apiNameRegexp.FindSubmatch([]byte(pstat.Name))
	if matches != nil {
		pstat.Url = string(matches[1])
	}
	return &pstat, nil
}

func (k *kubernetes) ToPoints() []*model.Point {
	points := make([]*model.Point, 9)

	points[0] = &model.Point{
		Name:        "kubernetes_namespace_metadata_success_total",
		Type:        model.Counter,
		Value:       k.NamespaceMetaSuccess,
		Description: "successful fetches of namespace metadata",
		LabelName:   "url",
		LabelValue:  k.Url,
	}

	points[1] = &model.Point{
		Name:        "kubernetes_namespace_metadata_notfound_total",
		Type:        model.Counter,
		Value:       k.NamespaceMetaNotFound,
		Description: "notfound fetches of namespace metadata",
		LabelName:   "url",
		LabelValue:  k.Url,
	}

	points[2] = &model.Point{
		Name:        "kubernetes_namespace_metadata_busy_total",
		Type:        model.Counter,
		Value:       k.NamespaceMetaBusy,
		Description: "busy fetches of namespace metadata",
		LabelName:   "url",
		LabelValue:  k.Url,
	}

	points[3] = &model.Point{
		Name:        "kubernetes_namespace_metadata_error_total",
		Type:        model.Counter,
		Value:       k.NamespaceMetaError,
		Description: "error fetches of namespace metadata",
		LabelName:   "url",
		LabelValue:  k.Url,
	}

	points[4] = &model.Point{
		Name:        "kubernetes_pod_metadata_success_total",
		Type:        model.Counter,
		Value:       k.PodMetaSuccess,
		Description: "successful fetches of pod metadata",
		LabelName:   "url",
		LabelValue:  k.Url,
	}

	points[5] = &model.Point{
		Name:        "kubernetes_pod_metadata_notfound_total",
		Type:        model.Counter,
		Value:       k.PodMetaNotFound,
		Description: "notfound fetches of pod metadata",
		LabelName:   "url",
		LabelValue:  k.Url,
	}

	points[6] = &model.Point{
		Name:        "kubernetes_pod_metadata_busy_total",
		Type:        model.Counter,
		Value:       k.PodMetaBusy,
		Description: "busy fetches of pod metadata",
		LabelName:   "url",
		LabelValue:  k.Url,
	}

	points[7] = &model.Point{
		Name:        "kubernetes_pod_metadata_error_total",
		Type:        model.Counter,
		Value:       k.PodMetaError,
		Description: "error fetches of pod metadata",
		LabelName:   "url",
		LabelValue:  k.Url,
	}

	points[8] = &model.Point{
		Name:        "kubernetes_record_seen_total",
		Type:        model.Counter,
		Value:       k.RecordSeen,
		Description: "records fetched from the api",
		LabelName:   "url",
		LabelValue:  k.Url,
	}

	return points
}
