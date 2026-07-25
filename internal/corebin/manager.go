package corebin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Installer interface {
	Detect(name string) (installed bool, version string, path string)
	Install(ctx context.Context, name string) error
	Update(ctx context.Context, name string) error
	Verify(name string) error
	Repair(ctx context.Context, name string) error
}

type Manager struct {
	installDir string
_cores      map[string]CoreDef
}

type CoreDef struct {
	Name        string
	Source      string // "github" or "pacman"
	Description string
}

func NewManager() *Manager {
	m := &Manager{
		installDir: filepath.Join("/usr/local/lib/netbridge/bin"),
		_cores:     make(map[string]CoreDef),
	}

	m._cores["xray"] = CoreDef{
		Name:        "xray",
		Source:      "github",
		Description: "Xray-core proxy tool",
	}
	m._cores["sing-box"] = CoreDef{
		Name:        "sing-box",
		Source:      "github",
		Description: "sing-box universal proxy platform",
	}
	m._cores["wireguard-tools"] = CoreDef{
		Name:        "wireguard-tools",
		Source:      "pacman",
		Description: "WireGuard userspace tools",
	}
	m._cores["openvpn"] = CoreDef{
		Name:        "openvpn",
		Source:      "pacman",
		Description: "OpenVPN client",
	}

	return m
}

func (m *Manager) ListCores() []CoreDef {
	var list []CoreDef
	for _, c := range m._cores {
		list = append(list, c)
	}
	return list
}

func (m *Manager) Detect(name string) (installed bool, version string, path string) {
	// Check our install dir first
	localPath := filepath.Join(m.installDir, name)
	if _, err := os.Stat(localPath); err == nil {
		return true, detectVersion(localPath), localPath
	}

	// Check system PATH
	if sysPath, err := exec.LookPath(name); err == nil {
		return true, detectVersion(sysPath), sysPath
	}

	return false, "", ""
}

func detectVersion(binaryPath string) string {
	out, err := exec.Command(binaryPath, "version").CombinedOutput()
	if err != nil {
		return "unknown"
	}
	// Extract version from output (first line usually)
	lines := strings.SplitN(string(out), "\n", 2)
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return "unknown"
}

func (m *Manager) GetCoreDef(name string) (CoreDef, error) {
	c, ok := m._cores[name]
	if !ok {
		return CoreDef{}, fmt.Errorf("unknown core: %s", name)
	}
	return c, nil
}

func (m *Manager) InstallDir() string {
	return m.installDir
}

func platform() string {
	return runtime.GOOS
}

func arch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return runtime.GOARCH
	}
}
