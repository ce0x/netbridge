package profile

import (
	"context"
	"fmt"

	netbridge "github.com/netbridge/netbridge"
)

type Validator struct{}

func (v *Validator) Validate(ctx context.Context, p *netbridge.Profile) error {
	if p == nil {
		return fmt.Errorf("profile is nil")
	}
	if p.Name == "" {
		return fmt.Errorf("profile name is required")
	}
	if p.Server == "" {
		return fmt.Errorf("server address is required")
	}
	if p.Port <= 0 || p.Port > 65535 {
		return fmt.Errorf("invalid port: %d", p.Port)
	}
	if p.Protocol == "" {
		return fmt.Errorf("protocol is required")
	}
	return nil
}
