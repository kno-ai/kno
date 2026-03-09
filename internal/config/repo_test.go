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
	content := `page = "cloud-infra"

[skill]
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
	if rc.Page != "cloud-infra" {
		t.Errorf("expected page = cloud-infra, got %q", rc.Page)
	}
	if rc.Skill.NudgeLevel == nil || *rc.Skill.NudgeLevel != "active" {
		t.Errorf("expected nudge_level = active, got %v", rc.Skill.NudgeLevel)
	}
}

func TestLoadRepoConfig_PageOnly(t *testing.T) {
	dir := t.TempDir()
	content := `page = "my-project"
`
	if err := os.WriteFile(filepath.Join(dir, ".kno"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rc, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc.Page != "my-project" {
		t.Errorf("expected page = my-project, got %q", rc.Page)
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
	if rc.Page != "" {
		t.Errorf("expected empty page, got %q", rc.Page)
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

	level := "active"
	rc := &RepoConfig{
		Page: "cloud-infra",
		Skill: RepoSkillConfig{
			NudgeLevel: &level,
		},
	}

	if err := SaveRepoConfig(dir, rc); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Page != "cloud-infra" {
		t.Errorf("round-trip: expected page = cloud-infra, got %q", loaded.Page)
	}
	if loaded.Skill.NudgeLevel == nil || *loaded.Skill.NudgeLevel != "active" {
		t.Error("round-trip: expected nudge_level = active")
	}
}

func TestSaveRepoConfig_PageOnlyOmitsSkill(t *testing.T) {
	dir := t.TempDir()

	rc := &RepoConfig{
		Page: "my-project",
	}

	if err := SaveRepoConfig(dir, rc); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Page != "my-project" {
		t.Errorf("expected page = my-project, got %q", loaded.Page)
	}
	if loaded.Skill.NudgeLevel != nil {
		t.Error("expected nudge_level to remain nil after round-trip")
	}
}
