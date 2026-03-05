package cli

import (
	"fmt"
	"os"

	"github.com/kno-ai/kno/internal/config"
	"github.com/kno-ai/kno/internal/vault/fs"
	"github.com/spf13/cobra"
)

func defaultVaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/kno"
}

func newSetupCmd() *cobra.Command {
	var (
		vaultPath string
		subdir    string
		mcpConfig string
		noMCP     bool
	)

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure kno with a knowledge vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			if vaultPath == "" {
				vaultPath = defaultVaultPath()
			}
			if vaultPath == "" {
				return fmt.Errorf("could not determine default vault path; use --vault")
			}

			if err := os.MkdirAll(vaultPath, 0o755); err != nil {
				return fmt.Errorf("creating vault directory %q: %w", vaultPath, err)
			}

			cfg := config.DefaultConfig()
			cfg.VaultPath = vaultPath
			if subdir != "" {
				cfg.KnoSubdir = subdir
			}

			v := fs.New(cfg.VaultPath, cfg.KnoSubdir)
			if err := v.EnsureLayout(); err != nil {
				return fmt.Errorf("creating vault layout: %w", err)
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Vault: %s\n", vaultPath)
			if configPath, err := config.Path(); err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Config: %s\n", configPath)
			}

			if !noMCP {
				results := registerMCP(mcpConfig)
				for _, r := range results {
					fmt.Fprintf(cmd.OutOrStdout(), "Registered with %s\n", r)
				}
				if len(results) > 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "\nRestart Claude Desktop to activate.")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&vaultPath, "vault", "", "Path to the knowledge vault directory (default: ~/kno)")
	cmd.Flags().StringVar(&subdir, "subdir", "", "Subdirectory within vault for kno data (default: vault root; use e.g. 'kno' for shared vaults)")
	cmd.Flags().StringVar(&mcpConfig, "mcp-config", "", "Explicit path to an MCP client config file to register with")
	cmd.Flags().BoolVar(&noMCP, "no-mcp", false, "Skip MCP client registration")

	return cmd
}
