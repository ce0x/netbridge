package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/adapters/openvpn"
	"github.com/netbridge/netbridge/adapters/singbox"
	"github.com/netbridge/netbridge/adapters/wireguard"
	"github.com/netbridge/netbridge/adapters/xray"
	"github.com/netbridge/netbridge/internal/benchmark"
	"github.com/netbridge/netbridge/internal/config"
	"github.com/netbridge/netbridge/internal/dns"
	"github.com/netbridge/netbridge/internal/health"
	"github.com/netbridge/netbridge/internal/profile"
	"github.com/netbridge/netbridge/internal/routing"
	"github.com/netbridge/netbridge/internal/session"
	"github.com/netbridge/netbridge/internal/stats"
	"github.com/netbridge/netbridge/plugins"
)

type Engine struct {
	cfg            *config.Config
	profiles       netbridge.ProfileManager
	sessions       netbridge.SessionManager
	routing        netbridge.RoutingEngine
	health         netbridge.HealthEngine
	benchmark      netbridge.BenchmarkEngine
	dnsEngine      netbridge.DNSEngine
	plugins        netbridge.PluginManager
	statsCollector *stats.Collector
	mu             sync.RWMutex
}

type builtinPlugin struct {
	name      string
	protocols []netbridge.Protocol
	factory   func() netbridge.Backend
}

func (p *builtinPlugin) Name() string                        { return p.name }
func (p *builtinPlugin) Version() string                     { return "builtin" }
func (p *builtinPlugin) Protocols() []netbridge.Protocol      { return p.protocols }
func (p *builtinPlugin) NewBackend(_ netbridge.Profile) (netbridge.Backend, error) {
	return p.factory(), nil
}

func registerBuiltinPlugins(reg *plugins.Registry) {
	reg.Register(&builtinPlugin{
		name:      "xray",
		protocols: []netbridge.Protocol{
			netbridge.ProtocolVLESS, netbridge.ProtocolVMess,
			netbridge.ProtocolTrojan, netbridge.ProtocolShadowsocks,
			netbridge.ProtocolSOCKS, netbridge.ProtocolHTTP,
		},
		factory: func() netbridge.Backend { return xray.New() },
	})
	reg.Register(&builtinPlugin{
		name:      "singbox",
		protocols: []netbridge.Protocol{
			netbridge.ProtocolVLESS, netbridge.ProtocolVMess,
			netbridge.ProtocolTrojan, netbridge.ProtocolShadowsocks,
		},
		factory: func() netbridge.Backend { return singbox.New() },
	})
	reg.Register(&builtinPlugin{
		name:      "wireguard",
		protocols: []netbridge.Protocol{netbridge.ProtocolWireGuard},
		factory:   func() netbridge.Backend { return wireguard.New() },
	})
	reg.Register(&builtinPlugin{
		name:      "openvpn",
		protocols: []netbridge.Protocol{netbridge.ProtocolOpenVPN},
		factory:   func() netbridge.Backend { return openvpn.New() },
	})
}

func New(cfg *config.Config) (*Engine, error) {
	pm := profile.NewManager(cfg)
	plm := plugins.NewRegistry()
	registerBuiltinPlugins(plm)
	sm := session.NewManager(pm, plm, cfg)
	re := routing.NewEngine()
	he := health.NewEngine(pm)
	be := benchmark.NewEngine(pm, he)
	de := dns.NewEngine()
	sc := stats.NewCollector()

	e := &Engine{
		cfg:            cfg,
		profiles:       pm,
		sessions:       sm,
		routing:        re,
		health:         he,
		benchmark:      be,
		dnsEngine:      de,
		plugins:        plm,
		statsCollector: sc,
	}

	// Recover session state from previous CLI run
	_ = sm.Recover(context.Background())

	return e, nil
}

func (e *Engine) ProfileManager() netbridge.ProfileManager {
	return e.profiles
}

func (e *Engine) SessionManager() netbridge.SessionManager {
	return e.sessions
}

func (e *Engine) RoutingEngine() netbridge.RoutingEngine {
	return e.routing
}

func (e *Engine) HealthEngine() netbridge.HealthEngine {
	return e.health
}

func (e *Engine) BenchmarkEngine() netbridge.BenchmarkEngine {
	return e.benchmark
}

func (e *Engine) DNSEngine() netbridge.DNSEngine {
	return e.dnsEngine
}

func (e *Engine) PluginManager() netbridge.PluginManager {
	return e.plugins
}

func (e *Engine) RunCommand(ctx context.Context, profileID string, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("no command specified")
	}

	envVars := e.EnvVars()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = os.Environ()
	for k, v := range envVars {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *Engine) EnvVars() map[string]string {
	return map[string]string{
		"http_proxy":  "http://127.0.0.1:8080",
		"https_proxy": "http://127.0.0.1:8080",
		"all_proxy":   "socks5://127.0.0.1:10808",
		"no_proxy":    "localhost,127.0.0.1,::1",
	}
}

func (e *Engine) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.sessions.Status() == netbridge.StatusConnected {
		_ = e.sessions.Disconnect(context.Background())
	}

	return nil
}
