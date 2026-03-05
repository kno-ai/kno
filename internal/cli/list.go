package cli

import (
	"fmt"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/config"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent captures",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config (have you run 'kno setup'?): %w", err)
			}

			a, err := app.FromConfig(cfg)
			if err != nil {
				return err
			}

			names, err := a.Vault.ListCaptures(limit)
			if err != nil {
				return err
			}

			if len(names) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No captures yet.")
				return nil
			}

			for _, name := range names {
				fmt.Fprintln(cmd.OutOrStdout(), name)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Number of recent captures to show")

	return cmd
}
