package cli

import (
	"fmt"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/config"
	"github.com/kno-ai/kno/internal/publish"
	"github.com/spf13/cobra"
)

func newPublishCmd() *cobra.Command {
	var (
		pageID  string
		format  string
		dryRun  bool
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish pages to configured targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := loadApp(cmd)
			if err != nil {
				return err
			}

			targets := a.CollectPublishTargets()
			if len(targets) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No publish targets configured.")
				fmt.Fprintln(cmd.OutOrStdout(), "")
				fmt.Fprintln(cmd.OutOrStdout(), "Publish to Obsidian or any markdown directory:")
				fmt.Fprintln(cmd.OutOrStdout(), "  kno setup --publish ~/obsidian/kno")
				fmt.Fprintln(cmd.OutOrStdout(), "")
				fmt.Fprintln(cmd.OutOrStdout(), "This adds a target to your user config (~/.kno/config.toml)")
				fmt.Fprintln(cmd.OutOrStdout(), "so pages from all your vaults publish there.")
				fmt.Fprintln(cmd.OutOrStdout(), "")
				fmt.Fprintln(cmd.OutOrStdout(), "For project-specific targets, add to your vault's config.toml:")
				fmt.Fprintln(cmd.OutOrStdout(), "  [[publish.targets]]")
				fmt.Fprintln(cmd.OutOrStdout(), `  path = "docs/kno"`)
				fmt.Fprintln(cmd.OutOrStdout(), `  format = "markdown"`)
				return nil
			}

			// Override format if specified (copy to avoid mutating config).
			if format != "" {
				if !config.ValidPublishFormat(format) {
					return fmt.Errorf("invalid format %q; use \"markdown\" or \"frontmatter\"", format)
				}
				override := make([]config.PublishTarget, len(targets))
				copy(override, targets)
				for i := range override {
					override[i].Format = format
				}
				targets = override
			}

			var pageIDs []string
			if pageID != "" {
				pageIDs = []string{pageID}
			}

			if dryRun {
				return publishDryRun(cmd, a, targets, pageIDs)
			}

			results, err := publish.PublishPages(a.Vault, targets, a.ProjectName, pageIDs)
			if err != nil {
				return err
			}

			if jsonOut {
				type jsonResult struct {
					PageID string `json:"page_id"`
					Target string `json:"target"`
					Path   string `json:"path,omitempty"`
					Error  string `json:"error,omitempty"`
				}
				var items []jsonResult
				for _, r := range results {
					item := jsonResult{PageID: r.PageID, Target: r.Target, Path: r.Path}
					if r.Err != nil {
						item.Error = r.Err.Error()
					}
					items = append(items, item)
				}
				if items == nil {
					items = []jsonResult{}
				}
				return printJSON(cmd.OutOrStdout(), items)
			}

			succeeded := 0
			failed := 0
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "  ✗ %s → %s: %v\n", r.PageName, r.Target, r.Err)
					failed++
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "  ✓ %s → %s\n", r.PageName, r.Target)
					succeeded++
				}
			}

			if failed > 0 {
				return fmt.Errorf("%d of %d pages failed to publish", failed, succeeded+failed)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nPublished %d page(s).\n", succeeded)
			return nil
		},
	}

	cmd.Flags().StringVar(&pageID, "page", "", "Publish a single page by ID")
	cmd.Flags().StringVar(&format, "format", "", "Override format (markdown or frontmatter)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be published")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")

	return cmd
}

func publishDryRun(cmd *cobra.Command, a *app.App, targets []config.PublishTarget, pageIDs []string) error {
	metas, err := a.Vault.ListPages()
	if err != nil {
		return err
	}

	var pages []string
	if pageIDs != nil {
		pages = pageIDs
	} else {
		for _, m := range metas {
			pages = append(pages, m.ID)
		}
	}

	if len(pages) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No pages to publish.")
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Would publish:")
	for _, target := range targets {
		targetPath := target.Path
		if publish.ShouldGroup(target, a.ProjectName) {
			targetPath = targetPath + "/" + a.ProjectName
		}
		for _, id := range pages {
			name := id
			for _, m := range metas {
				if m.ID == id {
					name = m.Name
					break
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s → %s/%s.md  [%s]\n", name, targetPath, id, target.Format)
		}
	}
	return nil
}
