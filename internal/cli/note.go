package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/search"
	"github.com/spf13/cobra"
)

func newNoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "note",
		Short: "Manage notes",
	}

	cmd.AddCommand(
		newNoteCreateCmd(),
		newNoteListCmd(),
		newNoteShowCmd(),
		newNoteUpdateCmd(),
		newNoteDeleteCmd(),
		newNoteSearchCmd(),
		newNotePruneCmd(),
	)

	return cmd
}

func newNoteCreateCmd() *cobra.Command {
	var (
		title     string
		metaPairs []string
		jsonOut   bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new note from stdin",
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return fmt.Errorf("--title is required")
			}

			content, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			if len(content) == 0 {
				return fmt.Errorf("no content provided on stdin")
			}

			meta, err := model.ParseMetaFlags(metaPairs)
			if err != nil {
				return err
			}

			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			if err := a.ValidateNoteContent(string(content)); err != nil {
				return err
			}

			removed, err := a.AutoRemoveOldestNote()
			if err != nil {
				return err
			}

			now := time.Now()
			note := model.Note{
				CreatedAt: now,
				Title:     title,
				Content:   string(content),
				Metadata:  meta,
			}

			id, err := a.Vault.WriteNote(note)
			if err != nil {
				return err
			}
			note.ID = id
			a.IndexNote(note)

			if jsonOut {
				out := map[string]any{
					"id":           id,
					"title":        title,
					"created_at":   now.Format(time.RFC3339),
					"auto_removed": nil,
				}
				if removed != nil {
					out["auto_removed"] = removed.ID
					if removed.Uncurated {
						out["auto_removed_uncurated"] = true
					}
				}
				return printJSON(cmd.OutOrStdout(), out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s  [%s]\n", title, id)
			if removed != nil {
				displayTitle := removed.Title
				if displayTitle == "" {
					displayTitle = removed.ID
				}
				if removed.Uncurated {
					fmt.Fprintf(cmd.ErrOrStderr(), "Removed: %s  [%s]  (oldest — UNCURATED, knowledge may be lost. Run /kno.curate)\n", displayTitle, removed.ID)
				} else {
					fmt.Fprintf(cmd.ErrOrStderr(), "Removed: %s  [%s]  (oldest curated — curate backlog reminder)\n", displayTitle, removed.ID)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Title for the note (required)")
	cmd.Flags().StringArrayVar(&metaPairs, "meta", nil, "Metadata key=value pair (repeatable)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")

	return cmd
}

func newNoteListCmd() *cobra.Command {
	var (
		filterPairs []string
		limit       int
		jsonOut     bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List notes",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			filters, err := model.ParseFilterFlags(filterPairs)
			if err != nil {
				return err
			}

			if limit == 0 {
				limit = a.Config.Notes.DefaultListLimit
			}

			metas, err := a.Vault.ListNotes(0)
			if err != nil {
				return err
			}

			var filtered []model.NoteMeta
			for _, m := range metas {
				if filters != nil && !m.Metadata.MatchesFilter(filters) {
					continue
				}
				filtered = append(filtered, m)
				if limit > 0 && len(filtered) >= limit {
					break
				}
			}

			if jsonOut {
				type noteListItem struct {
					ID        string         `json:"id"`
					Title     string         `json:"title"`
					Metadata  map[string]any `json:"metadata"`
					CreatedAt string         `json:"created_at"`
				}
				var items []noteListItem
				for _, m := range filtered {
					items = append(items, noteListItem{
						ID:        m.ID,
						Title:     m.Title,
						Metadata:  noteMetaJSON(m.Metadata),
						CreatedAt: m.CreatedAt,
					})
				}
				if items == nil {
					items = []noteListItem{}
				}
				return printJSON(cmd.OutOrStdout(), items)
			}

			if len(filtered) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No notes found.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-40s  %-12s  %s\n", "ID", "TITLE", "CREATED", "STATUS")
			for _, m := range filtered {
				status := "not curated"
				if m.Metadata.Has("curated_at") {
					status = "curated"
				}
				created := m.CreatedAt
				if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
					created = t.Format("2006-01-02")
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-40s  %-12s  %s\n", m.ID, m.Title, created, status)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVar(&filterPairs, "filter", nil, "Filter key=value (repeatable, use null for absent)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of notes to return")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")

	return cmd
}

func newNoteShowCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "show <id> [<id>...]",
		Short: "Show note details",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			type showResult struct {
				ID        string         `json:"id"`
				Title     string         `json:"title"`
				Content   string         `json:"content"`
				Metadata  map[string]any `json:"metadata"`
				CreatedAt string         `json:"created_at"`
			}

			var results []showResult
			for _, id := range args {
				note, err := a.Vault.ReadNote(id)
				if err != nil {
					return fmt.Errorf("not found: note %s", id)
				}
				results = append(results, showResult{
					ID:        note.ID,
					Title:     note.Title,
					Content:   note.Content,
					Metadata:  noteMetaJSON(note.Metadata),
					CreatedAt: note.CreatedAt.Format(time.RFC3339),
				})
			}

			if jsonOut {
				return printJSON(cmd.OutOrStdout(), results)
			}

			for i, r := range results {
				if i > 0 {
					fmt.Fprintln(cmd.OutOrStdout())
				}
				created := r.CreatedAt
				if t, err := time.Parse(time.RFC3339, r.CreatedAt); err == nil {
					created = t.Format("2006-01-02")
				}
				header := decorativeHeader(fmt.Sprintf("━━━ %s  [%s]  %s ", r.Title, r.ID, created), 72)
				fmt.Fprintln(cmd.OutOrStdout(), header)
				fmt.Fprintln(cmd.OutOrStdout())
				fmt.Fprint(cmd.OutOrStdout(), r.Content)
				if r.Content != "" && r.Content[len(r.Content)-1] != '\n' {
					fmt.Fprintln(cmd.OutOrStdout())
				}
				fmt.Fprintln(cmd.OutOrStdout())
				for k, v := range r.Metadata {
					if v == nil {
						fmt.Fprintf(cmd.OutOrStdout(), "%s: —\n", k)
					} else if s, ok := v.(string); ok {
						fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", k, s)
					} else if arr, ok := v.([]string); ok {
						fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", k, joinMeta(arr))
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newNoteUpdateCmd() *cobra.Command {
	var (
		metaPairs []string
		jsonOut   bool
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a note's content or metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			meta, err := model.ParseMetaFlags(metaPairs)
			if err != nil {
				return err
			}

			var content *string
			stdin, err := readStdin()
			if err != nil {
				return err
			}
			if stdin != "" {
				content = &stdin
			}

			if content == nil && meta == nil {
				return fmt.Errorf("nothing to update; provide content on stdin or --meta flags")
			}

			if content != nil {
				if err := a.ValidateNoteContent(*content); err != nil {
					return err
				}
			}

			id := args[0]

			// Read title before update for output
			noteMeta, err := a.Vault.ReadNoteMeta(id)
			if err != nil {
				return fmt.Errorf("not found: note %s", id)
			}

			if err := a.Vault.UpdateNote(id, content, meta); err != nil {
				return fmt.Errorf("updating note: %w", err)
			}

			// Update search index
			if note, err := a.Vault.ReadNote(id); err == nil {
				a.IndexNote(note)
			}

			if jsonOut {
				return printJSON(cmd.OutOrStdout(), map[string]any{
					"id":         id,
					"updated_at": time.Now().Format(time.RFC3339),
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updated: %s  [%s]\n", noteMeta.Title, id)
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&metaPairs, "meta", nil, "Metadata key=value pair (repeatable)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newNoteSearchCmd() *cobra.Command {
	var (
		filterPairs []string
		limit       int
		jsonOut     bool
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search notes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			filters, err := model.ParseFilterFlags(filterPairs)
			if err != nil {
				return err
			}

			if limit == 0 {
				limit = a.Config.Search.DefaultLimit
			}

			idx, err := search.Open(a.Vault.IndexDir())
			if err != nil {
				return fmt.Errorf("opening search index (try 'kno vault rebuild-index'): %w", err)
			}
			defer idx.Close()

			results, err := idx.SearchNotes(args[0], limit*2)
			if err != nil {
				return err
			}

			type searchResult struct {
				ID        string         `json:"id"`
				Title     string         `json:"title"`
				Score     float64        `json:"score"`
				Metadata  map[string]any `json:"metadata"`
				CreatedAt string         `json:"created_at"`
			}

			var output []searchResult
			for _, r := range results {
				if limit > 0 && len(output) >= limit {
					break
				}
				noteMeta, err := a.Vault.ReadNoteMeta(r.ID)
				if err != nil {
					continue
				}
				if filters != nil && !noteMeta.Metadata.MatchesFilter(filters) {
					continue
				}
				output = append(output, searchResult{
					ID:        r.ID,
					Title:     noteMeta.Title,
					Score:     r.Score,
					Metadata:  noteMetaJSON(noteMeta.Metadata),
					CreatedAt: noteMeta.CreatedAt,
				})
			}

			if jsonOut {
				if output == nil {
					output = []searchResult{}
				}
				return printJSON(cmd.OutOrStdout(), output)
			}

			if len(output) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No results found.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-40s  %-6s  %s\n", "ID", "TITLE", "SCORE", "STATUS")
			for _, r := range output {
				status := "not curated"
				if r.Metadata["curated_at"] != nil {
					status = "curated"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-40s  %-6.2f  %s\n", r.ID, r.Title, r.Score, status)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&filterPairs, "filter", nil, "Filter key=value (repeatable)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum results")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newNoteDeleteCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Permanently delete a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			id := args[0]

			noteMeta, err := a.Vault.ReadNoteMeta(id)
			if err != nil {
				return fmt.Errorf("not found: note %s", id)
			}

			if err := a.Vault.DeleteNote(id); err != nil {
				return err
			}
			a.RemoveNoteFromIndex(id)

			if jsonOut {
				return printJSON(cmd.OutOrStdout(), map[string]any{
					"id":      id,
					"title":   noteMeta.Title,
					"deleted": true,
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted: %s  [%s]\n", noteMeta.Title, id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newNotePruneCmd() *cobra.Command {
	var (
		count   int
		dryRun  bool
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove oldest notes regardless of curate status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if count <= 0 {
				return fmt.Errorf("--count is required")
			}

			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

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
					status := "not curated"
					if m.Metadata.Has("curated_at") {
						status = "curated"
					}
					created := m.CreatedAt
					if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
						created = t.Format("2006-01-02")
					}
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
				status := "not curated"
				if m.Metadata.Has("curated_at") {
					status = "curated"
				}
				created := m.CreatedAt
				if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
					created = t.Format("2006-01-02")
				}
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

func joinMeta(vals []string) string {
	if len(vals) == 1 {
		return vals[0]
	}
	result := ""
	for i, v := range vals {
		if i > 0 {
			result += ", "
		}
		result += v
	}
	return result
}
