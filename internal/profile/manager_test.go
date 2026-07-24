package profile

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/config"
)

func tempConfig(t *testing.T) *config.Config {
	t.Helper()
	dir := t.TempDir()
	return &config.Config{
		DataDir: dir,
	}
}

func TestProfileManagerImport(t *testing.T) {
	cfg := tempConfig(t)
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
	cfg := tempConfig(t)
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
	cfg := tempConfig(t)
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

func TestProfilePersistence(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{DataDir: dir}

	// First manager: save a profile
	mgr1 := NewManager(cfg)
	ctx := context.Background()

	_, err := mgr1.Import(ctx, "vless://user@server.example.com:443?security=tls&sni=server.example.com#persist-test")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	// Verify file exists on disk
	profileDir := filepath.Join(dir, "profiles")
	entries, err := os.ReadDir(profileDir)
	if err != nil {
		t.Fatalf("failed to read profiles dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file on disk, got %d", len(entries))
	}

	// Second manager: should load the profile from disk
	mgr2 := NewManager(cfg)
	profiles, err := mgr2.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile loaded from disk, got %d", len(profiles))
	}
	if profiles[0].Name != "persist-test" {
		t.Errorf("expected name 'persist-test', got '%s'", profiles[0].Name)
	}
}

func TestProfileDeleteFromDisk(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{DataDir: dir}
	mgr := NewManager(cfg)
	ctx := context.Background()

	p, err := mgr.Import(ctx, "vless://user@del.example.com:443?security=tls&sni=del.example.com#delete-test")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	err = mgr.Delete(ctx, p.ID)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Verify file removed from disk
	profileDir := filepath.Join(dir, "profiles")
	entries, err := os.ReadDir(profileDir)
	if err != nil {
		t.Fatalf("failed to read profiles dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 files on disk after delete, got %d", len(entries))
	}

	// Second manager should see nothing
	mgr2 := NewManager(cfg)
	profiles, err := mgr2.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(profiles) != 0 {
		t.Fatalf("expected 0 profiles after delete, got %d", len(profiles))
	}
}
