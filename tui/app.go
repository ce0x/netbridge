package tui

import (
	"fmt"

	netbridge "github.com/netbridge/netbridge"
)

type App struct {
	engine    netbridge.CoreEngine
	currentView string
	quitting  bool
}

func NewApp(engine netbridge.CoreEngine) *App {
	return &App{
		engine:      engine,
		currentView: "menu",
	}
}

func (a *App) Init() {}

func (a *App) Update(msg string) (string, error) {
	switch msg {
	case "quit":
		a.quitting = true
		return "", nil
	case "menu":
		a.currentView = "menu"
	case "profiles":
		a.currentView = "profiles"
	case "connect":
		a.currentView = "connect"
	}
	return a.currentView, nil
}

func (a *App) View() string {
	if a.quitting {
		return "Goodbye!\n"
	}

	switch a.currentView {
	case "menu":
		return a.menuView()
	case "profiles":
		return a.profilesView()
	case "connect":
		return a.connectView()
	default:
		return a.menuView()
	}
}

func (a *App) menuView() string {
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

func (a *App) profilesView() string {
	return "Profiles view\n"
}

func (a *App) connectView() string {
	return "Connect view\n"
}
