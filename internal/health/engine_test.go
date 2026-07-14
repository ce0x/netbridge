package health

import (
	"context"
	"testing"

	"github.com/netbridge/netbridge/internal/config"
	"github.com/netbridge/netbridge/internal/profile"
)

func TestHealthEngineCheck(t *testing.T) {
	cfg := config.DefaultConfig()
	pm := profile.NewManager(cfg)
	engine := NewEngine(pm)

	ctx := context.Background()

	_, err := engine.Check(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestHealthEngineCheckAll(t *testing.T) {
	cfg := config.DefaultConfig()
	pm := profile.NewManager(cfg)
	engine := NewEngine(pm)

	ctx := context.Background()

	results, err := engine.CheckAll(ctx)
	if err != nil {
		t.Fatalf("check all failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty profile list, got %d", len(results))
	}
}

func TestHealthEngineWatchdog(t *testing.T) {
	cfg := config.DefaultConfig()
	pm := profile.NewManager(cfg)
	engine := NewEngine(pm)

	ctx := context.Background()

	err := engine.StartWatchdog(ctx, 1)
	if err != nil {
		t.Fatalf("start watchdog failed: %v", err)
	}

	err = engine.StopWatchdog()
	if err != nil {
		t.Fatalf("stop watchdog failed: %v", err)
	}
}
