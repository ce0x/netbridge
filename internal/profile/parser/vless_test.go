package parser

import (
	"testing"
)

func TestParseVLESS_FlowVision(t *testing.T) {
	uri := `vless://uuid-1234@server.example.com:443?security=reality&sni=example.com&fp=chrome&pbk=TEST_PBK&sid=abc123&flow=xtls-rprx-vision&type=tcp#TestServer`

	p, err := ParseVLESS(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Flow != "xtls-rprx-vision" {
		t.Errorf("Flow = %q, want %q", p.Flow, "xtls-rprx-vision")
	}
	if !p.TLS.Enabled {
		t.Error("TLS should be enabled for Reality")
	}
	if p.Server != "server.example.com" {
		t.Errorf("Server = %q, want %q", p.Server, "server.example.com")
	}
	if p.Port != 443 {
		t.Errorf("Port = %d, want 443", p.Port)
	}
}

func TestParseVLESS_NoFlow(t *testing.T) {
	uri := `vless://uuid-1234@server.example.com:443?security=tls&sni=example.com&type=ws&path=/ws#NoFlow`

	p, err := ParseVLESS(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Flow != "" {
		t.Errorf("Flow = %q, want empty", p.Flow)
	}
}

func TestParseVLESS_FlowWithoutTLS(t *testing.T) {
	uri := `vless://uuid-1234@server.example.com:443?security=none&flow=xtls-rprx-vision&type=tcp#BadConfig`

	_, err := ParseVLESS(uri)
	if err == nil {
		t.Fatal("expected error for flow with security=none, got nil")
	}
}

func TestParseVLESS_Encryption(t *testing.T) {
	uri := `vless://uuid-1234@server.example.com:443?encryption=none&security=tls&sni=example.com&type=tcp#EncTest`

	p, err := ParseVLESS(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Encryption != "none" {
		t.Errorf("Encryption = %q, want %q", p.Encryption, "none")
	}
}

func TestParseVLESS_EncryptionNative(t *testing.T) {
	uri := `vless://uuid-1234@server.example.com:443?encryption=mlkem768x25519plus.native.0rtt&security=tls&sni=example.com&type=tcp#PQTest`

	p, err := ParseVLESS(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Encryption != "mlkem768x25519plus.native.0rtt" {
		t.Errorf("Encryption = %q, want %q", p.Encryption, "mlkem768x25519plus.native.0rtt")
	}
}

func TestParseVLESS_RealityWithFlow(t *testing.T) {
	uri := `vless://uuid-1234@server.example.com:443?security=reality&sni=example.com&fp=chrome&pbk=PBK123&sid=SID456&flow=xtls-rprx-vision&type=tcp#RealityFlow`

	p, err := ParseVLESS(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Flow != "xtls-rprx-vision" {
		t.Errorf("Flow = %q, want xtls-rprx-vision", p.Flow)
	}
	if p.TLS.RealityPublicKey != "PBK123" {
		t.Errorf("RealityPublicKey = %q, want PBK123", p.TLS.RealityPublicKey)
	}
	if p.TLS.RealityShortID != "SID456" {
		t.Errorf("RealityShortID = %q, want SID456", p.TLS.RealityShortID)
	}
}

func TestParseVLESS_UnknownFlowKeptAsIs(t *testing.T) {
	uri := `vless://uuid-1234@server.example.com:443?security=tls&sni=example.com&flow=xtls-rprx-vision-future&type=tcp#FutureFlow`

	p, err := ParseVLESS(uri)
	if err != nil {
		t.Fatalf("unexpected error for unknown flow value: %v", err)
	}
	if p.Flow != "xtls-rprx-vision-future" {
		t.Errorf("Flow = %q, want xtls-rprx-vision-future (kept as-is)", p.Flow)
	}
}

func TestParseVLESS_InvalidURI(t *testing.T) {
	_, err := ParseVLESS("not-a-vless-uri")
	if err == nil {
		t.Fatal("expected error for invalid URI, got nil")
	}
}
