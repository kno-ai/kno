package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kno-ai/kno/internal/config"
	"github.com/kno-ai/kno/internal/vault/fs"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	var (
		name            string
		noClaudeDesktop bool
		publishPath     string
	)

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Initialize a kno vault",
		Args:  cobra.NoArgs,
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

			// Add publish target if requested.
			if publishPath != "" {
				cfg, err := config.Load(vaultPath)
				if err != nil {
					return fmt.Errorf("loading config: %w", err)
				}

				// Check if target already exists.
				alreadyExists := false
				for _, t := range cfg.Publish.Targets {
					if t.Path == publishPath {
						alreadyExists = true
						break
					}
				}

				if !alreadyExists {
					cfg.Publish.Targets = append(cfg.Publish.Targets, config.PublishTarget{
						Path:   publishPath,
						Format: "frontmatter",
					})
					if err := config.Save(vaultPath, cfg); err != nil {
						return fmt.Errorf("saving config: %w", err)
					}
					// Create target directory.
					expanded := publishPath
					if len(expanded) > 1 && expanded[:2] == "~/" {
						if home, herr := os.UserHomeDir(); herr == nil {
							expanded = filepath.Join(home, expanded[2:])
						}
					}
					if err := os.MkdirAll(expanded, 0o755); err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not create publish directory: %v\n", err)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "✓  Publish target added: %s\n", publishPath)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "—  Publish target already configured: %s\n", publishPath)
				}
			}

			if !noClaudeDesktop {
				serverName := "kno"
				if name != "" {
					serverName = name
				}
				results := registerMCP("", vaultPath, serverName)
				if len(results) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "✓  MCP server %q registered with Claude Desktop\n", serverName)
					fmt.Fprintf(cmd.OutOrStdout(), "\nRestart Claude Desktop, then enter /kno in a chat to connect.\n")
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

			// Show publish tip if no publish targets configured.
			if publishPath == "" {
				cfg, _ := config.Load(vaultPath)
				if len(cfg.Publish.Targets) == 0 {
					fmt.Fprintln(cmd.OutOrStdout())
					fmt.Fprintln(cmd.OutOrStdout(), "Tip: Publish curated pages to Obsidian or any markdown viewer")
					fmt.Fprintln(cmd.OutOrStdout(), "that supports frontmatter:")
					fmt.Fprintln(cmd.OutOrStdout(), "  kno setup --publish ~/obsidian/kno")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "MCP server name (default: kno)")
	cmd.Flags().BoolVar(&noClaudeDesktop, "no-claude-desktop", false, "Skip Claude Desktop MCP registration")
	cmd.Flags().StringVar(&publishPath, "publish", "", "Publish curated pages to this directory")

	return cmd
}
