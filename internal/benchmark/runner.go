package benchmark

import (
	"fmt"
	"net"
	"time"

	netbridge "github.com/netbridge/netbridge"
)

type Runner struct{}

func (r *Runner) MeasureLatency(host string, port int) (time.Duration, error) {
	start := time.Now()
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return 0, err
	}
	conn.Close()
	return time.Since(start), nil
}

func (r *Runner) MeasureJitter(host string, port int, samples int) (time.Duration, error) {
	var latencies []time.Duration
	for i := 0; i < samples; i++ {
		lat, err := r.MeasureLatency(host, port)
		if err != nil {
			return 0, err
		}
		latencies = append(latencies, lat)
		time.Sleep(100 * time.Millisecond)
	}

	if len(latencies) < 2 {
		return 0, nil
	}

	var totalDiff time.Duration
	for i := 1; i < len(latencies); i++ {
		diff := latencies[i] - latencies[i-1]
		if diff < 0 {
			diff = -diff
		}
		totalDiff += diff
	}

	return totalDiff / time.Duration(len(latencies)-1), nil
}

func (r *Runner) RunBenchmark(host string, port int) (*netbridge.BenchmarkResult, error) {
	latency, err := r.MeasureLatency(host, port)
	if err != nil {
		return nil, err
	}

	jitter, _ := r.MeasureJitter(host, port, 5)

	return &netbridge.BenchmarkResult{
		Latency:    latency,
		Jitter:     jitter,
		Throughput: 0,
		PacketLoss: 0,
		Score:      0,
	}, nil
}
