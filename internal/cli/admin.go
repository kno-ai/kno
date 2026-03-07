package cli

import (
	"fmt"
	"time"

	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/search"
	"github.com/spf13/cobra"
)

func formatDateOnly(rfc3339 string) string {
	if t, err := time.Parse(time.RFC3339, rfc3339); err == nil {
		return t.Format("2006-01-02")
	}
	return rfc3339
}

func newAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Administrative commands",
	}

	pageCmd := &cobra.Command{
		Use:   "page",
		Short: "Page admin operations",
	}
	pageCmd.AddCommand(newAdminPageDeleteCmd())

	indexCmd := &cobra.Command{
		Use:   "index",
		Short: "Search index operations",
	}
	indexCmd.AddCommand(newAdminIndexRebuildCmd())

	cmd.AddCommand(
		newAdminPruneCmd(),
		pageCmd,
		indexCmd,
	)

	return cmd
}

func newAdminPruneCmd() *cobra.Command {
	var (
		count   int
		dryRun  bool
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove oldest notes regardless of distill status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if count <= 0 {
				return fmt.Errorf("--count is required")
			}

			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			// Get all notes sorted oldest first (ListNotes returns newest first)
			metas, err := a.Vault.ListNotes(0)
			if err != nil {
				return err
			}

			// Reverse to get oldest first
			for i, j := 0, len(metas)-1; i < j; i, j = i+1, j-1 {
				metas[i], metas[j] = metas[j], metas[i]
			}

			if count > len(metas) {
				count = len(metas)
			}
			toRemove := metas[:count]

			if dryRun {
				if jsonOut {
					var ids []string
					for _, m := range toRemove {
						ids = append(ids, m.ID)
					}
					return printJSON(cmd.OutOrStdout(), map[string]any{
						"dry_run":      true,
						"would_remove": len(toRemove),
						"ids":          ids,
					})
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Would remove %d notes (oldest first):\n\n", len(toRemove))
				for _, m := range toRemove {
					status := "not distilled"
					if m.Metadata.Has("distilled_at") {
						status = "distilled"
					}
					created := formatDateOnly(m.CreatedAt)
					fmt.Fprintf(cmd.OutOrStdout(), "  %s  %-30s  %s  %s\n", m.ID, m.Title, created, status)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "\nRun without --dry-run to proceed.")
				return nil
			}

			var ids []string
			for _, m := range toRemove {
				if err := a.Vault.DeleteNote(m.ID); err != nil {
					return fmt.Errorf("deleting %s: %w", m.ID, err)
				}
				a.RemoveNoteFromIndex(m.ID)
				ids = append(ids, m.ID)
			}

			if jsonOut {
				return printJSON(cmd.OutOrStdout(), map[string]any{
					"removed": len(ids),
					"ids":     ids,
				})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed %d notes (oldest first):\n\n", len(toRemove))
			for _, m := range toRemove {
				status := "not distilled"
				if m.Metadata.Has("distilled_at") {
					status = "distilled"
				}
				created := formatDateOnly(m.CreatedAt)
				fmt.Fprintf(cmd.OutOrStdout(), "  %s  %-30s  %s  %s\n", m.ID, m.Title, created, status)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&count, "count", 0, "Number of oldest notes to remove (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be removed without removing")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newAdminPageDeleteCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Permanently delete a page",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			pageID := args[0]

			pageMeta, err := a.Vault.ReadPageMeta(pageID)
			if err != nil {
				return fmt.Errorf("Not found: page %s", pageID)
			}

			if err := a.Vault.DeletePage(pageID); err != nil {
				return err
			}
			a.RemovePageFromIndex(pageID)

			// Update notes that reference this page in distilled_into
			allNotes, err := a.Vault.ListNotes(0)
			if err != nil {
				return err
			}

			for _, nm := range allNotes {
				if nm.Metadata == nil || !nm.Metadata.Has("distilled_into") {
					continue
				}

				vals := nm.Metadata["distilled_into"]
				var remaining []string
				for _, v := range vals {
					if v != pageID {
						remaining = append(remaining, v)
					}
				}

				if len(remaining) == len(vals) {
					continue
				}

				updateMeta := make(model.MetaMap)
				if len(remaining) == 0 {
					updateMeta["distilled_into"] = nil
					updateMeta["distilled_at"] = nil
				} else {
					updateMeta["distilled_into"] = remaining
				}

				if err := a.Vault.UpdateNote(nm.ID, nil, updateMeta); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to update note %s: %v\n", nm.ID, err)
				}
			}

			if jsonOut {
				return printJSON(cmd.OutOrStdout(), map[string]any{
					"id":      pageID,
					"deleted": true,
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted: %s  [%s]\n", pageMeta.Name, pageID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newAdminIndexRebuildCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "rebuild",
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
