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

	"github.com/prometheus-community/rsyslog_exporter/internal/model"
)

var (
	dynafileCacheLog = []byte(`{ "name": "dynafile cache cluster", "origin": "omfile", "requests": 1783254, "level0": 1470906, "missed": 2625, "evicted": 2525, "maxused": 100, "closetimeouts": 10 }`)
)

func TestNewDynafileCacheFromJSON(t *testing.T) {
	if got := StatType(dynafileCacheLog); got != TypeDynafileCache {
		t.Errorf(th.DetectedStatTypeFmt, TypeDynafileCache, got)
	}
	pstat, err := NewDynafileCacheFromJSON(dynafileCacheLog)
	if err != nil {
		t.Fatalf("parse dynafile cache stat failed: %v", err)
	}
	th.AssertEqString(t, "name", th.Cluster, pstat.Name)
	nums := []struct {
		ctx       string
		want, got int64
	}{
		{"requests", 1783254, pstat.Requests},
		{"level0", 1470906, pstat.Level0},
		{"missed", 2625, pstat.Missed},
		{"evicted", 2525, pstat.Evicted},
		{"maxused", 100, pstat.MaxUsed},
		{"closetimeouts", 10, pstat.CloseTimeouts},
	}
	for _, n := range nums {
		th.AssertEqInt(t, n.ctx, n.want, n.got)
	}
}

func TestDynafileCacheToPoints(t *testing.T) {
	expected := []model.Point{
		{Name: "dynafile_cache_requests", Type: model.Counter, Value: 1783254, Description: "number of requests made to obtain a dynafile", LabelName: "cache", LabelValue: th.Cluster},
		{Name: "dynafile_cache_level0", Type: model.Counter, Value: 1470906, Description: "number of requests for the current active file", LabelName: "cache", LabelValue: th.Cluster},
		{Name: "dynafile_cache_missed", Type: model.Counter, Value: 2625, Description: "number of cache misses", LabelName: "cache", LabelValue: th.Cluster},
		{Name: "dynafile_cache_evicted", Type: model.Counter, Value: 2525, Description: "number of times a file needed to be evicted from cache", LabelName: "cache", LabelValue: th.Cluster},
		{Name: "dynafile_cache_maxused", Type: model.Counter, Value: 100, Description: "maximum number of cache entries ever used", LabelName: "cache", LabelValue: th.Cluster},
		{Name: "dynafile_cache_closetimeouts", Type: model.Counter, Value: 10, Description: "number of times a file was closed due to timeout settings", LabelName: "cache", LabelValue: th.Cluster},
	}
	pstat, err := NewDynafileCacheFromJSON(dynafileCacheLog)
	if err != nil {
		t.Fatalf("parse dynafile cache stat failed: %v", err)
	}
	points := pstat.ToPoints()
	if len(points) != len(expected) {
		t.Fatalf(th.ExpectedPointsFmt, len(expected), len(points))
	}
	// Indexing stable because ToPoints has deterministic order.
	for i, exp := range expected {
		got := points[i]
		if exp.Name != got.Name {
			t.Errorf(th.WantStringFmt, exp.Name, got.Name)
		}
		th.AssertEqInt(t, exp.Name+" value", exp.Value, got.Value)
		th.AssertEqString(t, exp.Name+" label", exp.LabelValue, got.LabelValue)
		if exp.Type != got.Type {
			t.Errorf(exp.Name+": expected type %v, got %v", exp.Type, got.Type)
		}
		if exp.Description != got.Description {
			t.Errorf(th.WantStringFmt, exp.Description, got.Description)
		}
	}
}
