package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kno-ai/kno/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
)

func listHandler(a *app.App) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := 10
		if l, ok := request.GetArguments()["limit"]; ok {
			if f, ok := l.(float64); ok && f > 0 && f <= 100 {
				limit = int(f)
			}
		}

		names, err := a.Vault.ListCaptures(limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing captures: %v", err)), nil
		}

		if len(names) == 0 {
			return mcp.NewToolResultText("No captures found."), nil
		}

		// Read meta for each to provide titles and dates.
		type entry struct {
			Name    string `json:"name"`
			Title   string `json:"title,omitempty"`
			Created string `json:"created"`
		}

		entries := make([]entry, 0, len(names))
		for _, name := range names {
			meta, _, err := a.Vault.ReadCapture(name)
			if err != nil {
				entries = append(entries, entry{Name: name})
				continue
			}
			entries = append(entries, entry{
				Name:    name,
				Title:   meta.Title,
				Created: meta.Created,
			})
		}

		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("encoding captures: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
}

func readHandler(a *app.App) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}

		meta, content, err := a.Vault.ReadCapture(name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("reading capture: %v", err)), nil
		}

		var b strings.Builder
		if meta.Title != "" {
			b.WriteString(fmt.Sprintf("Title: %s\n", meta.Title))
		}
		b.WriteString(fmt.Sprintf("Created: %s\n", meta.Created))
		b.WriteString(fmt.Sprintf("ID: %s\n", meta.ID))
		if len(meta.Meta) > 0 {
			metaJSON, err := json.Marshal(meta.Meta)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("encoding meta: %v", err)), nil
			}
			b.WriteString(fmt.Sprintf("Meta: %s\n", string(metaJSON)))
		}
		b.WriteString("\n---\n\n")
		b.WriteString(content)

		return mcp.NewToolResultText(b.String()), nil
	}
}
