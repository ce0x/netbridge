package stats

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	netbridge "github.com/netbridge/netbridge"
)

type Store struct {
	mu   sync.RWMutex
	dir  string
	data []netbridge.TrafficStats
}

func NewStore(dataDir string) *Store {
	return &Store{
		dir: filepath.Join(dataDir, "stats"),
	}
}

func (s *Store) Record(stats netbridge.TrafficStats) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = append(s.data, stats)

	if len(s.data) > 1000 {
		s.data = s.data[len(s.data)-1000:]
	}

	return nil
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, "history.json")
	return os.WriteFile(path, data, 0640)
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dir, "history.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.data)
}
