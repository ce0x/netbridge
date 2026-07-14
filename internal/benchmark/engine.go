package benchmark

import (
	"context"
	"fmt"
	"math"
	"time"

	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/health"
	"github.com/netbridge/netbridge/internal/profile"
)

type Engine struct {
	profileMgr *profile.Manager
	health     *health.Engine
}

func NewEngine(pm *profile.Manager, he *health.Engine) *Engine {
	return &Engine{
		profileMgr: pm,
		health:     he,
	}
}

func (e *Engine) Run(ctx context.Context, profileID string) (*netbridge.BenchmarkResult, error) {
	p, err := e.profileMgr.Get(ctx, profileID)
	if err != nil {
		return nil, err
	}

	result := &netbridge.BenchmarkResult{
		ProfileID: profileID,
	}

	healthResult, err := e.health.Check(ctx, profileID)
	if err == nil {
		result.Latency = healthResult.Latency
		result.PacketLoss = healthResult.PacketLoss
	}

	result.Jitter = time.Duration(math.Abs(float64(result.Latency) * 0.1))
	result.Throughput = float64(1024*1024*10) / math.Max(float64(result.Latency.Milliseconds()), 1)
	result.Score = e.calculateScore(result)

	_ = p
	return result, nil
}

func (e *Engine) RunAll(ctx context.Context) ([]*netbridge.BenchmarkResult, error) {
	profiles, err := e.profileMgr.List(ctx)
	if err != nil {
		return nil, err
	}

	var results []*netbridge.BenchmarkResult
	for _, p := range profiles {
		result, err := e.Run(ctx, p.ID)
		if err != nil {
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

func (e *Engine) Best(ctx context.Context) (string, error) {
	results, err := e.RunAll(ctx)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no profiles available")
	}

	best := results[0]
	for _, r := range results[1:] {
		if r.Score > best.Score {
			best = r
		}
	}
	return best.ProfileID, nil
}

func (e *Engine) calculateScore(r *netbridge.BenchmarkResult) int {
	latencyScore := math.Max(0, 100-float64(r.Latency.Milliseconds())/10)
	throughputScore := math.Min(100, r.Throughput/1024/1024*10)
	lossScore := math.Max(0, 100-r.PacketLoss*100)

	score := (latencyScore * 0.4) + (throughputScore * 0.3) + (lossScore * 0.3)
	return int(math.Round(math.Min(100, math.Max(0, score))))
}
