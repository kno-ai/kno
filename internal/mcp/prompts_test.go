package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/skills/embedded"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestPagePromptHandler_GeneralTemplate(t *testing.T) {
	a := &app.App{
		Skills: embedded.New(),
	}
	sc := &SessionContext{} // no git context

	handler := pagePromptHandler(a, sc)
	result, err := handler(context.Background(), mcp.GetPromptRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Messages[0].Content.(mcp.TextContent).Text

	if !strings.Contains(text, "## Page Guidance Template") {
		t.Error("expected template section in output")
	}
	// General template should not contain developer-specific content
	if strings.Contains(text, "{{repo_name}}") {
		t.Error("unreplaced {{repo_name}} placeholder in general template")
	}
	// Should not contain developer-specific keywords from developer template
	if strings.Contains(text, "Known Issues") && strings.Contains(text, "debt") {
		t.Error("general template should not contain developer-specific sections")
	}
}

func TestPagePromptHandler_DeveloperTemplate(t *testing.T) {
	a := &app.App{
		Skills: embedded.New(),
	}
	sc := &SessionContext{
		Git: &GitContext{RepoRoot: "/tmp/test", RepoName: "cloud-infra"},
	}

	handler := pagePromptHandler(a, sc)
	result, err := handler(context.Background(), mcp.GetPromptRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Messages[0].Content.(mcp.TextContent).Text

	if !strings.Contains(text, "## Page Guidance Template") {
		t.Error("expected template section in output")
	}
	// Should have repo name substituted
	if strings.Contains(text, "{{repo_name}}") {
		t.Error("{{repo_name}} placeholder should be replaced")
	}
	if !strings.Contains(text, "cloud-infra") {
		t.Error("expected repo name 'cloud-infra' in developer template")
	}
}

func TestPagePromptHandler_WithAction(t *testing.T) {
	a := &app.App{
		Skills: embedded.New(),
	}
	sc := &SessionContext{}

	handler := pagePromptHandler(a, sc)
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
