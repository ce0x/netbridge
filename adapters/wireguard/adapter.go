package wireguard

import (
	"context"
	"fmt"
	"sync"
	"time"

	netbridge "github.com/netbridge/netbridge"
)

type Adapter struct {
	config   netbridge.BackendConfig
	running  bool
	iface    string
	startAt  time.Time
	mu       sync.Mutex
}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Name() string {
	return "wireguard"
}

func (a *Adapter) SupportedProtocols() []netbridge.Protocol {
	return []netbridge.Protocol{netbridge.ProtocolWireGuard}
}

func (a *Adapter) Start(ctx context.Context, cfg netbridge.BackendConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config = cfg
	a.running = true
	a.startAt = time.Now()
	a.iface = "wg0"
	return nil
}

func (a *Adapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.running = false
	return nil
}

func (a *Adapter) Status() netbridge.BackendStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	return netbridge.BackendStatus{
		Running: a.running,
		Uptime:  time.Since(a.startAt),
	}
}

func (a *Adapter) Stats() netbridge.TrafficStats {
	return netbridge.TrafficStats{}
}

func (a *Adapter) Configure(cfg netbridge.BackendConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config = cfg
	return nil
}

func (a *Adapter) HealthCheck(ctx context.Context) error {
	if !a.running {
		return fmt.Errorf("wireguard not running")
	}
	return nil
}

func (a *Adapter) LocalEndpoints() []netbridge.Endpoint {
	return []netbridge.Endpoint{
		{Type: netbridge.ModeTUN, Address: a.iface, Iface: a.iface},
	}
}
