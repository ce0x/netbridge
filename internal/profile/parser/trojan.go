package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	netbridge "github.com/netbridge/netbridge"
)

func ParseTrojan(raw string) (*netbridge.Profile, error) {
	if !strings.HasPrefix(raw, "trojan://") {
		return nil, fmt.Errorf("not a trojan URI")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	password := u.User.Username()
	server := u.Hostname()
	portStr := u.Port()
	port := 443
	if portStr != "" {
		port, _ = strconv.Atoi(portStr)
	}

	query := u.Query()

	tls := netbridge.TLSConfig{
		Enabled:    true,
		ServerName: query.Get("sni"),
	}

	transport := netbridge.TransportConfig{
		Type: query.Get("type"),
		Path: query.Get("path"),
		Host: query.Get("host"),
	}

	outbound := map[string]any{
		"password": password,
	}

	return &netbridge.Profile{
		Name:      u.Fragment,
		Protocol:  netbridge.ProtocolTrojan,
		Backend:   "xray",
		RawURI:    raw,
		Server:    server,
		Port:      port,
		Transport: transport,
		TLS:       tls,
		Outbound:  outbound,
	}, nil
}
