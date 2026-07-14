package security

import (
	"crypto/sha256"
	"fmt"
	"os"
	"runtime"
)

type Keyring struct {
	machineID string
}

func NewKeyring() *Keyring {
	return &Keyring{
		machineID: getMachineID(),
	}
}

func (k *Keyring) DeriveKey(passphrase string) []byte {
	input := k.machineID + passphrase
	hash := sha256.Sum256([]byte(input))
	return hash[:]
}

func (k *Keyring) MachineID() string {
	return k.machineID
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
	case "darwin":
		return "darwin-fallback"
	case "windows":
		return "windows-fallback"
	}
	return fmt.Sprintf("fallback-%d", os.Getpid())
}
