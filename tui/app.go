package tui

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/selfupdate"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Align(lipgloss.Center)

	statusOK = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	statusErr = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	menuStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)

type App struct {
	engine      netbridge.CoreEngine
	currentView string
	selected    int
	quitting    bool
	width       int
	height      int
	lastUpdate  time.Time
	stats       netbridge.TrafficStats
	status      netbridge.ConnectionStatus
	profileName string
	serverInfo  string
	updateInfo  string
	tickCount   int
}

type tickMsg struct{}

func NewApp(engine netbridge.CoreEngine) *App {
	return &App{
		engine:      engine,
		currentView: "home",
		selected:    1,
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(tickCmd(), tea.EnterAltScreen)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tickMsg:
		a.tickCount++
		a.refreshData()
		return a, tickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			a.quitting = true
			return a, tea.Quit

		case "1":
			return a, a.doImport()
		case "2":
			return a, a.doConnect()
		case "3":
			return a, a.doStatus()
		case "4":
			return a, a.doDisconnect()
		case "5":
			return a, a.doUpdateCheck()
		case "h", "H":
			a.currentView = "help"
		case "esc":
			a.currentView = "home"
		}
	}

	return a, nil
}

func (a *App) refreshData() {
	if a.engine == nil {
		return
	}
	ctx := context.Background()

	a.status = a.engine.SessionManager().Status()

	if a.status == netbridge.StatusConnected {
		if sess, err := a.engine.SessionManager().Current(); err == nil {
			a.profileName = sess.ProfileID
			if p, err := a.engine.ProfileManager().Get(ctx, sess.ProfileID); err == nil {
				a.profileName = p.Name
				a.serverInfo = fmt.Sprintf("%s:%d", p.Server, p.Port)
			}
			a.stats = a.engine.SessionManager().Stats()
		}
	}
}

func (a *App) View() string {
	if a.quitting {
		return "\nGoodbye!\n"
	}

	var sb strings.Builder

	// Header
	header := titleStyle.Render("NETBRIDGE")
	sb.WriteString(header + "\n")

	// Separator
	sb.WriteString(strings.Repeat("─", 50) + "\n")

	// System info
	now := time.Now().Format("15:04:05")
	username := "unknown"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	sb.WriteString(fmt.Sprintf("  Time: %s    OS: %s/%s    User: %s\n",
		now, runtime.GOOS, runtime.GOARCH, username))

	// Status
	statusStr := statusErr.Render("● Disconnected")
	if a.status == netbridge.StatusConnected {
		statusStr = statusOK.Render("● Connected")
	}
	sb.WriteString(fmt.Sprintf("  Status: %s\n", statusStr))

	sb.WriteString(strings.Repeat("─", 50) + "\n")

	// Profile info
	if a.status == netbridge.StatusConnected && a.profileName != "" {
		sb.WriteString(fmt.Sprintf("  Profile: %s (%s)\n", a.profileName, a.serverInfo))
		sb.WriteString(fmt.Sprintf("  Uptime:  %s\n", a.stats.Uptime.Round(time.Second)))
	} else {
		sb.WriteString("  No active session\n")
	}

	sb.WriteString(strings.Repeat("─", 50) + "\n")

	// Menu
	menu := a.buildMenu()
	sb.WriteString(menu)

	// Help hint
	sb.WriteString("\n" + helpStyle.Render("  Press 1-5 for actions, H for help, Q to quit"))

	return sb.String()
}

func (a *App) buildMenu() string {
	items := []struct {
		key  string
		text string
	}{
		{"1", "Import Profile"},
		{"2", "Connect"},
		{"3", "Status"},
		{"4", "Disconnect"},
		{"5", "Check Update"},
	}

	var lines []string
	for _, item := range items {
		prefix := menuStyle.Render(fmt.Sprintf("  [%s]", item.key))
		text := item.text
		lines = append(lines, prefix+" "+text)
	}

	return strings.Join(lines, "\n")
}

func (a *App) doImport() tea.Cmd {
	return func() tea.Msg {
		return tea.Msg("import")
	}
}

func (a *App) doConnect() tea.Cmd {
	return func() tea.Msg {
		if a.engine == nil {
			return tea.Msg("no engine")
		}
		ctx := context.Background()
		mgr := a.engine.ProfileManager()
		active, err := mgr.GetActive(ctx)
		if err != nil {
			return tea.Msg("no active profile")
		}
		_, err = a.engine.SessionManager().Connect(ctx, active.ID, netbridge.ModeSOCKS)
		if err != nil {
			return tea.Msg(fmt.Sprintf("connect error: %v", err))
		}
		return tea.Msg("connected")
	}
}

func (a *App) doStatus() tea.Cmd {
	return func() tea.Msg {
		return tea.Msg("status")
	}
}

func (a *App) doDisconnect() tea.Cmd {
	return func() tea.Msg {
		if a.engine == nil {
			return tea.Msg("no engine")
		}
		_ = a.engine.SessionManager().Disconnect(context.Background())
		return tea.Msg("disconnected")
	}
}

func (a *App) doUpdateCheck() tea.Cmd {
	return func() tea.Msg {
		info, err := selfupdate.CheckLatest(context.Background())
		if err != nil {
			return tea.Msg(fmt.Sprintf("update check error: %v", err))
		}
		if info.UpdateAvailable {
			return tea.Msg(fmt.Sprintf("Update available: %s → %s", info.Current, info.Latest))
		}
		return tea.Msg("Already up to date")
	}
}

func RunApp(engine netbridge.CoreEngine) error {
	app := NewApp(engine)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func init() {
	_ = fmt.Sprintf
	_ = os.Exit
}
