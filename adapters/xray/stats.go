package xray

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type StatsClient struct {
	binaryPath string
	serverAddr string
}

// xrayStatResponse matches xray-core v26+ JSON output format.
// "value" is int64 (number), not string.
type xrayStatResponse struct {
	Stat []struct {
		Name  string `json:"name"`
		Value int64  `json:"value"`
	} `json:"stat"`
}

func NewStatsClient(binaryPath string, apiPort int) *StatsClient {
	return &StatsClient{
		binaryPath: binaryPath,
		serverAddr: fmt.Sprintf("127.0.0.1:%d", apiPort),
	}
}

type TrafficStats struct {
	Up   int64
	Down int64
}

func (c *StatsClient) QueryStats(ctx context.Context) (*TrafficStats, error) {
	// Query outbound>>>proxy>>>traffic (matches outbound tag "proxy" in builder.go)
	// Each direction requires a separate query (uplink and downlink are separate patterns)
	up, err := c.queryPattern(ctx, "outbound>>>proxy>>>traffic>>>uplink")
	if err != nil {
		return nil, fmt.Errorf("query uplink: %w", err)
	}

	down, err := c.queryPattern(ctx, "outbound>>>proxy>>>traffic>>>downlink")
	if err != nil {
		return nil, fmt.Errorf("query downlink: %w", err)
	}

	return &TrafficStats{Up: up, Down: down}, nil
}

func (c *StatsClient) queryPattern(ctx context.Context, pattern string) (int64, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(queryCtx, c.binaryPath,
		"api", "statsquery",
		"--server="+c.serverAddr,
		"-pattern", pattern,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("xray api statsquery: %w (output: %s)", err, string(output))
	}

	return parseStatsOutput(output)
}

func parseStatsOutput(data []byte) (int64, error) {
	// Try JSON parsing first (xray 1.8+, value is int64 number)
	var resp xrayStatResponse
	if err := json.Unmarshal(data, &resp); err == nil {
		var total int64
		for _, s := range resp.Stat {
			total += s.Value
		}
		return total, nil
	}

	// Fallback: parse line-based output for older xray versions
	var total int64
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			valStr := strings.TrimSpace(line[idx+1:])
			if val, err := strconv.ParseInt(valStr, 10, 64); err == nil {
				total += val
				continue
			}
		}
		if val, err := strconv.ParseInt(line, 10, 64); err == nil {
			total += val
		}
	}

	return total, nil
}
