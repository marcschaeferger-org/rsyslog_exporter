package testhelpers

// Shared format strings used across tests to avoid duplicated literals.
const (
	WantStringFmt       = "wanted '%s', got '%s'"
	WantIntFmt          = "want '%d', got '%d'"
	WantFloatFmt        = "%s: want '%f', got '%f'"
	DetectedTypeFmt     = "detected pstat type should be %d but is %d"
	ExpectedIndexFmt    = "expected point index %d to exist"
	ExpectedPointsFmt   = "expected %d points, got %d"
	ExpectedParseErrFmt = "expected parsing %s not to fail, got: %v"
	DynStatDesc         = "dynamic statistic bucket global"
)
