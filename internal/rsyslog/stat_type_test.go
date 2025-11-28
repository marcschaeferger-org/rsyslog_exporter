package rsyslog

import "testing"

func TestStatTypeUnknown(t *testing.T) {
	if got := StatType([]byte(`{"foo":"bar"}`)); got != TypeUnknown {
		t.Fatalf("expected TypeUnknown, got %v", got)
	}
}
