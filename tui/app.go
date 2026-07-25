package tui

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/selfupdate"
	"golang.org/x/net/proxy"
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

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	updateBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true)

	resultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
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
	statusMsg   string
	msgAge      int
	ipv4        string
	ipv6        string
	egressIP    string
	egressAge   int
	tickCount   int
	ping        time.Duration
	lastBytesUp int64
	lastBytesDn int64
	lastTick    time.Time
	rateUp      float64
	rateDown    float64
}

type tickMsg struct{}

func NewApp(engine netbridge.CoreEngine) *App {
	a := &App{
		engine:    engine,
		localAddr: "127.0.0.1:10808",
		lastTick:  time.Now(),
	}
	a.detectNetwork()
	return a
}

func (a *App) detectNetwork() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	var fallback4, fallback6 string
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		// Handle IPv4
		if ip4 := ipNet.IP.To4(); ip4 != nil {
			if !ip4.IsLoopback() {
				if ip4.IsPrivate() {
					a.ipv4 = ip4.String()
				} else if fallback4 == "" {
					fallback4 = ip4.String()
				}
			}
			continue // do NOT fall through to IPv6 branch
		}
		// Handle real IPv6
		if ip6 := ipNet.IP.To16(); ip6 != nil && !ip6.IsLoopback() {
			if isULAPrefix(ip6) {
				a.ipv6 = ip6.String()
			} else if fallback6 == "" {
				fallback6 = ip6.String()
			}
		}
	}
	if a.ipv4 == "" {
		a.ipv4 = fallback4
	}
	if a.ipv6 == "" {
		a.ipv6 = fallback6
	}
}

func isULAPrefix(ip net.IP) bool {
	return len(ip) == 16 && ip[0] == 0xfc
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
		a.msgAge++
		if a.msgAge > 8 {
			a.statusMsg = ""
		}
		a.refreshData()
		return a, tickCmd()

	case string:
		a.statusMsg = msg
		a.msgAge = 0
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			a.quitting = true
			return a, tea.Quit

		case "1":
			a.statusMsg = ""
			return a, a.doImport()
		case "2":
			a.statusMsg = ""
			return a, a.doConnect()
		case "3":
			a.statusMsg = ""
			return a, a.doStatus()
		case "4":
			a.statusMsg = ""
			return a, a.doDisconnect()
		case "5":
			a.statusMsg = ""
			return a, a.doUpdateCheck()
		case "h", "H":
			a.statusMsg = ""
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

			// Calculate transfer rate
			now := time.Now()
			elapsed := now.Sub(a.lastTick).Seconds()
			if elapsed > 0 {
				a.rateUp = float64(a.stats.BytesUp-a.lastBytesUp) / elapsed
				a.rateDown = float64(a.stats.BytesDown-a.lastBytesDn) / elapsed
				a.lastBytesUp = a.stats.BytesUp
				a.lastBytesDn = a.stats.BytesDown
				a.lastTick = now
			}

			// Check egress IP every 30 seconds
			a.egressAge++
			if a.egressIP == "" || a.egressAge >= 30 {
				a.egressAge = 0
				go a.checkEgressIP(sess.LocalAddr)
			}
		}
	} else {
		a.profileName = ""
		a.serverInfo = ""
		a.port = 0
		a.mode = ""
		a.localAddr = ""
		a.ping = 0
		a.rateUp = 0
		a.rateDown = 0
		a.egressIP = ""
	}
}

func (a *App) checkEgressIP(localAddr string) {
	dialer, err := proxy.SOCKS5("tcp", localAddr, nil, proxy.Direct)
	if err != nil {
		a.egressIP = "unreachable"
		return
	}

	transport := &http.Transport{
		DialContext: dialer.(interface {
			DialContext(ctx context.Context, network, addr string) (net.Conn, error)
		}).DialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		a.egressIP = "unreachable"
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.egressIP = "unreachable"
		return
	}

	a.egressIP = strings.TrimSpace(string(body))
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

		// Ping (real if measured, otherwise shows N/A until first measurement)
		pingStr := "N/A"
		if a.ping > 0 {
			pingStr = fmt.Sprintf("%dms", a.ping.Milliseconds())
		}
		sb.WriteString(fmt.Sprintf("  Ping: %s   ↑ %s   ↓ %s   Uptime: %s\n",
			pingStr,
			formatRate(a.rateUp),
			formatRate(a.rateDown),
			a.stats.Uptime.Round(time.Second)))

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

		// Egress IP (through tunnel)
		egressStr := "checking..."
		if a.egressIP == "unreachable" {
			egressStr = statusErr.Render("unreachable")
		} else if a.egressIP != "" {
			egressStr = a.egressIP + " (via tunnel)"
		}
		sb.WriteString(fmt.Sprintf("  Egress IP:  %s\n", egressStr))

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

	// ─── Status message from actions ───
	if a.statusMsg != "" {
		sb.WriteString("\n" + resultStyle.Render("  "+a.statusMsg))
	}

	return sb.String()
}

func formatRate(bps float64) string {
	if bps < 1024 {
		return fmt.Sprintf("%.0f B/s", bps)
	} else if bps < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", bps/1024)
	}
	return fmt.Sprintf("%.1f MB/s", bps/(1024*1024))
}

// ─── Actions ───

func (a *App) doImport() tea.Cmd {
	return func() tea.Msg {
		if a.engine == nil {
			return "no engine"
		}
		profiles, err := a.engine.ProfileManager().List(context.Background())
		if err != nil {
			return fmt.Sprintf("Error listing profiles: %v", err)
		}
		if len(profiles) == 0 {
			return "No profiles. Run 'netbridge profile import <uri>' to add one."
		}
		var names []string
		for _, p := range profiles {
			names = append(names, fmt.Sprintf("%s (%s:%d)", p.Name, p.Server, p.Port))
		}
		return fmt.Sprintf("Profiles (%d): %s", len(profiles), strings.Join(names, ", "))
	}
}

func (a *App) doConnect() tea.Cmd {
	return func() tea.Msg {
		if a.engine == nil {
			return "no engine"
		}
		ctx := context.Background()
		mgr := a.engine.ProfileManager()

		active, err := mgr.GetActive(ctx)
		if err != nil {
			profiles, listErr := mgr.List(ctx)
			if listErr != nil || len(profiles) == 0 {
				return "No profiles available — import one first."
			}
			active = profiles[0]
		}

		_, err = a.engine.SessionManager().Connect(ctx, active.ID, netbridge.ModeSOCKS)
		if err != nil {
			return fmt.Sprintf("Connect error: %v", err)
		}
		_ = mgr.SetActive(ctx, active.ID)
		return fmt.Sprintf("Connected to %s", active.Name)
	}
}

func (a *App) doStatus() tea.Cmd {
	return func() tea.Msg {
		if a.engine == nil {
			return "no engine"
		}
		sm := a.engine.SessionManager()
		status := sm.Status()
		if status == netbridge.StatusDisconnected {
			return "Status: Disconnected — no active session"
		}
		sess, err := sm.Current()
		if err != nil {
			return fmt.Sprintf("Status: %s", status)
		}
		stats := sm.Stats()
		return fmt.Sprintf("Status: %s | Session: %s | Uptime: %s",
			status, sess.ID, stats.Uptime.Round(time.Second))
	}
}

func (a *App) doDisconnect() tea.Cmd {
	return func() tea.Msg {
		if a.engine == nil {
			return "no engine"
		}
		if err := a.engine.SessionManager().Disconnect(context.Background()); err != nil {
			return fmt.Sprintf("Disconnect error: %v", err)
		}
		return "Disconnected"
	}
}

func (a *App) doUpdateCheck() tea.Cmd {
	return func() tea.Msg {
		info, err := selfupdate.CheckLatest(context.Background())
		if err != nil {
			return fmt.Sprintf("Update check error: %v", err)
		}
		if info.UpdateAvailable {
			return fmt.Sprintf("Update available: %s → %s (run 'netbridge update install')", info.Current, info.Latest)
		}
		return fmt.Sprintf("Already up to date (%s)", info.Current)
	}
}

func (a *App) doHelp() tea.Cmd {
	return func() tea.Msg {
		return `Commands:
  netbridge tui                    Launch this dashboard
  netbridge profile import <uri>   Import a VLESS/VMess/Trojan/SS profile
  netbridge profile list           List all profiles
  netbridge connect [profile]      Connect (uses active if no arg)
  netbridge disconnect             Disconnect current session
  netbridge status                 Show connection status
  netbridge health [profile]       Health check a profile
  netbridge core install --all     Install all core backends
  netbridge update check           Check for updates
  netbridge update install         Install latest version`
	}
}

func RunApp(engine netbridge.CoreEngine) error {
	app := NewApp(engine)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
