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
package testhelpers

// Shared format strings used across tests to avoid duplicated literals.
const (
	WantStringFmt       = "wanted %s, got %s"
	WantIntFmt          = "wanted %d, got %d"
	WantFloatFmt        = "wanted %f, got %f"
	DetectedStatTypeFmt = "detected stat type should be %d but is %d"
	ExpectedIndexFmt    = "expected point index %d to exist"
	ExpectedPointsFmt   = "expected %d points, got %d"
	ExpectedParseErrFmt = "expected parsing %s not to fail, got: %v"
	DynamicStatDesc     = "dynamic statistic bucket global"
	// Common label/value literals used across tests.
	ResourceUsage  = "resource-usage"
	MainQueueLabel = "main Q"

	MsgPerHostOpsOverflow   = "msg_per_host.ops_overflow"
	MsgPerHostNewMetricAdd  = "msg_per_host.new_metric_add"
	MsgPerHostNoMetric      = "msg_per_host.no_metric"
	MsgPerHostMetricsPurged = "msg_per_host.metrics_purged"
	MsgPerHostOpsIgnored    = "msg_per_host.ops_ignored"
	// Common test label values
	TestAction     = "test_action"
	TestInput      = "test_input"
	TestInputIMUDP = "test_input_imudp"
	Cluster        = "cluster"
)
