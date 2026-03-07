package cli

import (
	"fmt"
	"time"

	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/search"
	"github.com/spf13/cobra"
)

func newPageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "page",
		Short: "Manage pages",
	}

	cmd.AddCommand(
		newPageCreateCmd(),
		newPageListCmd(),
		newPageShowCmd(),
		newPageUpdateCmd(),
		newPageSearchCmd(),
	)

	return cmd
}

func newPageCreateCmd() *cobra.Command {
	var (
		name      string
		metaPairs []string
		jsonOut   bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new page",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			meta, err := model.ParseMetaFlags(metaPairs)
			if err != nil {
				return err
			}

			content, err := readStdin()
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}

			now := time.Now()
			page := model.Page{
				Name:      name,
				CreatedAt: now,
				Content:   content,
				Metadata:  meta,
			}

			id, err := a.Vault.WritePage(page)
			if err != nil {
				return err
			}
			page.ID = id
			a.IndexPage(page)

			if jsonOut {
				return printJSON(cmd.OutOrStdout(), map[string]any{
					"id":         id,
					"name":       name,
					"created_at": now.Format(time.RFC3339),
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s  [%s]\n", name, id)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Page name (required)")
	cmd.Flags().StringArrayVar(&metaPairs, "meta", nil, "Metadata key=value pair (repeatable)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newPageListCmd() *cobra.Command {
	var (
		filterPairs []string
		jsonOut     bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pages",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			filters, err := model.ParseFilterFlags(filterPairs)
			if err != nil {
				return err
			}

			metas, err := a.Vault.ListPages()
			if err != nil {
				return err
			}

			if jsonOut {
				type pageListItem struct {
					ID        string         `json:"id"`
					Name      string         `json:"name"`
					Metadata  map[string]any `json:"metadata"`
					CreatedAt string         `json:"created_at"`
				}
				var items []pageListItem
				for _, m := range metas {
					if filters != nil && !m.Metadata.MatchesFilter(filters) {
						continue
					}
					item := pageListItem{
						ID:        m.ID,
						Name:      m.Name,
						Metadata:  metaMapToJSON(m.Metadata),
						CreatedAt: m.CreatedAt,
					}
					items = append(items, item)
				}
				if items == nil {
					items = []pageListItem{}
				}
				return printJSON(cmd.OutOrStdout(), items)
			}

			var filtered []model.PageMeta
			for _, m := range metas {
				if filters != nil && !m.Metadata.MatchesFilter(filters) {
					continue
				}
				filtered = append(filtered, m)
			}

			if len(filtered) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No pages found.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-30s  %s\n", "ID", "NAME", "LAST DISTILLED")
			for _, m := range filtered {
				lastDistilled := "—"
				if m.Metadata != nil {
					if v := m.Metadata.Get("last_distilled_at"); v != "" {
						if t, err := time.Parse(time.RFC3339, v); err == nil {
							lastDistilled = t.Format("2006-01-02")
						} else {
							lastDistilled = v
						}
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-30s  %s\n", m.ID, m.Name, lastDistilled)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVar(&filterPairs, "filter", nil, "Filter key=value (repeatable)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newPageShowCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a page",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			page, err := a.Vault.ReadPage(args[0])
			if err != nil {
				return fmt.Errorf("Not found: page %s", args[0])
			}

			if jsonOut {
				out := map[string]any{
					"id":         page.ID,
					"name":       page.Name,
					"content":    page.Content,
					"metadata":   metaMapToJSON(page.Metadata),
					"created_at": page.CreatedAt.Format(time.RFC3339),
				}
				return printJSON(cmd.OutOrStdout(), out)
			}

			lastDistilled := formatPageLastDistilled(model.PageMeta{
				Metadata: page.Metadata,
			})
			header := decorativeHeader(fmt.Sprintf("━━━ %s  [%s]  %s ", page.Name, page.ID, lastDistilled), 72)
			fmt.Fprintln(cmd.OutOrStdout(), header)
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprint(cmd.OutOrStdout(), page.Content)
			if page.Content != "" && page.Content[len(page.Content)-1] != '\n' {
				fmt.Fprintln(cmd.OutOrStdout())
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newPageUpdateCmd() *cobra.Command {
	var (
		metaPairs []string
		jsonOut   bool
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a page's content or metadata",
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
				return fmt.Errorf("Nothing to update; provide content on stdin or --meta flags")
			}

			id := args[0]

			pageMeta, err := a.Vault.ReadPageMeta(id)
			if err != nil {
				return fmt.Errorf("Not found: page %s", id)
			}

			if err := a.Vault.UpdatePage(id, content, meta); err != nil {
				return fmt.Errorf("updating page: %w", err)
			}

			if page, err := a.Vault.ReadPage(id); err == nil {
				a.IndexPage(page)
			}

			if jsonOut {
				return printJSON(cmd.OutOrStdout(), map[string]any{
					"id":         id,
					"updated_at": time.Now().Format(time.RFC3339),
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updated: %s  [%s]\n", pageMeta.Name, id)
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&metaPairs, "meta", nil, "Metadata key=value pair (repeatable)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newPageSearchCmd() *cobra.Command {
	var (
		filterPairs []string
		limit       int
		jsonOut     bool
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search pages",
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

			results, err := idx.SearchPages(args[0], limit*2)
			if err != nil {
				return err
			}

			type searchResult struct {
				ID      string  `json:"id"`
				Name    string  `json:"name"`
				Score   float64 `json:"score"`
				Excerpt string  `json:"excerpt,omitempty"`
			}

			var output []searchResult
			for _, r := range results {
				if limit > 0 && len(output) >= limit {
					break
				}
				pageMeta, err := a.Vault.ReadPageMeta(r.ID)
				if err != nil {
					continue
				}
				if filters != nil && !pageMeta.Metadata.MatchesFilter(filters) {
					continue
				}
				sr := searchResult{
					ID:    r.ID,
					Name:  pageMeta.Name,
					Score: r.Score,
				}
				if p, err := a.Vault.ReadPage(r.ID); err == nil && p.Content != "" {
					excerpt := p.Content
					if len(excerpt) > 200 {
						excerpt = excerpt[:200] + "..."
					}
					sr.Excerpt = excerpt
				}
				output = append(output, sr)
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

			fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-30s  %s\n", "ID", "NAME", "SCORE")
			for _, r := range output {
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %-30s  %.2f\n", r.ID, r.Name, r.Score)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&filterPairs, "filter", nil, "Filter key=value (repeatable)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum results")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
