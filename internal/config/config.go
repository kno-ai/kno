package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	VaultPath string        `json:"vault_path"`
	KnoSubdir string        `json:"kno_subdir"`
	Capture   CaptureConfig `json:"capture"`
}

type CaptureConfig struct {
	MaxBodyBytes int `json:"max_body_bytes"`
}

func DefaultConfig() Config {
	return Config{
		KnoSubdir: ".",
		Capture: CaptureConfig{
			MaxBodyBytes: 60000,
		},
	}
}

func Dir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving config dir: %w", err)
	}
	return filepath.Join(configDir, "kno"), nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (Config, error) {
	p, err := Path()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.KnoSubdir == "" {
		cfg.KnoSubdir = "."
	}
	if cfg.Capture.MaxBodyBytes == 0 {
		cfg.Capture.MaxBodyBytes = 60000
	}
	return cfg, nil
}

func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	return os.WriteFile(p, data, 0o644)
}
