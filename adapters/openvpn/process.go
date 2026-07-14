package openvpn

import (
	"fmt"
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

	p.cmd = exec.Command("openvpn", "--config", configPath)
	p.cmd.Stdout = nil
	p.cmd.Stderr = nil

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("start openvpn: %w", err)
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