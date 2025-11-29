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
		t.Errorf(ctx+": "+ExpectedActualStringFmt, want, got)
	}
}

// AssertEqInt reports an error if want != got with context label.
func AssertEqInt(t tester, ctx string, want, got int64) {
	t.Helper()
	if want != got {
		t.Errorf(ctx+": "+ExpectedActualIntFmt, want, got)
	}
}

// AssertPointFields compares a Point's fields against expected values and
// reports errors. It accepts primitive types to avoid importing
// the model package and creating import cycles.
// PointExpectation groups expected or actual point attributes to avoid a long
// parameter list. This replaces the previous AssertPointFields signature that
// had 10 parameters (tester + 9 values) which triggered a code smell about
// excessive function parameters. Using a struct improves readability and
// future adaptability without impacting test clarity.
type PointExpectation struct {
	Name  string
	Type  int
	Value int64
	Label string
}

// AssertPointFields compares expected vs actual point snapshots. The idx is
// kept separate for clearer failure messages referencing ordering.
func AssertPointFields(t tester, idx int, want PointExpectation, got PointExpectation) {
	t.Helper()
	if got.Name != want.Name {
		t.Errorf("idx %d: want name %s got %s", idx, want.Name, got.Name)
	}
	if got.Type != want.Type {
		t.Errorf("%s: want type %d got %d", want.Name, want.Type, got.Type)
	}
	if got.Value != want.Value {
		t.Errorf("%s: want value %d got %d", want.Name, want.Value, got.Value)
	}
	if got.Label != want.Label {
		t.Errorf("%s: want label %s got %s", want.Name, want.Label, got.Label)
	}
}

// simple helper for testers to format errors in tests when needed
func formatErr(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
