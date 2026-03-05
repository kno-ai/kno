package cli

import (
	"fmt"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/config"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <capture>",
		Short: "Show a capture's content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config (have you run 'kno setup'?): %w", err)
			}

			a, err := app.FromConfig(cfg)
			if err != nil {
				return err
			}

			meta, content, err := a.Vault.ReadCapture(args[0])
			if err != nil {
				return fmt.Errorf("capture %q not found: %w", args[0], err)
			}

			if meta.Title != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Title:   %s\n", meta.Title)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", meta.Created)
			fmt.Fprintf(cmd.OutOrStdout(), "ID:      %s\n", meta.ID)
			if len(meta.Meta) > 0 {
				for k, v := range meta.Meta {
					fmt.Fprintf(cmd.OutOrStdout(), "%-8s %s\n", k+":", v)
				}
			}
			fmt.Fprintln(cmd.OutOrStdout(), "---")
			fmt.Fprint(cmd.OutOrStdout(), content)

			return nil
		},
	}

	return cmd
}
