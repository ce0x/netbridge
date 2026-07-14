package xray

import (
	"encoding/json"
	"strings"
	"testing"

	netbridge "github.com/netbridge/netbridge"
)

func TestBuildConfig_Reality(t *testing.T) {
	b := &Builder{}
	cfg := netbridge.BackendConfig{
		Mode:      netbridge.ModeSOCKS,
		LocalPort: 1080,
		Profile: netbridge.Profile{
			Protocol: netbridge.ProtocolVLESS,
			Server:   "server.example.com",
			Port:     443,
			Flow:     "xtls-rprx-vision",
			TLS: netbridge.TLSConfig{
				Enabled:         true,
				ServerName:      "example.com",
				Fingerprint:     "chrome",
				RealityPublicKey: "PUBLIC_KEY_123",
				RealityShortID:  "SHORT_ID_456",
			},
			Transport: netbridge.TransportConfig{Type: "tcp"},
		},
	}

	data, err := b.BuildConfig(cfg)
	if err != nil {
		t.Fatalf("BuildConfig failed: %v", err)
	}

	jsonStr := string(data)

	if !strings.Contains(jsonStr, `"security": "reality"`) {
		t.Error("expected security=reality in streamSettings")
	}
	if !strings.Contains(jsonStr, `"realitySettings"`) {
		t.Error("expected realitySettings block")
	}
	if !strings.Contains(jsonStr, `"publicKey": "PUBLIC_KEY_123"`) {
		t.Error("expected publicKey in realitySettings")
	}
	if !strings.Contains(jsonStr, `"shortId": "SHORT_ID_456"`) {
		t.Error("expected shortId in realitySettings")
	}
	if !strings.Contains(jsonStr, `"flow": "xtls-rprx-vision"`) {
		t.Error("expected flow in vnext user settings")
	}
}

func TestBuildConfig_TLSOnly(t *testing.T) {
	b := &Builder{}
	cfg := netbridge.BackendConfig{
		Mode:      netbridge.ModeSOCKS,
		LocalPort: 1080,
		Profile: netbridge.Profile{
			Protocol: netbridge.ProtocolVLESS,
			Server:   "server.example.com",
			Port:     443,
			TLS: netbridge.TLSConfig{
				Enabled:    true,
				ServerName: "example.com",
			},
			Transport: netbridge.TransportConfig{Type: "ws", Path: "/ws"},
		},
	}

	data, err := b.BuildConfig(cfg)
	if err != nil {
		t.Fatalf("BuildConfig failed: %v", err)
	}

	jsonStr := string(data)

	if strings.Contains(jsonStr, "realitySettings") {
		t.Error("should not have realitySettings for TLS-only profile")
	}
	if !strings.Contains(jsonStr, `"security": "tls"`) {
		t.Error("expected security=tls")
	}
	if !strings.Contains(jsonStr, `"wsSettings"`) {
		t.Error("expected wsSettings for ws transport")
	}
}

func TestBuildConfig_RealityDefaultFingerprint(t *testing.T) {
	b := &Builder{}
	cfg := netbridge.BackendConfig{
		Mode:      netbridge.ModeSOCKS,
		LocalPort: 1080,
		Profile: netbridge.Profile{
			Protocol: netbridge.ProtocolVLESS,
			Server:   "server.example.com",
			Port:     443,
			TLS: netbridge.TLSConfig{
				Enabled:         true,
				RealityPublicKey: "PK",
				RealityShortID:  "SID",
			},
			Transport: netbridge.TransportConfig{Type: "tcp"},
		},
	}

	data, err := b.BuildConfig(cfg)
	if err != nil {
		t.Fatalf("BuildConfig failed: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"fingerprint": "chrome"`) {
		t.Error("expected default fingerprint=chrome when fp is empty")
	}
}

func TestBuildConfig_RealityWithMLDSA65(t *testing.T) {
	b := &Builder{}
	cfg := netbridge.BackendConfig{
		Mode:      netbridge.ModeSOCKS,
		LocalPort: 1080,
		Profile: netbridge.Profile{
			Protocol: netbridge.ProtocolVLESS,
			Server:   "server.example.com",
			Port:     443,
			TLS: netbridge.TLSConfig{
				Enabled:         true,
				RealityPublicKey: "PK",
				RealityShortID:  "SID",
				MLDSA65Verify:   "MLDSA_KEY",
			},
			Transport: netbridge.TransportConfig{Type: "tcp"},
		},
	}

	data, err := b.BuildConfig(cfg)
	if err != nil {
		t.Fatalf("BuildConfig failed: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"mldsa65Verify": "MLDSA_KEY"`) {
		t.Error("expected mldsa65Verify in realitySettings")
	}
}

func TestBuildConfig_NoFlow(t *testing.T) {
	b := &Builder{}
	cfg := netbridge.BackendConfig{
		Mode:      netbridge.ModeSOCKS,
		LocalPort: 1080,
		Profile: netbridge.Profile{
			Protocol: netbridge.ProtocolVLESS,
			Server:   "server.example.com",
			Port:     443,
			TLS: netbridge.TLSConfig{
				Enabled:    true,
				ServerName: "example.com",
			},
			Transport: netbridge.TransportConfig{Type: "tcp"},
		},
	}

	data, err := b.BuildConfig(cfg)
	if err != nil {
		t.Fatalf("BuildConfig failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	outbounds := parsed["outbounds"].([]any)
	outbound := outbounds[0].(map[string]any)
	settings := outbound["settings"].(map[string]any)
	vnext := settings["vnext"].([]any)
	server := vnext[0].(map[string]any)
	users := server["users"].([]any)
	user := users[0].(map[string]any)

	if _, exists := user["flow"]; exists {
		t.Error("flow should not be present when empty")
	}
}
