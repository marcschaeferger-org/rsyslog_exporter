package testhelpers

import (
	"testing"
)

// AssertEqString reports an error if want != got with context label.
func AssertEqString(t *testing.T, ctx, want, got string) {
	t.Helper()
	if want != got {
		t.Errorf(ctx+": "+WantStringFmt, want, got)
	}
}

// AssertEqInt reports an error if want != got with context label.
func AssertEqInt(t *testing.T, ctx string, want, got int64) {
	t.Helper()
	if want != got {
		t.Errorf(ctx+": "+WantIntFmt, want, got)
	}
}

// AssertPointFields compares a Point's fields against expected values and
// reports errors. It accepts primitive types to avoid importing
// the model package and creating import cycles.
func AssertPointFields(t *testing.T, idx int, wantName string, wantType int, wantValue int64, wantLabel string, gotName string, gotType int, gotValue int64, gotLabel string) {
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
