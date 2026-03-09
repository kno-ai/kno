package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kno-ai/kno/internal/config"
	"github.com/kno-ai/kno/internal/vault/fs"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	var (
		name        string
		registerCSV string
		noRegister  bool
		publishPath string
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

			// Set publish target if requested — writes to user config
			// (~/.kno/config.toml) so the target applies across all vaults.
			// Always replaces any previous target set via setup --publish.
			// For multiple targets, edit config.toml directly.
			if publishPath != "" {
				userDir := config.UserConfigDir()
				if userDir == "" {
					return fmt.Errorf("could not determine user config directory")
				}
				if err := os.MkdirAll(userDir, 0o755); err != nil {
					return fmt.Errorf("creating user config directory: %w", err)
				}

				userCfg, err := config.Load(userDir)
				if err != nil {
					return fmt.Errorf("loading user config: %w", err)
				}

				// Replace all existing targets with the new one.
				newTarget := config.PublishTarget{
					Path:   publishPath,
					Format: "frontmatter",
				}
				userCfg.Publish.Targets = []config.PublishTarget{newTarget}
				if err := config.Save(userDir, userCfg); err != nil {
					return fmt.Errorf("saving user config: %w", err)
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
				fmt.Fprintf(cmd.OutOrStdout(), "✓  Publish target set: %s (in %s)\n", publishPath, config.ConfigPath(userDir))
			}

			if !noRegister {
				serverName := "kno"
				if name != "" {
					serverName = name
				}

				// Determine which clients to register with.
				clients := knownMCPClients()
				if registerCSV != "" {
					// Filter to only the requested clients.
					requested := parseCSV(registerCSV)
					var filtered []mcpClient
					for _, c := range clients {
						for _, r := range requested {
							if c.Name == r {
								filtered = append(filtered, c)
								break
							}
						}
					}
					clients = filtered
				}

				registered, regErrors := registerMCPClients(clients, vaultPath, serverName)
				for _, e := range regErrors {
					fmt.Fprintf(cmd.ErrOrStderr(), "⚠  MCP registration failed: %v\n", e)
				}
				if len(registered) > 0 {
					for _, c := range registered {
						fmt.Fprintf(cmd.OutOrStdout(), "✓  MCP server %q registered with %s\n", serverName, c.Name)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "\nRestart your client, then enter /%s.start in a chat to connect.\n", serverName)
				} else if len(regErrors) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "—  No supported clients found — skipping MCP registration")
					fmt.Fprintln(cmd.OutOrStdout(), "\nTo register manually, add the following to your client's MCP config:")
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

			// Show publish tip if no publish targets configured anywhere.
			if publishPath == "" {
				hasTargets := false
				if cfg, err := config.Load(vaultPath); err == nil && len(cfg.Publish.Targets) > 0 {
					hasTargets = true
				}
				if !hasTargets && len(config.LoadUserPublishTargets()) > 0 {
					hasTargets = true
				}
				if !hasTargets {
					fmt.Fprintln(cmd.OutOrStdout())
					fmt.Fprintln(cmd.OutOrStdout(), "Tip: Publish pages to Obsidian or any markdown viewer:")
					fmt.Fprintln(cmd.OutOrStdout(), "  kno setup --publish ~/obsidian/kno")
					fmt.Fprintln(cmd.OutOrStdout(), "Pages from all your vaults will publish there automatically.")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "MCP server name (default: kno)")
	cmd.Flags().StringVar(&registerCSV, "register", "", "Register with specific clients only (comma-separated: claude-desktop,claude-code)")
	cmd.Flags().BoolVar(&noRegister, "no-register", false, "Skip all client MCP registration")
	cmd.Flags().StringVar(&publishPath, "publish", "", "Publish curated pages to this directory")

	return cmd
}

// parseCSV splits a comma-separated string into trimmed, non-empty tokens.
func parseCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
