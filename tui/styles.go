package tui

var (
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[90m"
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
)

func StatusColor(status string) string {
	switch status {
	case "connected":
		return ColorGreen
	case "connecting", "reconnecting":
		return ColorYellow
	case "error", "disconnected":
		return ColorRed
	default:
		return ColorGray
	}
}
