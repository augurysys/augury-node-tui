package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	AuguryNodeRoot   string   `toml:"augury_node_root"`
	BinaryInstalled  bool     `toml:"binary_installed"`
	NixVerified      bool     `toml:"nix_verified"`
	SetupCompletedAt string   `toml:"setup_completed_at"`
	CompletedSteps   []string `toml:"completed_steps"`
	SkippedSteps     []string `toml:"skipped_steps"`
	CircleToken      string   `toml:"circle_token,omitempty"`
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "augury-node-tui", "config.toml"), nil
}

func Read(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Write(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
