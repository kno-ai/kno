package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Notes   NotesConfig   `toml:"notes" json:"notes"`
	Pages   PagesConfig   `toml:"pages" json:"pages"`
	Curate  CurateConfig  `toml:"curate" json:"curate"`
	Search  SearchConfig  `toml:"search" json:"search"`
	Skill   SkillConfig   `toml:"skill" json:"skill"`
	Publish PublishConfig `toml:"publish" json:"publish"`
}

type PublishTarget struct {
	Path   string `toml:"path" json:"path"`
	Format string `toml:"format" json:"format"`
}

type PublishConfig struct {
	Targets []PublishTarget `toml:"targets" json:"targets"`
}

// ValidPublishFormat reports whether format is a recognized publish format.
func ValidPublishFormat(format string) bool {
	switch format {
	case "markdown", "frontmatter":
		return true
	}
	return false
}

type NotesConfig struct {
	MaxCount         int `toml:"max_count" json:"max_count"`
	DefaultListLimit int `toml:"default_list_limit" json:"default_list_limit"`
	SummaryMaxTokens int `toml:"summary_max_tokens" json:"summary_max_tokens"`
	MaxContentTokens int `toml:"max_content_tokens" json:"max_content_tokens"`
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

type SkillConfig struct {
	NudgeLevel         string `toml:"nudge_level" json:"nudge_level"`
	PromptProjectSetup *bool  `toml:"prompt_project_setup,omitempty" json:"prompt_project_setup"`
}

// ValidNudgeLevel reports whether level is a recognized nudge setting.
func ValidNudgeLevel(level string) bool {
	switch level {
	case "off", "light", "active":
		return true
	}
	return false
}

func DefaultConfig() Config {
	return Config{
		Notes: NotesConfig{
			MaxCount:         500,
			DefaultListLimit: 50,
			SummaryMaxTokens: 100,
			MaxContentTokens: 3000,
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
		Skill: SkillConfig{
			NudgeLevel: "active",
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
	if cfg.Notes.MaxContentTokens == 0 {
		cfg.Notes.MaxContentTokens = d.Notes.MaxContentTokens
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
	if cfg.Skill.NudgeLevel == "" {
		cfg.Skill.NudgeLevel = d.Skill.NudgeLevel
	}
}
