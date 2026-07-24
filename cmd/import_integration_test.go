package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/netbridge/netbridge/internal/config"
	"github.com/netbridge/netbridge/internal/profile"
)

const testVLESSURI = "vless://abc123def456@server.example.com:443?encryption=none&security=tls&sni=server.example.com&fp=chrome&flow=xtls-rprx-vision&type=tcp#MyTestProfile"

func TestIntegrationImportAndPersist(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{DataDir: dir}

	// --- Manager 1: import ---
	mgr1 := profile.NewManager(cfg)
	ctx := context.Background()

	p, err := mgr1.Import(ctx, testVLESSURI)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected non-empty profile ID")
	}
	if p.Name != "MyTestProfile" {
		t.Fatalf("expected name 'MyTestProfile', got '%s'", p.Name)
	}
	if p.Server != "server.example.com" {
		t.Fatalf("expected server 'server.example.com', got '%s'", p.Server)
	}
	if p.Port != 443 {
		t.Fatalf("expected port 443, got %d", p.Port)
	}

	// --- Verify JSON file on disk ---
	profileDir := filepath.Join(dir, "profiles")
	entries, err := os.ReadDir(profileDir)
	if err != nil {
		t.Fatalf("failed to read profiles dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 JSON file on disk, got %d", len(entries))
	}
	if filepath.Ext(entries[0].Name()) != ".json" {
		t.Fatalf("expected .json extension, got %s", filepath.Ext(entries[0].Name()))
	}

	// Verify JSON content is valid and contains expected fields
	raw, err := os.ReadFile(filepath.Join(profileDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("failed to read profile JSON: %v", err)
	}
	var diskProfile struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Server   string `json:"server"`
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
	}
	if err := json.Unmarshal(raw, &diskProfile); err != nil {
		t.Fatalf("failed to unmarshal profile JSON: %v", err)
	}
	if diskProfile.Name != "MyTestProfile" {
		t.Fatalf("disk profile name mismatch: got '%s'", diskProfile.Name)
	}
	if diskProfile.Server != "server.example.com" {
		t.Fatalf("disk profile server mismatch: got '%s'", diskProfile.Server)
	}

	// --- Manager 2: should load from disk ---
	mgr2 := profile.NewManager(cfg)
	profiles, err := mgr2.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile loaded from disk, got %d", len(profiles))
	}

	loaded := profiles[0]
	if loaded.ID != p.ID {
		t.Fatalf("loaded profile ID mismatch: got '%s', expected '%s'", loaded.ID, p.ID)
	}
	if loaded.Name != "MyTestProfile" {
		t.Fatalf("loaded profile name mismatch: got '%s'", loaded.Name)
	}
	if loaded.RawURI != testVLESSURI {
		t.Fatalf("loaded profile RawURI mismatch")
	}

	// --- Manager 2: resolve by name ---
	resolved, err := mgr2.GetByName(ctx, "MyTestProfile")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if resolved.ID != p.ID {
		t.Fatalf("GetByName ID mismatch: got '%s'", resolved.ID)
	}

	// --- Manager 2: export ---
	uri, err := mgr2.Export(ctx, resolved.ID)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if uri != testVLESSURI {
		t.Fatalf("Export URI mismatch: got '%s'", uri)
	}
}

func TestIntegrationDeleteAndVerifyDisk(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{DataDir: dir}
	ctx := context.Background()

	mgr := profile.NewManager(cfg)
	p, err := mgr.Import(ctx, testVLESSURI)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	// Delete
	if err := mgr.Delete(ctx, p.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Verify file removed
	profileDir := filepath.Join(dir, "profiles")
	entries, err := os.ReadDir(profileDir)
	if err != nil {
		t.Fatalf("failed to read profiles dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 files after delete, got %d", len(entries))
	}

	// New manager should see nothing
	mgr2 := profile.NewManager(cfg)
	profiles, err := mgr2.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(profiles) != 0 {
		t.Fatalf("expected 0 profiles after delete, got %d", len(profiles))
	}
}
