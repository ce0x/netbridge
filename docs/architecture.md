# NetBridge Architecture

## Overview

NetBridge follows a clean architecture with strict separation of concerns.

## Layers

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

## Key Design Rules

1. Business logic never imports adapters
2. No layer may import from a layer above it
3. Core Engine does not know which backend is running
4. All backends implement the same `Backend` interface
5. Three interface modes (CLI, TUI, JSON) share the same core engine

## Backend Interface

Every transport adapter must implement:

```go
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
```

## Data Flow

1. CLI/TUI parses user command
2. Calls appropriate method on CoreEngine
3. CoreEngine delegates to service layer
4. Service layer interacts with backend adapters
5. Results flow back up through the layers
6. Output formatted for the interface mode (text/JSON)
