package testhelpers

import "testing"

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
