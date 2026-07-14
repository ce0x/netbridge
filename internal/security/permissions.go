package security

import (
	"fmt"
	"os"
	"runtime"
)

type PermissionChecker struct{}

func (p *PermissionChecker) CheckConfigDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		mode := info.Mode().Perm()
		if mode&0077 != 0 {
			return fmt.Errorf("insecure permissions on %s: %o (should be 700 or stricter)", dir, mode)
		}
	}
	return nil
}

func (p *PermissionChecker) FixPermissions(dir string) error {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return os.Chmod(dir, 0700)
	}
	return nil
}
