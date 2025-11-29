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

func TestStatTypeUnknown(t *testing.T) {
	if got := StatType([]byte(`{"foo":"bar"}`)); got != TypeUnknown {
		t.Fatalf("expected TypeUnknown, got %v", got)
	}
}

func TestStatTypeJSONNameOmkafka(t *testing.T) {
	if got := StatType([]byte(`{"name":"omkafka","submitted":1}`)); got != TypeOmkafka {
		t.Fatalf("expected TypeOmkafka, got %v", got)
	}
}

func TestStatTypeJSONNameOmfwd(t *testing.T) {
	if got := StatType([]byte(`{"name":"omfwd"}`)); got != TypeForward {
		t.Fatalf("expected TypeForward, got %v", got)
	}
}

func TestDetectByNameNonStringAndKubernetesPrefix(t *testing.T) {
	// name is non-string -> should be unknown
	if got := StatType([]byte(`{"name":123}`)); got != TypeUnknown {
		t.Fatalf("expected TypeUnknown for non-string name, got %v", got)
	}

	// name prefixed with mmkubernetes should detect Kubernetes type
	if got := StatType([]byte(`{"name":"mmkubernetes.svc"}`)); got != TypeKubernetes {
		t.Fatalf("expected TypeKubernetes for mmkubernetes prefix, got %v", got)
	}
}

func TestDetectBySubstringHeuristics(t *testing.T) {
	cases := []struct {
		line string
		want Type
	}{
		{line: "some text \"name\": \"omkafka\" more", want: TypeUnknown},
		{line: "blah submitted stuff", want: TypeInput},
		{line: "called.recvmmsg here", want: TypeInputIMDUP},
		{line: "queue enqueued event", want: TypeQueue},
		{line: "cpu utime usage", want: TypeResource},
		{line: "something dynstats event", want: TypeDynStat},
		{line: "contains dynafile cache entry", want: TypeDynafileCache},
		{line: "contains omfwd marker", want: TypeForward},
		{line: "mmkubernetes appears as substring", want: TypeKubernetes},
	}

	for i, c := range cases {
		if got := StatType([]byte(c.line)); got != c.want {
			t.Fatalf("case %d: expected %v for line %q, got %v", i, c.want, c.line, got)
		}
	}
}

func TestStatTypeProcessedShortcut(t *testing.T) {
	// presence of the substring "processed" should short-circuit to TypeAction
	if got := StatType([]byte(`{"processed":42}`)); got != TypeAction {
		t.Fatalf("expected TypeAction for processed substring, got %v", got)
	}
}
