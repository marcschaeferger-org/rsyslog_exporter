package rsyslog

import "testing"

func TestNewFromJSONErrorPaths(t *testing.T) {
	bad := []byte("notjson")
	// Each constructor should fail with invalid JSON. We intentionally ignore
	// the successful return value (first value) focusing only on error presence.
	cases := []struct {
		name string
		fn   func([]byte) error
	}{
		{"action", func(b []byte) error { _, err := NewActionFromJSON(b); return err }},
		{"dynstat", func(b []byte) error { _, err := NewDynStatFromJSON(b); return err }},
		{"dynafilecache", func(b []byte) error { _, err := NewDynafileCacheFromJSON(b); return err }},
		{"forward", func(b []byte) error { _, err := NewForwardFromJSON(b); return err }},
		{"imudp", func(b []byte) error { _, err := NewInputIMUDPFromJSON(b); return err }},
		{"input", func(b []byte) error { _, err := NewInputFromJSON(b); return err }},
		{"k8s", func(b []byte) error { _, err := NewKubernetesFromJSON(b); return err }},
		{"omkafka", func(b []byte) error { _, err := NewOmkafkaFromJSON(b); return err }},
		{"queue", func(b []byte) error { _, err := NewQueueFromJSON(b); return err }},
		{"resource", func(b []byte) error { _, err := NewResourceFromJSON(b); return err }},
	}
	for _, c := range cases {
		if c.fn(bad) == nil {
			t.Fatalf("expected %s error", c.name)
		}
	}
}
