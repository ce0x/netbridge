package profile

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/config"
	"github.com/netbridge/netbridge/internal/profile/parser"
)

type Manager struct {
	cfg     *config.Config
	profiles map[string]*netbridge.Profile
	activeID string
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg:      cfg,
		profiles: make(map[string]*netbridge.Profile),
	}
}

func (m *Manager) Import(ctx context.Context, raw string) (*netbridge.Profile, error) {
	p, err := parser.ParseURI(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", netbridge.ErrInvalidURI, err)
	}
	if err := m.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (m *Manager) ImportFile(ctx context.Context, path string) ([]*netbridge.Profile, error) {
	profiles, err := parser.ParseFile(path)
	if err != nil {
		return nil, err
	}
	for _, p := range profiles {
		if err := m.Save(ctx, p); err != nil {
			return nil, err
		}
	}
	return profiles, nil
}

func (m *Manager) ImportSubscription(ctx context.Context, url string) ([]*netbridge.Profile, error) {
	profiles, err := parser.ParseSubscription(url)
	if err != nil {
		return nil, err
	}
	for _, p := range profiles {
		if err := m.Save(ctx, p); err != nil {
			return nil, err
		}
	}
	return profiles, nil
}

func (m *Manager) Save(ctx context.Context, p *netbridge.Profile) error {
	if p.ID == "" {
		p.ID = generateID()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	p.UpdatedAt = time.Now()
	m.profiles[p.ID] = p
	return nil
}

func (m *Manager) Get(ctx context.Context, id string) (*netbridge.Profile, error) {
	p, ok := m.profiles[id]
	if !ok {
		return nil, netbridge.ErrProfileNotFound
	}
	return p, nil
}

func (m *Manager) GetByName(ctx context.Context, name string) (*netbridge.Profile, error) {
	for _, p := range m.profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, netbridge.ErrProfileNotFound
}

func (m *Manager) List(ctx context.Context) ([]*netbridge.Profile, error) {
	list := make([]*netbridge.Profile, 0, len(m.profiles))
	for _, p := range m.profiles {
		list = append(list, p)
	}
	return list, nil
}

func (m *Manager) Delete(ctx context.Context, id string) error {
	if _, ok := m.profiles[id]; !ok {
		return netbridge.ErrProfileNotFound
	}
	delete(m.profiles, id)
	return nil
}

func (m *Manager) Rename(ctx context.Context, id, newName string) error {
	p, ok := m.profiles[id]
	if !ok {
		return netbridge.ErrProfileNotFound
	}
	p.Name = newName
	p.UpdatedAt = time.Now()
	return nil
}

func (m *Manager) Clone(ctx context.Context, id, newName string) (*netbridge.Profile, error) {
	src, ok := m.profiles[id]
	if !ok {
		return nil, netbridge.ErrProfileNotFound
	}
	clone := *src
	clone.ID = generateID()
	clone.Name = newName
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = time.Now()
	m.profiles[clone.ID] = &clone
	return &clone, nil
}

func (m *Manager) Export(ctx context.Context, id string) (string, error) {
	p, ok := m.profiles[id]
	if !ok {
		return "", netbridge.ErrProfileNotFound
	}
	if p.RawURI != "" {
		return p.RawURI, nil
	}
	return fmt.Sprintf("netbridge://%s@%s:%d", p.Protocol, p.Server, p.Port), nil
}

func (m *Manager) Validate(ctx context.Context, p *netbridge.Profile) error {
	if p.Server == "" {
		return fmt.Errorf("server is required")
	}
	if p.Port <= 0 || p.Port > 65535 {
		return fmt.Errorf("invalid port: %d", p.Port)
	}
	return nil
}

func (m *Manager) SetActive(ctx context.Context, id string) error {
	if _, ok := m.profiles[id]; !ok {
		return netbridge.ErrProfileNotFound
	}
	m.activeID = id
	return nil
}

func (m *Manager) GetActive(ctx context.Context) (*netbridge.Profile, error) {
	if m.activeID == "" {
		return nil, netbridge.ErrProfileNotFound
	}
	return m.Get(ctx, m.activeID)
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
