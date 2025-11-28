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

package exporter

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/prometheus-community/rsyslog_exporter/internal/model"
	th "github.com/prometheus-community/rsyslog_exporter/internal/testhelpers"
	"github.com/prometheus/client_golang/prometheus"
)

// Build a fake log line as the exporter expects: 4 columns with the JSON in the 4th.
func resourceLineJSON(name string, utime int64) []byte {
	js := fmt.Sprintf(`{"name":"%s","utime":%d,"stime":0,"maxrss":0,"minflt":0,"majflt":0,"inblock":0,"oublock":0,"nvcsw":0,"nivcsw":0}`, name, utime)
	// prefix three columns separated by space to mimic the format processed by handleStatLine
	return []byte("col1 col2 col3 " + js)
}

func TestHandleStatLineResource(t *testing.T) {
	re := New()
	line := resourceLineJSON("myres", 42)
	if err := re.handleStatLine(line); err != nil {
		t.Fatalf("handleStatLine failed: %v", err)
	}

	// verify store has the expected point key (name.label)
	key := "resource_utime.myres"
	p, err := re.Get(key)
	if err != nil {
		t.Fatalf("expected point for key %s: %v", key, err)
	}
	if p.Name != "resource_utime" {
		t.Fatalf("unexpected point name: %s", p.Name)
	}
	if p.Value != 42 {
		t.Fatalf("unexpected value: %d", p.Value)
	}
	if p.Type != model.Counter {
		t.Fatalf("unexpected type: %v", p.Type)
	}
}

func testHelper(t *testing.T, line []byte, testCase []*testUnit) {
	exporter := New()
	// nolint:errcheck
	exporter.handleStatLine(line)

	for _, k := range exporter.Keys() {
		t.Logf("have key: '%s'", k)
	}

	for _, item := range testCase {
		p, err := exporter.Get(item.key())
		if err != nil {
			t.Error(err)
		}

		if want, got := item.Val, p.PromValue(); want != got {
			t.Errorf(th.WantFloatFmt, item.Name, want, got)
		}
	}

	// nolint:errcheck
	exporter.handleStatLine(line)

	for _, item := range testCase {
		p, err := exporter.Get(item.key())
		if err != nil {
			t.Error(err)
		}

		var wanted float64
		switch p.Type {
		case model.Counter:
			wanted = item.Val
		case model.Gauge:
			wanted = item.Val
		default:
			t.Errorf("%d is not a valid metric type", p.Type)
			continue
		}

		if want, got := wanted, p.PromValue(); want != got {
			t.Errorf("%s: want '%f', got '%f'", item.Name, want, got)
		}
	}
}

type testUnit struct {
	Name       string
	Val        float64
	LabelValue string
}

func (t *testUnit) key() string {
	return fmt.Sprintf("%s.%s", t.Name, t.LabelValue)
}

func TestHandleLineWithAction(t *testing.T) {
	tests := []*testUnit{
		{
			Name:       "action_processed",
			Val:        100000,
			LabelValue: th.TestAction,
		},
		{
			Name:       "action_failed",
			Val:        2,
			LabelValue: th.TestAction,
		},
		{
			Name:       "action_suspended",
			Val:        1,
			LabelValue: th.TestAction,
		},
		{
			Name:       "action_suspended_duration",
			Val:        1000,
			LabelValue: th.TestAction,
		},
		{
			Name:       "action_resumed",
			Val:        1,
			LabelValue: th.TestAction,
		},
	}

	actionLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {"name":"test_action","processed":100000,"failed":2,"suspended":1,"suspended.duration":1000,"resumed":1}`)
	testHelper(t, actionLog, tests)
}

func TestHandleLineWithResource(t *testing.T) {
	tests := []*testUnit{
		{
			Name:       "resource_utime",
			Val:        10,
			LabelValue: th.ResourceUsage,
		},
		{
			Name:       "resource_stime",
			Val:        20,
			LabelValue: th.ResourceUsage,
		},
		{
			Name:       "resource_maxrss",
			Val:        30,
			LabelValue: th.ResourceUsage,
		},
		{
			Name:       "resource_minflt",
			Val:        40,
			LabelValue: th.ResourceUsage,
		},
		{
			Name:       "resource_majflt",
			Val:        50,
			LabelValue: th.ResourceUsage,
		},
		{
			Name:       "resource_inblock",
			Val:        60,
			LabelValue: th.ResourceUsage,
		},
		{
			Name:       "resource_oublock",
			Val:        70,
			LabelValue: th.ResourceUsage,
		},
		{
			Name:       "resource_nvcsw",
			Val:        80,
			LabelValue: th.ResourceUsage,
		},
		{
			Name:       "resource_nivcsw",
			Val:        90,
			LabelValue: "resource-usage",
		},
	}

	resourceLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {"name":"` + th.ResourceUsage + `","utime":10,"stime":20,"maxrss":30,"minflt":40,"majflt":50,"inblock":60,"oublock":70,"nvcsw":80,"nivcsw":90}`)
	testHelper(t, resourceLog, tests)
}

func TestHandleLineWithInput(t *testing.T) {
	tests := []*testUnit{
		{
			Name:       "input_submitted",
			Val:        1000,
			LabelValue: th.TestInput,
		},
	}

	inputLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {"name":"` + th.TestInput + `", "origin":"imuxsock", "submitted":1000}`)
	testHelper(t, inputLog, tests)
}

func TestHandleLineWithQueue(t *testing.T) {
	tests := []*testUnit{
		{
			Name:       "queue_size",
			Val:        10,
			LabelValue: th.MainQ,
		},
		{
			Name:       "queue_enqueued",
			Val:        20,
			LabelValue: th.MainQ,
		},
		{
			Name:       "queue_full",
			Val:        30,
			LabelValue: th.MainQ,
		},
		{
			Name:       "queue_discarded_full",
			Val:        40,
			LabelValue: th.MainQ,
		},
		{
			Name:       "queue_discarded_not_full",
			Val:        50,
			LabelValue: th.MainQ,
		},
		{
			Name:       "queue_max_size",
			Val:        60,
			LabelValue: th.MainQ,
		},
	}

	queueLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {"name":"` + th.MainQ + `","size":10,"enqueued":20,"full":30,"discarded.full":40,"discarded.nf":50,"maxqsize":60}`)
	testHelper(t, queueLog, tests)
}

func TestHandleLineWithGlobal(t *testing.T) {
	tests := []*testUnit{
		{
			Name:       "dynstat_global",
			Val:        1,
			LabelValue: th.MsgPerHostOpsOverflow,
		},
		{
			Name:       "dynstat_global",
			Val:        3,
			LabelValue: th.MsgPerHostNewMetricAdd,
		},
		{
			Name:       "dynstat_global",
			Val:        0,
			LabelValue: th.MsgPerHostNoMetric,
		},
		{
			Name:       "dynstat_global",
			Val:        0,
			LabelValue: th.MsgPerHostMetricsPurged,
		},
		{
			Name:       "dynstat_global",
			Val:        0,
			LabelValue: th.MsgPerHostOpsIgnored,
		},
	}

	log := []byte(`2018-01-18T09:39:12.763025+00:00 some-node.example.org rsyslogd-pstats: { "name": "global", "origin": "dynstats", "values": { "` + th.MsgPerHostOpsOverflow + `": 1, "` + th.MsgPerHostNewMetricAdd + `": 3, "` + th.MsgPerHostNoMetric + `": 0, "` + th.MsgPerHostMetricsPurged + `": 0, "` + th.MsgPerHostOpsIgnored + `": 0 } }`)

	testHelper(t, log, tests)
}

func TestHandleLineWithDynafileCache(t *testing.T) {
	tests := []*testUnit{
		{
			Name:       "dynafile_cache_requests",
			Val:        412044,
			LabelValue: "cluster",
		},
		{
			Name:       "dynafile_cache_level0",
			Val:        294002,
			LabelValue: "cluster",
		},
		{
			Name:       "dynafile_cache_missed",
			Val:        210,
			LabelValue: "cluster",
		},
		{
			Name:       "dynafile_cache_evicted",
			Val:        14,
			LabelValue: "cluster",
		},
	}

	dynafileCacheLog := []byte(`2019-07-03T17:04:01.312432+00:00 some-node.example.org rsyslogd-pstats: { "name": "dynafile cache cluster", "origin": "omfile", "requests": 412044, "level0": 294002, "missed": 210, "evicted": 14, "maxused": 100, "closetimeouts": 0 }`)
	testHelper(t, dynafileCacheLog, tests)
}

func TestHandleUnknown(t *testing.T) {
	unknownLog := []byte(`2017-08-30T08:10:04.786350+00:00 some-node.example.org rsyslogd-pstats: {"a":"b"}`)

	exporter := New()
	// nolint:errcheck
	exporter.handleStatLine(unknownLog)

	if want, got := 0, len(exporter.Keys()); want != got {
		t.Errorf(th.WantIntFmt, want, got)
	}
}

func TestDescribeAndCollect(t *testing.T) {
	re := New()

	// add a point to the store
	p := &model.Point{Name: "my_metric", Type: model.Gauge, Value: 5}
	re.Set(p)

	descCh := make(chan *prometheus.Desc, 10)
	re.Describe(descCh)
	if len(descCh) == 0 {
		t.Errorf("expected at least one descriptor from Describe")
	}
	for len(descCh) > 0 {
		<-descCh
	}

	metricCh := make(chan prometheus.Metric, 10)
	re.Collect(metricCh)
	if len(metricCh) == 0 {
		t.Errorf("expected at least one metric from Collect")
	}
	for len(metricCh) > 0 {
		<-metricCh
	}
}

func TestHandleAdditionalStatTypes(t *testing.T) {
	cases := []struct{ line string }{
		{line: `2025-01-01T00:00:00Z host rsyslogd-pstats: {"name":"omfwd","omfwd.sent":1}`},                  // forward
		{line: `2025-01-01T00:00:00Z host rsyslogd-pstats: {"name":"mmkubernetes","mmkubernetes.dropped":2}`}, // kubernetes
		{line: `2025-01-01T00:00:00Z host rsyslogd-pstats: {"name":"test_input","called.recvmmsg":3}`},        // input imudp
		{line: `2025-01-01T00:00:00Z host rsyslogd-pstats: {"name":"omkafka","submitted":4}`},                 // omkafka
	}
	for i, c := range cases {
		re := New()
		// nolint:errcheck
		re.handleStatLine([]byte(c.line))
		if len(re.Keys()) == 0 {
			t.Fatalf("case %d: expected at least one point for line %s", i, c.line)
		}
	}
}

// brokenReader returns data for the first Read call then returns an error on subsequent calls.
type brokenReader struct {
	data []byte
	used bool
}

func (b *brokenReader) Read(p []byte) (int, error) {
	if !b.used {
		b.used = true
		n := copy(p, b.data)
		return n, nil
	}
	return 0, fmt.Errorf("read error")
}

func TestRunLoopCountsErrorsAndHandlesScannerErr(t *testing.T) {
	re := New()

	// malformed line (too few columns) should cause handleStatLine to return error
	malformed := []byte("too few columns")
	// well-formed line for resource with small JSON
	good := resourceLineJSON("r1", 1)
	// prepare combined input: malformed then good
	buf := bytes.NewBuffer(nil)
	buf.Write(malformed)
	buf.WriteByte('\n')
	buf.Write(good)
	buf.WriteByte('\n')

	// set scanner to buf
	re.scanner = bufio.NewScanner(buf)

	// run loop with silent=false so it logs error but we don't assert logs
	if err := re.runLoop(false); err != nil {
		t.Fatalf("runLoop failed: %v", err)
	}

	// expect stats_line_errors to be present and value >=1
	p, err := re.Get("stats_line_errors")
	if err != nil {
		t.Fatalf("expected stats_line_errors point: %v", err)
	}
	if p.Value < 1 {
		t.Fatalf("expected stats_line_errors >= 1, got %d", p.Value)
	}

	// Now test scanner.Err path using brokenReader
	br := &brokenReader{data: []byte("col1 col2 col3 {\"name\":\"global\"}")}
	re2 := New()
	re2.scanner = bufio.NewScanner(br)
	if err := re2.runLoop(true); err == nil {
		t.Fatalf("expected runLoop to return scanner error")
	}
}

func TestHandleStatLineInvalidSplit(t *testing.T) {
	re := New()
	if err := re.handleStatLine([]byte("one two three")); err == nil {
		t.Fatalf("expected split error")
	}
}

func TestHandleStatLineDecodeError(t *testing.T) {
	re := New()
	// Force TypeAction via marker and malformed JSON
	line := []byte("c1 c2 c3 {invalid}")
	// We need a 'processed' substring to pick TypeAction; include it
	line = []byte("c1 c2 c3 {\"processed\":notjson}")
	if err := re.handleStatLine(line); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestDescribeErrorBranch(t *testing.T) {
	re := New()
	p := &model.Point{Name: "x", Type: model.Gauge, Value: 1}
	_ = re.Set(p)
	// Arrange hook to delete the point before Get to trigger error branch
	orig := describeBeforeGetHook
	defer func() { describeBeforeGetHook = orig }()
	describeBeforeGetHook = func() { re.Delete(p.Key()) }

	ch := make(chan *prometheus.Desc, 10)
	re.Describe(ch)
	// We should at least have the 'scrapes' descriptor; and not panic
	if len(ch) == 0 {
		t.Fatalf("expected at least one descriptor")
	}
}

func TestCollectCoversLabelAndNoLabel(t *testing.T) {
	re := New()
	// no-label point
	_ = re.Set(&model.Point{Name: "a", Type: model.Gauge, Value: 1})
	// with label
	_ = re.Set(&model.Point{Name: "b", Type: model.Counter, Value: 2, LabelName: "x", LabelValue: "y"})
	ch := make(chan prometheus.Metric, 10)
	re.Collect(ch)
	if len(ch) < 2 {
		t.Fatalf("expected metrics for both points")
	}
}

func TestCollectErrorBranch(t *testing.T) {
	re := New()
	p := &model.Point{Name: "gone", Type: model.Gauge, Value: 1}
	_ = re.Set(p)
	orig := collectBeforeGetHook
	defer func() { collectBeforeGetHook = orig }()
	collectBeforeGetHook = func() { re.Delete(p.Key()) }
	ch := make(chan prometheus.Metric, 10)
	re.Collect(ch)
	// if we reached here without panic, the error branch was exercised via continue
}

func TestRunExported(t *testing.T) {
	re := New()
	buf := bytes.NewBufferString("col1 col2 col3 {\"submitted\":1}\n")
	re.scanner = bufio.NewScanner(buf)
	if err := re.Run(true); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}
