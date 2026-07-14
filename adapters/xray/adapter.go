package xray

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
	pid      int
	startAt  time.Time
	mu       sync.Mutex
}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Name() string {
	return "xray"
}

func (a *Adapter) SupportedProtocols() []netbridge.Protocol {
	return []netbridge.Protocol{
		netbridge.ProtocolVLESS,
		netbridge.ProtocolVMess,
		netbridge.ProtocolTrojan,
		netbridge.ProtocolShadowsocks,
		netbridge.ProtocolSOCKS,
		netbridge.ProtocolHTTP,
	}
}

func (a *Adapter) Start(ctx context.Context, cfg netbridge.BackendConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config = cfg
	a.running = true
	a.startAt = time.Now()

	return nil
}

func (a *Adapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.running = false
	a.pid = 0
	return nil
}

func (a *Adapter) Status() netbridge.BackendStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	return netbridge.BackendStatus{
		Running: a.running,
		PID:     a.pid,
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
		return fmt.Errorf("xray not running")
	}
	return nil
}

func (a *Adapter) LocalEndpoints() []netbridge.Endpoint {
	var endpoints []netbridge.Endpoint
	switch a.config.Mode {
	case netbridge.ModeSOCKS:
		endpoints = append(endpoints, netbridge.Endpoint{
			Type:    netbridge.ModeSOCKS,
			Address: fmt.Sprintf("127.0.0.1:%d", a.config.LocalPort),
		})
	case netbridge.ModeHTTP:
		endpoints = append(endpoints, netbridge.Endpoint{
			Type:    netbridge.ModeHTTP,
			Address: fmt.Sprintf("127.0.0.1:%d", a.config.LocalPort),
		})
	case netbridge.ModeTUN:
		endpoints = append(endpoints, netbridge.Endpoint{
			Type:    netbridge.ModeTUN,
			Address: a.config.TUNName,
			Iface:   a.config.TUNName,
		})
	}
	return endpoints
}
