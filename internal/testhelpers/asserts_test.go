package testhelpers

import (
	"fmt"
	"strings"
	"testing"
)

func TestAssertEqStringAndInt(t *testing.T) {
	// These should not produce errors
	AssertEqString(t, "ctx1", "a", "a")
	AssertEqInt(t, "ctx2", 1, 1)
}

func TestAssertPointFields(t *testing.T) {
	// happy path: matches
	want := PointExpectation{Name: "name", Type: 1, Value: 123, Label: "lbl"}
	got := PointExpectation{Name: "name", Type: 1, Value: 123, Label: "lbl"}
	AssertPointFields(t, 0, want, got)
}

// fakeT captures Errorf calls for assertion testing
type fakeT struct {
	errs []string
}

func (*fakeT) Helper() {
	// Intentionally empty: satisfies testing.TB Helper interface to adjust stack depth in error reports.
}
func (f *fakeT) Errorf(format string, args ...interface{}) {
	f.errs = append(f.errs, fmt.Sprintf(format, args...))
}

func TestAssertErrors(t *testing.T) {
	ft := &fakeT{}

	// string mismatch
	AssertEqString(ft, "ctx", "a", "b")
	// int mismatch
	AssertEqInt(ft, "ctx", 1, 2)
	// point mismatches: name, type, value, label
	want := PointExpectation{Name: "name", Type: 1, Value: 10, Label: "lbl"}
	AssertPointFields(ft, 0, want, PointExpectation{Name: "different", Type: 1, Value: 10, Label: "lbl"})
	AssertPointFields(ft, 0, want, PointExpectation{Name: "name", Type: 2, Value: 10, Label: "lbl"})
	AssertPointFields(ft, 0, want, PointExpectation{Name: "name", Type: 1, Value: 11, Label: "lbl"})
	AssertPointFields(ft, 0, want, PointExpectation{Name: "name", Type: 1, Value: 10, Label: "other"})

	if len(ft.errs) != 6 { // 2 simple mismatches + 4 point field mismatches
		t.Fatalf("expected 6 errors captured, got %d", len(ft.errs))
	}
	// Assert presence of specific substrings for each mismatch category
	wantSubstrings := []string{
		"ctx: wanted 'a', got 'b'",            // string mismatch (uses WantedIntFmt pattern variant for strings?)
		"ctx: want '1', got '2'",              // int mismatch
		"idx 0: want name name got different", // name mismatch full prefix
		"name: want type 1 got 2",             // type mismatch
		"name: want value 10 got 11",          // value mismatch
		"name: want label lbl got other",      // label mismatch
	}
	for _, sub := range wantSubstrings {
		found := false
		for _, e := range ft.errs {
			if strings.Contains(e, sub) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected to find substring '%s' in errors: %v", sub, ft.errs)
		}
	}
}

func TestFormatErrHelper(t *testing.T) {
	got := formatErr("x=%d", 5)
	if got != "x=5" {
		t.Fatalf("formatErr wrong output: %s", got)
	}
	// ensure no side effects (calling again returns same)
	got2 := formatErr("x=%d", 5)
	if got2 != got {
		t.Fatalf("formatErr inconsistent output: %s vs %s", got, got2)
	}
}

func TestAssertEqStringAndIntNoErrorCapture(t *testing.T) {
	ft := &fakeT{}
	AssertEqString(ft, "ctx", "same", "same")
	AssertEqInt(ft, "ctx", 9, 9)
	if len(ft.errs) != 0 {
		t.Fatalf("expected 0 errors for equal values, got %d: %v", len(ft.errs), ft.errs)
	}
}
