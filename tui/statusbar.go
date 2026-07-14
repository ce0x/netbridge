package tui

import (
	"fmt"
	"time"
)

type StatusBar struct {
	profile    string
	status     string
	backend    string
	uptime     time.Duration
	bytesUp    int64
	bytesDown  int64
	rateUp     float64
	rateDown   float64
}

func NewStatusBar() *StatusBar {
	return &StatusBar{}
}

func (sb *StatusBar) Update(profile, status, backend string, uptime time.Duration, up, down int64, rateUp, rateDown float64) {
	sb.profile = profile
	sb.status = status
	sb.backend = backend
	sb.uptime = uptime
	sb.bytesUp = up
	sb.bytesDown = down
	sb.rateUp = rateUp
	sb.rateDown = rateDown
}

func (sb *StatusBar) View() string {
	statusColor := StatusColor(sb.status)
	return fmt.Sprintf(
		"\nActive Profile : %s\nStatus         : %s%s%s\nBackend        : %s\nUptime         : %s\n",
		sb.profile,
		statusColor, sb.status, ColorReset,
		sb.backend,
		sb.uptime,
	)
}
