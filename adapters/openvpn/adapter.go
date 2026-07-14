package openvpn

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	netbridge "github.com/netbridge/netbridge"
)

type Adapter struct {
	config    netbridge.BackendConfig
	process   *Process
	configDir string
	state     ConnectionState
	stateMu   sync.RWMutex
	mu        sync.Mutex
}

type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateFailed
)

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Name() string {
	return "openvpn"
}

func (a *Adapter) SupportedProtocols() []netbridge.Protocol {
	return []netbridge.Protocol{netbridge.ProtocolOpenVPN}
}

func (a *Adapter) Start(ctx context.Context, cfg netbridge.BackendConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.process != nil && a.process.Running() {
		return fmt.Errorf("openvpn already running (PID %d)", a.process.PID())
	}

	a.config = cfg

	if err := os.MkdirAll(a.configDir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	configPath := filepath.Join(a.configDir, "client.ovpn")
	if err := a.generateConfig(cfg, configPath); err != nil {
		return fmt.Errorf("generate config: %w", err)
	}

	a.process = NewProcess()
	if err := a.process.Start(configPath); err != nil {
		return fmt.Errorf("start openvpn: %w", err)
	}

	a.setState(StateConnecting)
	go a.monitorOutput()

	log.Printf("openvpn started (PID %d)", a.process.PID())
	return nil
}

func (a *Adapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.process == nil || !a.process.Running() {
		return nil
	}

	if err := a.process.Signal(syscall.SIGTERM); err != nil {
		_ = a.process.Kill()
	}

	done := make(chan struct{})
	go func() {
		a.process.WaitDone()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		_ = a.process.Kill()
	}

	a.setState(StateDisconnected)
	log.Printf("openvpn stopped")
	return nil
}

func (a *Adapter) Reload(configPath string) error {
	return fmt.Errorf("openvpn does not support hot reload, use Stop+Start")
}

func (a *Adapter) Status() netbridge.BackendStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	running := a.process != nil && a.process.Running()
	pid := 0
	if a.process != nil {
		pid = a.process.PID()
	}
	return netbridge.BackendStatus{
		Running: running,
		PID:     pid,
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
	a.stateMu.RLock()
	state := a.state
	a.stateMu.RUnlock()

	if state != StateConnected {
		return fmt.Errorf("openvpn not connected (state: %d)", state)
	}
	return nil
}

func (a *Adapter) LocalEndpoints() []netbridge.Endpoint {
	return nil
}

func (a *Adapter) GetState() ConnectionState {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()
	return a.state
}

func (a *Adapter) setState(s ConnectionState) {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	a.state = s
}

func (a *Adapter) monitorOutput() {
	if a.process == nil || a.process.cmd == nil {
		return
	}

	stdout, err := a.process.cmd.StdoutPipe()
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		a.parseLine(line)
	}
}

func (a *Adapter) parseLine(line string) {
	switch {
	case strings.Contains(line, "Initialization Sequence Completed"):
		a.setState(StateConnected)
		log.Printf("openvpn: connected")
	case strings.Contains(line, "AUTH_FAILED"):
		a.setState(StateFailed)
		log.Printf("openvpn: auth failed")
	case strings.Contains(line, "Connection reset"):
		a.setState(StateConnecting)
		log.Printf("openvpn: reconnecting")
	case strings.Contains(line, "SIGTERM[soft,exit]"):
		a.setState(StateDisconnected)
		log.Printf("openvpn: disconnected")
	}
}

func (a *Adapter) generateConfig(cfg netbridge.BackendConfig, path string) error {
	var b strings.Builder

	b.WriteString("client\n")
	b.WriteString("dev tun\n")
	b.WriteString("proto udp\n")

	if cfg.Profile.Outbound != nil {
		if proto, ok := cfg.Profile.Outbound["proto"].(string); ok && proto != "" {
			b.WriteString(fmt.Sprintf("proto %s\n", proto))
		}
	}

	b.WriteString(fmt.Sprintf("remote %s %d\n", cfg.Profile.Server, cfg.Profile.Port))
	b.WriteString("resolv-retry infinite\n")
	b.WriteString("nobind\n")
	b.WriteString("persist-key\n")
	b.WriteString("persist-tun\n")

	if cfg.Profile.Outbound != nil {
		if cipher, ok := cfg.Profile.Outbound["cipher"].(string); ok && cipher != "" {
			b.WriteString(fmt.Sprintf("cipher %s\n", cipher))
		}
		if auth, ok := cfg.Profile.Outbound["auth"].(string); ok && auth != "" {
			b.WriteString(fmt.Sprintf("auth %s\n", auth))
		}
		if ca, ok := cfg.Profile.Outbound["ca"].(string); ok && ca != "" {
			b.WriteString("<ca>\n")
			b.WriteString(ca)
			b.WriteString("\n</ca>\n")
		}
		if cert, ok := cfg.Profile.Outbound["cert"].(string); ok && cert != "" {
			b.WriteString("<cert>\n")
			b.WriteString(cert)
			b.WriteString("\n</cert>\n")
		}
		if key, ok := cfg.Profile.Outbound["key"].(string); ok && key != "" {
			b.WriteString("<key>\n")
			b.WriteString(key)
			b.WriteString("\n</key>\n")
		}
	}

	b.WriteString("verb 3\n")

	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func (a *Adapter) WaitForConnection(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		a.stateMu.RLock()
		state := a.state
		a.stateMu.RUnlock()

		switch state {
		case StateConnected:
			return nil
		case StateFailed:
			return fmt.Errorf("openvpn connection failed")
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for openvpn connection")
}