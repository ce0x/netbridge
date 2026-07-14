# NetBridge Commands Reference

## Profile Management

| Command | Description |
|---------|-------------|
| `netbridge import <url\|file>` | Import profile from URL or file |
| `netbridge export <name>` | Export profile |
| `netbridge delete <name>` | Delete profile |
| `netbridge rename <old> <new>` | Rename profile |
| `netbridge clone <name> <new>` | Clone profile |
| `netbridge list` | List all profiles |
| `netbridge use <name>` | Set active profile |
| `netbridge show <name>` | Show profile details |

## Connection Management

| Command | Description |
|---------|-------------|
| `netbridge connect [profile]` | Connect active or named profile |
| `netbridge disconnect` | Disconnect current session |
| `netbridge restart` | Restart current connection |
| `netbridge reload` | Reload config without disconnecting |
| `netbridge status` | Show current connection status |

## Session Modes

| Command | Description |
|---------|-------------|
| `netbridge connect --mode socks` | Local SOCKS5 at 127.0.0.1:10808 |
| `netbridge connect --mode http` | Local HTTP proxy at 127.0.0.1:8080 |
| `netbridge connect --mode tun` | Virtual TUN interface |
| `netbridge run <command>` | Run command through active profile |

## Testing & Health

| Command | Description |
|---------|-------------|
| `netbridge test [profile]` | Test profile connectivity |
| `netbridge health [profile]` | Health check for profile |
| `netbridge benchmark [--all]` | Benchmark and score profiles |

## Smart Routing

| Command | Description |
|---------|-------------|
| `netbridge route add <domain> <profile>` | Add routing rule |
| `netbridge route remove <domain>` | Remove routing rule |
| `netbridge route list` | List routing rules |
| `netbridge route clear` | Clear all rules |

## DNS Management

| Command | Description |
|---------|-------------|
| `netbridge dns list` | List DNS presets |
| `netbridge dns use <preset\|ip>` | Set active DNS resolver |
| `netbridge dns benchmark` | Benchmark DNS resolvers |
| `netbridge dns show` | Show current DNS config |
| `netbridge dns reset` | Restore system default DNS |

## Shell Integration

| Command | Description |
|---------|-------------|
| `netbridge env` | Print proxy env var exports |
| `netbridge unset` | Print proxy env var unsets |

## Service Management

| Command | Description |
|---------|-------------|
| `netbridge service install` | Install systemd unit |
| `netbridge service start` | Start service |
| `netbridge service stop` | Stop service |
| `netbridge service restart` | Restart service |
| `netbridge service status` | Show service status |
| `netbridge service uninstall` | Remove systemd unit |

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |
| `-q, --quiet` | Suppress non-essential output |
| `-v, --verbose` | Enable verbose output |
