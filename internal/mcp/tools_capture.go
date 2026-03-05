package mcp

import (
	"context"
	"fmt"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/capture"
	"github.com/mark3labs/mcp-go/mcp"
)

func captureHandler(a *app.App) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		markdown, err := request.RequireString("markdown")
		if err != nil {
			return mcp.NewToolResultError("markdown is required"), nil
		}

		title := request.GetString("title", "")

		// Extract meta as map[string]string from the arguments.
		var meta map[string]string
		args := request.GetArguments()
		if raw, ok := args["meta"]; ok {
			if m, ok := raw.(map[string]any); ok {
				meta = make(map[string]string, len(m))
				for k, v := range m {
					if s, ok := v.(string); ok {
						meta[k] = s
					}
				}
			}
		}

		result, err := a.Capture.Create(capture.CreateParams{
			Title:      title,
			BodyMD:     markdown,
			SourceKind: "claude_desktop",
			SourceTool: "claude",
			Meta:       meta,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("capture failed: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(
			"Captured successfully.\npath: %s\nid: %s\ncreated: %s",
			result.Path,
			result.ID,
			result.Created.Format("2006-01-02T15:04:05-07:00"),
		)), nil
	}
}
