package cli

import (
	"github.com/kno-ai/kno/internal/app"
	mcpserver "github.com/kno-ai/kno/internal/mcp"
	"github.com/spf13/cobra"
)

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run the MCP server over stdio",
		RunE: func(cmd *cobra.Command, args []string) error {
			vaultPath := resolveVault(cmd)
			if vaultPath == "" {
				return mcpserver.ServeUnconfigured(nil)
			}

			a, err := app.FromVaultPath(vaultPath)
			if err != nil {
				return mcpserver.ServeUnconfigured(err)
			}

			return mcpserver.Serve(a)
		},
	}

	return cmd
}
