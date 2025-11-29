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
		{"action", func(b []byte) (any, error) { return NewActionFromJSON(b) }},
		{"dynstat", func(b []byte) (any, error) { return NewDynStatFromJSON(b) }},
		{"dynafilecache", func(b []byte) (any, error) { return NewDynafileCacheFromJSON(b) }},
		{"forward", func(b []byte) (any, error) { return NewForwardFromJSON(b) }},
		{"imudp", func(b []byte) (any, error) { return NewInputIMUDPFromJSON(b) }},
		{"input", func(b []byte) (any, error) { return NewInputFromJSON(b) }},
		{"k8s", func(b []byte) (any, error) { return NewKubernetesFromJSON(b) }},
		{"omkafka", func(b []byte) (any, error) { return NewOmkafkaFromJSON(b) }},
		{"queue", func(b []byte) (any, error) { return NewQueueFromJSON(b) }},
		{"resource", func(b []byte) (any, error) { return NewResourceFromJSON(b) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.parseFunc(invalidJSON) //NOSONAR
			if err == nil {
				t.Fatalf("expected %s error", tc.name)
			}
		})
	}
}
