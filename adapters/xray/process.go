package xray

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type Process struct {
	cmd      *exec.Cmd
	pid      int
	running  bool
	startAt  time.Time
	logFile  *os.File
	logPath  string
	lines    []string
	exitedCh chan struct{}
	mu       sync.Mutex
}

const maxLogLines = 20

func NewProcess() *Process {
	return &Process{}
}

func (p *Process) Start(configPath string, logDir string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("already running (PID %d)", p.pid)
	}

	binary, err := FindBinary()
	if err != nil {
		return fmt.Errorf("xray binary not installed, run: netbridge core install xray")
	}

	if logDir == "" {
		logDir = filepath.Join(os.TempDir(), "netbridge-xray")
	}
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	logPath := filepath.Join(logDir, "xray.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}
	p.logFile = logFile

	p.cmd = exec.Command(binary, "run", "-c", configPath)
	p.cmd.Stdout = logFile
	p.cmd.Stderr = logFile

	if err := p.cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("start xray: %w", err)
	}

	p.pid = p.cmd.Process.Pid
	p.running = true
	p.startAt = time.Now()
	p.logPath = logPath
	p.exitedCh = make(chan struct{})

	go p.monitor()

	// Wait for xray to initialize
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (p *Process) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.Process != nil {
		p.running = false
		if err := p.cmd.Process.Kill(); err == nil {
			p.cmd.Wait()
		}
	}
	if p.logFile != nil {
		p.logFile.Close()
		p.logFile = nil
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
		p.readLogNow()
		if p.exitedCh != nil {
			close(p.exitedCh)
		}
	}
}

func (p *Process) readLogNow() {
	if p.logPath == "" {
		return
	}
	f, err := os.Open(p.logPath)
	if err != nil {
		return
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	p.mu.Lock()
	if len(lines) > maxLogLines {
		lines = lines[len(lines)-maxLogLines:]
	}
	p.lines = lines
	p.mu.Unlock()
}

func (p *Process) WaitExited() {
	if p.exitedCh != nil {
		<-p.exitedCh
	}
}

func (p *Process) LastLogLines() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lines
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

func FindBinary() (string, error) {
	paths := []string{
		"xray",
		"/usr/local/bin/xray",
		"/usr/bin/xray",
		"/usr/local/lib/netbridge/bin/xray",
	}
	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("xray binary not found in PATH or common locations")
}

func IsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
