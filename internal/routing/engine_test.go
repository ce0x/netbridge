package routing

import (
	"context"
	"testing"

	netbridge "github.com/netbridge/netbridge"
)

func TestRoutingEngineAddRemove(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	rule := netbridge.RouteRule{
		Pattern:   "github.com",
		RuleType:  "domain",
		ProfileID: "profile-1",
		Priority:  10,
	}

	err := engine.AddRule(ctx, rule)
	if err != nil {
		t.Fatalf("add rule failed: %v", err)
	}

	rules, err := engine.ListRules(ctx)
	if err != nil {
		t.Fatalf("list rules failed: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	err = engine.RemoveRule(ctx, rules[0].ID)
	if err != nil {
		t.Fatalf("remove rule failed: %v", err)
	}

	rules, _ = engine.ListRules(ctx)
	if len(rules) != 0 {
		t.Errorf("expected 0 rules after remove, got %d", len(rules))
	}
}

func TestRoutingEngineResolve(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	_ = engine.AddRule(ctx, netbridge.RouteRule{
		Pattern:   "github.com",
		RuleType:  "domain",
		ProfileID: "profile-1",
		Priority:  10,
	})

	_ = engine.AddRule(ctx, netbridge.RouteRule{
		Pattern:   "docker.io",
		RuleType:  "domain_suffix",
		ProfileID: "profile-2",
		Priority:  5,
	})

	profileID, err := engine.Resolve("github.com")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if profileID != "profile-1" {
		t.Errorf("expected profile-1, got %s", profileID)
	}

	profileID, _ = engine.Resolve("registry.docker.io")
	if profileID != "profile-2" {
		t.Errorf("expected profile-2 for docker.io suffix, got %s", profileID)
	}

	profileID, _ = engine.Resolve("docker.io")
	if profileID != "profile-2" {
		t.Errorf("expected profile-2 for exact docker.io, got %s", profileID)
	}

	profileID, _ = engine.Resolve("google.com")
	if profileID != "" {
		t.Errorf("expected direct (empty) for unmatched domain, got %s", profileID)
	}
}

func TestRoutingEngineClear(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	_ = engine.AddRule(ctx, netbridge.RouteRule{
		Pattern:   "github.com",
		RuleType:  "domain",
		ProfileID: "profile-1",
		Priority:  10,
	})

	err := engine.ClearRules(ctx)
	if err != nil {
		t.Fatalf("clear rules failed: %v", err)
	}

	rules, _ := engine.ListRules(ctx)
	if len(rules) != 0 {
		t.Errorf("expected 0 rules after clear, got %d", len(rules))
	}
}
