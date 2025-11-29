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

import "testing"

func TestNewFromJSONErrorPaths(t *testing.T) {
	invalidJSON := []byte("notjson")
	cases := []struct {
		name      string
		parseFunc func([]byte) (any, error)
	}{
		{"action", func(b []byte) error { _, err := NewActionFromJSON(b); return err }},
		{"dynstat", func(b []byte) error { _, err := NewDynStatFromJSON(b); return err }},
		{"dynafilecache", func(b []byte) error { _, err := NewDynafileCacheFromJSON(b); return err }},
		{"forward", func(b []byte) error { _, err := NewForwardFromJSON(b); return err }},
		{"imudp", func(b []byte) error { _, err := NewInputIMUDPFromJSON(b); return err }},
		{"input", func(b []byte) error { _, err := NewInputFromJSON(b); return err }},
		{"k8s", func(b []byte) error { _, err := NewKubernetesFromJSON(b); return err }},
		{"omkafka", func(b []byte) error { _, err := NewOmkafkaFromJSON(b); return err }},
		{"queue", func(b []byte) error { _, err := NewQueueFromJSON(b); return err }},
		{"resource", func(b []byte) error { _, err := NewResourceFromJSON(b); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.parseFunc(invalidJSON); err == nil {
				t.Fatalf("expected %s error", tc.name)
			}
		})
	}
}
