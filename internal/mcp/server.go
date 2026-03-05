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
	s := server.NewMCPServer(
		"kno",
		internal.Version,
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	registerTools(s, a)
	registerPrompts(s, a)

	return server.ServeStdio(s)
}

// ServeUnconfigured starts a degraded MCP server that returns setup
// instructions for every tool call. This ensures Claude can relay the
// problem to the user instead of showing an opaque connection error.
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

	s.AddTool(mcp.NewTool("kno_capture",
		mcp.WithDescription("Write a structured capture note into the configured vault."),
		mcp.WithString("markdown", mcp.Required(), mcp.Description("The structured capture content (markdown).")),
	), handler)
	s.AddTool(mcp.NewTool("kno_list",
		mcp.WithDescription("List recent captures from the knowledge vault."),
	), handler)
	s.AddTool(mcp.NewTool("kno_read",
		mcp.WithDescription("Read a specific capture's content and metadata."),
		mcp.WithString("name", mcp.Required(), mcp.Description("The capture directory name.")),
	), handler)
	s.AddTool(mcp.NewTool("kno_version",
		mcp.WithDescription("Return the kno version."),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(internal.Version + " (unconfigured)"), nil
	})

	return server.ServeStdio(s)
}

func registerTools(s *server.MCPServer, a *app.App) {
	s.AddTool(mcp.NewTool("kno_capture",
		mcp.WithDescription("Write a structured capture note into the configured vault."),
		mcp.WithString("markdown",
			mcp.Required(),
			mcp.Description("The structured capture content (markdown)."),
		),
		mcp.WithString("title",
			mcp.Description("Optional title for the capture."),
		),
		mcp.WithObject("meta",
			mcp.Description("Optional metadata key-value pairs (e.g. {\"topic\": \"aws/sqs\", \"project\": \"platform\"})."),
		),
	), captureHandler(a))

	s.AddTool(mcp.NewTool("kno_list",
		mcp.WithDescription("List recent captures from the knowledge vault."),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of captures to return (default: 10)."),
		),
	), listHandler(a))

	s.AddTool(mcp.NewTool("kno_version",
		mcp.WithDescription("Return the kno version."),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(internal.Version), nil
	})

	s.AddTool(mcp.NewTool("kno_read",
		mcp.WithDescription("Read a specific capture's content and metadata."),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The capture directory name (from kno_list)."),
		),
	), readHandler(a))
}
