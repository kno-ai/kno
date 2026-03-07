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
		Name:        "kno.save",
		Description: "Save the current session to your knowledge vault.",
	}, savePromptHandler(a))

	s.AddPrompt(mcp.Prompt{
		Name:        "kno.distill",
		Description: "Synthesize undistilled notes into page documents.",
	}, distillPromptHandler(a))

	s.AddPrompt(mcp.Prompt{
		Name:        "kno.load",
		Description: "Load relevant knowledge into this conversation.",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "hint",
				Description: "Optional page or keyword to guide what to load (e.g. 'aws infrastructure')",
				Required:    false,
			},
		},
	}, loadPromptHandler(a))

	s.AddPrompt(mcp.Prompt{
		Name:        "kno.page",
		Description: "Create or manage knowledge pages.",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "action",
				Description: "Action: new, list, or edit",
				Required:    false,
			},
		},
	}, pagePromptHandler(a))

	s.AddPrompt(mcp.Prompt{
		Name:        "kno.status",
		Description: "Check vault status.",
	}, statusPromptHandler(a))
}

func savePromptHandler(a *app.App) server.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		skill, err := a.Skills.Get("save.md")
		if err != nil {
			return nil, fmt.Errorf("loading save skill: %w", err)
		}
		return &mcp.GetPromptResult{
			Description: "Save the current session to your knowledge vault.",
			Messages:    []mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(skill))},
		}, nil
	}
}

func distillPromptHandler(a *app.App) server.PromptHandlerFunc {
	return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		skill, err := a.Skills.Get("distill.md")
		if err != nil {
			return nil, fmt.Errorf("loading distill skill: %w", err)
		}
		return &mcp.GetPromptResult{
			Description: "Synthesize undistilled notes into page documents.",
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
			Description: "Load relevant knowledge into this conversation.",
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

		return &mcp.GetPromptResult{
			Description: "Create or manage knowledge pages.",
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
			Description: "Check vault status.",
			Messages:    []mcp.PromptMessage{mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(skill))},
		}, nil
	}
}
