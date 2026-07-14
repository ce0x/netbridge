package wireguard

import (
	"context"
	"testing"
)

func TestAdapter_Name(t *testing.T) {
	a := New()
	if a.Name() != "wireguard" {
		t.Errorf("Name() = %q, want %q", a.Name(), "wireguard")
	}
}

func TestAdapter_SupportedProtocols(t *testing.T) {
	a := New()
	protocols := a.SupportedProtocols()
	if len(protocols) != 1 {
		t.Errorf("SupportedProtocols() returned %d, want 1", len(protocols))
	}
}

func TestAdapter_StopWhenNotRunning(t *testing.T) {
	a := New()
	err := a.Stop()
	if err != nil {
		t.Errorf("Stop() on non-running adapter should return nil, got: %v", err)
	}
}

func TestAdapter_StatusWhenNotRunning(t *testing.T) {
	a := New()
	status := a.Status()
	if status.Running {
		t.Error("Status().Running should be false when not started")
	}
}

func TestAdapter_HealthCheckNotRunning(t *testing.T) {
	a := New()
	err := a.HealthCheck(context.Background())
	if err == nil {
		t.Error("HealthCheck should error when not running")
	}
}

func TestAdapter_LocalEndpoints(t *testing.T) {
	a := New()
	a.ifaceName = "wg0"
	endpoints := a.LocalEndpoints()
	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(endpoints))
	}
	if endpoints[0].Iface != "wg0" {
		t.Errorf("endpoint iface = %q, want wg0", endpoints[0].Iface)
	}
}
