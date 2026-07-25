package tui

import (
	"context"
	"fmt"
	"net"
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

	menuKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	menuText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	updateBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true)

	separator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

type App struct {
	engine      netbridge.CoreEngine
	quitting    bool
	width       int
	height      int
	stats       netbridge.TrafficStats
	status      netbridge.ConnectionStatus
	profileName string
	serverInfo  string
	port        int
	mode        string
	localAddr   string
	updateMsg   string
	ipv4        string
	ipv6        string
	tickCount   int
}

type tickMsg struct{}

func NewApp(engine netbridge.CoreEngine) *App {
	a := &App{
		engine:   engine,
		localAddr: "127.0.0.1:10808",
	}
	a.detectNetwork()
	return a
}

func (a *App) detectNetwork() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
			a.ipv4 = ipNet.IP.String()
		}
		if ipNet.IP.To4() == nil && !ipNet.IP.IsLoopback() && ipNet.IP.String() != "::1" {
			a.ipv6 = ipNet.IP.String()
		}
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
			return a, a.doHelp()
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
			a.mode = string(sess.Mode)
			a.localAddr = sess.LocalAddr
			if p, err := a.engine.ProfileManager().Get(ctx, sess.ProfileID); err == nil {
				a.profileName = p.Name
				a.serverInfo = p.Server
				a.port = p.Port
			}
			a.stats = a.engine.SessionManager().Stats()
		}
	} else {
		a.profileName = ""
		a.serverInfo = ""
		a.port = 0
		a.mode = ""
		a.localAddr = ""
	}
}

func (a *App) View() string {
	if a.quitting {
		return "\nGoodbye!\n"
	}

	var sb strings.Builder

	// ─── Header with ASCII logo ───
	logo := `
  _   _    _  _____ ____  _____   __
 | \ | |  / \|_   _/ ___|| ____| / _|
 |  \| | / _ \ | | \___ \|  _|  | |_
 | |\  |/ ___ \| |  ___) | |___ |  _|
 |_| \_/_/   \_\_| |____/|_____||_|
`
	sb.WriteString(titleStyle.Render(logo))
	sb.WriteString("\n")

	// ─── Separator ───
	sb.WriteString(separator.Render(strings.Repeat("═", 52)) + "\n")

	// ─── System info line ───
	now := time.Now().Format("15:04:05")
	username := "unknown"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	statusStr := statusErr.Render("● Disconnected")
	if a.status == netbridge.StatusConnected {
		statusStr = statusOK.Render("● Connected")
	}
	sb.WriteString(fmt.Sprintf("  Time: %-20s OS: %s/%s\n", now, runtime.GOOS, runtime.GOARCH))
	sb.WriteString(fmt.Sprintf("  User: %-20s %s\n", username, statusStr))

	// ─── Separator ───
	sb.WriteString(separator.Render(strings.Repeat("─", 52)) + "\n")

	// ─── Profile info ───
	if a.status == netbridge.StatusConnected && a.profileName != "" {
		sb.WriteString(fmt.Sprintf("  Profile: %s (%s:%d)\n", a.profileName, a.serverInfo, a.port))

		// Stats line — upload/download are zero until T13 xray stats gRPC is wired
		uptime := a.stats.Uptime.Round(time.Second)
		sb.WriteString(fmt.Sprintf("  Ping: 0ms   ↑ 0 KB/s   ↓ 0 KB/s   Uptime: %s\n", uptime))

		// Network info
		ipv4Str := a.ipv4
		if ipv4Str == "" {
			ipv4Str = "N/A"
		}
		ipv6Str := a.ipv6
		if ipv6Str == "" {
			ipv6Str = "N/A"
		}
		sb.WriteString(fmt.Sprintf("  Local IPv4: %-20s IPv6: %s\n", ipv4Str, ipv6Str))

		// Mode info
		modeStr := "unknown"
		switch a.mode {
		case "socks":
			modeStr = fmt.Sprintf("socks mode, %s", a.localAddr)
		case "http":
			modeStr = fmt.Sprintf("http mode, %s", a.localAddr)
		case "tun":
			modeStr = "tun mode"
		}
		sb.WriteString(fmt.Sprintf("  TUN: OFF (%s)\n", modeStr))
	} else {
		sb.WriteString("  No active session\n")
		sb.WriteString("\n")
	}

	// ─── Separator ───
	sb.WriteString(separator.Render(strings.Repeat("─", 52)) + "\n")

	// ─── Menu grid ───
	line1 := fmt.Sprintf("  %s Import    %s Connect    %s Status",
		menuKey.Render("[1]"), menuKey.Render("[2]"), menuKey.Render("[3]"))
	line2 := fmt.Sprintf("  %s Disconnect  %s Update    %s Help",
		menuKey.Render("[4]"), menuKey.Render("[5]"), menuKey.Render("[H]"))
	line3 := fmt.Sprintf("  %s Quit",
		menuKey.Render("[Q]"))

	sb.WriteString(line1 + "\n")
	sb.WriteString(line2 + "\n")
	sb.WriteString(line3 + "\n")

	// ─── Update banner ───
	if a.updateMsg != "" {
		sb.WriteString("\n" + updateBanner.Render("  "+a.updateMsg))
	}

	return sb.String()
}

// ─── Actions ───

func (a *App) doImport() tea.Cmd {
	return func() tea.Msg {
		if a.engine == nil {
			return tea.Msg("no engine")
		}
		// List current profiles for feedback
		profiles, err := a.engine.ProfileManager().List(context.Background())
		if err != nil {
			return tea.Msg(fmt.Sprintf("error listing profiles: %v", err))
		}
		return tea.Msg(fmt.Sprintf("Profiles: %d total. Use 'netbridge profile import <uri>' to add.", len(profiles)))
	}
}

func (a *App) doConnect() tea.Cmd {
	return func() tea.Msg {
		if a.engine == nil {
			return tea.Msg("no engine")
		}
		ctx := context.Background()
		mgr := a.engine.ProfileManager()

		// Try active profile first
		active, err := mgr.GetActive(ctx)
		if err != nil {
			// Try first available profile
			profiles, listErr := mgr.List(ctx)
			if listErr != nil || len(profiles) == 0 {
				return tea.Msg("no profiles available — import one first")
			}
			active = profiles[0]
		}

		_, err = a.engine.SessionManager().Connect(ctx, active.ID, netbridge.ModeSOCKS)
		if err != nil {
			return tea.Msg(fmt.Sprintf("connect error: %v", err))
		}
		_ = mgr.SetActive(ctx, active.ID)
		return tea.Msg(fmt.Sprintf("Connected to %s", active.Name))
	}
}

func (a *App) doStatus() tea.Cmd {
	return func() tea.Msg {
		if a.engine == nil {
			return tea.Msg("no engine")
		}
		sm := a.engine.SessionManager()
		status := sm.Status()
		if status == netbridge.StatusDisconnected {
			return tea.Msg("Status: Disconnected — no active session")
		}
		sess, err := sm.Current()
		if err != nil {
			return tea.Msg(fmt.Sprintf("Status: %s", status))
		}
		stats := sm.Stats()
		return tea.Msg(fmt.Sprintf("Status: %s | Session: %s | Uptime: %s",
			status, sess.ID, stats.Uptime.Round(time.Second)))
	}
}

func (a *App) doDisconnect() tea.Cmd {
	return func() tea.Msg {
		if a.engine == nil {
			return tea.Msg("no engine")
		}
		if err := a.engine.SessionManager().Disconnect(context.Background()); err != nil {
			return tea.Msg(fmt.Sprintf("disconnect error: %v", err))
		}
		return tea.Msg("Disconnected")
	}
}

func (a *App) doUpdateCheck() tea.Cmd {
	return func() tea.Msg {
		info, err := selfupdate.CheckLatest(context.Background())
		if err != nil {
			return tea.Msg(fmt.Sprintf("update check error: %v", err))
		}
		if info.UpdateAvailable {
			return tea.Msg(fmt.Sprintf("Update available: %s → %s (run 'netbridge update install')", info.Current, info.Latest))
		}
		return tea.Msg(fmt.Sprintf("Already up to date (%s)", info.Current))
	}
}

func (a *App) doHelp() tea.Cmd {
	return func() tea.Msg {
		help := `
Commands:
  netbridge tui                    Launch this dashboard
  netbridge profile import <uri>   Import a VLESS/VMess/Trojan/SS profile
  netbridge profile list           List all profiles
  netbridge connect [profile]      Connect (uses active if no arg)
  netbridge disconnect             Disconnect current session
  netbridge status                 Show connection status
  netbridge health [profile]       Health check a profile
  netbridge core install --all     Install all core backends
  netbridge update check           Check for updates
  netbridge update install         Install latest version
`
		return tea.Msg(help)
	}
}

func RunApp(engine netbridge.CoreEngine) error {
	app := NewApp(engine)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
