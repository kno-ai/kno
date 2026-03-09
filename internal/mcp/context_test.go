package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kno-ai/kno/internal/config"
)

func TestParseRepoNameFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"git@github.com:org/cloud-infra.git", "cloud-infra"},
		{"https://github.com/org/cloud-infra.git", "cloud-infra"},
		{"git@github.com:org/cloud-infra", "cloud-infra"},
		{"https://github.com/org/cloud-infra", "cloud-infra"},
		{"ssh://git@github.com/org/repo.git", "repo"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := parseRepoNameFromURL(tt.url)
			if got != tt.want {
				t.Errorf("parseRepoNameFromURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestMergedNudgeLevel(t *testing.T) {
	// No repo config — use vault default.
	sc := &SessionContext{}
	if got := sc.MergedNudgeLevel("light"); got != "light" {
		t.Errorf("expected light, got %q", got)
	}

	// Repo config overrides.
	active := "active"
	sc = &SessionContext{
		RepoConfig: &config.RepoConfig{
			Skill: config.RepoSkillConfig{NudgeLevel: &active},
		},
	}
	if got := sc.MergedNudgeLevel("light"); got != "active" {
		t.Errorf("expected active, got %q", got)
	}

	// Invalid repo config value — fall back to vault default.
	invalid := "bogus"
	sc = &SessionContext{
		RepoConfig: &config.RepoConfig{
			Skill: config.RepoSkillConfig{NudgeLevel: &invalid},
		},
	}
	if got := sc.MergedNudgeLevel("light"); got != "light" {
		t.Errorf("expected light fallback, got %q", got)
	}
}

func TestBoundPage(t *testing.T) {
	// No repo config — empty.
	sc := &SessionContext{}
	if got := sc.BoundPage(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}

	// Page set.
	sc = &SessionContext{
		RepoConfig: &config.RepoConfig{Page: "cloud-infra"},
	}
	if got := sc.BoundPage(); got != "cloud-infra" {
		t.Errorf("expected cloud-infra, got %q", got)
	}

	// Empty page.
	sc = &SessionContext{
		RepoConfig: &config.RepoConfig{},
	}
	if got := sc.BoundPage(); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestFindRepoRoot(t *testing.T) {
	// Create a temp dir with nested structure: root/.git, root/sub/deep
	root := t.TempDir()
	gitDir := filepath.Join(root, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	deepDir := filepath.Join(root, "sub", "deep")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}

	// From deep nested dir, should find root
	if got := findRepoRoot(deepDir); got != root {
		t.Errorf("findRepoRoot(%q) = %q, want %q", deepDir, got, root)
	}

	// From root itself
	if got := findRepoRoot(root); got != root {
		t.Errorf("findRepoRoot(%q) = %q, want %q", root, got, root)
	}

	// From a dir with no .git anywhere above
	noGitDir := t.TempDir()
	if got := findRepoRoot(noGitDir); got != "" {
		t.Errorf("findRepoRoot(%q) = %q, want empty", noGitDir, got)
	}
}

func TestExtractRepoName_FallbackToDir(t *testing.T) {
	// A directory with no git remote — should fall back to directory name.
	dir := t.TempDir()
	got := extractRepoName(dir)
	want := filepath.Base(dir)
	if got != want {
		t.Errorf("extractRepoName(%q) = %q, want %q", dir, got, want)
	}
}
