package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kno-ai/kno/internal/config"
	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/skills/embedded"
	"github.com/kno-ai/kno/internal/vault/fs"
	"github.com/kno-ai/kno/internal/vault/sanitize"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a project vault in the current directory",
		Long:  "Creates a .kno/ project vault in the current directory (or git root). The vault is self-contained with its own config, notes, pages, and search index.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot determine working directory: %w", err)
			}

			// Use git root if in a repo, otherwise cwd.
			projectRoot := cwd
			if root := findRepoRoot(cwd); root != "" {
				projectRoot = root
			}

			vaultPath := filepath.Join(projectRoot, ".kno")

			// Check if already initialized.
			if info, err := os.Stat(vaultPath); err == nil && info.IsDir() {
				return fmt.Errorf("project vault already exists at %s", vaultPath)
			}

			// Create vault directory.
			if err := os.MkdirAll(vaultPath, 0o755); err != nil {
				return fmt.Errorf("creating .kno directory: %w", err)
			}

			// Create vault layout (notes/, pages/).
			v := fs.New(vaultPath)
			if err := v.EnsureLayout(); err != nil {
				// Clean up on failure.
				os.RemoveAll(vaultPath)
				return fmt.Errorf("creating vault layout: %w", err)
			}

			pageName := filepath.Base(projectRoot)

			// Write default config with page binding.
			cfg := config.DefaultConfig()
			cfg.Page = sanitize.Slugify(pageName)
			if err := config.Save(vaultPath, cfg); err != nil {
				os.RemoveAll(vaultPath)
				return fmt.Errorf("writing config: %w", err)
			}

			// Create the default page with guidance template.
			guidance := loadPageTemplate()
			page := model.Page{
				Name:      pageName,
				CreatedAt: time.Now(),
				Content:   guidance,
			}
			if _, err := v.WritePage(page); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not create default page: %v\n", err)
			}

			// Write .gitignore inside .kno/.
			gitignorePath := filepath.Join(vaultPath, ".gitignore")
			gitignoreContent := "index/\nnotes/\n"
			if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o644); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not write .gitignore: %v\n", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓  Project vault created at %s\n", vaultPath)
			fmt.Fprintf(cmd.OutOrStdout(), "✓  Page %q created and bound for auto-load\n", pageName)
			fmt.Fprintf(cmd.OutOrStdout(), "✓  .gitignore created (index/ and notes/ excluded from git)\n")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Commit .kno/ to share curated pages with your team.")
			fmt.Fprintln(cmd.OutOrStdout(), "Notes are local to each developer — pages are the shared artifact.")

			return nil
		},
	}

	return cmd
}

// loadPageTemplate returns the guidance content for a new project page.
func loadPageTemplate() string {
	store := embedded.New()

	content, err := store.Get("templates/project-page.md")
	if err != nil {
		return ""
	}

	return "<!-- Guidance -->\n" + content
}
