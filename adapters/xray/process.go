package xray

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type Process struct {
	cmd     *exec.Cmd
	pid     int
	running bool
	startAt time.Time
	mu      sync.Mutex
}

func NewProcess() *Process {
	return &Process{}
}

func (p *Process) Start(configPath string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("already running (PID %d)", p.pid)
	}

	p.cmd = exec.Command("xray", "run", "-c", configPath)
	p.cmd.Stdout = nil
	p.cmd.Stderr = nil

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("start xray: %w", err)
	}

	p.pid = p.cmd.Process.Pid
	p.running = true
	p.startAt = time.Now()

	go p.monitor()
	return nil
}

func (p *Process) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.Process != nil {
		p.running = false
		return p.cmd.Process.Kill()
	}
	return nil
}

func (p *Process) Signal(sig syscall.Signal) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("no process")
	}
	return p.cmd.Process.Signal(sig)
}

func (p *Process) Kill() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.Process != nil {
		p.running = false
		return p.cmd.Process.Kill()
	}
	return nil
}

func (p *Process) monitor() {
	if p.cmd != nil {
		p.cmd.Wait()
		p.mu.Lock()
		p.running = false
		p.mu.Unlock()
	}
}

func (p *Process) WaitDone() {
	for {
		p.mu.Lock()
		running := p.running
		p.mu.Unlock()
		if !running {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (p *Process) PID() int {
	return p.pid
}

func (p *Process) Running() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

func (p *Process) Uptime() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.running {
		return 0
	}
	return time.Since(p.startAt)
}

// FindBinary searches for the xray binary in common locations.
func FindBinary() (string, error) {
	paths := []string{
		"xray",
		"/usr/local/bin/xray",
		"/usr/bin/xray",
	}
	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("xray binary not found in PATH or common locations")
}

// IsRunning checks if a process with the given PID is alive.
func IsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
