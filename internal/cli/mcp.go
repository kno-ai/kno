package cli

import (
	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/config"
	mcpserver "github.com/kno-ai/kno/internal/mcp"
	"github.com/spf13/cobra"
)

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run the MCP server for Claude Desktop",
		Long:  "Starts an MCP server over stdio, exposing kno tools to Claude Desktop.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return mcpserver.ServeUnconfigured(err)
			}

			a, err := app.FromConfig(cfg)
			if err != nil {
				return mcpserver.ServeUnconfigured(err)
			}

			return mcpserver.Serve(a)
		},
	}

	return cmd
}
