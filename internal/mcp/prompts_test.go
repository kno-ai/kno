package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/skills/embedded"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestPagePromptHandler_IncludesTemplate(t *testing.T) {
	a := &app.App{
		Skills: embedded.New(),
	}

	handler := pagePromptHandler(a)
	result, err := handler(context.Background(), mcp.GetPromptRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Messages[0].Content.(mcp.TextContent).Text

	if !strings.Contains(text, "## Page Guidance Template") {
		t.Error("expected template section in output")
	}
	if !strings.Contains(text, "CAPTURE PRIORITIES") {
		t.Error("expected project template content")
	}
}

func TestPagePromptHandler_WithAction(t *testing.T) {
	a := &app.App{
		Skills: embedded.New(),
	}

	handler := pagePromptHandler(a)
	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{"action": "new"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Messages[0].Content.(mcp.TextContent).Text
	if !strings.Contains(text, "Requested action: new") {
		t.Error("expected action to be appended")
	}
}
