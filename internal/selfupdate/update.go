package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
 CurrentVersion = "dev"
 RepoOwner      = "ce0x"
 RepoName       = "netbridge"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type UpdateInfo struct {
	Current    string
	Latest     string
	DownloadURL string
	UpdateAvailable bool
}

func CheckLatest(ctx context.Context) (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("GitHub API rate limit exceeded")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(CurrentVersion, "v")

	downloadURL := findBinaryAsset(release)

	return &UpdateInfo{
		Current:         current,
		Latest:          latest,
		DownloadURL:     downloadURL,
		UpdateAvailable: latest != current && downloadURL != "",
	}, nil
}

func findBinaryAsset(release githubRelease) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)

		switch {
		case goos == "linux" && goarch == "amd64":
			if strings.Contains(name, "linux") && strings.Contains(name, "amd64") && strings.HasSuffix(name, ".tar.gz") {
				return asset.BrowserDownloadURL
			}
		case goos == "linux" && goarch == "arm64":
			if strings.Contains(name, "linux") && strings.Contains(name, "arm64") && strings.HasSuffix(name, ".tar.gz") {
				return asset.BrowserDownloadURL
			}
		case goos == "darwin" && goarch == "amd64":
			if strings.Contains(name, "macos") && strings.Contains(name, "amd64") && strings.HasSuffix(name, ".tar.gz") {
				return asset.BrowserDownloadURL
			}
		case goos == "darwin" && goarch == "arm64":
			if strings.Contains(name, "macos") && strings.Contains(name, "arm64") && strings.HasSuffix(name, ".tar.gz") {
				return asset.BrowserDownloadURL
			}
		case goos == "windows" && goarch == "amd64":
			if strings.Contains(name, "windows") && strings.Contains(name, "amd64") && strings.HasSuffix(name, ".zip") {
				return asset.BrowserDownloadURL
			}
		}
	}
	return ""
}

func Install(ctx context.Context, info *UpdateInfo) error {
	if info == nil || !info.UpdateAvailable {
		return fmt.Errorf("no update available")
	}

	// Get current binary path
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	self, _ = filepath.EvalSymlinks(self)

	// Download to temp file
	tmpFile, err := os.CreateTemp("", "netbridge-update-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", info.DownloadURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("write download: %w", err)
	}
	tmpFile.Close()

	// Extract binary from archive
	tmpDir, err := os.MkdirTemp("", "netbridge-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if strings.HasSuffix(info.DownloadURL, ".tar.gz") {
		if err := extractTarGz(tmpFile.Name(), tmpDir); err != nil {
			return fmt.Errorf("extract: %w", err)
		}
	} else if strings.HasSuffix(info.DownloadURL, ".zip") {
		if err := extractZip(tmpFile.Name(), tmpDir); err != nil {
			return fmt.Errorf("extract: %w", err)
		}
	}

	// Find extracted binary
	var newBinary string
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "netbridge" || info.Name() == "netbridge.exe" {
			newBinary = path
		}
		return nil
	})

	if newBinary == "" {
		return fmt.Errorf("netbridge binary not found in archive")
	}

	// Atomic replace
	if err := os.Rename(newBinary, self); err != nil {
		// Fallback: copy
		src, err := os.Open(newBinary)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(self)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
	}

	os.Chmod(self, 0755)
	return nil
}

func extractTarGz(tgzPath, destDir string) error {
	// Simplified extraction - in production use archive/tar
	return fmt.Errorf("tar.gz extraction not implemented for %s", runtime.GOOS)
}

func extractZip(zipPath, destDir string) error {
	// Simplified extraction - in production use archive/zip
	return fmt.Errorf("zip extraction not implemented for %s", runtime.GOOS)
}
