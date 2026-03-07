package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/search"
	"github.com/spf13/cobra"
)

func newVaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Vault operations",
	}

	cmd.AddCommand(newVaultStatusCmd(), newVaultRebuildIndexCmd())
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

			// Count curated vs uncurated
			allNotes, err := a.Vault.ListNotes(0)
			if err != nil {
				return err
			}
			curated := 0
			for _, c := range allNotes {
				if c.Metadata.Has("curated_at") {
					curated++
				}
			}
			uncurated := total - curated
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
						"total":     total,
						"max_count": a.Config.Notes.MaxCount,
						"remaining": remaining,
						"curated":   curated,
						"uncurated": uncurated,
					},
					"pages":  pageInfos,
					"config": a.Config,
				}
				return printJSON(cmd.OutOrStdout(), out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Vault: %s\n\n", absPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Notes: %d / %d  (%d remaining)\n", total, a.Config.Notes.MaxCount, remaining)
			fmt.Fprintf(cmd.OutOrStdout(), "  Curated:   %d\n", curated)
			fmt.Fprintf(cmd.OutOrStdout(), "  Uncurated: %d\n", uncurated)

			if len(pages) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\nPages:\n")
				for _, p := range pages {
					lastCurated := formatPageLastCurated(p)
					fmt.Fprintf(cmd.OutOrStdout(), "  %s  %-25s  %s\n", p.ID, p.Name, lastCurated)
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "\nNo pages yet.")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nConfig:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  notes.max_count              %d\n", a.Config.Notes.MaxCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  notes.default_list_limit     %d\n", a.Config.Notes.DefaultListLimit)
			fmt.Fprintf(cmd.OutOrStdout(), "  notes.summary_max_tokens     %d\n", a.Config.Notes.SummaryMaxTokens)
			fmt.Fprintf(cmd.OutOrStdout(), "  notes.max_content_tokens     %d\n", a.Config.Notes.MaxContentTokens)
			fmt.Fprintf(cmd.OutOrStdout(), "  pages.max_content_tokens     %d\n", a.Config.Pages.MaxContentTokens)
			fmt.Fprintf(cmd.OutOrStdout(), "  curate.max_notes_per_run     %d\n", a.Config.Curate.MaxNotesPerRun)
			fmt.Fprintf(cmd.OutOrStdout(), "  search.default_limit         %d\n", a.Config.Search.DefaultLimit)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newVaultRebuildIndexCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "rebuild-index",
		Short: "Rebuild the full-text search index",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			if !jsonOut {
				fmt.Fprintln(cmd.OutOrStdout(), "Rebuilding index...")
			}

			idx, err := search.Rebuild(a.Vault)
			if err != nil {
				return fmt.Errorf("rebuilding index: %w", err)
			}
			defer idx.Close()

			notes, _ := a.Vault.ListNotes(0)
			pages, _ := a.Vault.ListPages()

			if jsonOut {
				return printJSON(cmd.OutOrStdout(), map[string]any{
					"status": "ok",
					"notes":  len(notes),
					"pages":  len(pages),
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Indexed %d notes, %d pages.\n", len(notes), len(pages))
			fmt.Fprintln(cmd.OutOrStdout(), "Done.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func formatPageLastCurated(p model.PageMeta) string {
	if p.Metadata == nil {
		return "(never curated)"
	}
	lastAt := p.Metadata.Get("last_curated_at")
	if lastAt == "" {
		return "(never curated)"
	}
	parsed, err := time.Parse(time.RFC3339, lastAt)
	if err != nil {
		return lastAt
	}
	return "last curated " + parsed.Format("2006-01-02")
}
