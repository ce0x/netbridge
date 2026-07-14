package uri

import (
	"fmt"
	"net/url"
	"strings"
)

type ParsedURI struct {
	Scheme   string
	User     string
	Password string
	Host     string
	Port     string
	Path     string
	Fragment string
	Params   url.Values
}

func Parse(raw string) (*ParsedURI, error) {
	raw = strings.TrimSpace(raw)

	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("parse uri: %w", err)
		}
		return &ParsedURI{
			Scheme:   u.Scheme,
			User:     u.User.Username(),
			Password: func() string { p, _ := u.User.Password(); return p }(),
			Host:     u.Hostname(),
			Port:     u.Port(),
			Path:     u.Path,
			Fragment: u.Fragment,
			Params:   u.Query(),
		}, nil
	}

	return &ParsedURI{Host: raw}, nil
}
