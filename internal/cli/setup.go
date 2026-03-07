package cli

import (
	"fmt"
	"os"

	"github.com/kno-ai/kno/internal/config"
	"github.com/kno-ai/kno/internal/vault/fs"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	var (
		name            string
		noClaudeDesktop bool
	)

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Initialize a kno vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			vaultPath := resolveVault(cmd)
			if vaultPath == "" {
				return fmt.Errorf("could not determine default vault path; use --vault")
			}

			if err := os.MkdirAll(vaultPath, 0o755); err != nil {
				return fmt.Errorf("creating vault directory: %w", err)
			}

			v := fs.New(vaultPath)
			if err := v.EnsureLayout(); err != nil {
				return fmt.Errorf("creating vault layout: %w", err)
			}

			// Write default config if none exists.
			cfgPath := config.ConfigPath(vaultPath)
			if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
				if err := config.Save(vaultPath, config.DefaultConfig()); err != nil {
					return fmt.Errorf("writing config: %w", err)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓  Vault created at %s\n", vaultPath)
			fmt.Fprintf(cmd.OutOrStdout(), "✓  Config written to %s\n", cfgPath)

			if !noClaudeDesktop {
				serverName := "kno"
				if name != "" {
					serverName = name
				}
				results := registerMCP("", vaultPath, serverName)
				if len(results) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "✓  MCP server %q registered with Claude Desktop\n", serverName)
					fmt.Fprintf(cmd.OutOrStdout(), "\nRestart Claude Desktop to activate %s skills.\n", serverName)
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "—  Claude Desktop not found — skipping MCP registration")
					fmt.Fprintln(cmd.OutOrStdout(), "\nTo register manually, add the following to your Claude Desktop config:")
					fmt.Fprintln(cmd.OutOrStdout())
					fmt.Fprintln(cmd.OutOrStdout(), "  {")
					fmt.Fprintln(cmd.OutOrStdout(), "    \"mcpServers\": {")
					fmt.Fprintf(cmd.OutOrStdout(), "      %q: {\n", serverName)
					fmt.Fprintln(cmd.OutOrStdout(), "        \"command\": \"kno\",")
					fmt.Fprintf(cmd.OutOrStdout(), "        \"args\": [\"--vault\", %q, \"mcp\"]\n", vaultPath)
					fmt.Fprintln(cmd.OutOrStdout(), "      }")
					fmt.Fprintln(cmd.OutOrStdout(), "    }")
					fmt.Fprintln(cmd.OutOrStdout(), "  }")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "MCP server name (default: kno)")
	cmd.Flags().BoolVar(&noClaudeDesktop, "no-claude-desktop", false, "Skip Claude Desktop MCP registration")

	return cmd
}
