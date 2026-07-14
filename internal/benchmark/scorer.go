package benchmark

import (
	"math"

	netbridge "github.com/netbridge/netbridge"
)

type Scorer struct{}

func (s *Scorer) Score(r *netbridge.BenchmarkResult) int {
	latencyScore := 100.0
	if r.Latency.Milliseconds() > 0 {
		latencyScore = math.Max(0, 100-float64(r.Latency.Milliseconds())/10)
	}

	throughputScore := 0.0
	if r.Throughput > 0 {
		throughputScore = math.Min(100, r.Throughput/1024/1024*10)
	}

	lossScore := math.Max(0, 100-r.PacketLoss*100)

	score := (latencyScore * 0.4) + (throughputScore * 0.3) + (lossScore * 0.3)
	return int(math.Round(math.Min(100, math.Max(0, score))))
}

func (s *Scorer) Rank(results []*netbridge.BenchmarkResult) []*netbridge.BenchmarkResult {
	for _, r := range results {
		r.Score = s.Score(r)
	}

	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	return results
}
