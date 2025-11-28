package testhelpers

import (
	"fmt"
)

// tester is a minimal subset of testing.T used by these helpers. Using a
// narrow interface allows tests to provide a fake implementation that
// captures Errorf calls for negative-path assertions.
type tester interface {
	Helper()
	Errorf(format string, args ...interface{})
}

// AssertEqString reports an error if want != got with context label.
func AssertEqString(t tester, ctx, want, got string) {
	t.Helper()
	if want != got {
		t.Errorf(ctx+": "+WantStringFmt, want, got)
	}
}

// AssertEqInt reports an error if want != got with context label.
func AssertEqInt(t tester, ctx string, want, got int64) {
	t.Helper()
	if want != got {
		t.Errorf(ctx+": "+WantIntFmt, want, got)
	}
}

// AssertPointFields compares a Point's fields against expected values and
// reports errors. It accepts primitive types to avoid importing
// the model package and creating import cycles.
func AssertPointFields(t tester, idx int, wantName string, wantType int, wantValue int64, wantLabel string, gotName string, gotType int, gotValue int64, gotLabel string) {
	t.Helper()
	if gotName != wantName {
		t.Errorf("idx %d: want name %s got %s", idx, wantName, gotName)
	}
	if gotType != wantType {
		t.Errorf("%s: want type %d got %d", wantName, wantType, gotType)
	}
	if gotValue != wantValue {
		t.Errorf("%s: want value %d got %d", wantName, wantValue, gotValue)
	}
	if gotLabel != wantLabel {
		t.Errorf("%s: want label %s got %s", wantName, wantLabel, gotLabel)
	}
}

// simple helper for testers to format errors in tests when needed
func formatErr(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
