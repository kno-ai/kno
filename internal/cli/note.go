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
		newNoteSearchCmd(),
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

			// Check capacity and auto-remove if needed.
			count, err := a.Vault.CountNotes()
			if err != nil {
				return err
			}

			var autoRemoved string
			var autoRemovedTitle string
			var autoRemovedUndistilled bool
			if count >= a.Config.Notes.MaxCount {
				oldest, err := a.Vault.OldestDistilledNoteID()
				if err != nil {
					return err
				}
				if oldest == "" {
					// No distilled notes — fall back to oldest note overall
					oldest, err = a.Vault.OldestNoteID()
					if err != nil {
						return err
					}
					if oldest == "" {
						return fmt.Errorf("vault at capacity (%d notes) with nothing to remove", a.Config.Notes.MaxCount)
					}
					autoRemovedUndistilled = true
				}
				if rm, err := a.Vault.ReadNoteMeta(oldest); err == nil {
					autoRemovedTitle = rm.Title
				}
				if err := a.Vault.DeleteNote(oldest); err != nil {
					return fmt.Errorf("auto-removing note: %w", err)
				}
				a.RemoveNoteFromIndex(oldest)
				autoRemoved = oldest
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
					"auto_removed": nilIfEmpty(autoRemoved),
				}
				if autoRemovedUndistilled {
					out["auto_removed_undistilled"] = true
				}
				return printJSON(cmd.OutOrStdout(), out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s  [%s]\n", title, id)
			if autoRemoved != "" {
				displayTitle := autoRemovedTitle
				if displayTitle == "" {
					displayTitle = autoRemoved
				}
				if autoRemovedUndistilled {
					fmt.Fprintf(cmd.ErrOrStderr(), "Removed: %s  [%s]  (oldest — UNDISTILLED, knowledge may be lost. Run /kno.distill)\n", displayTitle, autoRemoved)
				} else {
					fmt.Fprintf(cmd.ErrOrStderr(), "Removed: %s  [%s]  (oldest distilled — distill backlog reminder)\n", displayTitle, autoRemoved)
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
				status := "not distilled"
				if m.Metadata.Has("distilled_at") {
					status = "distilled"
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
				return fmt.Errorf("opening search index (try 'kno admin index-rebuild'): %w", err)
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
				status := "not distilled"
				if r.Metadata["distilled_at"] != nil {
					status = "distilled"
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

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
