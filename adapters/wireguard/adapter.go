package wireguard

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	netbridge "github.com/netbridge/netbridge"
)

type Adapter struct {
	config     netbridge.BackendConfig
	iface      *Interface
	routes     *RouteManager
	running    bool
	ifaceName  string
	allowedIPs []net.IPNet
	startAt    time.Time
	mu         sync.Mutex
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

func getFieldString(outbound map[string]any, key string) (string, error) {
	if outbound == nil {
		return "", fmt.Errorf("missing field %q", key)
	}
	val, ok := outbound[key]
	if !ok || val == nil {
		return "", fmt.Errorf("missing field %q", key)
	}
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("field %q is not a string", key)
	}
	if s == "" {
		return "", fmt.Errorf("field %q is empty", key)
	}
	return s, nil
}

func parseAllowedIPs(raw string) ([]net.IPNet, error) {
	var result []net.IPNet
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		_, cidr, err := net.ParseCIDR(part)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", part, err)
		}
		result = append(result, *cidr)
	}
	return result, nil
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

	privateKeyStr, err := getFieldString(cfg.Profile.Outbound, "private_key")
	if err != nil {
		return fmt.Errorf("private_key: %w", err)
	}
	privateKey, err := wgtypes.ParseKey(privateKeyStr)
	if err != nil {
		return fmt.Errorf("parse private key: %w", err)
	}

	publicKeyStr, err := getFieldString(cfg.Profile.Outbound, "public_key")
	if err != nil {
		return fmt.Errorf("public_key: %w", err)
	}
	peerKey, err := wgtypes.ParseKey(publicKeyStr)
	if err != nil {
		return fmt.Errorf("parse public key: %w", err)
	}

	var endpoint *net.UDPAddr
	if epStr, err := getFieldString(cfg.Profile.Outbound, "endpoint"); err == nil {
		resolver := &net.Resolver{}
		ips, resolveErr := resolver.LookupIPAddr(ctx, epStr)
		if resolveErr != nil {
			return fmt.Errorf("resolve endpoint %q: %w", epStr, resolveErr)
		}
		if len(ips) == 0 {
			return fmt.Errorf("resolve endpoint %q: no addresses found", epStr)
		}
		port := 51820
		if idx := strings.LastIndex(epStr, ":"); idx != -1 {
			fmt.Sscanf(epStr[idx:], ":%d", &port)
		}
		endpoint = &net.UDPAddr{IP: ips[0].IP, Port: port}
	}

	var allowedIPs []net.IPNet
	if aipsStr, err := getFieldString(cfg.Profile.Outbound, "allowed_ips"); err == nil {
		aips, parseErr := parseAllowedIPs(aipsStr)
		if parseErr != nil {
			return fmt.Errorf("parse allowed_ips: %w", parseErr)
		}
		allowedIPs = aips
	}

	a.iface = NewInterface(a.ifaceName)
	if err := a.iface.Create(nil); err != nil {
		return fmt.Errorf("create interface: %w", err)
	}

	rollback := func() {
		if a.iface != nil {
			_ = a.iface.Delete()
			a.iface = nil
		}
	}

	if err := a.iface.ConfigurePrivateKey(privateKey); err != nil {
		rollback()
		return fmt.Errorf("set private key: %w", err)
	}

	keepalive := 25 * time.Second
	if err := a.iface.AddPeerWithKeepAlive(peerKey, endpoint, allowedIPs, keepalive); err != nil {
		rollback()
		return fmt.Errorf("add peer: %w", err)
	}

	addrStr, err := getFieldString(cfg.Profile.Outbound, "address")
	if err != nil {
		rollback()
		return fmt.Errorf("address (required): %w", err)
	}
	if err := a.iface.SetAddr(addrStr); err != nil {
		rollback()
		return fmt.Errorf("set address: %w", err)
	}

	mtu := 1420
	if mtuStr, err := getFieldString(cfg.Profile.Outbound, "mtu"); err == nil {
		fmt.Sscanf(mtuStr, "%d", &mtu)
	}
	if err := a.iface.SetMTU(mtu); err != nil {
		log.Printf("wireguard: set MTU %d failed (non-fatal): %v", mtu, err)
	}

	if err := a.iface.SetUp(); err != nil {
		rollback()
		return fmt.Errorf("interface up: %w", err)
	}

	a.allowedIPs = allowedIPs
	for _, aip := range allowedIPs {
		if err := a.routes.AddRoute(aip.String(), a.ifaceName); err != nil {
			log.Printf("wireguard: add route %s: %v", aip.String(), err)
		}
	}

	a.running = true
	a.startAt = time.Now()
	log.Printf("wireguard started on %s (mtu %d, keepalive %v)", a.ifaceName, mtu, keepalive)
	return nil
}

func (a *Adapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	for _, aip := range a.allowedIPs {
		if err := a.routes.RemoveRoute(aip.String()); err != nil {
			log.Printf("wireguard: remove route %s: %v", aip.String(), err)
		}
	}

	if a.iface != nil {
		if err := a.iface.Down(); err != nil {
			log.Printf("wireguard: down: %v", err)
		}
		if err := a.iface.Delete(); err != nil {
			log.Printf("wireguard: delete: %v", err)
		}
	}

	a.running = false
	a.allowedIPs = nil
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
