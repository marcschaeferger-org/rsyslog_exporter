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

import "testing"

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
		t.Errorf("want '%d', got '%d'", want, got)
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
		t.Errorf("want '%d', got '%d'", want, got)
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
		t.Errorf("want '%d', got '%d'", want, got)
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
		t.Errorf("want '%d', got '%d'", want, got)
	}

	_, err = ps.Get("no point")
	if err != ErrPointNotFound {
		t.Error("getting non existent point should raise error")
	}
}
