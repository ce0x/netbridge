package wireguard

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	netbridge "github.com/netbridge/netbridge"
)

type Adapter struct {
	config    netbridge.BackendConfig
	iface     *Interface
	routes    *RouteManager
	running   bool
	ifaceName string
	startAt   time.Time
	mu        sync.Mutex
}

func New() *Adapter {
	return &Adapter{
		routes: NewRouteManager(),
	}
}

func (a *Adapter) Name() string {
	return "wireguard"
}

func (a *Adapter) SupportedProtocols() []netbridge.Protocol {
	return []netbridge.Protocol{netbridge.ProtocolWireGuard}
}

func (a *Adapter) Start(ctx context.Context, cfg netbridge.BackendConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return fmt.Errorf("wireguard already running")
	}

	a.config = cfg
	a.ifaceName = "wg0"
	if cfg.TUNName != "" {
		a.ifaceName = cfg.TUNName
	}

	a.iface = NewInterface(a.ifaceName)

	if err := a.iface.Create(nil); err != nil {
		return fmt.Errorf("create interface: %w", err)
	}

	peerKey, err := wgtypes.ParseKey(fmt.Sprintf("%v", cfg.Profile.Outbound["public_key"]))
	if err != nil {
		return fmt.Errorf("parse public key: %w", err)
	}

	var endpoint *net.UDPAddr
	if epStr, ok := cfg.Profile.Outbound["endpoint"].(string); ok && epStr != "" {
		endpoint, err = net.ResolveUDPAddr("udp", epStr)
		if err != nil {
			return fmt.Errorf("resolve endpoint: %w", err)
		}
	}

	var allowedIPs []net.IPNet
	if aips, ok := cfg.Profile.Outbound["allowed_ips"].(string); ok && aips != "" {
		_, cidr, err := net.ParseCIDR(aips)
		if err == nil {
			allowedIPs = append(allowedIPs, *cidr)
		}
	}

	if err := a.iface.AddPeer(peerKey, endpoint, allowedIPs); err != nil {
		return fmt.Errorf("add peer: %w", err)
	}

	for _, aip := range allowedIPs {
		if err := a.routes.AddRoute(aip.String(), a.ifaceName); err != nil {
			log.Printf("add route %s: %v", aip.String(), err)
		}
	}

	a.running = true
	a.startAt = time.Now()
	log.Printf("wireguard started on %s", a.ifaceName)
	return nil
}

func (a *Adapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	if a.iface != nil {
		_ = a.iface.Down()
		_ = a.iface.Delete()
	}

	a.running = false
	log.Printf("wireguard stopped")
	return nil
}

func (a *Adapter) Status() netbridge.BackendStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	return netbridge.BackendStatus{
		Running: a.running,
		Uptime:  time.Since(a.startAt),
	}
}

func (a *Adapter) Stats() netbridge.TrafficStats {
	return netbridge.TrafficStats{}
}

func (a *Adapter) Configure(cfg netbridge.BackendConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config = cfg
	return nil
}

func (a *Adapter) HealthCheck(ctx context.Context) error {
	a.mu.Lock()
	running := a.running
	a.mu.Unlock()

	if !running {
		return fmt.Errorf("wireguard not running")
	}

	if a.iface != nil && a.iface.IsRunning() {
		return nil
	}

	return fmt.Errorf("wireguard interface not healthy")
}

func (a *Adapter) LocalEndpoints() []netbridge.Endpoint {
	return []netbridge.Endpoint{
		{Type: netbridge.ModeTUN, Address: a.ifaceName, Iface: a.ifaceName},
	}
}