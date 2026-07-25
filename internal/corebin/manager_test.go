package corebin

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestManager_ListCores(t *testing.T) {
	m := NewManager()
	cores := m.ListCores()
	if len(cores) != 4 {
		t.Errorf("expected 4 cores, got %d", len(cores))
	}
}

func TestManager_GetCoreDef(t *testing.T) {
	m := NewManager()

	def, err := m.GetCoreDef("xray")
	if err != nil {
		t.Fatalf("GetCoreDef(xray) failed: %v", err)
	}
	if def.Source != "github" {
		t.Errorf("expected source 'github', got '%s'", def.Source)
	}

	def, err = m.GetCoreDef("openvpn")
	if err != nil {
		t.Fatalf("GetCoreDef(openvpn) failed: %v", err)
	}
	if def.Source != "pacman" {
		t.Errorf("expected source 'pacman', got '%s'", def.Source)
	}

	_, err = m.GetCoreDef("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent core")
	}
}

func TestManager_Detect_NotInstalled(t *testing.T) {
	m := NewManager()
	installed, version, path := m.Detect("nonexistent-binary-12345")
	if installed {
		t.Error("expected not installed")
	}
	if version != "" {
		t.Errorf("expected empty version, got '%s'", version)
	}
	if path != "" {
		t.Errorf("expected empty path, got '%s'", path)
	}
}

func TestManager_Detect_Installed(t *testing.T) {
	m := NewManager()
	// go is always available in test environment
	installed, version, path := m.Detect("go")
	if !installed {
		t.Skip("go not found in PATH")
	}
	if version == "" {
		t.Error("expected non-empty version")
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestManager_Detect_LocalBinary(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{
		installDir: dir,
		_cores:     make(map[string]CoreDef),
	}

	// Create a fake binary
	binaryPath := filepath.Join(dir, "fake-binary")
	os.WriteFile(binaryPath, []byte("#!/bin/sh\necho fake 1.0"), 0o755)

	installed, version, path := m.Detect("fake-binary")
	if !installed {
		t.Error("expected installed from local dir")
	}
	if path != binaryPath {
		t.Errorf("expected path '%s', got '%s'", binaryPath, path)
	}
	_ = version // version may be "unknown" for fake binary
}

func TestGitHubInstaller_findAsset_Xray(t *testing.T) {
	g := NewGitHubInstaller("/tmp/test")

	release := githubRelease{
		TagName: "v1.8.0",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "Xray-linux-64.zip", BrowserDownloadURL: "https://example.com/linux-64.zip"},
			{Name: "Xray-linux-arm64-v8a.zip", BrowserDownloadURL: "https://example.com/linux-arm64.zip"},
			{Name: "Xray-macos-64.zip", BrowserDownloadURL: "https://example.com/macos-64.zip"},
			{Name: "Xray-macos-arm64-v8a.zip", BrowserDownloadURL: "https://example.com/macos-arm64.zip"},
			{Name: "Xray-windows-64.zip", BrowserDownloadURL: "https://example.com/win-64.zip"},
		},
	}

	url, err := g.findAsset(release, "xray")
	if err != nil {
		t.Fatalf("findAsset failed: %v", err)
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	var expected string
	switch {
	case goos == "linux" && goarch == "amd64":
		expected = "https://example.com/linux-64.zip"
	case goos == "linux" && goarch == "arm64":
		expected = "https://example.com/linux-arm64.zip"
	case goos == "darwin" && goarch == "amd64":
		expected = "https://example.com/macos-64.zip"
	case goos == "darwin" && goarch == "arm64":
		expected = "https://example.com/macos-arm64.zip"
	case goos == "windows" && goarch == "amd64":
		expected = "https://example.com/win-64.zip"
	default:
		t.Skip("no matching asset for this platform")
	}
	if url != expected {
		t.Errorf("expected URL '%s', got '%s'", expected, url)
	}
}

func TestGitHubInstaller_findAsset_SingBox(t *testing.T) {
	g := NewGitHubInstaller("/tmp/test")

	release := githubRelease{
		TagName: "v1.10.0",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "sing-box-1.10.0-linux-amd64.tar.gz", BrowserDownloadURL: "https://example.com/sb-linux-amd64.tar.gz"},
			{Name: "sing-box-1.10.0-linux-arm64.tar.gz", BrowserDownloadURL: "https://example.com/sb-linux-arm64.tar.gz"},
			{Name: "sing-box-1.10.0-darwin-amd64.tar.gz", BrowserDownloadURL: "https://example.com/sb-darwin-amd64.tar.gz"},
			{Name: "sing-box-1.10.0-windows-amd64.tar.gz", BrowserDownloadURL: "https://example.com/sb-windows-amd64.tar.gz"},
		},
	}

	url, err := g.findAsset(release, "sing-box")
	if err != nil {
		t.Fatalf("findAsset failed: %v", err)
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	var expectedURL string
	switch {
	case goos == "linux" && goarch == "amd64":
		expectedURL = "https://example.com/sb-linux-amd64.tar.gz"
	case goos == "linux" && goarch == "arm64":
		expectedURL = "https://example.com/sb-linux-arm64.tar.gz"
	case goos == "darwin" && goarch == "amd64":
		expectedURL = "https://example.com/sb-darwin-amd64.tar.gz"
	case goos == "windows" && goarch == "amd64":
		expectedURL = "https://example.com/sb-windows-amd64.tar.gz"
	default:
		t.Skip("no matching asset for this platform")
	}
	if url != expectedURL {
		t.Errorf("expected URL '%s', got '%s'", expectedURL, url)
	}
}

func TestGitHubInstaller_findAsset_NoMatch(t *testing.T) {
	g := NewGitHubInstaller("/tmp/test")

	release := githubRelease{
		TagName: "v1.0.0",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "Xray-windows-64.zip", BrowserDownloadURL: "https://example.com/win.zip"},
		},
	}

	// On non-Windows, this should fail
	if runtime.GOOS != "windows" {
		_, err := g.findAsset(release, "xray")
		if err == nil {
			t.Error("expected error for no matching asset")
		}
	}
}

func TestDetectPackageManager(t *testing.T) {
	pm := detectPackageManager()
	// Should return something, even if "unknown"
	if pm == "" {
		t.Error("expected non-empty package manager")
	}
}

func TestGetPackageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"wireguard-tools", "wireguard-tools"},
		{"openvpn", "openvpn"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		result := getPackageName(tt.input)
		if result != tt.expected {
			t.Errorf("getPackageName(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "dpkg output",
			input:    "Package: wireguard-tools\nStatus: install ok installed\nVersion: 1.0.20210914-1",
			expected: "1.0.20210914-1",
		},
		{
			name:     "simple version",
			input:    "xray 1.8.0",
			expected: "xray 1.8.0",
		},
		{
			name:     "empty",
			input:    "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersion(tt.input)
			if result != tt.expected {
				t.Errorf("extractVersion() = %s, want %s", result, tt.expected)
			}
		})
	}
}
