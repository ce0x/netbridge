package benchmark

import (
	"context"
	"testing"

	"github.com/netbridge/netbridge/internal/config"
	"github.com/netbridge/netbridge/internal/health"
	"github.com/netbridge/netbridge/internal/profile"
)

func TestBenchmarkEngineRunAll(t *testing.T) {
	cfg := config.DefaultConfig()
	pm := profile.NewManager(cfg)
	he := health.NewEngine(pm)
	engine := NewEngine(pm, he)

	ctx := context.Background()

	results, err := engine.RunAll(ctx)
	if err != nil {
		t.Fatalf("run all failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty profile list, got %d", len(results))
	}
}

func TestBenchmarkEngineBest(t *testing.T) {
	cfg := config.DefaultConfig()
	pm := profile.NewManager(cfg)
	he := health.NewEngine(pm)
	engine := NewEngine(pm, he)

	ctx := context.Background()

	_, err := engine.Best(ctx)
	if err == nil {
		t.Error("expected error for no profiles")
	}
}
