package singbox

import (
	"fmt"
	"os/exec"
	"sync"
)

type Process struct {
	cmd     *exec.Cmd
	pid     int
	running bool
	mu      sync.Mutex
}

func NewProcess() *Process {
	return &Process{}
}

func (p *Process) Start(configPath string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cmd = exec.Command("sing-box", "run", "-c", configPath)
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("start sing-box: %w", err)
	}
	p.pid = p.cmd.Process.Pid
	p.running = true
	go p.wait()
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

func (p *Process) wait() {
	if p.cmd != nil {
		p.cmd.Wait()
		p.mu.Lock()
		p.running = false
		p.mu.Unlock()
	}
}
