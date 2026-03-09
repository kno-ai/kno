package mcp

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SessionContext holds session-scoped state detected at MCP server init.
// It is distinct from *app.App which holds vault-scoped state.
type SessionContext struct {
	Git              *GitContext // nil if no .git detected
	IsProjectVault   bool        // true if session is using a .kno/ project vault
	ProjectVaultPath string      // path to .kno/ directory, empty if personal vault
}

// GitContext holds information about the detected git repository.
type GitContext struct {
	RepoRoot string // absolute path to the repo root
	RepoName string // extracted repo name (from remote URL or directory name)
}

// DetectSessionContext detects git and project vault from the current working
// directory. Walks up from cwd to find .git, then checks for .kno/ directory
// (project vault). Returns a valid (possibly empty) SessionContext; never
// returns an error.
func DetectSessionContext() *SessionContext {
	sc := &SessionContext{}

	cwd, err := os.Getwd()
	if err != nil {
		return sc
	}

	projectRoot := cwd
	repoRoot := findRepoRoot(cwd)
	if repoRoot != "" {
		projectRoot = repoRoot
		sc.Git = &GitContext{
			RepoRoot: repoRoot,
			RepoName: extractRepoName(repoRoot),
		}
	}

	// Check for .kno/ project vault.
	knoDir := filepath.Join(projectRoot, ".kno")
	if info, err := os.Stat(knoDir); err == nil && info.IsDir() {
		sc.IsProjectVault = true
		sc.ProjectVaultPath = knoDir
	}

	return sc
}

// findRepoRoot walks up from dir looking for a .git directory.
// Returns the directory containing .git, or empty string if not found.
func findRepoRoot(dir string) string {
	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// extractRepoName extracts a human-readable repo name.
// Prefers the git remote origin URL; falls back to directory name.
func extractRepoName(repoRoot string) string {
	if name := repoNameFromRemote(repoRoot); name != "" {
		return name
	}
	return filepath.Base(repoRoot)
}

// repoNameFromRemote extracts the repo name from `git remote get-url origin`.
func repoNameFromRemote(repoRoot string) string {
	cmd := exec.Command("git", "-C", repoRoot, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return parseRepoNameFromURL(strings.TrimSpace(string(out)))
}

// parseRepoNameFromURL extracts the repo name from a git URL.
// Handles SSH (git@github.com:org/repo.git) and HTTPS (https://github.com/org/repo.git).
func parseRepoNameFromURL(url string) string {
	if url == "" {
		return ""
	}

	// Strip trailing .git
	url = strings.TrimSuffix(url, ".git")

	// SSH format: git@github.com:org/repo
	if i := strings.LastIndex(url, ":"); i >= 0 && !strings.Contains(url, "://") {
		url = url[i+1:]
	}

	// HTTPS format: https://github.com/org/repo
	if i := strings.LastIndex(url, "/"); i >= 0 {
		return url[i+1:]
	}

	return url
}
