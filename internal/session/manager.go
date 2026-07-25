package session

import (
	"context"
	"fmt"
	"time"

	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/profile"
)

type Manager struct {
	profileMgr    *profile.Manager
	pluginMgr     netbridge.PluginManager
	current       *netbridge.Session
	currentBackend netbridge.Backend
	status        netbridge.ConnectionStatus
}

func NewManager(pm *profile.Manager, plm netbridge.PluginManager) *Manager {
	return &Manager{
		profileMgr: pm,
		pluginMgr:  plm,
		status:     netbridge.StatusDisconnected,
	}
}

func (m *Manager) Connect(ctx context.Context, profileID string, mode netbridge.SessionMode) (*netbridge.Session, error) {
	if m.status == netbridge.StatusConnected {
		return nil, netbridge.ErrAlreadyConnected
	}

	p, err := m.profileMgr.Get(ctx, profileID)
	if err != nil {
		return nil, err
	}

	m.status = netbridge.StatusConnecting

	backend, err := m.pluginMgr.BackendFor(p.Protocol)
	if err != nil {
		m.status = netbridge.StatusDisconnected
		return nil, fmt.Errorf("no backend for %s: %w", p.Protocol, err)
	}

	bcfg := netbridge.BackendConfig{
		Profile:   *p,
		Mode:      mode,
		LocalPort: resolveLocalPort(mode),
		TUNName:   "tun0",
	}
	if err := backend.Start(ctx, bcfg); err != nil {
		m.status = netbridge.StatusDisconnected
		return nil, fmt.Errorf("backend start: %w", err)
	}

	// Retry health check up to 5 times with 1s delay
	var healthErr error
	for i := 0; i < 5; i++ {
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		healthErr = backend.HealthCheck(checkCtx)
		cancel()
		if healthErr == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if healthErr != nil {
		_ = backend.Stop()
		m.status = netbridge.StatusDisconnected
		return nil, fmt.Errorf("health check after start: %w", healthErr)
	}

	session := &netbridge.Session{
		ID:        generateSessionID(),
		ProfileID: p.ID,
		Mode:      mode,
		LocalAddr: resolveLocalAddr(mode),
		Status:    netbridge.StatusConnected,
		StartedAt: time.Now(),
	}

	m.current = session
	m.currentBackend = backend
	m.status = netbridge.StatusConnected

	return session, nil
}

func (m *Manager) Disconnect(ctx context.Context) error {
	if m.currentBackend != nil {
		_ = m.currentBackend.Stop()
		m.currentBackend = nil
	}
	if m.current != nil {
		now := time.Now()
		m.current.EndedAt = &now
		m.current.Status = netbridge.StatusDisconnected
	}
	m.status = netbridge.StatusDisconnected
	m.current = nil
	return nil
}

func (m *Manager) Restart(ctx context.Context) error {
	if m.current == nil {
		return netbridge.ErrNoActiveSession
	}
	profileID := m.current.ProfileID
	mode := m.current.Mode
	_ = m.Disconnect(ctx)
	_, err := m.Connect(ctx, profileID, mode)
	return err
}

func (m *Manager) Reload(ctx context.Context) error {
	if m.current == nil {
		return netbridge.ErrNoActiveSession
	}
	return nil
}

func (m *Manager) Current() (*netbridge.Session, error) {
	if m.current == nil {
		return nil, netbridge.ErrNoActiveSession
	}
	return m.current, nil
}

func (m *Manager) Status() netbridge.ConnectionStatus {
	return m.status
}

func (m *Manager) Stats() netbridge.TrafficStats {
	if m.current == nil {
		return netbridge.TrafficStats{}
	}
	return netbridge.TrafficStats{
		BytesUp:   m.current.BytesUp,
		BytesDown: m.current.BytesDown,
		Uptime:    time.Since(m.current.StartedAt),
	}
}

func (m *Manager) Persist(ctx context.Context) error {
	return nil
}

func (m *Manager) Recover(ctx context.Context) error {
	return nil
}

func generateSessionID() string {
	return fmt.Sprintf("sess-%d", time.Now().UnixNano())
}

func resolveLocalAddr(mode netbridge.SessionMode) string {
	switch mode {
	case netbridge.ModeSOCKS:
		return "127.0.0.1:10808"
	case netbridge.ModeHTTP:
		return "127.0.0.1:8080"
	case netbridge.ModeTUN:
		return "tun0"
	default:
		return "127.0.0.1:10808"
	}
}

func resolveLocalPort(mode netbridge.SessionMode) int {
	switch mode {
	case netbridge.ModeSOCKS:
		return 10808
	case netbridge.ModeHTTP:
		return 8080
	case netbridge.ModeTUN:
		return 12345
	default:
		return 10808
	}
}
