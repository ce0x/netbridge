package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/profile"
)

type FailoverManager struct {
	chain      *netbridge.FailoverChain
	profileMgr *profile.Manager
	health     *Engine
	mu         sync.Mutex
}

func NewFailoverManager(pm *profile.Manager, he *Engine) *FailoverManager {
	return &FailoverManager{
		profileMgr: pm,
		health:     he,
	}
}

func (f *FailoverManager) Create(name string, profileIDs []string) *netbridge.FailoverChain {
	f.chain = &netbridge.FailoverChain{
		ID:                 fmt.Sprintf("fo-%s", name),
		Name:               name,
		ProfileIDs:         profileIDs,
		CurrentIndex:       0,
		HealthCheckInterval: 30 * time.Second,
		FailThreshold:      3,
	}
	return f.chain
}

func (f *FailoverManager) GetCurrent() (*netbridge.Profile, error) {
	if f.chain == nil || len(f.chain.ProfileIDs) == 0 {
		return nil, fmt.Errorf("no failover chain configured")
	}
	return f.profileMgr.Get(context.Background(), f.chain.ProfileIDs[f.chain.CurrentIndex])
}

func (f *FailoverManager) SwitchToNext() (*netbridge.Profile, error) {
	if f.chain == nil {
		return nil, fmt.Errorf("no failover chain")
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	f.chain.CurrentIndex++
	if f.chain.CurrentIndex >= len(f.chain.ProfileIDs) {
		f.chain.CurrentIndex = 0
	}

	return f.profileMgr.Get(context.Background(), f.chain.ProfileIDs[f.chain.CurrentIndex])
}

func (f *FailoverManager) Status() map[string]any {
	if f.chain == nil {
		return map[string]any{"active": false}
	}
	current := ""
	if p, err := f.GetCurrent(); err == nil {
		current = p.Name
	}
	return map[string]any{
		"active":        true,
		"name":          f.chain.Name,
		"current":       current,
		"currentIndex":  f.chain.CurrentIndex,
		"totalProfiles": len(f.chain.ProfileIDs),
	}
}
