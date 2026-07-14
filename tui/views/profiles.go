package tui

import (
	"fmt"

	netbridge "github.com/netbridge/netbridge"
)

type ProfilesView struct {
	engine netbridge.CoreEngine
}

func NewProfilesView(engine netbridge.CoreEngine) *ProfilesView {
	return &ProfilesView{engine: engine}
}

func (p *ProfilesView) View() string {
	profiles, err := p.engine.ProfileManager().List(nil)
	if err != nil || len(profiles) == 0 {
		return "\n  No profiles configured.\n\n  Press 'i' to import a profile, or 'b' to go back.\n"
	}

	output := "\n  Profiles:\n  ─────────────────────────────────────────\n"
	for i, prof := range profiles {
		output += fmt.Sprintf("  %d. %s (%s)\n", i+1, prof.Name, prof.Protocol)
	}
	output += "\n  Press 'b' to go back.\n"
	return output
}

var _ = netbridge.CoreEngine(nil)
