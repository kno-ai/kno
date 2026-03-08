package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// RepoConfig represents the .kno file at a repository root.
// It holds project-specific skill settings that override vault defaults.
type RepoConfig struct {
	Skill RepoSkillConfig `toml:"skill"`
}

// RepoSkillConfig holds skill settings from a .kno file.
// Pointer fields distinguish "not set" (nil) from "set to zero value".
type RepoSkillConfig struct {
	NudgeLevel        *string `toml:"nudge_level"`
	AutoLoadOnConfirm *bool   `toml:"auto_load_on_confirm"`
}

// LoadRepoConfig reads a .kno file from the given directory.
// Returns nil without error if the file does not exist.
func LoadRepoConfig(dir string) (*RepoConfig, error) {
	p := filepath.Join(dir, ".kno")
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var rc RepoConfig
	if err := toml.Unmarshal(data, &rc); err != nil {
		return nil, err
	}
	return &rc, nil
}

// SaveRepoConfig writes a .kno file to the given directory.
func SaveRepoConfig(dir string, rc *RepoConfig) error {
	p := filepath.Join(dir, ".kno")
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(rc)
}
