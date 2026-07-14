package tui

import (
	netbridge "github.com/netbridge/netbridge"
)

type SettingsView struct {
	engine netbridge.CoreEngine
}

func NewSettingsView(engine netbridge.CoreEngine) *SettingsView {
	return &SettingsView{engine: engine}
}

func (s *SettingsView) View() string {
	return `
  Settings:
  ─────────────────────────────────────────────
  Default Mode:     socks
  Default Port:     10808
  Watchdog:         enabled
  Auto-reconnect:   enabled
  Log Level:        info
  
  Press 'b' to go back.
`
}
