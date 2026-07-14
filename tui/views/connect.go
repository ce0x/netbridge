package tui

import (
	"fmt"

	netbridge "github.com/netbridge/netbridge"
)

type ConnectView struct {
	engine netbridge.CoreEngine
}

func NewConnectView(engine netbridge.CoreEngine) *ConnectView {
	return &ConnectView{engine: engine}
}

func (c *ConnectView) View() string {
	status := c.engine.SessionManager().Status()
	if status == netbridge.StatusConnected {
		stats := c.engine.SessionManager().Stats()
		return fmt.Sprintf(`
  ● Connected
    Uptime    : %s
    ↑ Upload  : %.1f KB/s
    ↓ Download: %.1f KB/s

  Press 'd' to disconnect, 'b' to go back.
`, stats.Uptime, stats.RateUp/1024, stats.RateDown/1024)
	}

	return `
  ○ Disconnected
  
  Select a profile to connect, or press 'b' to go back.
`
}
