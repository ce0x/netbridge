package profile

import (
	"context"
	"testing"

	"github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/config"
)

func TestProfileManagerImport(t *testing.T) {
	cfg := config.DefaultConfig()
	mgr := NewManager(cfg)

	ctx := context.Background()

	_, err := mgr.Import(ctx, "vless://test@test.com:443?security=tls&sni=test.com#test-profile")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	profiles, err := mgr.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}

	if profiles[0].Name != "test-profile" {
		t.Errorf("expected name 'test-profile', got '%s'", profiles[0].Name)
	}
}

func TestProfileManagerCRUD(t *testing.T) {
	cfg := config.DefaultConfig()
	mgr := NewManager(cfg)
	ctx := context.Background()

	profile := &netbridge.Profile{
		Name:     "test",
		Protocol: "vless",
		Server:   "test.com",
		Port:     443,
	}

	err := mgr.Save(ctx, profile)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := mgr.Get(ctx, profile.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if got.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", got.Name)
	}

	err = mgr.Delete(ctx, profile.ID)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = mgr.Get(ctx, profile.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestProfileManagerSetActive(t *testing.T) {
	cfg := config.DefaultConfig()
	mgr := NewManager(cfg)
	ctx := context.Background()

	profile := &netbridge.Profile{
		Name:     "test",
		Protocol: "vless",
		Server:   "test.com",
		Port:     443,
	}

	err := mgr.Save(ctx, profile)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	err = mgr.SetActive(ctx, profile.ID)
	if err != nil {
		t.Fatalf("set active failed: %v", err)
	}

	active, err := mgr.GetActive(ctx)
	if err != nil {
		t.Fatalf("get active failed: %v", err)
	}

	if active.ID != profile.ID {
		t.Errorf("expected active profile ID %s, got %s", profile.ID, active.ID)
	}
}
