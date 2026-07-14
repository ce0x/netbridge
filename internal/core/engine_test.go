package core

import (
	"context"
	"testing"

	"github.com/netbridge/netbridge/internal/config"
)

func TestEngineNew(t *testing.T) {
	cfg := config.DefaultConfig()
	engine, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	if engine == nil {
		t.Fatal("engine is nil")
	}

	if engine.ProfileManager() == nil {
		t.Error("ProfileManager is nil")
	}

	if engine.SessionManager() == nil {
		t.Error("SessionManager is nil")
	}

	if engine.RoutingEngine() == nil {
		t.Error("RoutingEngine is nil")
	}

	if engine.HealthEngine() == nil {
		t.Error("HealthEngine is nil")
	}

	if engine.BenchmarkEngine() == nil {
		t.Error("BenchmarkEngine is nil")
	}

	if engine.DNSEngine() == nil {
		t.Error("DNSEngine is nil")
	}

	if engine.PluginManager() == nil {
		t.Error("PluginManager is nil")
	}
}

func TestEngineEnvVars(t *testing.T) {
	cfg := config.DefaultConfig()
	engine, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	vars := engine.EnvVars()
	if _, ok := vars["http_proxy"]; !ok {
		t.Error("missing http_proxy")
	}
	if _, ok := vars["https_proxy"]; !ok {
		t.Error("missing https_proxy")
	}
	if _, ok := vars["all_proxy"]; !ok {
		t.Error("missing all_proxy")
	}
}

func TestEngineShutdown(t *testing.T) {
	cfg := config.DefaultConfig()
	engine, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	err = engine.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}
