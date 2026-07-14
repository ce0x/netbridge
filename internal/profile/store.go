package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/config"
)

type Store struct {
	dir string
}

func NewStore(cfg *config.Config) *Store {
	return &Store{
		dir: filepath.Join(cfg.DataDir, "profiles"),
	}
}

func (s *Store) Write(p *netbridge.Profile) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("create profiles dir: %w", err)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, p.ID+".json")
	return os.WriteFile(path, data, 0600)
}

func (s *Store) Read(id string) (*netbridge.Profile, error) {
	path := filepath.Join(s.dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p netbridge.Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) List() ([]*netbridge.Profile, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var profiles []*netbridge.Profile
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-5]
		p, err := s.Read(id)
		if err != nil {
			continue
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (s *Store) Delete(id string) error {
	path := filepath.Join(s.dir, id+".json")
	return os.Remove(path)
}
