package plugins

import (
	"fmt"
	"sync"

	netbridge "github.com/netbridge/netbridge"
)

type PluginEntry struct {
	Name     string
	Version  string
	Protocol netbridge.Protocol
	Factory  func(netbridge.Profile) (netbridge.Backend, error)
}

type PluginRegistry struct {
	mu      sync.RWMutex
	plugins map[string]*PluginEntry
}

func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]*PluginEntry),
	}
}

func (pr *PluginRegistry) Register(entry *PluginEntry) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.plugins[entry.Name] = entry
}

func (pr *PluginRegistry) Get(name string) (*PluginEntry, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	e, ok := pr.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return e, nil
}

func (pr *PluginRegistry) List() []*PluginEntry {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	list := make([]*PluginEntry, 0, len(pr.plugins))
	for _, e := range pr.plugins {
		list = append(list, e)
	}
	return list
}
