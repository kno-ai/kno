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
		mcp.WithDescription("Set a project-specific skill option. Writes to .kno in the repo root. Requires git context."),
		mcp.WithString("key", mcp.Required(), mcp.Description("Option key to set (e.g. auto_load_on_confirm).")),
		mcp.WithBoolean("value", mcp.Required(), mcp.Description("Value to set.")),
	), setOptionHandler(sc))
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
		if aloc := sc.AutoLoadOnConfirm(); aloc != nil {
			mergedSkill["auto_load_on_confirm"] = *aloc
		} else {
			mergedSkill["auto_load_on_confirm"] = nil
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

// allowedOptionKeys defines which keys kno_set_option can write.
var allowedOptionKeys = map[string]bool{
	"auto_load_on_confirm": true,
}

func setOptionHandler(sc *SessionContext) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if sc.Git == nil {
			return mcp.NewToolResultError("No git context detected. kno_set_option only works inside a git repository."), nil
		}

		key, err := req.RequireString("key")
		if err != nil {
			return mcp.NewToolResultError("key is required"), nil
		}

		if !allowedOptionKeys[key] {
			return mcp.NewToolResultError(fmt.Sprintf("unknown option key %q — allowed keys: auto_load_on_confirm", key)), nil
		}

		// Parse the value from arguments.
		rawValue, ok := req.GetArguments()["value"]
		if !ok {
			return mcp.NewToolResultError("value is required"), nil
		}

		repoRoot := sc.Git.RepoRoot

		// Load existing .kno or create new one.
		rc, _ := config.LoadRepoConfig(repoRoot)
		created := rc == nil
		if created {
			rc = &config.RepoConfig{}
		}

		switch key {
		case "auto_load_on_confirm":
			bval, ok := rawValue.(bool)
			if !ok {
				return mcp.NewToolResultError("auto_load_on_confirm value must be a boolean"), nil
			}
			rc.Skill.AutoLoadOnConfirm = &bval
		}

		if err := config.SaveRepoConfig(repoRoot, rc); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("writing .kno: %v", err)), nil
		}

		// Update in-memory session context so the setting takes effect immediately.
		sc.RepoConfig = rc

		result := map[string]any{
			"key":     key,
			"value":   rawValue,
			"created": created,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}
