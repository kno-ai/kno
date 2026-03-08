package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRepoConfig_NotFound(t *testing.T) {
	dir := t.TempDir()
	rc, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc != nil {
		t.Error("expected nil for missing .kno file")
	}
}

func TestLoadRepoConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `[skill]
auto_load_on_confirm = true
nudge_level = "active"
`
	if err := os.WriteFile(filepath.Join(dir, ".kno"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc == nil {
		t.Fatal("expected non-nil RepoConfig")
	}
	if rc.Skill.AutoLoadOnConfirm == nil || !*rc.Skill.AutoLoadOnConfirm {
		t.Error("expected auto_load_on_confirm = true")
	}
	if rc.Skill.NudgeLevel == nil || *rc.Skill.NudgeLevel != "active" {
		t.Errorf("expected nudge_level = active, got %v", rc.Skill.NudgeLevel)
	}
}

func TestLoadRepoConfig_PartialFields(t *testing.T) {
	dir := t.TempDir()
	// Only auto_load_on_confirm set, nudge_level omitted
	content := `[skill]
auto_load_on_confirm = false
`
	if err := os.WriteFile(filepath.Join(dir, ".kno"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc.Skill.AutoLoadOnConfirm == nil || *rc.Skill.AutoLoadOnConfirm {
		t.Error("expected auto_load_on_confirm = false")
	}
	if rc.Skill.NudgeLevel != nil {
		t.Error("expected nudge_level to be nil (unset)")
	}
}

func TestLoadRepoConfig_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".kno"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc == nil {
		t.Fatal("expected non-nil RepoConfig for empty file")
	}
	if rc.Skill.AutoLoadOnConfirm != nil {
		t.Error("expected nil auto_load_on_confirm")
	}
	if rc.Skill.NudgeLevel != nil {
		t.Error("expected nil nudge_level")
	}
}

func TestLoadRepoConfig_MalformedTOML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".kno"), []byte("not valid toml {{{}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadRepoConfig(dir)
	if err == nil {
		t.Error("expected error for malformed TOML")
	}
}

func TestSaveAndLoadRepoConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	tr := true
	level := "active"
	rc := &RepoConfig{
		Skill: RepoSkillConfig{
			AutoLoadOnConfirm: &tr,
			NudgeLevel:        &level,
		},
	}

	if err := SaveRepoConfig(dir, rc); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Skill.AutoLoadOnConfirm == nil || !*loaded.Skill.AutoLoadOnConfirm {
		t.Error("round-trip: expected auto_load_on_confirm = true")
	}
	if loaded.Skill.NudgeLevel == nil || *loaded.Skill.NudgeLevel != "active" {
		t.Error("round-trip: expected nudge_level = active")
	}
}

func TestSaveRepoConfig_NilFields(t *testing.T) {
	dir := t.TempDir()

	// Save with only auto_load_on_confirm set
	fa := false
	rc := &RepoConfig{
		Skill: RepoSkillConfig{
			AutoLoadOnConfirm: &fa,
		},
	}

	if err := SaveRepoConfig(dir, rc); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Skill.AutoLoadOnConfirm == nil || *loaded.Skill.AutoLoadOnConfirm {
		t.Error("expected auto_load_on_confirm = false")
	}
	if loaded.Skill.NudgeLevel != nil {
		t.Error("expected nudge_level to remain nil after round-trip")
	}
}
