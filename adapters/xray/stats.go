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

type xrayStatResponse struct {
	Stat []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
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
	up, err := c.queryPattern(ctx, "user>>>traffic>>>uplink")
	if err != nil {
		return nil, fmt.Errorf("query uplink: %w", err)
	}

	down, err := c.queryPattern(ctx, "user>>>traffic>>>downlink")
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
	// Try JSON parsing first (xray 1.8+)
	var resp xrayStatResponse
	if err := json.Unmarshal(data, &resp); err == nil {
		var total int64
		for _, s := range resp.Stat {
			if val, err := strconv.ParseInt(s.Value, 10, 64); err == nil {
				total += val
			}
		}
		return total, nil
	}

	// Fallback: parse line-based output
	// Format varies by xray version, try common patterns:
	// "name: value" or "value: 12345" or just a number
	var total int64
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try "name: value" format
		if idx := strings.Index(line, ":"); idx > 0 {
			valStr := strings.TrimSpace(line[idx+1:])
			if val, err := strconv.ParseInt(valStr, 10, 64); err == nil {
				total += val
				continue
			}
		}

		// Try plain number
		if val, err := strconv.ParseInt(line, 10, 64); err == nil {
			total += val
		}
	}

	return total, nil
}
