package wireguard

import (
	"context"
	"strings"
	"testing"

	netbridge "github.com/netbridge/netbridge"
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

func TestAdapter_StartMissingPrivateKey(t *testing.T) {
	a := New()
	cfg := netbridge.BackendConfig{
		Profile: netbridge.Profile{
			Outbound: map[string]any{
				"public_key": "somekey",
			},
		},
	}
	err := a.Start(context.Background(), cfg)
	if err == nil {
		t.Fatal("Start should fail without private_key")
	}
	if !strings.Contains(err.Error(), "private_key") {
		t.Errorf("error should mention private_key, got: %v", err)
	}
}

func TestAdapter_StartMissingPublicKey(t *testing.T) {
	a := New()
	cfg := netbridge.BackendConfig{
		Profile: netbridge.Profile{
			Outbound: map[string]any{
				"private_key": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			},
		},
	}
	err := a.Start(context.Background(), cfg)
	if err == nil {
		t.Fatal("Start should fail without public_key")
	}
	if !strings.Contains(err.Error(), "public_key") {
		t.Errorf("error should mention public_key, got: %v", err)
	}
}

func TestAdapter_StartMissingAddress(t *testing.T) {
	a := New()
	cfg := netbridge.BackendConfig{
		Profile: netbridge.Profile{
			Outbound: map[string]any{
				"private_key": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				"public_key":  "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
			},
		},
	}
	err := a.Start(context.Background(), cfg)
	if err == nil {
		t.Fatal("Start should fail without address")
	}
	if !strings.Contains(err.Error(), "address") {
		t.Errorf("error should mention address, got: %v", err)
	}
}

func TestAdapter_AlreadyRunning(t *testing.T) {
	a := New()
	a.running = true
	err := a.Start(context.Background(), netbridge.BackendConfig{})
	if err == nil {
		t.Error("Start should fail when already running")
	}
}

func TestGetFieldString_NilMap(t *testing.T) {
	_, err := getFieldString(nil, "key")
	if err == nil {
		t.Error("should error on nil map")
	}
}

func TestGetFieldString_Missing(t *testing.T) {
	_, err := getFieldString(map[string]any{}, "key")
	if err == nil {
		t.Error("should error on missing key")
	}
}

func TestGetFieldString_NilValue(t *testing.T) {
	_, err := getFieldString(map[string]any{"key": nil}, "key")
	if err == nil {
		t.Error("should error on nil value")
	}
}

func TestGetFieldString_Empty(t *testing.T) {
	_, err := getFieldString(map[string]any{"key": ""}, "key")
	if err == nil {
		t.Error("should error on empty string")
	}
}

func TestGetFieldString_OK(t *testing.T) {
	val, err := getFieldString(map[string]any{"key": "value"}, "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Errorf("val = %q, want %q", val, "value")
	}
}

func TestParseAllowedIPs_Single(t *testing.T) {
	result, err := parseAllowedIPs("10.0.0.0/8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 CIDR, got %d", len(result))
	}
}

func TestParseAllowedIPs_Multiple(t *testing.T) {
	result, err := parseAllowedIPs("0.0.0.0/0,::/0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 CIDRs, got %d", len(result))
	}
}

func TestParseAllowedIPs_WithSpaces(t *testing.T) {
	result, err := parseAllowedIPs("10.0.0.0/8 , 192.168.0.0/16")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 CIDRs, got %d", len(result))
	}
}

func TestParseAllowedIPs_Invalid(t *testing.T) {
	_, err := parseAllowedIPs("not-a-cidr")
	if err == nil {
		t.Error("should error on invalid CIDR")
	}
}