package plugins

import (
	"fmt"
	"sync"

	netbridge "github.com/netbridge/netbridge"
)

type Registry struct {
	mu      sync.RWMutex
	plugins map[string]netbridge.Plugin
}

func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]netbridge.Plugin),
	}
}

func (r *Registry) Load(path string) (netbridge.Plugin, error) {
	return nil, fmt.Errorf("plugin loading not yet implemented: %s", path)
}

func (r *Registry) Unload(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.plugins, name)
	return nil
}

func (r *Registry) List() []netbridge.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]netbridge.Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		list = append(list, p)
	}
	return list
}

func (r *Registry) Get(name string) (netbridge.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return p, nil
}

func (r *Registry) BackendFor(protocol netbridge.Protocol) (netbridge.Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.plugins {
		for _, pp := range p.Protocols() {
			if pp == protocol {
				return p.NewBackend(netbridge.Profile{Protocol: protocol})
			}
		}
	}
	return nil, netbridge.ErrBackendNotFound
}

func (r *Registry) Register(p netbridge.Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins[p.Name()] = p
}
