// Package netbridge defines the core contracts between all NetBridge layers.
// No layer may import a layer above it. Business logic never imports adapters.
package netbridge

import (
	"context"
	"time"
)

type Protocol string

const (
	ProtocolVLESS        Protocol = "vless"
	ProtocolVMess        Protocol = "vmess"
	ProtocolTrojan       Protocol = "trojan"
	ProtocolShadowsocks  Protocol = "shadowsocks"
	ProtocolWireGuard    Protocol = "wireguard"
	ProtocolOpenVPN      Protocol = "openvpn"
	ProtocolSOCKS        Protocol = "socks"
	ProtocolHTTP         Protocol = "http"
)

type SessionMode string

const (
	ModeSOCKS   SessionMode = "socks"
	ModeHTTP    SessionMode = "http"
	ModeTUN     SessionMode = "tun"
	ModeProcess SessionMode = "process"
)

type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusConnected    ConnectionStatus = "connected"
	StatusReconnecting ConnectionStatus = "reconnecting"
	StatusError        ConnectionStatus = "error"
)

type Profile struct {
	ID        string            `json:"id" yaml:"id"`
	Name      string            `json:"name" yaml:"name"`
	Protocol  Protocol          `json:"protocol" yaml:"protocol"`
	Backend   string            `json:"backend" yaml:"backend"`
	RawURI    string            `json:"raw_uri,omitempty" yaml:"raw_uri,omitempty"`
	Server    string            `json:"server" yaml:"server"`
	Port      int               `json:"port" yaml:"port"`
	Transport TransportConfig   `json:"transport" yaml:"transport"`
	TLS       TLSConfig         `json:"tls" yaml:"tls"`
	Outbound  map[string]any    `json:"outbound" yaml:"outbound"`
	Score     int               `json:"score" yaml:"score"`
	Tags      []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	CreatedAt time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" yaml:"updated_at"`
}

type TransportConfig struct {
	Type    string            `json:"type" yaml:"type"`
	Path    string            `json:"path,omitempty" yaml:"path,omitempty"`
	Host    string            `json:"host,omitempty" yaml:"host,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

type TLSConfig struct {
	Enabled         bool     `json:"enabled" yaml:"enabled"`
	ServerName      string   `json:"server_name,omitempty" yaml:"server_name,omitempty"`
	Fingerprint     string   `json:"fingerprint,omitempty" yaml:"fingerprint,omitempty"`
	ALPN            []string `json:"alpn,omitempty" yaml:"alpn,omitempty"`
	AllowInsecure   bool     `json:"allow_insecure,omitempty" yaml:"allow_insecure,omitempty"`
	RealityPublicKey string  `json:"reality_public_key,omitempty" yaml:"reality_public_key,omitempty"`
	RealityShortID   string  `json:"reality_short_id,omitempty" yaml:"reality_short_id,omitempty"`
}

type Session struct {
	ID        string           `json:"id"`
	ProfileID string           `json:"profile_id"`
	Mode      SessionMode      `json:"mode"`
	LocalAddr string           `json:"local_addr"`
	TUNIface  string           `json:"tun_iface,omitempty"`
	PID       int              `json:"pid"`
	Status    ConnectionStatus `json:"status"`
	BytesUp   int64            `json:"bytes_up"`
	BytesDown int64            `json:"bytes_down"`
	StartedAt time.Time        `json:"started_at"`
	EndedAt   *time.Time       `json:"ended_at,omitempty"`
}

type RouteRule struct {
	ID        string `json:"id"`
	Pattern   string `json:"pattern"`
	RuleType  string `json:"rule_type"`
	ProfileID string `json:"profile_id"`
	Priority  int    `json:"priority"`
	Enabled   bool   `json:"enabled"`
}

type FailoverChain struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	ProfileIDs         []string      `json:"profile_ids"`
	CurrentIndex       int           `json:"current_index"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	FailThreshold      int           `json:"fail_threshold"`
}

type TrafficStats struct {
	BytesUp   int64         `json:"bytes_up"`
	BytesDown int64         `json:"bytes_down"`
	RateUp    float64       `json:"rate_up_bps"`
	RateDown  float64       `json:"rate_down_bps"`
	Latency   time.Duration `json:"latency"`
	Uptime    time.Duration `json:"uptime"`
}

type HealthResult struct {
	ProfileID  string        `json:"profile_id"`
	Reachable  bool          `json:"reachable"`
	Latency    time.Duration `json:"latency"`
	PacketLoss float64       `json:"packet_loss"`
	Error      string        `json:"error,omitempty"`
	CheckedAt  time.Time     `json:"checked_at"`
}

type BenchmarkResult struct {
	ProfileID  string        `json:"profile_id"`
	Latency    time.Duration `json:"latency"`
	Jitter     time.Duration `json:"jitter"`
	Throughput float64       `json:"throughput_bps"`
	PacketLoss float64       `json:"packet_loss"`
	Score      int           `json:"score"`
}

type Endpoint struct {
	Type    SessionMode `json:"type"`
	Address string      `json:"address"`
	Iface   string      `json:"iface,omitempty"`
}

type BackendConfig struct {
	Profile   Profile
	Mode      SessionMode
	LocalPort int
	TUNName   string
	ExtraArgs map[string]any
}

type BackendStatus struct {
	Running bool
	PID     int
	Uptime  time.Duration
	Error   error
}

type Backend interface {
	Name() string
	SupportedProtocols() []Protocol
	Start(ctx context.Context, cfg BackendConfig) error
	Stop() error
	Status() BackendStatus
	Stats() TrafficStats
	Configure(cfg BackendConfig) error
	HealthCheck(ctx context.Context) error
	LocalEndpoints() []Endpoint
}

type ProfileManager interface {
	Import(ctx context.Context, raw string) (*Profile, error)
	ImportFile(ctx context.Context, path string) ([]*Profile, error)
	ImportSubscription(ctx context.Context, url string) ([]*Profile, error)
	Save(ctx context.Context, p *Profile) error
	Get(ctx context.Context, id string) (*Profile, error)
	GetByName(ctx context.Context, name string) (*Profile, error)
	List(ctx context.Context) ([]*Profile, error)
	Delete(ctx context.Context, id string) error
	Rename(ctx context.Context, id, newName string) error
	Clone(ctx context.Context, id, newName string) (*Profile, error)
	Export(ctx context.Context, id string) (string, error)
	Validate(ctx context.Context, p *Profile) error
	SetActive(ctx context.Context, id string) error
	GetActive(ctx context.Context) (*Profile, error)
}

type SessionManager interface {
	Connect(ctx context.Context, profileID string, mode SessionMode) (*Session, error)
	Disconnect(ctx context.Context) error
	Restart(ctx context.Context) error
	Reload(ctx context.Context) error
	Current() (*Session, error)
	Status() ConnectionStatus
	Stats() TrafficStats
	Persist(ctx context.Context) error
	Recover(ctx context.Context) error
}

type RoutingEngine interface {
	AddRule(ctx context.Context, rule RouteRule) error
	RemoveRule(ctx context.Context, id string) error
	ListRules(ctx context.Context) ([]*RouteRule, error)
	ClearRules(ctx context.Context) error
	Resolve(destination string) (profileID string, err error)
	Apply(ctx context.Context) error
}

type HealthEngine interface {
	Check(ctx context.Context, profileID string) (*HealthResult, error)
	CheckAll(ctx context.Context) ([]*HealthResult, error)
	StartWatchdog(ctx context.Context, interval time.Duration) error
	StopWatchdog() error
	OnFailure(fn func(profileID string, result *HealthResult))
}

type BenchmarkEngine interface {
	Run(ctx context.Context, profileID string) (*BenchmarkResult, error)
	RunAll(ctx context.Context) ([]*BenchmarkResult, error)
	Best(ctx context.Context) (string, error)
}

type DNSPreset struct {
	Name    string
	Servers []string
}

type DNSBenchResult struct {
	Name    string
	Server  string
	Latency time.Duration
	Error   error
}

type DNSEngine interface {
	ListPresets() []DNSPreset
	SetResolver(ctx context.Context, nameOrAddr string) error
	CurrentResolver() string
	Benchmark(ctx context.Context) ([]DNSBenchResult, error)
	Reset(ctx context.Context) error
}

type Plugin interface {
	Name() string
	Version() string
	Protocols() []Protocol
	NewBackend(profile Profile) (Backend, error)
}

type PluginManager interface {
	Load(path string) (Plugin, error)
	Unload(name string) error
	List() []Plugin
	Get(name string) (Plugin, error)
	BackendFor(protocol Protocol) (Backend, error)
}

type CoreEngine interface {
	ProfileManager() ProfileManager
	SessionManager() SessionManager
	RoutingEngine() RoutingEngine
	HealthEngine() HealthEngine
	BenchmarkEngine() BenchmarkEngine
	DNSEngine() DNSEngine
	PluginManager() PluginManager
	RunCommand(ctx context.Context, profileID string, argv []string) error
	EnvVars() map[string]string
	Shutdown(ctx context.Context) error
}

var (
	ErrProfileNotFound  = errorf("profile not found")
	ErrNoActiveSession  = errorf("no active session")
	ErrAlreadyConnected = errorf("already connected")
	ErrNoHotReload      = errorf("backend does not support hot reload")
	ErrBackendNotFound  = errorf("no backend found for protocol")
	ErrInvalidURI       = errorf("invalid profile URI")
	ErrPermissionDenied = errorf("permission denied — run as root or with CAP_NET_ADMIN")
)

type sentinelError string

func (e sentinelError) Error() string { return string(e) }
func errorf(s string) sentinelError   { return sentinelError(s) }
