package xray

import (
	"encoding/json"

	netbridge "github.com/netbridge/netbridge"
)

type Builder struct{}

func (b *Builder) BuildConfig(cfg netbridge.BackendConfig) ([]byte, error) {
	xrayCfg := map[string]any{
		"log": map[string]any{
			"loglevel": "warning",
		},
		"inbounds":  b.buildInbounds(cfg),
		"outbounds": b.buildOutbounds(cfg),
	}

	return json.MarshalIndent(xrayCfg, "", "  ")
}

func (b *Builder) buildInbounds(cfg netbridge.BackendConfig) []map[string]any {
	var inbounds []map[string]any

	switch cfg.Mode {
	case netbridge.ModeSOCKS:
		inbounds = append(inbounds, map[string]any{
			"tag":      "socks-in",
			"protocol": "socks",
			"port":     cfg.LocalPort,
			"settings": map[string]any{
				"udp": true,
			},
		})
	case netbridge.ModeHTTP:
		inbounds = append(inbounds, map[string]any{
			"tag":      "http-in",
			"protocol": "http",
			"port":     cfg.LocalPort,
		})
	case netbridge.ModeTUN:
		inbounds = append(inbounds, map[string]any{
			"tag":      "tun-in",
			"protocol": "dokodemo-door",
			"port":     12345,
			"settings": map[string]any{
				"network":        "tcp,udp",
				"followRedirect": true,
			},
		})
	}

	return inbounds
}

func (b *Builder) buildOutbounds(cfg netbridge.BackendConfig) []map[string]any {
	protocol := string(cfg.Profile.Protocol)

	outbound := map[string]any{
		"tag":      "proxy",
		"protocol": protocol,
		"settings": b.buildOutboundSettings(cfg),
	}

	if cfg.Profile.TLS.Enabled {
		outbound["streamSettings"] = b.buildStreamSettings(cfg)
	}

	return []map[string]any{outbound}
}

func (b *Builder) buildOutboundSettings(cfg netbridge.BackendConfig) map[string]any {
	protocol := string(cfg.Profile.Protocol)

	switch protocol {
	case "vless":
		user := map[string]any{}
		if id, ok := cfg.Profile.Outbound["id"]; ok {
			user["id"] = id
		}
		if cfg.Profile.Flow != "" {
			user["flow"] = cfg.Profile.Flow
		}
		if cfg.Profile.Encryption != "" {
			user["encryption"] = cfg.Profile.Encryption
		} else {
			user["encryption"] = "none"
		}
		return map[string]any{
			"vnext": []map[string]any{
				{
					"address": cfg.Profile.Server,
					"port":    cfg.Profile.Port,
					"users":   []map[string]any{user},
				},
			},
		}

	case "vmess":
		user := map[string]any{}
		if id, ok := cfg.Profile.Outbound["id"]; ok {
			user["id"] = id
		}
		if aid, ok := cfg.Profile.Outbound["alter_id"]; ok {
			user["alterId"] = aid
		}
		return map[string]any{
			"vnext": []map[string]any{
				{
					"address": cfg.Profile.Server,
					"port":    cfg.Profile.Port,
					"users":   []map[string]any{user},
				},
			},
		}

	case "trojan":
		server := map[string]any{
			"address": cfg.Profile.Server,
			"port":    cfg.Profile.Port,
		}
		if pw, ok := cfg.Profile.Outbound["password"]; ok {
			server["password"] = pw
		}
		return map[string]any{
			"servers": []map[string]any{server},
		}

	case "shadowsocks":
		server := map[string]any{
			"address": cfg.Profile.Server,
			"port":    cfg.Profile.Port,
		}
		if method, ok := cfg.Profile.Outbound["method"]; ok {
			server["method"] = method
		}
		if pw, ok := cfg.Profile.Outbound["password"]; ok {
			server["password"] = pw
		}
		return map[string]any{
			"servers": []map[string]any{server},
		}

	default:
		return map[string]any{
			"servers": []map[string]any{
				{
					"address": cfg.Profile.Server,
					"port":    cfg.Profile.Port,
				},
			},
		}
	}
}

func (b *Builder) buildStreamSettings(cfg netbridge.BackendConfig) map[string]any {
	stream := map[string]any{
		"network": cfg.Profile.Transport.Type,
	}

	if cfg.Profile.TLS.RealityPublicKey != "" {
		stream["security"] = "reality"
		fp := cfg.Profile.TLS.Fingerprint
		if fp == "" {
			fp = "chrome"
		}
		realitySettings := map[string]any{
			"show":        false,
			"fingerprint": fp,
			"serverName":  cfg.Profile.TLS.ServerName,
			"publicKey":   cfg.Profile.TLS.RealityPublicKey,
			"shortId":     cfg.Profile.TLS.RealityShortID,
			"spiderX":     "",
		}
		if cfg.Profile.TLS.MLDSA65Verify != "" {
			realitySettings["mldsa65Verify"] = cfg.Profile.TLS.MLDSA65Verify
		}
		stream["realitySettings"] = realitySettings
	} else if cfg.Profile.TLS.Enabled {
		stream["security"] = "tls"
		stream["tlsSettings"] = map[string]any{
			"serverName":    cfg.Profile.TLS.ServerName,
			"fingerprint":   cfg.Profile.TLS.Fingerprint,
			"allowInsecure": cfg.Profile.TLS.AllowInsecure,
		}
	}

	if cfg.Profile.Transport.Path != "" {
		switch cfg.Profile.Transport.Type {
		case "ws":
			stream["wsSettings"] = map[string]any{
				"path": cfg.Profile.Transport.Path,
				"host": cfg.Profile.Transport.Host,
			}
		case "grpc":
			stream["grpcSettings"] = map[string]any{
				"serviceName": cfg.Profile.Transport.Path,
			}
		}
	}

	return stream
}
