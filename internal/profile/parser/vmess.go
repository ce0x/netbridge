package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	netbridge "github.com/netbridge/netbridge"
)

type vmessConfig struct {
	V    string `json:"v"`
	PS   string `json:"ps"`
	Add  string `json:"add"`
	Port string `json:"port"`
	ID   string `json:"id"`
	Aid  string `json:"aid"`
	Net  string `json:"net"`
	Type string `json:"type"`
	Host string `json:"host"`
	Path string `json:"path"`
	TLS  string `json:"tls"`
	SNI  string `json:"sni"`
}

func ParseVMess(raw string) (*netbridge.Profile, error) {
	if len(raw) < 9 || raw[:8] != "vmess://" {
		return nil, fmt.Errorf("not a vmess URI")
	}

	encoded := raw[8:]
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid vmess base64: %w", err)
	}

	var cfg vmessConfig
	if err := json.Unmarshal(decoded, &cfg); err != nil {
		return nil, fmt.Errorf("invalid vmess json: %w", err)
	}

	port, _ := strconv.Atoi(cfg.Port)

	tls := netbridge.TLSConfig{
		Enabled: cfg.TLS == "tls",
	}
	if cfg.SNI != "" {
		tls.ServerName = cfg.SNI
	}

	transport := netbridge.TransportConfig{
		Type: cfg.Net,
		Path: cfg.Path,
		Host: cfg.Host,
	}

	return &netbridge.Profile{
		Name:      cfg.PS,
		Protocol:  netbridge.ProtocolVMess,
		Backend:   "xray",
		RawURI:    raw,
		Server:    cfg.Add,
		Port:      port,
		Transport: transport,
		TLS:       tls,
	}, nil
}
