package tui

import (
	netbridge "github.com/netbridge/netbridge"
)

type RoutingView struct {
	engine netbridge.CoreEngine
}

func NewRoutingView(engine netbridge.CoreEngine) *RoutingView {
	return &RoutingView{engine: engine}
}

func (r *RoutingView) View() string {
	return `
  Routing Rules:
  ─────────────────────────────────────────────
  (no rules configured)
  
  Press 'a' to add a rule, 'b' to go back.
`
}
