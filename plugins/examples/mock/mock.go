package mock

import (
	"context"
	"fmt"
	"time"

	netbridge "github.com/netbridge/netbridge"
)

type MockBackend struct {
	name    string
	running bool
	config  netbridge.BackendConfig
}

func NewMockBackend() *MockBackend {
	return &MockBackend{name: "mock"}
}

func (m *MockBackend) Name() string {
	return m.name
}

func (m *MockBackend) SupportedProtocols() []netbridge.Protocol {
	return []netbridge.Protocol{
		netbridge.ProtocolVLESS,
		netbridge.ProtocolVMess,
		netbridge.ProtocolTrojan,
		netbridge.ProtocolShadowsocks,
		netbridge.ProtocolSOCKS,
		netbridge.ProtocolHTTP,
	}
}

func (m *MockBackend) Start(ctx context.Context, cfg netbridge.BackendConfig) error {
	m.config = cfg
	m.running = true
	return nil
}

func (m *MockBackend) Stop() error {
	m.running = false
	return nil
}

func (m *MockBackend) Status() netbridge.BackendStatus {
	return netbridge.BackendStatus{
		Running: m.running,
		Uptime:  time.Since(time.Now()),
	}
}

func (m *MockBackend) Stats() netbridge.TrafficStats {
	return netbridge.TrafficStats{}
}

func (m *MockBackend) Configure(cfg netbridge.BackendConfig) error {
	m.config = cfg
	return nil
}

func (m *MockBackend) HealthCheck(ctx context.Context) error {
	if !m.running {
		return fmt.Errorf("mock backend not running")
	}
	return nil
}

func (m *MockBackend) LocalEndpoints() []netbridge.Endpoint {
	return []netbridge.Endpoint{
		{Type: netbridge.ModeSOCKS, Address: "127.0.0.1:10808"},
	}
}
