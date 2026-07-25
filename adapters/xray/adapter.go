package xray

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	netbridge "github.com/netbridge/netbridge"
)

type Adapter struct {
	config      netbridge.BackendConfig
	builder     *Builder
	process     *Process
	configDir   string
	statsClient *StatsClient
	mu          sync.Mutex
}

func New() *Adapter {
	return &Adapter{
		builder: &Builder{},
	}
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

	if a.process != nil && a.process.Running() {
		return fmt.Errorf("already running (PID %d)", a.process.PID())
	}

	a.config = cfg
	a.configDir = filepath.Join(os.TempDir(), "netbridge-xray", cfg.Profile.ID)
	logDir := filepath.Join(a.configDir, "logs")

	if err := os.MkdirAll(a.configDir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := a.builder.BuildConfig(cfg)
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}

	configPath := filepath.Join(a.configDir, "xray.json")
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	a.process = NewProcess()
	if err := a.process.Start(configPath, logDir); err != nil {
		return fmt.Errorf("start xray: %w", err)
	}

	// Wait briefly for xray to start or crash
	time.Sleep(500 * time.Millisecond)
	if !a.process.Running() {
		// Process exited — wait for monitor() to collect logs
		a.process.WaitExited()
		lines := a.process.LastLogLines()
		errMsg := "xray crashed immediately after start"
		if len(lines) > 0 {
			errMsg += ":\n"
			for _, line := range lines {
				errMsg += "  " + line + "\n"
			}
		}
		return fmt.Errorf("%s", errMsg)
	}

	log.Printf("xray started (PID %d)", a.process.PID())

	// Initialize stats client for traffic monitoring
	apiPort := cfg.LocalPort + 1000
	a.statsClient = NewStatsClient(apiPort)

	return nil
}

func (a *Adapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.process == nil || !a.process.Running() {
		return nil
	}

	if err := a.process.Signal(syscall.SIGTERM); err != nil {
		log.Printf("SIGTERM failed, sending SIGKILL: %v", err)
		_ = a.process.Kill()
	}

	done := make(chan struct{})
	go func() {
		a.process.WaitDone()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = a.process.Kill()
	}

	log.Printf("xray stopped")
	return nil
}

func (a *Adapter) Reload(configPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.process == nil || !a.process.Running() {
		return fmt.Errorf("xray not running, cannot reload")
	}

	data, err := a.builder.BuildConfig(a.config)
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}

	path := filepath.Join(a.configDir, "xray.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if err := a.process.Signal(syscall.SIGHUP); err != nil {
		log.Printf("SIGHUP reload failed, falling back to restart: %v", err)
		if stopErr := a.Stop(); stopErr != nil {
			return fmt.Errorf("stop for restart: %w", stopErr)
		}
		return a.Start(context.Background(), a.config)
	}

	log.Printf("xray reloaded via SIGHUP")
	return nil
}

func (a *Adapter) Status() netbridge.BackendStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.process == nil {
		return netbridge.BackendStatus{Running: false}
	}

	return netbridge.BackendStatus{
		Running: a.process.Running(),
		PID:     a.process.PID(),
		Uptime:  a.process.Uptime(),
	}
}

func (a *Adapter) Stats() netbridge.TrafficStats {
	a.mu.Lock()
	running := a.process != nil && a.process.Running()
	process := a.process
	client := a.statsClient
	a.mu.Unlock()

	if !running || process == nil {
		return netbridge.TrafficStats{}
	}

	stats := netbridge.TrafficStats{
		Uptime: process.Uptime(),
	}

	// Query real traffic stats from xray gRPC API
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if ts, err := client.QueryStats(ctx); err == nil {
			stats.BytesUp = ts.Up
			stats.BytesDown = ts.Down
		}
	}

	return stats
}

func (a *Adapter) Configure(cfg netbridge.BackendConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config = cfg
	return nil
}

func (a *Adapter) HealthCheck(ctx context.Context) error {
	a.mu.Lock()
	running := a.process != nil && a.process.Running()
	config := a.config
	a.mu.Unlock()

	if !running {
		return fmt.Errorf("xray not running")
	}

	addr := fmt.Sprintf("127.0.0.1:%d", config.LocalPort)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	conn.Close()
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
