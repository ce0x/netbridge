package sysinfo

import (
	"fmt"
	"os"
	"runtime"
)

type MachineInfo struct {
	Hostname  string
	OS        string
	Arch      string
	GoVersion string
	MachineID string
}

func GetMachineInfo() (*MachineInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	machineID := getMachineID()

	return &MachineInfo{
		Hostname:  hostname,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		MachineID: machineID,
	}, nil
}

func getMachineID() string {
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/etc/machine-id")
		if err == nil {
			return string(data)
		}
		data, err = os.ReadFile("/var/lib/dbus/machine-id")
		if err == nil {
			return string(data)
		}
	}
	return fmt.Sprintf("fallback-%d", os.Getpid())
}
