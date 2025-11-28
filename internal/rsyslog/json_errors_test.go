package rsyslog

import "testing"

func TestNewFromJSONErrorPaths(t *testing.T) {
	bad := []byte("notjson")
	if _, err := NewActionFromJSON(bad); err == nil { t.Fatalf("expected action error") }
	if _, err := NewDynStatFromJSON(bad); err == nil { t.Fatalf("expected dynstat error") }
	if _, err := NewDynafileCacheFromJSON(bad); err == nil { t.Fatalf("expected dfc error") }
	if _, err := NewForwardFromJSON(bad); err == nil { t.Fatalf("expected forward error") }
	if _, err := NewInputIMUDPFromJSON(bad); err == nil { t.Fatalf("expected imudp error") }
	if _, err := NewInputFromJSON(bad); err == nil { t.Fatalf("expected input error") }
	if _, err := NewKubernetesFromJSON(bad); err == nil { t.Fatalf("expected k8s error") }
	if _, err := NewOmkafkaFromJSON(bad); err == nil { t.Fatalf("expected omkafka error") }
	if _, err := NewQueueFromJSON(bad); err == nil { t.Fatalf("expected queue error") }
	if _, err := NewResourceFromJSON(bad); err == nil { t.Fatalf("expected resource error") }
}