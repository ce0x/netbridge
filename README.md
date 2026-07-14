# NetBridge

Universal network access and connectivity toolkit for Linux servers.

## Features

- Multi-protocol support (VLESS, VMess, Trojan, Shadowsocks, WireGuard, OpenVPN)
- Three interface modes (CLI, TUI, JSON API)
- Smart routing engine
- Health monitoring and failover
- Benchmark and scoring
- DNS management
- Plugin SDK for extensibility

## Installation

```bash
# Quick install
curl -fsSL https://get.netbridge.dev | bash

# Or build from source
git clone https://github.com/netbridge/netbridge
cd netbridge
make build
sudo make install
```

## Quick Start

```bash
# Import a profile
netbridge import "vless://uuid@server:443?security=tls#my-server"

# List profiles
netbridge list

# Connect
netbridge connect my-server

# Check status
netbridge status
```

## Commands

```bash
netbridge import <url|file>     Import profile
netbridge list                  List profiles
netbridge connect [profile]     Connect
netbridge disconnect            Disconnect
netbridge status                Show status
netbridge test [profile]        Test connectivity
netbridge benchmark [--all]     Benchmark profiles
netbridge route add <domain> <profile>  Smart routing
netbridge dns list              DNS presets
netbridge env                   Proxy env vars
```

## JSON API

All commands support `--json` flag:

```bash
netbridge status --json
netbridge list --json
```

## TUI Mode

Launch interactive terminal UI:

```bash
netbridge tui
```

## Documentation

- [Architecture](docs/architecture.md)
- [Commands](docs/commands.md)
- [Protocols](docs/protocols.md)
- [Plugins](docs/plugins.md)
- [Security](docs/security.md)

## License

MIT License
