package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kno-ai/kno/internal"
	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerVaultTools(s *server.MCPServer, a *app.App, sc *SessionContext) {
	s.AddTool(mcp.NewTool("kno_vault_status",
		mcp.WithDescription("Show vault status including note counts, page list, and config."),
	), vaultStatusHandler(a, sc))

	s.AddTool(mcp.NewTool("kno_version",
		mcp.WithDescription("Return the kno version."),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(internal.Version), nil
	})

	s.AddTool(mcp.NewTool("kno_set_option",
		mcp.WithDescription("Set a project option. Writes to .kno in the project root (repo root if git detected, otherwise cwd)."),
		mcp.WithString("key", mcp.Required(), mcp.Description("Option key to set (e.g. page, nudge_level).")),
		mcp.WithString("value", mcp.Required(), mcp.Description("Value to set.")),
	), setOptionHandler(a, sc))
}

func vaultStatusHandler(a *app.App, sc *SessionContext) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		absPath, _ := filepath.Abs(a.VaultPath)

		total, err := a.Vault.CountNotes()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("counting: %v", err)), nil
		}

		allNotes, err := a.Vault.ListNotes(0)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing: %v", err)), nil
		}
		curated := 0
		for _, c := range allNotes {
			if c.Metadata.Has("curated_at") {
				curated++
			}
		}

		pages, err := a.Vault.ListPages()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing pages: %v", err)), nil
		}

		type pageInfo struct {
			ID       string         `json:"id"`
			Name     string         `json:"name"`
			Metadata map[string]any `json:"metadata"`
		}
		var pageInfos []pageInfo
		for _, p := range pages {
			pageInfos = append(pageInfos, pageInfo{
				ID:       p.ID,
				Name:     p.Name,
				Metadata: metaMapForMCP(p.Metadata),
			})
		}
		if pageInfos == nil {
			pageInfos = []pageInfo{}
		}

		remaining := a.Config.Notes.MaxCount - total
		if remaining < 0 {
			remaining = 0
		}

		// Build merged skill config for the session.
		mergedSkill := map[string]any{
			"nudge_level": sc.MergedNudgeLevel(a.Config.Skill.NudgeLevel),
		}
		if pps := a.Config.Skill.PromptProjectSetup; pps != nil {
			mergedSkill["prompt_project_setup"] = *pps
		}

		out := map[string]any{
			"vault_path": absPath,
			"notes": map[string]int{
				"total":     total,
				"max_count": a.Config.Notes.MaxCount,
				"remaining": remaining,
				"curated":   curated,
				"uncurated": total - curated,
			},
			"pages":  pageInfos,
			"config": a.Config,
			"skill":  mergedSkill,
		}

		// Add project binding if .kno has a page set.
		if page := sc.BoundPage(); page != "" {
			out["project"] = map[string]string{
				"page": page,
			}
		}

		// Add git context if detected.
		if sc.Git != nil {
			out["git"] = map[string]string{
				"repo_name": sc.Git.RepoName,
				"repo_root": sc.Git.RepoRoot,
			}
		}

		data, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

// projectOptionKeys are written to .kno in the project root.
var projectOptionKeys = map[string]bool{
	"page":        true,
	"nudge_level": true,
}

// vaultOptionKeys are written to config.toml in the vault.
var vaultOptionKeys = map[string]bool{
	"prompt_project_setup": true,
}

func setOptionHandler(a *app.App, sc *SessionContext) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		key, err := req.RequireString("key")
		if err != nil {
			return mcp.NewToolResultError("key is required"), nil
		}

		value, err := req.RequireString("value")
		if err != nil {
			return mcp.NewToolResultError("value is required"), nil
		}

		// Route to vault config or project config.
		if vaultOptionKeys[key] {
			return setVaultOption(a, key, value)
		}
		if projectOptionKeys[key] {
			return setProjectOption(sc, key, value)
		}

		return mcp.NewToolResultError(fmt.Sprintf("unknown option key %q — allowed keys: page, nudge_level, prompt_project_setup", key)), nil
	}
}

func setVaultOption(a *app.App, key, value string) (*mcp.CallToolResult, error) {
	switch key {
	case "prompt_project_setup":
		bval, ok := parseBool(value)
		if !ok {
			return mcp.NewToolResultError("prompt_project_setup value must be true or false"), nil
		}
		a.Config.Skill.PromptProjectSetup = &bval
	}

	if err := config.Save(a.VaultPath, a.Config); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("writing config: %v", err)), nil
	}

	result := map[string]any{"key": key, "value": value}
	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func setProjectOption(sc *SessionContext, key, value string) (*mcp.CallToolResult, error) {
	// Determine project root: repo root if git detected, otherwise cwd.
	var projectRoot string
	var err error
	if sc.Git != nil {
		projectRoot = sc.Git.RepoRoot
	} else {
		projectRoot, err = os.Getwd()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot determine project root: %v", err)), nil
		}
	}

	// Load existing .kno or create new one.
	rc, _ := config.LoadRepoConfig(projectRoot)
	created := rc == nil
	if created {
		rc = &config.RepoConfig{}
	}

	switch key {
	case "page":
		rc.Page = value
	case "nudge_level":
		if !config.ValidNudgeLevel(value) {
			return mcp.NewToolResultError(fmt.Sprintf("invalid nudge_level %q — must be off, light, or active", value)), nil
		}
		rc.Skill.NudgeLevel = &value
	}

	if err := config.SaveRepoConfig(projectRoot, rc); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("writing .kno: %v", err)), nil
	}

	// Update in-memory session context so the setting takes effect immediately.
	sc.RepoConfig = rc

	result := map[string]any{
		"key":     key,
		"value":   value,
		"created": created,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func parseBool(s string) (bool, bool) {
	switch s {
	case "true":
		return true, true
	case "false":
		return false, true
	}
	return false, false
}
