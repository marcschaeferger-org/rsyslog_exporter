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
	"strings"
)

// Type classifies rsyslog impstats messages by their content.
type Type int

const (
	TypeUnknown Type = iota
	TypeAction
	TypeInput
	TypeQueue
	TypeResource
	TypeDynStat
	TypeDynafileCache
	TypeInputIMDUP
	TypeForward
	TypeKubernetes
	TypeOmkafka
)

// StatType detects the impstats message type from the raw JSON buffer.
func StatType(buf []byte) Type {
	line := string(buf)
	if strings.Contains(line, "processed") {
		return TypeAction
	}
	// Use short variable declaration inside the condition to avoid calling
	// detectByName twice (which would re-unmarshal JSON and waste CPU).
	if t := detectByName(buf); t != TypeUnknown {
		return t
	}
	return detectBySubstring(line)
}

// detectByName parses JSON and classifies based on the "name" field.
// Returns TypeUnknown if parsing fails or no matching name is found.
func detectByName(buf []byte) Type {
	var obj map[string]any
	if json.Unmarshal(buf, &obj) != nil {
		// Unmarshal failed; classification falls back to substring heuristics.
		// Returning TypeUnknown here keeps parsing cheap without logging noise.
		return TypeUnknown
	}
	// Directly assert the "name" field to a string to avoid an extra
	// temporary variable for the intermediate map lookup.
	s, ok := obj["name"].(string)
	if !ok {
		return TypeUnknown
	}
	if strings.HasPrefix(s, "mmkubernetes") {
		return TypeKubernetes
	}
	switch s {
	case "omkafka":
		return TypeOmkafka
	case "omfwd":
		return TypeForward
	}
	return TypeUnknown
}

// detectBySubstring falls back to substring heuristics when JSON parsing isn't available.
func detectBySubstring(line string) Type {
	if strings.Contains(line, "\"name\": \"omkafka\"") {
		return TypeOmkafka
	}
	if strings.Contains(line, "submitted") {
		return TypeInput
	}
	if strings.Contains(line, "called.recvmmsg") {
		return TypeInputIMDUP
	}
	if strings.Contains(line, "enqueued") {
		return TypeQueue
	}
	if strings.Contains(line, "utime") {
		return TypeResource
	}
	if strings.Contains(line, "dynstats") {
		return TypeDynStat
	}
	if strings.Contains(line, "dynafile cache") {
		return TypeDynafileCache
	}
	if strings.Contains(line, "omfwd") {
		return TypeForward
	}
	if strings.Contains(line, "mmkubernetes") {
		return TypeKubernetes
	}
	return TypeUnknown
}
