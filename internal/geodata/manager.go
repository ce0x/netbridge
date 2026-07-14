package geodata

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Source struct {
	Name      string
	URL       string
	Format    string
	CacheDir  string
	FileName  string
	lastHash  string
	mu        sync.RWMutex
}

type Manager struct {
	sources  []*Source
	interval time.Duration
	stopCh   chan struct{}
}

func NewManager(cacheDir string, interval time.Duration) *Manager {
	return &Manager{
		sources: []*Source{
			{
				Name:     "geosite",
				URL:      "https://github.com/v2fly/domain-list-community/releases/latest/download/dlc.dat",
				Format:   "dat",
				CacheDir: cacheDir,
				FileName: "geosite.dat",
			},
			{
				Name:     "geoip",
				URL:      "https://github.com/v2fly/geoip/releases/latest/download/geoip.dat",
				Format:   "dat",
				CacheDir: cacheDir,
				FileName: "geoip.dat",
			},
		},
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (m *Manager) FetchAll() error {
	for _, src := range m.sources {
		if err := m.Fetch(src); err != nil {
			return fmt.Errorf("fetch %s: %w", src.Name, err)
		}
	}
	return nil
}

func (m *Manager) Fetch(src *Source) error {
	if err := os.MkdirAll(src.CacheDir, 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	tmpPath := filepath.Join(src.CacheDir, src.FileName+".tmp")
	finalPath := filepath.Join(src.CacheDir, src.FileName)

	resp, err := http.Get(src.URL)
	if err != nil {
		if _, statErr := os.Stat(finalPath); statErr == nil {
			return nil
		}
		return fmt.Errorf("download %s: %w", src.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if _, statErr := os.Stat(finalPath); statErr == nil {
			return nil
		}
		return fmt.Errorf("download %s: HTTP %d", src.Name, resp.StatusCode)
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(f, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("write file: %w", err)
	}

	f.Close()

	newHash := hex.EncodeToString(hasher.Sum(nil))

	src.mu.RLock()
	oldHash := src.lastHash
	src.mu.RUnlock()

	if newHash == oldHash {
		os.Remove(tmpPath)
		return nil
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename: %w", err)
	}

	src.mu.Lock()
	src.lastHash = newHash
	src.mu.Unlock()

	return nil
}

func (m *Manager) StartAutoUpdate() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = m.FetchAll()
		case <-m.stopCh:
			return
		}
	}
}

func (m *Manager) Stop() {
	close(m.stopCh)
}

func (m *Manager) GetPath(name string) string {
	for _, src := range m.sources {
		if src.Name == name {
			return filepath.Join(src.CacheDir, src.FileName)
		}
	}
	return ""
}

func (m *Manager) Exists(name string) bool {
	path := m.GetPath(name)
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func (m *Manager) CustomRulesPath() string {
	if len(m.sources) == 0 {
		return ""
	}
	return filepath.Join(m.sources[0].CacheDir, "custom-rules.txt")
}