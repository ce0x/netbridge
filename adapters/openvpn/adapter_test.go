package openvpn

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	netbridge "github.com/netbridge/netbridge"
)

func TestAdapter_Name(t *testing.T) {
	a := New()
	if a.Name() != "openvpn" {
		t.Errorf("Name() = %q, want %q", a.Name(), "openvpn")
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
		t.Errorf("Stop() should return nil, got: %v", err)
	}
}

func TestAdapter_StatusWhenNotRunning(t *testing.T) {
	a := New()
	status := a.Status()
	if status.Running {
		t.Error("Status().Running should be false")
	}
}

func TestAdapter_HealthCheckNotConnected(t *testing.T) {
	a := New()
	err := a.HealthCheck(context.Background())
	if err == nil {
		t.Error("HealthCheck should error when not connected")
	}
}

func TestAdapter_ReloadReturnsError(t *testing.T) {
	a := New()
	err := a.Reload("/tmp/config.ovpn")
	if err == nil {
		t.Error("Reload should return error for openvpn")
	}
}

func TestAdapter_GenerateConfig(t *testing.T) {
	a := New()
	dir := t.TempDir()
	cfg := netbridge.BackendConfig{
		Profile: netbridge.Profile{
			Server: "vpn.example.com",
			Port:   1194,
			Outbound: map[string]any{
				"proto": "udp",
				"cipher": "AES-256-GCM",
			},
		},
	}

	path := filepath.Join(dir, "test.ovpn")
	err := a.generateConfig(cfg, path)
	if err != nil {
		t.Fatalf("generateConfig failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	content := string(data)
	if !contains(content, "remote vpn.example.com 1194") {
		t.Error("config missing remote directive")
	}
	if !contains(content, "proto udp") {
		t.Error("config missing proto directive")
	}
	if !contains(content, "cipher AES-256-GCM") {
		t.Error("config missing cipher directive")
	}
	if !contains(content, "client") {
		t.Error("config missing client directive")
	}
}

func TestAdapter_GenerateConfigWithInlineCerts(t *testing.T) {
	a := New()
	dir := t.TempDir()
	cfg := netbridge.BackendConfig{
		Profile: netbridge.Profile{
			Server: "vpn.example.com",
			Port:   1194,
			Outbound: map[string]any{
				"ca":   "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
				"cert": "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
				"key":  "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
			},
		},
	}

	path := filepath.Join(dir, "test.ovpn")
	err := a.generateConfig(cfg, path)
	if err != nil {
		t.Fatalf("generateConfig failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	content := string(data)
	if !contains(content, "<ca>") {
		t.Error("config missing ca section")
	}
	if !contains(content, "<cert>") {
		t.Error("config missing cert section")
	}
	if !contains(content, "<key>") {
		t.Error("config missing key section")
	}
}

func TestAdapter_InitialState(t *testing.T) {
	a := New()
	if a.GetState() != StateDisconnected {
		t.Errorf("initial state = %d, want StateDisconnected", a.GetState())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}