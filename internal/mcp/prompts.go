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
		Name:        "kno.capture",
		Description: "Capture the current session into your knowledge vault.",
	}, capturePromptHandler(a))

	s.AddPrompt(mcp.Prompt{
		Name:        "kno.load",
		Description: "Load a previous capture into this conversation for context.",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "hint",
				Description: "Optional topic or keyword to help find the right capture (e.g. 'dogs', 'SQS debugging')",
				Required:    false,
			},
		},
	}, loadPromptHandler(a))
}

func capturePromptHandler(a *app.App) server.PromptHandlerFunc {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		skill, err := a.Skills.Get("capture/default.md")
		if err != nil {
			return nil, fmt.Errorf("loading capture skill: %w", err)
		}

		return &mcp.GetPromptResult{
			Description: "Capture the current session into your knowledge vault.",
			Messages: []mcp.PromptMessage{
				mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(skill)),
			},
		}, nil
	}
}

const loadInstructionsBase = `Load a previous capture into this conversation.

First, call kno_list to get recent captures.`

const loadInstructionsNoHint = `
This is a two-turn interaction. You MUST wait for the user to choose before reading any capture.

Turn 1 (now):
- Display the captures as a numbered list with title and date
- Ask: "Which capture would you like to load?"
- Then STOP. Do not call kno_read. Do not read any capture content. End your response and wait.

Turn 2 (after user replies):
- Call kno_read with the capture the user selected
- Summarize the key points
- Confirm the context is loaded

You must never call kno_read in the same turn as kno_list.`

const loadInstructionsWithHint = `
The user is looking for: %s

After listing captures, decide based on confidence:
- If exactly one capture clearly matches the hint (by title, topic, or metadata), load it immediately with kno_read. No need to ask.
- If a few captures could match, list the top matches and ask which one.
- If nothing matches, show all recent captures and let the user pick.

Prefer recent captures when confidence is similar between matches.`

func loadPromptHandler(a *app.App) server.PromptHandlerFunc {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		instructions := loadInstructionsBase
		if hint, ok := request.Params.Arguments["hint"]; ok && hint != "" {
			instructions += fmt.Sprintf(loadInstructionsWithHint, hint)
		} else {
			instructions += loadInstructionsNoHint
		}

		return &mcp.GetPromptResult{
			Description: "Load a previous capture into this conversation for context.",
			Messages: []mcp.PromptMessage{
				mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(instructions)),
			},
		}, nil
	}
}
