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
	"testing"

	th "github.com/prometheus-community/rsyslog_exporter/internal/testhelpers"
)

var (
	kubernetesLog = []byte(`{ "name": "mmkubernetes(https://host.domain.tld:6443)", "origin": "mmkubernetes", "recordseen": 477943, "namespacemetadatasuccess": 7, "namespacemetadatanotfound": 0, "namespacemetadatabusy": 0, "namespacemetadataerror": 0, "podmetadatasuccess": 26, "podmetadatanotfound": 0, "podmetadatabusy": 0, "podmetadataerror": 0 }`)
)

func TestNewKubernetesFromJSON(t *testing.T) {
	if got := GetStatType(kubernetesLog); got != TypeKubernetes {
		t.Errorf(th.DetectedTypeFmt, TypeKubernetes, got)
	}
	pstat, err := NewKubernetesFromJSON(kubernetesLog)
	if err != nil {
		t.Fatalf("parse kubernetes stat failed: %v", err)
	}
	// Table of string expectations.
	th.AssertEqString(t, "k8s name", "mmkubernetes(https://host.domain.tld:6443)", pstat.Name)
	th.AssertEqString(t, "k8s url", "https://host.domain.tld:6443", pstat.Url)

	// Numeric field expectations.
	numExpectations := []struct {
		ctx  string
		want int64
		got  int64
	}{
		{"record_seen", 477943, pstat.RecordSeen},
		{"ns_meta_success", 7, pstat.NamespaceMetaSuccess},
		{"ns_meta_notfound", 0, pstat.NamespaceMetaNotFound},
		{"ns_meta_busy", 0, pstat.NamespaceMetaBusy},
		{"ns_meta_error", 0, pstat.NamespaceMetaError},
		{"pod_meta_success", 26, pstat.PodMetaSuccess},
		{"pod_meta_notfound", 0, pstat.PodMetaNotFound},
		{"pod_meta_busy", 0, pstat.PodMetaBusy},
		{"pod_meta_error", 0, pstat.PodMetaError},
	}
	for _, ne := range numExpectations {
		th.AssertEqInt(t, ne.ctx, ne.want, ne.got)
	}
}

func TestKubernetesToPoints(t *testing.T) {
	pstat, err := NewKubernetesFromJSON(kubernetesLog)
	if err != nil {
		t.Fatalf("parse kubernetes stat failed: %v", err)
	}
	points := pstat.ToPoints()
	expectedNames := []string{
		"kubernetes_namespace_metadata_success_total",
		"kubernetes_namespace_metadata_notfound_total",
		"kubernetes_namespace_metadata_busy_total",
		"kubernetes_namespace_metadata_error_total",
		"kubernetes_pod_metadata_success_total",
		"kubernetes_pod_metadata_notfound_total",
		"kubernetes_pod_metadata_busy_total",
		"kubernetes_pod_metadata_error_total",
		"kubernetes_record_seen_total",
	}
	if len(points) != len(expectedNames) {
		t.Fatalf(th.ExpectedPointsFmt, len(expectedNames), len(points))
	}
	for i, name := range expectedNames {
		if points[i].Name != name {
			t.Errorf(th.WantStringFmt, name, points[i].Name)
		}
		th.AssertEqString(t, "label url", "https://host.domain.tld:6443", points[i].LabelValue)
	}
}
