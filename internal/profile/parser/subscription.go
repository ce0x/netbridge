package parser

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	netbridge "github.com/netbridge/netbridge"
)

func ParseSubscription(url string) ([]*netbridge.Profile, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(body)))
	if err != nil {
		decoded = body
	}

	lines := strings.Split(string(decoded), "\n")
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
