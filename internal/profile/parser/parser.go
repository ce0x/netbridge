package parser

import (
	"fmt"
	"os"
	"strings"

	netbridge "github.com/netbridge/netbridge"
)

func ParseURI(raw string) (*netbridge.Profile, error) {
	raw = strings.TrimSpace(raw)

	switch {
	case strings.HasPrefix(raw, "vless://"):
		return ParseVLESS(raw)
	case strings.HasPrefix(raw, "vmess://"):
		return ParseVMess(raw)
	case strings.HasPrefix(raw, "trojan://"):
		return ParseTrojan(raw)
	case strings.HasPrefix(raw, "ss://"):
		return ParseShadowsocks(raw)
	default:
		return nil, fmt.Errorf("unsupported URI scheme")
	}
}

func ParseFile(path string) ([]*netbridge.Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(path)
	switch {
	case strings.HasSuffix(ext, ".conf"):
		p, err := ParseWireGuardConf(string(data))
		if err != nil {
			return nil, err
		}
		return []*netbridge.Profile{p}, nil
	case strings.HasSuffix(ext, ".ovpn"):
		p, err := ParseOpenVPNConf(string(data))
		if err != nil {
			return nil, err
		}
		return []*netbridge.Profile{p}, nil
	default:
		lines := strings.Split(string(data), "\n")
		var profiles []*netbridge.Profile
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			p, err := ParseURI(line)
			if err != nil {
				continue
			}
			profiles = append(profiles, p)
		}
		return profiles, nil
	}
}
