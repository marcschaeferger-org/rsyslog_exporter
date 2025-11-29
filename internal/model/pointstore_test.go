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

package model

import (
	"testing"

	th "github.com/prometheus-community/rsyslog_exporter/internal/testhelpers"
)

func TestPointStore(t *testing.T) {
	ps := NewStore()

	s1 := &Point{
		Name:  "my counter",
		Type:  Counter,
		Value: int64(10),
	}

	s2 := &Point{
		Name:  "my counter",
		Type:  Counter,
		Value: int64(5),
	}

	err := ps.Set(s1)
	if err != nil {
		t.Error(err)
	}

	got, err := ps.Get(s1.Key())
	if err != nil {
		t.Error(err)
	}

	if want, got := int64(10), got.Value; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	err = ps.Set(s2)
	if err != nil {
		t.Error(err)
	}

	got, err = ps.Get(s2.Key())
	if err != nil {
		t.Error(err)
	}

	if want, got := int64(5), got.Value; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	s3 := &Point{
		Name:  "my gauge",
		Type:  Gauge,
		Value: int64(20),
	}

	err = ps.Set(s3)
	if err != nil {
		t.Error(err)
	}

	got, err = ps.Get(s3.Key())
	if err != nil {
		t.Error(err)
	}

	if want, got := int64(20), got.Value; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	s4 := &Point{
		Name:  "my gauge",
		Type:  Gauge,
		Value: int64(15),
	}

	err = ps.Set(s4)
	if err != nil {
		t.Error(err)
	}

	got, err = ps.Get(s4.Key())
	if err != nil {
		t.Error(err)
	}

	if want, got := int64(15), got.Value; want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}

	_, err = ps.Get("no point")
	if err != ErrPointNotFound {
		t.Error("getting non existent point should raise error")
	}
}

func TestKeysOrdering(t *testing.T) {
	ps := NewStore()

	const setFailedFmt = "Set failed: %v"
	if err := ps.Set(&Point{Name: "b", Type: Gauge, Value: 1}); err != nil {
		t.Fatalf(setFailedFmt, err)
	}
	if err := ps.Set(&Point{Name: "a", Type: Gauge, Value: 2}); err != nil {
		t.Fatalf(setFailedFmt, err)
	}
	if err := ps.Set(&Point{Name: "c", Type: Gauge, Value: 3}); err != nil {
		t.Fatalf(setFailedFmt, err)
	}

	keys := ps.Keys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Fatalf("keys not sorted: %v", keys)
	}
}

func TestDeleteRemovesKey(t *testing.T) {
	ps := NewStore()
	p := &Point{Name: "d", Type: Gauge, Value: 4}
	_ = ps.Set(p)
	if _, err := ps.Get(p.Key()); err != nil {
		t.Fatalf("expected point to exist before delete: %v", err)
	}
	ps.Delete(p.Key())
	if _, err := ps.Get(p.Key()); err != ErrPointNotFound {
		t.Fatalf("expected ErrPointNotFound after delete, got %v", err)
	}
}
