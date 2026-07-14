package config

import "fmt"

type Migration struct {
	FromVersion int
	ToVersion   int
	Apply       func(cfg *Config) error
}

var migrations = []Migration{}

func Migrate(cfg *Config, currentVersion, targetVersion int) error {
	for _, m := range migrations {
		if m.FromVersion >= currentVersion && m.ToVersion <= targetVersion {
			if err := m.Apply(cfg); err != nil {
				return fmt.Errorf("migration %d->%d failed: %w", m.FromVersion, m.ToVersion, err)
			}
		}
	}
	return nil
}
