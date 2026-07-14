package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	netbridge "github.com/netbridge/netbridge"
)

func ParseVLESS(raw string) (*netbridge.Profile, error) {
	if !strings.HasPrefix(raw, "vless://") {
		return nil, fmt.Errorf("not a vless URI")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	server := u.Hostname()
	portStr := u.Port()
	port := 443
	if portStr != "" {
		port, _ = strconv.Atoi(portStr)
	}

	tags := []string{}
	if u.Fragment != "" {
		tags = append(tags, u.Fragment)
	}

	query := u.Query()

	tls := netbridge.TLSConfig{
		Enabled:       true,
		ServerName:    query.Get("sni"),
		Fingerprint:   query.Get("fp"),
		AllowInsecure: query.Get("allowInsecure") == "1",
	}
	if query.Get("pbk") != "" {
		tls.RealityPublicKey = query.Get("pbk")
		tls.RealityShortID = query.Get("sid")
	}

	transport := netbridge.TransportConfig{
		Type: query.Get("type"),
		Path: query.Get("path"),
		Host: query.Get("host"),
	}

	return &netbridge.Profile{
		Name:      u.Fragment,
		Protocol:  netbridge.ProtocolVLESS,
		Backend:   "xray",
		RawURI:    raw,
		Server:    server,
		Port:      port,
		Transport: transport,
		TLS:       tls,
		Tags:      tags,
	}, nil
}
