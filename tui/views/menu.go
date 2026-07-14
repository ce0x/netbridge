package tui

import (
	"fmt"

	netbridge "github.com/netbridge/netbridge"
)

type MenuView struct {
	engine netbridge.CoreEngine
}

func NewMenuView(engine netbridge.CoreEngine) *MenuView {
	return &MenuView{engine: engine}
}

func (m *MenuView) View() string {
	return fmt.Sprintf(`
╔══════════════════════════════════════════════════════╗
║                   NetBridge CLI                     ║
╠══════════════════════════════════════════════════════╣
║  1. Profile Management                              ║
║  2. Import Profile                                  ║
║  3. List Profiles                                   ║
║  4. Test Profile                                    ║
║  5. Connect                                         ║
║  6. Disconnect                                      ║
║  7. Status                                          ║
║------------------------------------------------------║
║  8. Health Check                                    ║
║  9. Benchmark                                       ║
║ 10. Smart Routing                                   ║
║ 11. DNS Management                                  ║
║------------------------------------------------------║
║ 12. Service Management                              ║
║ 13. Logs                                            ║
║ 14. Statistics                                      ║
║------------------------------------------------------║
║ 15. Settings                                        ║
║ 16. Backup                                          ║
║ 17. Restore                                         ║
║------------------------------------------------------║
║  0. Exit                                            ║
╚══════════════════════════════════════════════════════╝
`)
}

func (m *MenuView) HandleInput(input string) string {
	switch input {
	case "1":
		return "profiles"
	case "2":
		return "import"
	case "5":
		return "connect"
	case "7":
		return "status"
	case "0":
		return "quit"
	default:
		return "menu"
	}
}
