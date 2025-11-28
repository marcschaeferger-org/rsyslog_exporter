package testhelpers

// Shared format strings used across tests to avoid duplicated literals.
const (
	WantStringFmt = "wanted '%s', got '%s'"
	WantIntFmt    = "want '%d', got '%d'"
	// Slight variant used in some tests; kept for compatibility.
	WantedIntFmt        = "wanted '%d', got '%d'"
	WantFloatFmt        = "%s: want '%f', got '%f'"
	DetectedTypeFmt     = "detected pstat type should be %d but is %d"
	ExpectedIndexFmt    = "expected point index %d to exist"
	ExpectedPointsFmt   = "expected %d points, got %d"
	ExpectedParseErrFmt = "expected parsing %s not to fail, got: %v"
	DynStatDesc         = "dynamic statistic bucket global"
	// Common label/value literals used across tests.
	ResourceUsage = "resource-usage"
	MainQ         = "main Q"

	MsgPerHostOpsOverflow   = "msg_per_host.ops_overflow"
	MsgPerHostNewMetricAdd  = "msg_per_host.new_metric_add"
	MsgPerHostNoMetric      = "msg_per_host.no_metric"
	MsgPerHostMetricsPurged = "msg_per_host.metrics_purged"
	MsgPerHostOpsIgnored    = "msg_per_host.ops_ignored"
)
