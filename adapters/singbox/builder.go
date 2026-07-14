package singbox

import (
	"encoding/json"
	"fmt"

	netbridge "github.com/netbridge/netbridge"
)

type Builder struct{}

func (b *Builder) BuildConfig(cfg netbridge.BackendConfig) ([]byte, error) {
	singCfg := map[string]any{
		"log": map[string]any{
			"level": "warn",
		},
		"inbounds": []map[string]any{
			{
				"tag":         "socks-in",
				"protocol":    "socks",
				"listen":      "127.0.0.1",
				"listen_port": cfg.LocalPort,
			},
		},
		"outbounds": []map[string]any{
			{
				"tag":         "proxy",
				"protocol":    string(cfg.Profile.Protocol),
				"server":      cfg.Profile.Server,
				"server_port": cfg.Profile.Port,
			},
		},
	}

	return json.MarshalIndent(singCfg, "", "  ")
}

func unused() {
	_ = fmt.Sprintf
}
