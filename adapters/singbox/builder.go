package singbox

import (
	"encoding/json"

	netbridge "github.com/netbridge/netbridge"
)

type Builder struct{}

func (b *Builder) BuildConfig(cfg netbridge.BackendConfig) ([]byte, error) {
	inbounds := b.buildInbounds(cfg)
	outbounds := b.buildOutbounds(cfg)

	singCfg := map[string]any{
		"log": map[string]any{
			"level": "warn",
		},
		"inbounds":  inbounds,
		"outbounds": outbounds,
	}

	return json.MarshalIndent(singCfg, "", "  ")
}

func (b *Builder) buildInbounds(cfg netbridge.BackendConfig) []map[string]any {
	switch cfg.Mode {
	case netbridge.ModeHTTP:
		return []map[string]any{{
			"tag":         "http-in",
			"type":        "http",
			"listen":      "127.0.0.1",
			"listen_port": cfg.LocalPort,
		}}
	case netbridge.ModeTUN:
		return []map[string]any{{
			"tag":             "tun-in",
			"type":            "tun",
			"interface_name": cfg.TUNName,
			"inet4_address": "172.19.0.1/30",
		}}
	default:
		return []map[string]any{{
			"tag":         "socks-in",
			"type":        "socks",
			"listen":      "127.0.0.1",
			"listen_port": cfg.LocalPort,
			"sniff":       true,
			"sniff_override_destination": true,
		}}
	}
}

func (b *Builder) buildOutbounds(cfg netbridge.BackendConfig) []map[string]any {
	protocol := string(cfg.Profile.Protocol)

	switch protocol {
	case "vless":
		outbound := map[string]any{
			"tag":       "proxy",
			"type":      "vless",
			"server":    cfg.Profile.Server,
			"server_port": cfg.Profile.Port,
		}
		if id, ok := cfg.Profile.Outbound["id"]; ok {
			outbound["uuid"] = id
		}
		if cfg.Profile.Flow != "" {
			outbound["flow"] = cfg.Profile.Flow
		}
		if cfg.Profile.TLS.Enabled {
			outbound["tls"] = b.buildTLS(cfg)
		}
		if cfg.Profile.Transport.Type != "" {
			outbound["transport"] = b.buildTransport(cfg)
		}
		return []map[string]any{outbound}

	case "vmess":
		outbound := map[string]any{
			"tag":       "proxy",
			"type":      "vmess",
			"server":    cfg.Profile.Server,
			"server_port": cfg.Profile.Port,
		}
		if id, ok := cfg.Profile.Outbound["id"]; ok {
			outbound["uuid"] = id
		}
		if cfg.Profile.TLS.Enabled {
			outbound["tls"] = b.buildTLS(cfg)
		}
		if cfg.Profile.Transport.Type != "" {
			outbound["transport"] = b.buildTransport(cfg)
		}
		return []map[string]any{outbound}

	case "trojan":
		outbound := map[string]any{
			"tag":       "proxy",
			"type":      "trojan",
			"server":    cfg.Profile.Server,
			"server_port": cfg.Profile.Port,
		}
		if pw, ok := cfg.Profile.Outbound["password"]; ok {
			outbound["password"] = pw
		}
		if cfg.Profile.TLS.Enabled {
			outbound["tls"] = b.buildTLS(cfg)
		}
		if cfg.Profile.Transport.Type != "" {
			outbound["transport"] = b.buildTransport(cfg)
		}
		return []map[string]any{outbound}

	case "shadowsocks":
		outbound := map[string]any{
			"tag":       "proxy",
			"type":      "shadowsocks",
			"server":    cfg.Profile.Server,
			"server_port": cfg.Profile.Port,
		}
		if method, ok := cfg.Profile.Outbound["method"]; ok {
			outbound["method"] = method
		}
		if pw, ok := cfg.Profile.Outbound["password"]; ok {
			outbound["password"] = pw
		}
		return []map[string]any{outbound}

	default:
		return []map[string]any{{
			"tag":  "direct",
			"type": "direct",
		}}
	}
}

func (b *Builder) buildTLS(cfg netbridge.BackendConfig) map[string]any {
	tls := map[string]any{
		"enabled":    true,
		"server_name": cfg.Profile.TLS.ServerName,
	}
	if cfg.Profile.TLS.Fingerprint != "" {
		tls["utls"] = map[string]any{
			"enabled":     true,
			"fingerprint": cfg.Profile.TLS.Fingerprint,
		}
	}
	if cfg.Profile.TLS.RealityPublicKey != "" {
		tls["reality"] = map[string]any{
			"enabled":     true,
			"public_key":  cfg.Profile.TLS.RealityPublicKey,
			"short_id":    cfg.Profile.TLS.RealityShortID,
		}
	}
	return tls
}

func (b *Builder) buildTransport(cfg netbridge.BackendConfig) map[string]any {
	t := map[string]any{
		"type": cfg.Profile.Transport.Type,
	}
	switch cfg.Profile.Transport.Type {
	case "ws":
		if cfg.Profile.Transport.Path != "" {
			t["path"] = cfg.Profile.Transport.Path
		}
		if cfg.Profile.Transport.Host != "" {
			t["headers"] = map[string]any{
				"Host": cfg.Profile.Transport.Host,
			}
		}
	case "grpc":
		if cfg.Profile.Transport.Path != "" {
			t["service_name"] = cfg.Profile.Transport.Path
		}
	case "h2", "http":
		if cfg.Profile.Transport.Path != "" {
			t["path"] = cfg.Profile.Transport.Path
		}
		if cfg.Profile.Transport.Host != "" {
			t["host"] = []string{cfg.Profile.Transport.Host}
		}
	}
	return t
}
