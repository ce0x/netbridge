package dns

import (
	"context"
	"net"
	"time"

	netbridge "github.com/netbridge/netbridge"
)

type Benchmarker struct{}

func (b *Benchmarker) BenchmarkAll(ctx context.Context) ([]netbridge.DNSBenchResult, error) {
	var results []netbridge.DNSBenchResult

	for _, preset := range DefaultPresets {
		if preset.Name == "system" {
			continue
		}
		for _, server := range preset.Servers {
			start := time.Now()
			resolver := &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{}
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
