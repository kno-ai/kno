package mcp

import (
	"context"
	"fmt"

	"github.com/kno-ai/kno/internal"
	"github.com/kno-ai/kno/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Serve starts the MCP server over stdio.
func Serve(a *app.App) error {
	opts := []server.ServerOption{
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	}

	if instructions := awarenessInstructions(a); instructions != "" {
		opts = append(opts, server.WithInstructions(instructions))
	}

	s := server.NewMCPServer("kno", internal.Version, opts...)

	registerNoteTools(s, a)
	registerPageTools(s, a)
	registerVaultTools(s, a)
	registerPrompts(s, a)

	return server.ServeStdio(s)
}

// awarenessInstructions returns standing instructions based on the nudge level.
// Returns empty string for "off".
func awarenessInstructions(a *app.App) string {
	level := a.Config.Nudges.Level
	if level == "off" {
		return ""
	}

	skill, err := a.Skills.Get("awareness.md")
	if err != nil {
		return ""
	}

	if level == "active" {
		return skill
	}

	// "light" — append a restraint note
	return skill + `

## Nudge level: light

You are in light mode. Be more conservative with nudges:

- Only nudge for knowledge checkpoints with very high signal — decisions
  with clear tradeoffs, hard-won debugging insights, or things the user
  explicitly called out as important.
- Prefer fewer nudges. When in doubt, stay quiet.
- Topic awareness (offering to load vault knowledge) is still active.
- Session-end capture nudges are still active.
`
}

// ServeUnconfigured starts a degraded MCP server with setup instructions.
func ServeUnconfigured(cause error) error {
	s := server.NewMCPServer(
		"kno",
		internal.Version,
		server.WithToolCapabilities(true),
	)

	msg := fmt.Sprintf(
		"kno is not configured: %v\n\nRun 'kno setup' in your terminal to get started.",
		cause,
	)
	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultError(msg), nil
	}

	// Register stubs so Claude can relay the setup message.
	for _, name := range []string{
		"kno_note_create", "kno_note_list", "kno_note_show",
		"kno_note_update", "kno_note_delete", "kno_note_search",
		"kno_page_create", "kno_page_list", "kno_page_show",
		"kno_page_update", "kno_page_rename", "kno_page_delete", "kno_page_search",
		"kno_vault_status",
	} {
		s.AddTool(mcp.NewTool(name, mcp.WithDescription("kno is not configured.")), handler)
	}

	s.AddTool(mcp.NewTool("kno_version",
		mcp.WithDescription("Return the kno version."),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(internal.Version + " (unconfigured)"), nil
	})

	return server.ServeStdio(s)
}
