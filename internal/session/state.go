package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	netbridge "github.com/netbridge/netbridge"
)

type State struct {
	mu       sync.RWMutex
	dir      string
	sessions []*netbridge.Session
}

func NewState(dataDir string) *State {
	return &State{
		dir: filepath.Join(dataDir, "state"),
	}
}

func (s *State) Save(session *netbridge.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, "active.json")
	return os.WriteFile(path, data, 0600)
}

func (s *State) Load() (*netbridge.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.dir, "active.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no saved session: %w", err)
	}
	var session netbridge.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *State) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := filepath.Join(s.dir, "active.json")
	return os.Remove(path)
}
