package core

import (
	"context"
	"sync"

	netbridge "github.com/netbridge/netbridge"
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

func New(cfg *config.Config) (*Engine, error) {
	pm := profile.NewManager(cfg)
	sm := session.NewManager(pm)
	re := routing.NewEngine()
	he := health.NewEngine(pm)
	be := benchmark.NewEngine(pm, he)
	de := dns.NewEngine()
	plm := plugins.NewRegistry()
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
	return nil
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
