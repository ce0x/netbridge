package parser

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	netbridge "github.com/netbridge/netbridge"
)

func ParseShadowsocks(raw string) (*netbridge.Profile, error) {
	if !strings.HasPrefix(raw, "ss://") {
		return nil, fmt.Errorf("not a shadowsocks URI")
	}

	rest := raw[5:]

	atIdx := strings.LastIndex(rest, "@")
	if atIdx == -1 {
		decoded, err := base64.StdEncoding.DecodeString(rest)
		if err != nil {
			return nil, fmt.Errorf("invalid ss base64: %w", err)
		}
		rest = string(decoded)
		atIdx = strings.LastIndex(rest, "@")
		if atIdx == -1 {
			return nil, fmt.Errorf("invalid ss uri format")
		}
	}

	userInfo := rest[:atIdx]
	serverPart := rest[atIdx+1:]

	decodedInfo, err := base64.StdEncoding.DecodeString(userInfo)
	if err != nil {
		decodedInfo = []byte(userInfo)
	}

	parts := strings.SplitN(string(decodedInfo), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ss userinfo")
	}
	method := parts[0]
	password := parts[1]

	serverHost := serverPart
	serverPort := 8388

	u, err := url.Parse("ss://" + serverPart)
	if err == nil {
		serverHost = u.Hostname()
		if p := u.Port(); p != "" {
			serverPort, _ = strconv.Atoi(p)
		}
	}

	name := fmt.Sprintf("ss-%s", serverHost)

	return &netbridge.Profile{
		Name:     name,
		Protocol: netbridge.ProtocolShadowsocks,
		Backend:  "xray",
		RawURI:   raw,
		Server:   serverHost,
		Port:     serverPort,
		Outbound: map[string]any{
			"method":   method,
			"password": password,
		},
	}, nil
}
