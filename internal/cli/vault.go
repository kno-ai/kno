package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/kno-ai/kno/internal/model"
	"github.com/spf13/cobra"
)

func newVaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Vault operations",
	}

	cmd.AddCommand(newVaultStatusCmd())
	return cmd
}

func newVaultStatusCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show vault status",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			absPath, _ := filepath.Abs(a.VaultPath)

			total, err := a.Vault.CountNotes()
			if err != nil {
				return err
			}

			// Count distilled vs undistilled
			allNotes, err := a.Vault.ListNotes(0)
			if err != nil {
				return err
			}
			distilled := 0
			for _, c := range allNotes {
				if c.Metadata.Has("distilled_at") {
					distilled++
				}
			}
			undistilled := total - distilled
			remaining := a.Config.Notes.MaxCount - total
			if remaining < 0 {
				remaining = 0
			}

			pages, err := a.Vault.ListPages()
			if err != nil {
				return err
			}

			if jsonOut {
				type pageInfo struct {
					ID       string         `json:"id"`
					Name     string         `json:"name"`
					Metadata map[string]any `json:"metadata"`
				}
				var pageInfos []pageInfo
				for _, p := range pages {
					pageInfos = append(pageInfos, pageInfo{
						ID:       p.ID,
						Name:     p.Name,
						Metadata: metaMapToJSON(p.Metadata),
					})
				}
				if pageInfos == nil {
					pageInfos = []pageInfo{}
				}

				out := map[string]any{
					"vault_path": absPath,
					"notes": map[string]int{
						"total":       total,
						"max_count":   a.Config.Notes.MaxCount,
						"remaining":   remaining,
						"distilled":   distilled,
						"undistilled": undistilled,
					},
					"pages":  pageInfos,
					"config": a.Config,
				}
				return printJSON(cmd.OutOrStdout(), out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Vault: %s\n\n", absPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Notes: %d / %d  (%d remaining)\n", total, a.Config.Notes.MaxCount, remaining)
			fmt.Fprintf(cmd.OutOrStdout(), "  Distilled:   %d\n", distilled)
			fmt.Fprintf(cmd.OutOrStdout(), "  Undistilled: %d\n", undistilled)

			if len(pages) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\nPages:\n")
				for _, p := range pages {
					lastDistilled := formatPageLastDistilled(p)
					fmt.Fprintf(cmd.OutOrStdout(), "  %s  %-25s  %s\n", p.ID, p.Name, lastDistilled)
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "\nNo pages yet.")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nConfig:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  notes.max_count              %d\n", a.Config.Notes.MaxCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  notes.default_list_limit     %d\n", a.Config.Notes.DefaultListLimit)
			fmt.Fprintf(cmd.OutOrStdout(), "  notes.summary_max_tokens     %d\n", a.Config.Notes.SummaryMaxTokens)
			fmt.Fprintf(cmd.OutOrStdout(), "  pages.max_content_tokens     %d\n", a.Config.Pages.MaxContentTokens)
			fmt.Fprintf(cmd.OutOrStdout(), "  distill.max_notes_per_run    %d\n", a.Config.Distill.MaxNotesPerRun)
			fmt.Fprintf(cmd.OutOrStdout(), "  search.default_limit         %d\n", a.Config.Search.DefaultLimit)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func formatPageLastDistilled(p model.PageMeta) string {
	if p.Metadata == nil {
		return "(never distilled)"
	}
	lastAt := p.Metadata.Get("last_distilled_at")
	if lastAt == "" {
		return "(never distilled)"
	}
	parsed, err := time.Parse(time.RFC3339, lastAt)
	if err != nil {
		return lastAt
	}
	return "last distilled " + parsed.Format("2006-01-02")
}
