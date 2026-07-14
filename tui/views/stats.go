package tui

import (
	netbridge "github.com/netbridge/netbridge"
)

type StatsView struct {
	engine netbridge.CoreEngine
}

func NewStatsView(engine netbridge.CoreEngine) *StatsView {
	return &StatsView{engine: engine}
}

func (s *StatsView) View() string {
	return `
  Traffic Statistics:
  ─────────────────────────────────────────────
  ↑ Total Upload:   0 bytes
  ↓ Total Download: 0 bytes
  Session Uptime:   0s
  
  Press 'b' to go back.
`
}
