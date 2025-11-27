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

import "strings"

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

// GetStatType detects the impstats message type from the raw JSON buffer.
func GetStatType(buf []byte) Type {
	line := string(buf)
	if strings.Contains(line, "processed") {
		return TypeAction
	} else if strings.Contains(line, "\"name\": \"omkafka\"") {
		// Not checking for just omkafka here as multiple actions may/will contain that word.
		// omkafka lines have a submitted field, so they need to be filtered before TypeInput
		return TypeOmkafka
	} else if strings.Contains(line, "submitted") {
		return TypeInput
	} else if strings.Contains(line, "called.recvmmsg") {
		return TypeInputIMDUP
	} else if strings.Contains(line, "enqueued") {
		return TypeQueue
	} else if strings.Contains(line, "utime") {
		return TypeResource
	} else if strings.Contains(line, "dynstats") {
		return TypeDynStat
	} else if strings.Contains(line, "dynafile cache") {
		return TypeDynafileCache
	} else if strings.Contains(line, "omfwd") {
		return TypeForward
	} else if strings.Contains(line, "mmkubernetes") {
		return TypeKubernetes
	}
	return TypeUnknown
}
