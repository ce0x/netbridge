package parser

import (
	"fmt"
	"strings"

	netbridge "github.com/netbridge/netbridge"
)

func ParseOpenVPNConf(content string) (*netbridge.Profile, error) {
	lines := strings.Split(content, "\n")
	server := ""
	port := 1194

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "remote ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				server = parts[1]
			}
			if len(parts) >= 3 {
				fmt.Sscanf(parts[2], "%d", &port)
			}
		}
	}

	if server == "" {
		return nil, fmt.Errorf("no remote server found in openvpn config")
	}

	return &netbridge.Profile{
		Name:     fmt.Sprintf("ovpn-%s", server),
		Protocol: netbridge.ProtocolOpenVPN,
		Backend:  "openvpn",
		Server:   server,
		Port:     port,
	}, nil
}
