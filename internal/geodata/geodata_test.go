package geodata

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestManager_Fetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test-geo-data"))
	}))
	defer server.Close()

	dir := t.TempDir()
	src := &Source{
		Name:     "test",
		URL:      server.URL,
		Format:   "dat",
		CacheDir: dir,
		FileName: "test.dat",
	}

	m := &Manager{sources: []*Source{src}}

	if err := m.Fetch(src); err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	path := filepath.Join(dir, "test.dat")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if string(data) != "test-geo-data" {
		t.Errorf("content = %q, want %q", string(data), "test-geo-data")
	}
}

func TestManager_FetchNetworkErrorFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.dat")
	os.WriteFile(path, []byte("cached-data"), 0o644)

	src := &Source{
		Name:     "test",
		URL:      "http://nonexistent.example.com",
		Format:   "dat",
		CacheDir: dir,
		FileName: "existing.dat",
	}

	m := &Manager{sources: []*Source{src}}

	if err := m.Fetch(src); err != nil {
		t.Errorf("Fetch should fallback to cache, got: %v", err)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "cached-data" {
		t.Error("cached file should be preserved")
	}
}

func TestManager_Exists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "geosite.dat")
	os.WriteFile(path, []byte("data"), 0o644)

	m := &Manager{
		sources: []*Source{
			{Name: "geosite", CacheDir: dir, FileName: "geosite.dat"},
		},
	}

	if !m.Exists("geosite") {
		t.Error("Exists should return true for existing file")
	}
	if m.Exists("nonexistent") {
		t.Error("Exists should return false for missing file")
	}
}

func TestManager_GetPath(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{
		sources: []*Source{
			{Name: "geosite", CacheDir: dir, FileName: "geosite.dat"},
		},
	}

	path := m.GetPath("geosite")
	expected := filepath.Join(dir, "geosite.dat")
	if path != expected {
		t.Errorf("GetPath = %q, want %q", path, expected)
	}

	if m.GetPath("nonexistent") != "" {
		t.Error("GetPath should return empty for unknown source")
	}
}

func TestResolver_LookupDomains(t *testing.T) {
	r := NewResolver()
	r.domains["category-ads"] = []string{"ads.example.com", "tracking.example.com"}

	domains, err := r.LookupDomains("category-ads")
	if err != nil {
		t.Fatalf("LookupDomains failed: %v", err)
	}

	if len(domains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(domains))
	}
}

func TestResolver_LookupDomainsNotFound(t *testing.T) {
	r := NewResolver()
	_, err := r.LookupDomains("nonexistent")
	if err == nil {
		t.Error("LookupDomains should error for unknown category")
	}
}

func TestResolver_LookupCIDRs(t *testing.T) {
	r := NewResolver()
	r.cidrs["cn"] = []string{"1.0.0.0/8", "2.0.0.0/8"}

	cidrs, err := r.LookupCIDRs("cn")
	if err != nil {
		t.Fatalf("LookupCIDRs failed: %v", err)
	}

	if len(cidrs) != 2 {
		t.Errorf("expected 2 cidrs, got %d", len(cidrs))
	}
}

func TestResolver_Categories(t *testing.T) {
	r := NewResolver()
	r.domains["cat1"] = []string{"a.com"}
	r.cidrs["cat2"] = []string{"1.0.0.0/8"}

	cats := r.Categories()
	if len(cats) != 2 {
		t.Errorf("expected 2 categories, got %d", len(cats))
	}
}

func TestResolver_LoadDomains(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("category-ads,ads.example.com\ncategory-ads,tracking.example.com\n# comment\n"), 0o644)

	r := NewResolver()
	if err := r.LoadDomains(path); err != nil {
		t.Fatalf("LoadDomains failed: %v", err)
	}

	domains, err := r.LookupDomains("category-ads")
	if err != nil {
		t.Fatalf("LookupDomains failed: %v", err)
	}

	if len(domains) != 2 {
		t.Errorf("expected 2 domains, got %d", len(domains))
	}
}