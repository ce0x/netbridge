package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DataDir          string `yaml:"data_dir"`
	LogDir           string `yaml:"log_dir"`
	DefaultMode      string `yaml:"default_mode"`
	DefaultPort      int    `yaml:"default_port"`
	WatchdogEnabled  bool   `yaml:"watchdog_enabled"`
	WatchdogInterval int    `yaml:"watchdog_interval_seconds"`
	AutoReconnect    bool   `yaml:"auto_reconnect"`
	ReconnectDelay   int    `yaml:"reconnect_delay_seconds"`
	MaxReconnect     int    `yaml:"max_reconnect_attempts"`
	PersistSession   bool   `yaml:"persist_session"`
	LogLevel         string `yaml:"log_level"`
	TUIRefreshRate   int    `yaml:"tui_refresh_rate_ms"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".netbridge")

	return &Config{
		DataDir:          dataDir,
		LogDir:           filepath.Join(dataDir, "logs"),
		DefaultMode:      "socks",
		DefaultPort:      10808,
		WatchdogEnabled:  true,
		WatchdogInterval: 30,
		AutoReconnect:    true,
		ReconnectDelay:   5,
		MaxReconnect:     10,
		PersistSession:   true,
		LogLevel:         "info",
		TUIRefreshRate:   2000,
	}
}
