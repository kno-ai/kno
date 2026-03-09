package mcp

import (
	"context"
	"encoding/json"
	"fmt"
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
		mcp.WithDescription("Set a vault option. Writes to the vault's config.toml. For project vaults (.kno/), this is .kno/config.toml. Requires a project vault for project options like page."),
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

		// Build skill config for the session.
		skillInfo := map[string]any{
			"nudge_level": a.Config.Skill.NudgeLevel,
		}
		if pps := a.Config.Skill.PromptProjectSetup; pps != nil {
			skillInfo["prompt_project_setup"] = *pps
		}

		vaultType := "personal"
		if sc.IsProjectVault {
			vaultType = "project"
		}

		out := map[string]any{
			"vault_path": absPath,
			"vault_type": vaultType,
			"notes": map[string]int{
				"total":     total,
				"max_count": a.Config.Notes.MaxCount,
				"remaining": remaining,
				"curated":   curated,
				"uncurated": total - curated,
			},
			"pages":  pageInfos,
			"config": a.Config,
			"skill":  skillInfo,
		}

		// Add project binding if config has a page set.
		if page := a.Config.Page; page != "" {
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

// projectOptionKeys are written to the vault's config.toml.
// For project options like page and nudge_level, a project vault is required.
var projectOptionKeys = map[string]bool{
	"page":        true,
	"nudge_level": true,
}

// vaultOptionKeys are written to config.toml in any vault.
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

		// Route to vault config.
		if vaultOptionKeys[key] {
			return setVaultOption(a, key, value)
		}
		if projectOptionKeys[key] {
			return setProjectOption(sc, a, key, value)
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

func setProjectOption(sc *SessionContext, a *app.App, key, value string) (*mcp.CallToolResult, error) {
	// Project options require a project vault.
	if !sc.IsProjectVault {
		return mcp.NewToolResultError(fmt.Sprintf("cannot set %q: no project vault found. Create a project vault with 'kno setup --vault .kno' in your project root.", key)), nil
	}

	switch key {
	case "page":
		a.Config.Page = value
	case "nudge_level":
		if !config.ValidNudgeLevel(value) {
			return mcp.NewToolResultError(fmt.Sprintf("invalid nudge_level %q — must be off, light, or active", value)), nil
		}
		a.Config.Skill.NudgeLevel = value
	}

	if err := config.Save(a.VaultPath, a.Config); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("writing config: %v", err)), nil
	}

	result := map[string]any{
		"key":   key,
		"value": value,
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
