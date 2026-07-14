package tui

import (
	"fmt"

	netbridge "github.com/netbridge/netbridge"
)

type DNSView struct {
	engine netbridge.CoreEngine
}

func NewDNSView(engine netbridge.CoreEngine) *DNSView {
	return &DNSView{engine: engine}
}

func (d *DNSView) View() string {
	presets := d.engine.DNSEngine().ListPresets()
	current := d.engine.DNSEngine().CurrentResolver()

	output := "\n  DNS Configuration:\n  ─────────────────────────────────────────\n"
	output += fmt.Sprintf("  Current: %s\n\n", current)
	for _, p := range presets {
		output += fmt.Sprintf("  • %s: %v\n", p.Name, p.Servers)
	}
	output += "\n  Press 'b' to go back.\n"
	return output
}

var _ = netbridge.CoreEngine(nil)
