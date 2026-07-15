package wireguard

import (
	"fmt"
	"net"
	"os/exec"
	"sync"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Interface struct {
	name    string
	config  []byte
	client  *wgctrl.Client
	running bool
	mu      sync.Mutex
}

func NewInterface(name string) *Interface {
	return &Interface{name: name}
}

func (i *Interface) Create(config []byte) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("wgctrl new: %w", err)
	}
	i.client = client
	i.config = config
	return nil
}

func (i *Interface) Delete() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.client != nil {
		i.client.Close()
		i.client = nil
	}
	i.running = false
	return nil
}

func (i *Interface) ConfigurePrivateKey(key wgtypes.Key) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.client == nil {
		return fmt.Errorf("interface not created")
	}

	cfg := wgtypes.Config{
		PrivateKey: &key,
	}
	return i.client.ConfigureDevice(i.name, cfg)
}

func (i *Interface) ConfigureDevice(cfg wgtypes.Config) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.client == nil {
		return fmt.Errorf("interface not created")
	}

	return i.client.ConfigureDevice(i.name, cfg)
}

func (i *Interface) Up(cfg *wgtypes.Config) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.client == nil {
		return fmt.Errorf("interface not created")
	}

	if err := i.client.ConfigureDevice(i.name, *cfg); err != nil {
		return fmt.Errorf("configure device: %w", err)
	}

	i.running = true
	return nil
}

func (i *Interface) SetUp() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.running = true
	return nil
}

func (i *Interface) Down() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.client == nil {
		return nil
	}

	emptyCfg := wgtypes.Config{ReplacePeers: true}
	if err := i.client.ConfigureDevice(i.name, emptyCfg); err != nil {
		return fmt.Errorf("down device: %w", err)
	}

	i.running = false
	return nil
}

func (i *Interface) Address() (net.IP, error) {
	iface, err := net.InterfaceByName(i.name)
	if err != nil {
		return nil, fmt.Errorf("interface not found: %w", err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("get addresses: %w", err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && ipNet.IP.To4() != nil {
			return ipNet.IP, nil
		}
	}

	return nil, fmt.Errorf("no IPv4 address found")
}

func (i *Interface) IsRunning() bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.running
}

func (i *Interface) DeviceConfig() (*wgtypes.Device, error) {
	if i.client == nil {
		return nil, fmt.Errorf("interface not created")
	}
	return i.client.Device(i.name)
}

func (i *Interface) SetKeepAlive(interval time.Duration) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.client == nil {
		return fmt.Errorf("interface not created")
	}

	device, err := i.client.Device(i.name)
	if err != nil {
		return fmt.Errorf("get device: %w", err)
	}

	peers := make([]wgtypes.PeerConfig, len(device.Peers))
	for idx, peer := range device.Peers {
		peers[idx] = wgtypes.PeerConfig{
			PublicKey:                   peer.PublicKey,
			PersistentKeepaliveInterval: &interval,
			ReplaceAllowedIPs:           false,
			AllowedIPs:                  peer.AllowedIPs,
		}
	}

	cfg := wgtypes.Config{
		Peers: peers,
	}
	return i.client.ConfigureDevice(i.name, cfg)
}

func (i *Interface) AddPeer(pubKey wgtypes.Key, endpoint *net.UDPAddr, allowedIPs []net.IPNet) error {
	return i.AddPeerWithKeepAlive(pubKey, endpoint, allowedIPs, 25*time.Second)
}

func (i *Interface) AddPeerWithKeepAlive(pubKey wgtypes.Key, endpoint *net.UDPAddr, allowedIPs []net.IPNet, keepalive time.Duration) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.client == nil {
		return fmt.Errorf("interface not created")
	}

	peerCfg := wgtypes.PeerConfig{
		PublicKey:                   pubKey,
		Endpoint:                   endpoint,
		AllowedIPs:                 allowedIPs,
		ReplaceAllowedIPs:          true,
		PersistentKeepaliveInterval: &keepalive,
	}

	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerCfg},
	}
	return i.client.ConfigureDevice(i.name, cfg)
}

func (i *Interface) RemovePeer(pubKey wgtypes.Key) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.client == nil {
		return fmt.Errorf("interface not created")
	}

	peerCfg := wgtypes.PeerConfig{
		PublicKey: pubKey,
		Remove:   true,
	}

	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerCfg},
	}
	return i.client.ConfigureDevice(i.name, cfg)
}

func (i *Interface) SetAddr(addr string) error {
	cmd := exec.Command("ip", "addr", "add", addr, "dev", i.name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("set address: %w", err)
	}
	return nil
}

func (i *Interface) SetMTU(mtu int) error {
	cmd := exec.Command("ip", "link", "set", "mtu", fmt.Sprintf("%d", mtu), "dev", i.name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("set mtu: %w", err)
	}
	return nil
}