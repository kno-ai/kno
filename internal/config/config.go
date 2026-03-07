package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Notes  NotesConfig  `toml:"notes" json:"notes"`
	Pages  PagesConfig  `toml:"pages" json:"pages"`
	Curate CurateConfig `toml:"curate" json:"curate"`
	Search SearchConfig `toml:"search" json:"search"`
}

type NotesConfig struct {
	MaxCount         int `toml:"max_count" json:"max_count"`
	DefaultListLimit int `toml:"default_list_limit" json:"default_list_limit"`
	SummaryMaxTokens int `toml:"summary_max_tokens" json:"summary_max_tokens"`
}

type PagesConfig struct {
	MaxContentTokens int `toml:"max_content_tokens" json:"max_content_tokens"`
}

type CurateConfig struct {
	MaxNotesPerRun int `toml:"max_notes_per_run" json:"max_notes_per_run"`
}

type SearchConfig struct {
	DefaultLimit int `toml:"default_limit" json:"default_limit"`
}

func DefaultConfig() Config {
	return Config{
		Notes: NotesConfig{
			MaxCount:         500,
			DefaultListLimit: 50,
			SummaryMaxTokens: 100,
		},
		Pages: PagesConfig{
			MaxContentTokens: 12000,
		},
		Curate: CurateConfig{
			MaxNotesPerRun: 50,
		},
		Search: SearchConfig{
			DefaultLimit: 10,
		},
	}
}

func ConfigPath(vaultPath string) string {
	return filepath.Join(vaultPath, "config.toml")
}

// Load reads config.toml from inside the vault directory.
// Missing keys get default values.
func Load(vaultPath string) (Config, error) {
	cfg := DefaultConfig()
	p := ConfigPath(vaultPath)

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	applyDefaults(&cfg)
	return cfg, nil
}

// Save writes config.toml inside the vault directory.
func Save(vaultPath string, cfg Config) error {
	p := ConfigPath(vaultPath)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}

func applyDefaults(cfg *Config) {
	d := DefaultConfig()
	if cfg.Notes.MaxCount == 0 {
		cfg.Notes.MaxCount = d.Notes.MaxCount
	}
	if cfg.Notes.DefaultListLimit == 0 {
		cfg.Notes.DefaultListLimit = d.Notes.DefaultListLimit
	}
	if cfg.Notes.SummaryMaxTokens == 0 {
		cfg.Notes.SummaryMaxTokens = d.Notes.SummaryMaxTokens
	}
	if cfg.Pages.MaxContentTokens == 0 {
		cfg.Pages.MaxContentTokens = d.Pages.MaxContentTokens
	}
	if cfg.Curate.MaxNotesPerRun == 0 {
		cfg.Curate.MaxNotesPerRun = d.Curate.MaxNotesPerRun
	}
	if cfg.Search.DefaultLimit == 0 {
		cfg.Search.DefaultLimit = d.Search.DefaultLimit
	}
}
