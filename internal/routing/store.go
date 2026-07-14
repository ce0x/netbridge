package routing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	netbridge "github.com/netbridge/netbridge"
)

type Store struct {
	mu    sync.RWMutex
	dir   string
	rules []*netbridge.RouteRule
}

func NewStore(dataDir string) *Store {
	return &Store{
		dir: filepath.Join(dataDir, "routes"),
	}
}

func (s *Store) Load() ([]*netbridge.RouteRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.dir, "rules.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var rules []*netbridge.RouteRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *Store) Save(rules []*netbridge.RouteRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, "rules.json")
	return os.WriteFile(path, data, 0600)
}
