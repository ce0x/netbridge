package corebin

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type PackageManagerInstaller struct{}

func NewPackageManagerInstaller() *PackageManagerInstaller {
	return &PackageManagerInstaller{}
}

func (p *PackageManagerInstaller) Detect(name string) (installed bool, version string, path string) {
	// Check if package is installed via dpkg/rpm/pacman/apk
	cmds := [][]string{
		{"dpkg", "-s", name},
		{"rpm", "-q", name},
		{"pacman", "-Qi", name},
		{"apk", "info", name},
	}

	for _, cmd := range cmds {
		if out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err == nil {
			ver := extractVersion(string(out))
			return true, ver, ""
		}
	}

	// Check if binary exists in PATH
	if sysPath, err := exec.LookPath(name); err == nil {
		return true, detectVersion(sysPath), sysPath
	}

	return false, "", ""
}

func (p *PackageManagerInstaller) Install(ctx context.Context, name string) error {
	pkgName := getPackageName(name)
	pm := detectPackageManager()

	var cmd *exec.Cmd
	switch pm {
	case "apt":
		cmd = exec.CommandContext(ctx, "apt-get", "install", "-y", pkgName)
	case "dnf", "yum":
		cmd = exec.CommandContext(ctx, pm, "install", "-y", pkgName)
	case "pacman":
		cmd = exec.CommandContext(ctx, "pacman", "-S", "--noconfirm", pkgName)
	case "apk":
		cmd = exec.CommandContext(ctx, "apk", "add", pkgName)
	default:
		return fmt.Errorf("unsupported package manager: %s", pm)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install %s via %s: %w\n%s", pkgName, pm, err, stderr.String())
	}

	fmt.Printf("Installed %s via %s\n", name, pm)
	return nil
}

func (p *PackageManagerInstaller) Update(ctx context.Context, name string) error {
	pkgName := getPackageName(name)
	pm := detectPackageManager()

	var cmd *exec.Cmd
	switch pm {
	case "apt":
		cmd = exec.CommandContext(ctx, "apt-get", "install", "-y", "--only-upgrade", pkgName)
	case "dnf", "yum":
		cmd = exec.CommandContext(ctx, pm, "update", "-y", pkgName)
	case "pacman":
		cmd = exec.CommandContext(ctx, "pacman", "-Syu", "--noconfirm", pkgName)
	case "apk":
		cmd = exec.CommandContext(ctx, "apk", "upgrade", pkgName)
	default:
		return fmt.Errorf("unsupported package manager: %s", pm)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update %s: %w\n%s", pkgName, err, stderr.String())
	}

	fmt.Printf("Updated %s via %s\n", name, pm)
	return nil
}

func (p *PackageManagerInstaller) Verify(name string) error {
	installed, _, _ := p.Detect(name)
	if !installed {
		return fmt.Errorf("%s is not installed", name)
	}
	return nil
}

func (p *PackageManagerInstaller) Repair(ctx context.Context, name string) error {
	return p.Install(ctx, name)
}

func detectPackageManager() string {
	cmds := []string{"apt", "dnf", "yum", "pacman", "apk"}
	for _, cmd := range cmds {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}
	return "unknown"
}

func getPackageName(name string) string {
	switch name {
	case "wireguard-tools":
		return "wireguard-tools"
	case "openvpn":
		return "openvpn"
	default:
		return name
	}
}

func extractVersion(output string) string {
	// Try to extract version from dpkg/rpm output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		}
		if strings.HasPrefix(line, "Name:") {
			continue
		}
	}
	// Fallback: return first line
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return "unknown"
}
