package tui

import (
	netbridge "github.com/netbridge/netbridge"
)

type LogsView struct {
	engine netbridge.CoreEngine
}

func NewLogsView(engine netbridge.CoreEngine) *LogsView {
	return &LogsView{engine: engine}
}

func (l *LogsView) View() string {
	return `
  Logs:
  ─────────────────────────────────────────────
  (no log entries)
  
  Press 'f' to follow, 'b' to go back.
`
}
