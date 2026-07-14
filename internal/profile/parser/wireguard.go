package parser

import (
	"fmt"
	"strings"

	netbridge "github.com/netbridge/netbridge"
)

func ParseWireGuardConf(content string) (*netbridge.Profile, error) {
	sections := make(map[string]map[string]string)
	var currentSection string

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			end := strings.Index(line, "]")
			if end > 0 {
				currentSection = line[1:end]
				sections[currentSection] = make(map[string]string)
			}
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && currentSection != "" {
			sections[currentSection][strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	peer, ok := sections["Peer"]
	if !ok {
		return nil, fmt.Errorf("no [Peer] section found")
	}

	endpoint := peer["Endpoint"]
	server := endpoint
	port := 51820

	if idx := strings.LastIndex(endpoint, ":"); idx > 0 {
		server = endpoint[:idx]
		fmt.Sscanf(endpoint[idx+1:], "%d", &port)
	}

	return &netbridge.Profile{
		Name:     fmt.Sprintf("wg-%s", server),
		Protocol: netbridge.ProtocolWireGuard,
		Backend:  "wireguard",
		Server:   server,
		Port:     port,
		Outbound: map[string]any{
			"allowed_ips": peer["AllowedIPs"],
			"public_key":  peer["PublicKey"],
			"preshared":   peer["PresharedKey"],
		},
	}, nil
}
