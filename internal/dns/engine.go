package dns

import (
	"context"
	"net"
	"time"

	netbridge "github.com/netbridge/netbridge"
)

type Engine struct {
	currentResolver string
	presets         []netbridge.DNSPreset
}

func NewEngine() *Engine {
	return &Engine{
		currentResolver: "system",
		presets: []netbridge.DNSPreset{
			{Name: "cloudflare", Servers: []string{"1.1.1.1", "1.0.0.1"}},
			{Name: "google", Servers: []string{"8.8.8.8", "8.8.4.4"}},
			{Name: "quad9", Servers: []string{"9.9.9.9", "149.112.112.112"}},
			{Name: "adguard", Servers: []string{"94.140.14.14", "94.140.15.15"}},
			{Name: "system", Servers: []string{}},
		},
	}
}

func (e *Engine) ListPresets() []netbridge.DNSPreset {
	return e.presets
}

func (e *Engine) SetResolver(ctx context.Context, nameOrAddr string) error {
	for _, p := range e.presets {
		if p.Name == nameOrAddr {
			e.currentResolver = p.Name
			return nil
		}
	}
	e.currentResolver = nameOrAddr
	return nil
}

func (e *Engine) CurrentResolver() string {
	return e.currentResolver
}

func (e *Engine) Benchmark(ctx context.Context) ([]netbridge.DNSBenchResult, error) {
	var results []netbridge.DNSBenchResult

	for _, preset := range e.presets {
		if preset.Name == "system" {
			continue
		}
		for _, server := range preset.Servers {
			start := time.Now()
			resolver := &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{Timeout: 5 * time.Second}
					return d.DialContext(ctx, "udp", server+":53")
				},
			}
			_, err := resolver.LookupHost(ctx, "google.com")
			latency := time.Since(start)

			result := netbridge.DNSBenchResult{
				Name:    preset.Name,
				Server:  server,
				Latency: latency,
			}
			if err != nil {
				result.Error = err
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func (e *Engine) Reset(ctx context.Context) error {
	e.currentResolver = "system"
	return nil
}
