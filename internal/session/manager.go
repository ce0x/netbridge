package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/config"
	"github.com/netbridge/netbridge/internal/profile"
)

type sessionState struct {
	ID           string               `json:"session_id"`
	ProfileID    string               `json:"profile_id"`
	Mode         netbridge.SessionMode `json:"mode"`
	LocalAddr    string               `json:"local_addr"`
	BackendName  string               `json:"backend_name"`
	BackendPID   int                  `json:"backend_pid"`
	StartedAt    time.Time            `json:"started_at"`
}

type Manager struct {
	profileMgr     *profile.Manager
	pluginMgr      netbridge.PluginManager
	current        *netbridge.Session
	currentBackend netbridge.Backend
	status         netbridge.ConnectionStatus
	stateFile      string
	lastKnownPID   int
}

func NewManager(pm *profile.Manager, plm netbridge.PluginManager, cfg *config.Config) *Manager {
	m := &Manager{
		profileMgr: pm,
		pluginMgr:  plm,
		status:     netbridge.StatusDisconnected,
	}
	if cfg != nil {
		m.stateFile = filepath.Join(cfg.DataDir, "session.json")
		os.MkdirAll(filepath.Dir(m.stateFile), 0o700)
	}
	return m
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

	// Polling health check: try every 200ms until context expires
	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var healthErr error
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-checkCtx.Done():
			if healthErr != nil {
				_ = backend.Stop()
				m.status = netbridge.StatusDisconnected
				return nil, fmt.Errorf("health check after start: %w", healthErr)
			}
			_ = backend.Stop()
			m.status = netbridge.StatusDisconnected
			return nil, fmt.Errorf("health check timeout")
		case <-ticker.C:
			healthErr = backend.HealthCheck(checkCtx)
			if healthErr == nil {
				break
			}
		}
		if healthErr == nil {
			break
		}
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

	// Persist state to disk
	_ = m.Persist(ctx)

	return session, nil
}

func (m *Manager) Disconnect(ctx context.Context) error {
	stoppedSomething := false

	if m.currentBackend != nil {
		_ = m.currentBackend.Stop()
		m.currentBackend = nil
		stoppedSomething = true
	} else if m.current != nil && m.lastKnownPID > 0 {
		// Recovered session — kill orphaned process directly
		if p, err := os.FindProcess(m.lastKnownPID); err == nil {
			_ = p.Kill()
		}
		m.lastKnownPID = 0
		stoppedSomething = true
	}

	if m.current != nil {
		now := time.Now()
		m.current.EndedAt = &now
		m.current.Status = netbridge.StatusDisconnected
	}
	m.status = netbridge.StatusDisconnected
	m.current = nil

	// Remove state file
	if m.stateFile != "" {
		os.Remove(m.stateFile)
	}

	if !stoppedSomething {
		return netbridge.ErrNoActiveSession
	}
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
	if m.stateFile == "" || m.current == nil {
		return nil
	}

	pid := 0
	if m.currentBackend != nil {
		pid = m.currentBackend.Status().PID
	}

	st := sessionState{
		ID:          m.current.ID,
		ProfileID:   m.current.ProfileID,
		Mode:        m.current.Mode,
		LocalAddr:   m.current.LocalAddr,
		BackendName: "xray",
		BackendPID:  pid,
		StartedAt:   m.current.StartedAt,
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.stateFile, data, 0o600)
}

func (m *Manager) Recover(ctx context.Context) error {
	if m.stateFile == "" {
		return nil
	}

	data, err := os.ReadFile(m.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var st sessionState
	if err := json.Unmarshal(data, &st); err != nil {
		os.Remove(m.stateFile)
		return nil
	}

	// Check if backend PID is still alive
	if st.BackendPID > 0 {
		if !pidAlive(st.BackendPID) {
			os.Remove(m.stateFile)
			return nil
		}
		m.lastKnownPID = st.BackendPID
	}

	// Reconstruct session
	m.current = &netbridge.Session{
		ID:        st.ID,
		ProfileID: st.ProfileID,
		Mode:      st.Mode,
		LocalAddr: st.LocalAddr,
		Status:    netbridge.StatusConnected,
		StartedAt: st.StartedAt,
	}
	m.status = netbridge.StatusConnected
	return nil
}

func pidAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
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
