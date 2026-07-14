# NetBridge Plugin Development Guide

## Plugin Interface

Every plugin must implement the `Plugin` interface:

```go
type Plugin interface {
    Name() string
    Version() string
    Protocols() []Protocol
    NewBackend(profile Profile) (Backend, error)
}
```

## Creating a Plugin

1. Create a new directory under `plugins/`
2. Implement the `Plugin` interface
3. Register the plugin with the registry

## Example Plugin

```go
package myplugin

import netbridge "github.com/netbridge/netbridge"

type MyPlugin struct{}

func (p *MyPlugin) Name() string { return "myplugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) Protocols() []netbridge.Protocol {
    return []netbridge.Protocol{"myproto"}
}
func (p *MyPlugin) NewBackend(profile netbridge.Profile) (netbridge.Backend, error) {
    return &MyBackend{}, nil
}
```

## Backend Interface

Your backend must implement:

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

## Loading Plugins

Plugins are loaded from the `plugins/` directory at startup.

```bash
netbridge plugin load ./plugins/myplugin
netbridge plugin list
netbridge plugin unload myplugin
```
