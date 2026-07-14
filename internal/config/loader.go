package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load() (*Config, error) {
	cfg := DefaultConfig()

	configPath := filepath.Join(cfg.DataDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	configPath := filepath.Join(cfg.DataDir, "config.yaml")
	return os.WriteFile(configPath, data, 0600)
}
