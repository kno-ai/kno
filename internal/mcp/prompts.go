package mcp

import (
	"context"
	"fmt"

	"github.com/kno-ai/kno/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerPrompts(s *server.MCPServer, a *app.App) {
	s.AddPrompt(mcp.Prompt{
		Name:        "start",
		Description: "Start here — show pages, offer to load",
	}, startPromptHandler(a))

	s.AddPrompt(mcp.Prompt{
		Name:        "capture",
		Description: "Save this session's insights to your vault",
	}, capturePromptHandler(a))

	s.AddPrompt(mcp.Prompt{
		Name:        "curate",
		Description: "Turn captured notes into lasting pages",
	}, curatePromptHandler(a))

	s.AddPrompt(mcp.Prompt{
		Name:        "load",
		Description: "Load a specific page or topic",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "hint",
				Description: "Page or keyword (e.g. 'aws infrastructure')",
				Required:    false,
			},
		},
	}, loadPromptHandler(a))

	s.AddPrompt(mcp.Prompt{
		Name:        "page",
		Description: "Create or edit a knowledge page",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "action",
				Description: "Action: new, list, or edit",
				Required:    false,
			},
		},
	}, pagePromptHandler(a))

	s.AddPrompt(mcp.Prompt{
		Name:        "status",
		Description: "Notes, pages, and vault health",
	}, statusPromptHandler(a))
}

func startPromptHandler(a *app.App) server.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		skill, err := a.Skills.Get("start.md")
		if err != nil {
			return nil, fmt.Errorf("loading start skill: %w", err)
		}
		return &mcp.GetPromptResult{
			Description: "Start here — show pages, offer to load",
			Messages:    []mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(skill))},
		}, nil
	}
}

func capturePromptHandler(a *app.App) server.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		skill, err := a.Skills.Get("capture.md")
		if err != nil {
			return nil, fmt.Errorf("loading capture skill: %w", err)
		}

		instructions := skill

		// Inject page guidance template so capture can create pages with
		// guidance on first save (same template the page skill uses).
		if tmpl, err := a.Skills.Get("templates/project-page.md"); err == nil {
			instructions += "\n\n## Page Guidance Template\n\nWhen creating a page on first save, prepend this guidance (wrapped in `<!-- Guidance -->`) before the curated content:\n\n```\n" + tmpl + "```"
		}

		return &mcp.GetPromptResult{
			Description: "Save this session's insights to your vault",
			Messages:    []mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(instructions))},
		}, nil
	}
}

func curatePromptHandler(a *app.App) server.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		skill, err := a.Skills.Get("curate.md")
		if err != nil {
			return nil, fmt.Errorf("loading curate skill: %w", err)
		}
		return &mcp.GetPromptResult{
			Description: "Turn captured notes into lasting pages",
			Messages:    []mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(skill))},
		}, nil
	}
}

func loadPromptHandler(a *app.App) server.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		skill, err := a.Skills.Get("load.md")
		if err != nil {
			return nil, fmt.Errorf("loading load skill: %w", err)
		}

		instructions := skill
		if hint, ok := req.Params.Arguments["hint"]; ok && hint != "" {
			instructions += fmt.Sprintf("\n\nThe user is looking for: %s", hint)
		}

		return &mcp.GetPromptResult{
			Description: "Load a specific page or topic",
			Messages:    []mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(instructions))},
		}, nil
	}
}

func pagePromptHandler(a *app.App) server.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		skill, err := a.Skills.Get("page.md")
		if err != nil {
			return nil, fmt.Errorf("loading page skill: %w", err)
		}

		instructions := skill
		if action, ok := req.Params.Arguments["action"]; ok && action != "" {
			instructions += fmt.Sprintf("\n\nRequested action: %s", action)
		}

		// Inject page guidance template.
		if tmpl, err := a.Skills.Get("templates/project-page.md"); err == nil {
			instructions += "\n\n## Page Guidance Template\n\nWhen creating a new page, use this as the default guidance (the user can adjust):\n\n```\n" + tmpl + "```"
		}

		return &mcp.GetPromptResult{
			Description: "Create or edit a knowledge page",
			Messages:    []mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(instructions))},
		}, nil
	}
}

func statusPromptHandler(a *app.App) server.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		skill, err := a.Skills.Get("status.md")
		if err != nil {
			return nil, fmt.Errorf("loading status skill: %w", err)
		}
		return &mcp.GetPromptResult{
			Description: "Notes, pages, and vault health",
			Messages:    []mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(skill))},
		}, nil
	}
}
