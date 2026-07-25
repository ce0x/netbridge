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

	outbound := map[string]any{}

	var currentBlock string
	var blockContent strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "<") && strings.HasSuffix(line, ">") && !strings.HasPrefix(line, "</") {
			tag := line[1 : len(line)-1]
			currentBlock = tag
			blockContent.Reset()
			continue
		}
		if strings.HasPrefix(line, "</") && strings.HasSuffix(line, ">") {
			if currentBlock != "" {
				outbound[currentBlock] = blockContent.String()
				currentBlock = ""
			}
			continue
		}
		if currentBlock != "" {
			blockContent.WriteString(line)
			blockContent.WriteString("\n")
			continue
		}

		if strings.HasPrefix(line, "remote ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				server = parts[1]
			}
			if len(parts) >= 3 {
				fmt.Sscanf(parts[2], "%d", &port)
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			key := parts[0]
			val := strings.TrimSpace(parts[1])
			switch key {
			case "cipher":
				outbound["cipher"] = val
			case "auth":
				outbound["auth"] = val
			case "proto":
				outbound["proto"] = val
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
		Outbound: outbound,
	}, nil
}
