# Project: NetBridge CLI — Complete Specification

## Vision

Design and implement a **production-grade Linux CLI application** called **NetBridge**.

NetBridge is a universal network access and connectivity toolkit for Linux servers. Its purpose is to allow administrators and DevOps engineers to easily create, manage, test, activate, deactivate, and utilize outbound connectivity profiles using modern tunneling, proxy, VPN, and transport technologies.

NetBridge acts as a **unified connectivity layer** between Linux applications and external networks.

### Core Design Principles

- Simplicity and predictability
- Reliability on production servers
- Minimal system impact
- Fast activation/deactivation
- Multi-profile management
- Safe operation — zero interference with existing services unless explicitly requested
- Full scriptability and automation support

---

## Supported Protocols

The architecture must be **plugin-based**.

### Xray Family (initial support)

- VLESS
- VMess
- Trojan
- Shadowsocks
- SOCKS
- HTTP Proxy
- Freedom
- WireGuard (through Xray)
- Hysteria2 (if supported by current ecosystem)

### Native VPN

- WireGuard (native kernel module)
- OpenVPN

### Future Extensions (must be addable without modifying core)

- TUIC
- AnyTLS
- SSH Tunnel
- Cloudflare Tunnel
- Hysteria
- Custom transports

---

## Architecture Requirements

The project must follow a **clean architecture** with strict separation of concerns.

### Layers

```
┌─────────────────────────────────────────────┐
│           Interface Layer                   │
│  CLI Mode │ Interactive TUI │ JSON API Mode │
├─────────────────────────────────────────────┤
│               Core Engine                   │
│  Orchestration · State · Event Bus          │
├─────────────────────────────────────────────┤
│              Service Layer                  │
│  Profile Manager   │  Session Manager       │
│  Routing Engine    │  Health Engine         │
│  DNS Engine        │  Plugin Manager        │
│  Benchmark Engine  │  Stats Engine          │
├─────────────────────────────────────────────┤
│       Backend Abstraction Layer             │
│  Common Interface: Start/Stop/Status/Stats  │
├─────────────────────────────────────────────┤
│           Backend Adapters                  │
│  Xray Core │ Sing-box │ WireGuard │ OpenVPN │
├─────────────────────────────────────────────┤
│              Plugin SDK                     │
│  TUIC │ Hysteria │ SSH │ CF Tunnel │ Custom │
├─────────────────────────────────────────────┤
│           Storage Layer                     │
│  /etc/netbridge — profiles/sessions/logs    │
└─────────────────────────────────────────────┘
```

### Key Architecture Rules

- Business logic must remain **completely independent** from transport implementations
- Every Backend must implement a **common interface**: `Start()`, `Stop()`, `Status()`, `Stats()`, `Configure()`
- The application must expose **three interface modes** simultaneously:
  1. **CLI Mode** — for direct terminal use
  2. **Interactive TUI Mode** — for visual menu-driven operation
  3. **JSON API Mode** (`--json` flag on all commands) — for scripting and future Web UI integration
- No layer may import from a layer above it
- The Core Engine must not know which backend is running

---

## Three Interface Modes

### 1. CLI Mode

Standard command-line interface, inspired by `git`, `docker`, `kubectl`.

All commands:
- Short and predictable
- Support `--json` flag for machine-readable output
- Support `--quiet` and `--verbose` flags
- Return proper exit codes for scripting

### 2. Interactive TUI Mode

Launched by running `netbridge` or `netbridge tui` with no arguments.

Must display a full-screen terminal UI:

```
╔══════════════════════════════════════════════════════╗
║                   NetBridge CLI                     ║
╠══════════════════════════════════════════════════════╣
║  1. Profile Management                              ║
║  2. Import Profile                                  ║
║  3. List Profiles                                   ║
║  4. Test Profile                                    ║
║  5. Connect                                          ║
║  6. Disconnect                                       ║
║  7. Status                                           ║
║------------------------------------------------------║
║  8. Health Check                                     ║
║  9. Benchmark                                        ║
║ 10. Smart Routing                                    ║
║ 11. DNS Management                                   ║
║------------------------------------------------------║
║ 12. Service Management                               ║
║ 13. Logs                                             ║
║ 14. Statistics                                       ║
║------------------------------------------------------║
║ 15. Settings                                         ║
║ 16. Backup                                           ║
║ 17. Restore                                          ║
║------------------------------------------------------║
║  0. Exit                                             ║
╚══════════════════════════════════════════════════════╝

Active Profile : main
Status         : Connected
Backend        : Xray
Uptime         : 2h 14m
```

TUI requirements:
- Keyboard navigation (arrow keys + number shortcuts)
- Real-time status bar showing active profile, connection state, uptime, traffic
- Sub-menus for each section
- Color coding: green = connected, yellow = connecting, red = error
- Refresh interval for live stats (configurable, default 2s)

### 3. JSON API Mode

Every command must support `--json` flag:

```
netbridge status --json
netbridge list --json
netbridge test profile-a --json
netbridge benchmark --json
```

JSON output must follow a consistent envelope:

```json
{
  "success": true,
  "command": "status",
  "timestamp": "2024-01-01T00:00:00Z",
  "data": { ... },
  "error": null
}
```

This mode enables future Web UI or management panel integration without any core changes.

---

## First Run Wizard

On first launch with no profiles configured:

```
Welcome to NetBridge!

No profiles found. Let's get started.

Import a profile:
  1) VLESS URL
  2) VMess URL
  3) Trojan URL
  4) WireGuard config
  5) OpenVPN config
  6) Import from file
  7) Skip for now

Select:
```

---

## Built-in Help System

```
netbridge help
netbridge help <command>
netbridge help <topic>
```

Example output for `netbridge help connect`:

```
Command:
    netbridge connect [profile]

Description:
    Connects the selected profile. If no profile is specified,
    uses the currently active profile.

Options:
    --mode socks|http|tun     Connection mode (default: socks)
    --port <number>           Local port override
    --no-watchdog             Disable auto-reconnect for this session

Examples:
    netbridge connect
    netbridge connect profile-a
    netbridge connect profile-a --mode tun

Related:
    netbridge disconnect
    netbridge status
    netbridge restart
```

---

## CLI Commands

### Profile Management

```
netbridge import <url|file>     Import profile from URL or file
netbridge export <name>         Export profile
netbridge delete <name>         Delete profile
netbridge rename <old> <new>    Rename profile
netbridge clone <name> <new>    Clone profile
netbridge list                  List all profiles
netbridge use <name>            Set active profile
netbridge show <name>           Show profile details (sensitive values masked)
```

### Connection Management

```
netbridge connect [profile]     Connect active or named profile
netbridge disconnect            Disconnect current session
netbridge restart               Restart current connection
netbridge reload                Reload config without disconnecting
netbridge status                Show current connection status
```

### Session Modes

```
netbridge connect --mode socks     Local SOCKS5 at 127.0.0.1:10808
netbridge connect --mode http      Local HTTP proxy at 127.0.0.1:8080
netbridge connect --mode tun       Virtual TUN interface (full tunnel)
netbridge run <command>            Run command through active profile
```

Examples of process mode:

```
netbridge run curl https://example.com
netbridge run apt update
netbridge run docker pull nginx
```

### Testing

```
netbridge test [profile]
```

Verifies:
- DNS resolution
- TCP reachability
- TLS handshake
- Latency measurement
- Download throughput
- Upload throughput

Output formats: human-readable table and `--json`.

### Health Monitoring

```
netbridge health [profile]
```

Reports:
- Reachability to configurable targets
- Latency (avg, min, max)
- Packet loss percentage
- Protocol-level verification

### Benchmark & Scoring

```
netbridge benchmark [--all] [profile]
```

Measures for each profile:
- Latency
- Throughput
- Jitter
- Stability over time

Output with scores:

```
Profile         Latency    Throughput    Score
─────────────────────────────────────────────
profile-a       42ms       18 MB/s       94
profile-b       88ms       12 MB/s       82
profile-c       210ms      6 MB/s        67
```

### Smart Routing Engine

```
netbridge route add <domain|pattern> <profile>
netbridge route remove <domain|pattern>
netbridge route list
netbridge route clear
```

Examples:

```
netbridge route add github.com profile-a
netbridge route add docker.io profile-b
netbridge route add *.example.com profile-c
netbridge route add 10.0.0.0/8 direct
```

Supports:
- Exact domain match
- Wildcard patterns (`*.example.com`)
- CIDR ranges
- Keyword matching
- GeoIP-based rules (future)

### Failover Chain

```
netbridge failover create <chain-name> <profile-a> <profile-b> <profile-c>
netbridge failover list
netbridge failover delete <chain-name>
netbridge failover status
```

Behavior:
- Automatic switching on failure: A → B → C
- Configurable health-check interval
- Configurable failure threshold before switching
- Notification on failover event (log + optional webhook)

### Auto Best Profile

```
netbridge auto
```

Tests all available profiles, scores them, and automatically activates the best one. Respects failover chains if configured.

### DNS Manager

```
netbridge dns list                  List available DNS presets
netbridge dns use <preset|ip>       Set active DNS resolver
netbridge dns benchmark             Benchmark all presets and show latency
netbridge dns show                  Show current DNS configuration
netbridge dns reset                 Restore system default DNS
```

Built-in presets: `cloudflare`, `google`, `quad9`, `adguard`, `system`

### Shell Integration

```
netbridge env        Print export commands for proxy environment variables
netbridge unset      Print unset commands to remove proxy variables
```

Usage:

```bash
eval $(netbridge env)     # Activate proxy env vars in current shell
eval $(netbridge unset)   # Remove proxy env vars from current shell
```

### Monitoring & Statistics

```
netbridge logs [--follow] [--level debug|info|warn|error]
netbridge stats [profile]
netbridge top
```

`netbridge top` displays:
- Active profile and backend
- Uptime
- Current upload/download rate
- Total traffic this session
- Total traffic all-time
- Active connections count

### Service Management

```
netbridge service install     Install and enable systemd unit
netbridge service start       Start the service
netbridge service stop        Stop the service
netbridge service restart     Restart the service
netbridge service status      Show systemd service status
netbridge service uninstall   Remove systemd unit
```

### Backup & Restore

```
netbridge backup [--output <file>]     Create encrypted backup archive
netbridge restore <file>               Restore from backup archive
```

Backup includes:
- All profiles
- Route rules
- DNS configuration
- Application settings
- Session history (optional, --include-history)

---

## Profile Format Support

Import from:
- VLESS links (`vless://...`)
- VMess links (`vmess://...`)
- Trojan links (`trojan://...`)
- Shadowsocks links (`ss://...`)
- WireGuard config files (`.conf`)
- OpenVPN config files (`.ovpn`)
- JSON format (native NetBridge format)
- YAML format (alternative native format)
- QR code data (decoded string input)
- Subscription URLs (auto-fetch and parse list)

---

## Configuration Storage

All data stored in `/etc/netbridge/`:

```
/etc/netbridge/
├── profiles/        # Encrypted profile files
├── sessions/        # Session state and history
├── cache/           # DNS cache, benchmark results
├── logs/            # Rotating log files
├── state/           # Runtime state (active profile, pid, etc.)
├── routes/          # Smart routing rules
└── config.yaml      # Global application configuration
```

File permissions:
- Config directory: `700` (root or dedicated user)
- Profile files: `600`
- Log files: `640`

---

## Security Requirements

- All profile files encrypted at rest using AES-256
- Encryption key derived from machine ID + optional passphrase
- Sensitive values (passwords, keys, tokens) masked in all log output
- Sensitive values masked in `netbridge show` output (use `--reveal` to show)
- No credentials written to shell history
- File permission enforcement on startup (warn or fix if too permissive)
- Memory-safe credential handling (zero out sensitive buffers after use)

---

## Reliability Requirements

- **Connection Watchdog**: background goroutine monitoring connection health
- **Auto-reconnect**: configurable retry with exponential backoff
- **Health Recovery**: restart backend process if it crashes
- **Crash Recovery**: restore last active session on startup if flag set (`--persist`)
- **State Persistence**: write state to disk atomically, survive kill -9
- **Graceful Shutdown**: on SIGTERM/SIGINT, close connections cleanly, flush logs

---

## Plugin SDK

Plugins live in `plugins/` directory (or system plugin path):

```
plugins/
├── tuic/
│   ├── plugin.go
│   └── README.md
├── hysteria/
└── ssh-tunnel/
```

Plugin interface (Go):

```go
type Plugin interface {
    Name() string
    Version() string
    Protocols() []string
    NewBackend(config Config) (Backend, error)
}
```

Plugins are loaded at startup without modifying core application code.

---

## Backend Interface

All backends must implement:

```go
type Backend interface {
    Name() string
    Start(ctx context.Context, cfg BackendConfig) error
    Stop() error
    Status() BackendStatus
    Stats() TrafficStats
    Configure(cfg BackendConfig) error
    HealthCheck() error
    LocalEndpoints() []Endpoint  // returns SOCKS/HTTP/TUN endpoints
}
```

---

## Implementation Stack

### Language

**Go**

Reasons:
- Single static binary, no runtime dependency
- Excellent networking and concurrency support
- Fast startup time
- Strong standard library for system integration
- Easy cross-compilation

### Key Libraries (suggested, not mandatory)

| Purpose | Library |
|---|---|
| CLI parsing | `cobra` + `viper` |
| TUI rendering | `bubbletea` + `lipgloss` |
| Xray integration | `xtls/xray-core` (embedded or subprocess) |
| WireGuard | `wireguard-go` or kernel wgctrl |
| Config format | `gopkg.in/yaml.v3` + `encoding/json` |
| Encryption | `crypto/aes` + `crypto/cipher` (stdlib) |
| Logging | `zap` or `zerolog` |
| Testing | stdlib `testing` + `testify` |

### Build Requirements

- Single binary output: `netbridge`
- Binary size target: under 30MB including embedded Xray
- Must run on: Ubuntu 20.04+, Debian 11+, CentOS 8+, Alpine 3.16+
- Architecture: amd64, arm64
- No external dependencies required at runtime

---

## UX Standards

CLI feel must match `git`, `docker`, `kubectl`:

```
# Good output style
$ netbridge status
● Connected
  Profile   : main (VLESS + Reality)
  Backend   : Xray 1.8.4
  Mode      : SOCKS  127.0.0.1:10808
  Uptime    : 2h 14m 33s
  ↑ Upload  : 1.2 MB/s  (total: 245 MB)
  ↓ Download: 3.8 MB/s  (total: 891 MB)
```

Color conventions:
- Green: connected, success
- Yellow: warning, connecting, partial
- Red: error, disconnected
- Gray: inactive, secondary info
- Cyan: profile names, highlights

Error messages must be:
- Human-readable ("Cannot connect: timeout after 10s")
- Actionable ("Try: netbridge test to diagnose")
- Include exit code documentation

---

## Testing Requirements

The project must include:

- Unit tests for all service layer components
- Integration tests for each backend adapter
- End-to-end tests for critical CLI commands
- Benchmark tests for routing engine
- Mock backend for testing without real VPN setup

Test coverage target: ≥80% for core engine and service layer.

---

## Documentation Requirements

The repository must include:

- `README.md` — quick start, installation, basic usage
- `docs/architecture.md` — architecture overview with diagrams
- `docs/commands.md` — full command reference
- `docs/protocols.md` — supported protocols and import formats
- `docs/plugins.md` — plugin development guide
- `docs/security.md` — security model and recommendations
- Inline `--help` text for every command (auto-generated from code)

---

## Project Structure

```
netbridge/
├── cmd/                    # CLI command definitions (cobra)
│   ├── root.go
│   ├── connect.go
│   ├── profile.go
│   └── ...
├── internal/
│   ├── core/               # Core engine, event bus, orchestrator
│   ├── profile/            # Profile manager, parser, storage
│   ├── session/            # Session manager, state persistence
│   ├── routing/            # Smart routing engine
│   ├── health/             # Health engine, watchdog
│   ├── dns/                # DNS manager
│   ├── benchmark/          # Benchmark and scoring engine
│   ├── stats/              # Traffic statistics
│   ├── security/           # Encryption, credential handling
│   └── config/             # Global configuration
├── adapters/               # Backend adapter implementations
│   ├── xray/
│   ├── singbox/
│   ├── wireguard/
│   └── openvpn/
├── tui/                    # TUI layer (bubbletea components)
│   ├── app.go
│   ├── menu.go
│   ├── statusbar.go
│   └── views/
├── plugins/                # Plugin SDK and built-in plugins
│   ├── sdk.go
│   └── examples/
├── docs/
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── scripts/
│   ├── install.sh
│   └── build.sh
├── Makefile
├── go.mod
└── README.md
```

---

## Delivery Milestones (suggested phases)

### Phase 1 — Core Foundation
- Project structure and clean architecture skeleton
- CLI framework (cobra/viper)
- Profile import/export (VLESS, VMess, Trojan, SS)
- Xray backend adapter
- SOCKS/HTTP session modes
- Basic `connect`, `disconnect`, `status`, `list` commands
- Configuration storage with encryption

### Phase 2 — Operations
- Health engine and watchdog
- Auto-reconnect with backoff
- Smart routing engine
- DNS manager
- Benchmark engine and scoring
- `netbridge run` process mode
- TUN mode (system-wide tunnel)

### Phase 3 — TUI & Polish
- Interactive TUI (bubbletea)
- First-run wizard
- Built-in help system
- Failover chain
- `netbridge auto`
- Backup/restore
- systemd service integration

### Phase 4 — Extensibility
- Plugin SDK
- Additional backend adapters (Sing-box, native WireGuard)
- JSON API mode polish
- Shell completion (bash, zsh, fish)
- Performance optimization
- Full test coverage

---

## Summary

NetBridge must be built as a long-term, open-source, production-grade tool.
It is not a prototype — every component must be designed for maintainability, extensibility, and reliability.

The three non-negotiable architectural pillars are:

1. **Clean separation** — business logic never depends on transport layer
2. **Three interfaces** — CLI, TUI, and JSON API from day one
3. **Backend abstraction** — adding a new VPN/proxy backend requires only implementing the `Backend` interface, nothing else
