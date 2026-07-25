package corebin

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type GitHubInstaller struct {
	installDir string
}

func NewGitHubInstaller(installDir string) *GitHubInstaller {
	return &GitHubInstaller{installDir: installDir}
}

func (g *GitHubInstaller) Detect(name string) (installed bool, version string, path string) {
	binaryPath := filepath.Join(g.installDir, name)
	if _, err := os.Stat(binaryPath); err == nil {
		return true, detectVersion(binaryPath), binaryPath
	}
	return false, "", ""
}

func (g *GitHubInstaller) LatestVersion(ctx context.Context, repo string) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return "", "", fmt.Errorf("GitHub API rate limit exceeded")
	}
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	return release.TagName, release.TagName, nil
}

func (g *GitHubInstaller) findAsset(release githubRelease, name string) (string, error) {
	goos := platform()
	goarch := arch()

	for _, asset := range release.Assets {
		assetName := strings.ToLower(asset.Name)

		switch name {
		case "xray":
			// Xray-linux-64.zip / Xray-linux-arm64-v8a.zip
			if goos == "linux" && goarch == "amd64" && strings.Contains(assetName, "linux") && strings.Contains(assetName, "64") {
				return asset.BrowserDownloadURL, nil
			}
			if goos == "linux" && goarch == "arm64" && strings.Contains(assetName, "linux") && strings.Contains(assetName, "arm64") {
				return asset.BrowserDownloadURL, nil
			}
			if goos == "darwin" && goarch == "amd64" && strings.Contains(assetName, "macos") && strings.Contains(assetName, "64") {
				return asset.BrowserDownloadURL, nil
			}
			if goos == "darwin" && goarch == "arm64" && strings.Contains(assetName, "macos") && strings.Contains(assetName, "arm64") {
				return asset.BrowserDownloadURL, nil
			}

		case "sing-box":
			// sing-box-<ver>-linux-amd64.tar.gz
			ver := strings.TrimPrefix(release.TagName, "v")
			expected := fmt.Sprintf("sing-box-%s-%s-%s", ver, goos, goarch)
			if strings.Contains(assetName, expected) {
				return asset.BrowserDownloadURL, nil
			}
		}
	}

	return "", fmt.Errorf("no matching asset found for %s on %s/%s", name, goos, goarch)
}

func (g *GitHubInstaller) Install(ctx context.Context, name string) error {
	repo := getRepo(name)
	if repo == "" {
		return fmt.Errorf("no GitHub repo configured for %s", name)
	}

	_, tag, err := g.LatestVersion(ctx, repo)
	if err != nil {
		return fmt.Errorf("get latest version: %w", err)
	}

	url, err := g.getDownloadURL(ctx, repo, tag, name)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "netbridge-core-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "download")
	if err := downloadFile(ctx, url, zipPath); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	if err := os.MkdirAll(g.installDir, 0o755); err != nil {
		return fmt.Errorf("create install dir: %w", err)
	}

	if strings.HasSuffix(url, ".zip") {
		if err := extractZip(zipPath, g.installDir, name); err != nil {
			return fmt.Errorf("extract: %w", err)
		}
	} else if strings.HasSuffix(url, ".tar.gz") {
		if err := extractTarGz(zipPath, g.installDir, name); err != nil {
			return fmt.Errorf("extract: %w", err)
		}
	}

	binaryPath := filepath.Join(g.installDir, name)
	if err := os.Chmod(binaryPath, 0o755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	fmt.Printf("Installed %s to %s\n", name, binaryPath)
	return nil
}

func (g *GitHubInstaller) Update(ctx context.Context, name string) error {
	return g.Install(ctx, name)
}

func (g *GitHubInstaller) Verify(name string) error {
	binaryPath := filepath.Join(g.installDir, name)
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("binary not found: %w", err)
	}
	// Basic check: try to run version
	out, err := exec.Command(binaryPath, "version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("binary verification failed: %w\nOutput: %s", err, string(out))
	}
	return nil
}

func (g *GitHubInstaller) Repair(ctx context.Context, name string) error {
	// Remove existing and reinstall
	binaryPath := filepath.Join(g.installDir, name)
	os.Remove(binaryPath)
	return g.Install(ctx, name)
}

func (g *GitHubInstaller) getDownloadURL(ctx context.Context, repo, tag, name string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, tag)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return g.findAsset(release, name)
}

func getRepo(name string) string {
	switch name {
	case "xray":
		return "XTLS/Xray-core"
	case "sing-box":
		return "SagerNet/sing-box"
	default:
		return ""
	}
}

func downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractZip(zipPath, destDir, binaryName string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		base := filepath.Base(f.Name)
		// For xray: the binary is named "xray" inside the zip
		if base == binaryName || base == binaryName+".exe" {
			outPath := filepath.Join(destDir, binaryName)
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			in, err := f.Open()
			if err != nil {
				return err
			}
			defer in.Close()

			_, err = io.Copy(outFile, in)
			return err
		}
	}

	return fmt.Errorf("binary %s not found in zip", binaryName)
}

func extractTarGz(tgzPath, destDir, binaryName string) error {
	f, err := os.Open(tgzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		base := filepath.Base(header.Name)
		if base == binaryName {
			outPath := filepath.Join(destDir, binaryName)
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, tr)
			return err
		}
	}

	return fmt.Errorf("binary %s not found in tar.gz", binaryName)
}
