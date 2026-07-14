package xray

import (
	"context"
	"testing"

	netbridge "github.com/netbridge/netbridge"
)

func TestAdapter_Name(t *testing.T) {
	a := New()
	if a.Name() != "xray" {
		t.Errorf("Name() = %q, want %q", a.Name(), "xray")
	}
}

func TestAdapter_SupportedProtocols(t *testing.T) {
	a := New()
	protocols := a.SupportedProtocols()
	if len(protocols) != 6 {
		t.Errorf("SupportedProtocols() returned %d, want 6", len(protocols))
	}
}

func TestAdapter_StartAlreadyRunning(t *testing.T) {
	a := New()
	a.process = NewProcess()
	a.process.running = true
	a.process.pid = 99999

	err := a.Start(context.Background(), netbridge.BackendConfig{})
	if err == nil {
		t.Error("expected error when starting already running adapter")
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

func TestAdapter_ReloadNotRunning(t *testing.T) {
	a := New()
	err := a.Reload("/tmp/config.json")
	if err == nil {
		t.Error("Reload should error when not running")
	}
}

func TestAdapter_LocalEndpointsSOCKS(t *testing.T) {
	a := New()
	a.config = netbridge.BackendConfig{
		Mode:      netbridge.ModeSOCKS,
		LocalPort: 1080,
	}
	endpoints := a.LocalEndpoints()
	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(endpoints))
	}
	if endpoints[0].Address != "127.0.0.1:1080" {
		t.Errorf("endpoint address = %q, want 127.0.0.1:1080", endpoints[0].Address)
	}
}

func TestAdapter_LocalEndpointsHTTP(t *testing.T) {
	a := New()
	a.config = netbridge.BackendConfig{
		Mode:      netbridge.ModeHTTP,
		LocalPort: 8080,
	}
	endpoints := a.LocalEndpoints()
	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(endpoints))
	}
	if endpoints[0].Address != "127.0.0.1:8080" {
		t.Errorf("endpoint address = %q, want 127.0.0.1:8080", endpoints[0].Address)
	}
}
