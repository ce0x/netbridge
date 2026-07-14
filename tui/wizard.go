package tui

import (
	"fmt"

	netbridge "github.com/netbridge/netbridge"
)

type Wizard struct {
	engine netbridge.CoreEngine
	step   int
}

func NewWizard(engine netbridge.CoreEngine) *Wizard {
	return &Wizard{
		engine: engine,
		step:   0,
	}
}

func (w *Wizard) View() string {
	return fmt.Sprintf(`
Welcome to NetBridge!

No profiles found. Let's get started.

Import a profile:
  1) VLESS URL
  2) VMess URL
  3) Trojan URL
  4) WireGuard config
  5) OpenVPN config
  6) Import from file
  7) Skip for now

Select: 
`)
}

func (w *Wizard) HandleInput(input string) string {
	switch input {
	case "1":
		return "Enter VLESS URL: "
	case "2":
		return "Enter VMess URL: "
	case "3":
		return "Enter Trojan URL: "
	case "4":
		return "Path to WireGuard .conf file: "
	case "5":
		return "Path to OpenVPN .ovpn file: "
	case "6":
		return "Path to file: "
	case "7":
		return "Skipping..."
	default:
		return "Invalid selection. Try again."
	}
}
