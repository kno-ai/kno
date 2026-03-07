package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/kno-ai/kno/internal"
	"github.com/kno-ai/kno/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerVaultTools(s *server.MCPServer, a *app.App) {
	s.AddTool(mcp.NewTool("kno_vault_status",
		mcp.WithDescription("Show vault status including note counts, page list, and config."),
	), vaultStatusHandler(a))

	s.AddTool(mcp.NewTool("kno_version",
		mcp.WithDescription("Return the kno version."),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(internal.Version), nil
	})
}

func vaultStatusHandler(a *app.App) server.ToolHandlerFunc {
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
		distilled := 0
		for _, c := range allNotes {
			if c.Metadata.Has("distilled_at") {
				distilled++
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

		out := map[string]any{
			"vault_path": absPath,
			"notes": map[string]int{
				"total":       total,
				"max_count":   a.Config.Notes.MaxCount,
				"remaining":   remaining,
				"distilled":   distilled,
				"undistilled": total - distilled,
			},
			"pages":  pageInfos,
			"config": a.Config,
		}

		data, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}
