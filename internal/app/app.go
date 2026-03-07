package app

import (
	"github.com/kno-ai/kno/internal/config"
	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/search"
	"github.com/kno-ai/kno/internal/skills"
	"github.com/kno-ai/kno/internal/skills/embedded"
	"github.com/kno-ai/kno/internal/vault"
	"github.com/kno-ai/kno/internal/vault/fs"
)

// App wires together the core services. Both CLI and MCP use this.
type App struct {
	VaultPath string
	Config    config.Config
	Vault     vault.Vault
	Skills    skills.Store
}

// FromVaultPath builds an App from a vault directory path.
func FromVaultPath(vaultPath string) (*App, error) {
	cfg, err := config.Load(vaultPath)
	if err != nil {
		return nil, err
	}

	v := fs.New(vaultPath)
	skillStore := embedded.New()

	return &App{
		VaultPath: vaultPath,
		Config:    cfg,
		Vault:     v,
		Skills:    skillStore,
	}, nil
}

// IndexNote updates the search index for a note (no-op if index not built yet).
func (a *App) IndexNote(note model.Note) {
	idx, err := search.TryOpen(a.Vault.IndexDir())
	if err != nil || idx == nil {
		return
	}
	defer idx.Close()
	idx.IndexNote(note)
}

// IndexPage updates the search index for a page (no-op if index not built yet).
func (a *App) IndexPage(page model.Page) {
	idx, err := search.TryOpen(a.Vault.IndexDir())
	if err != nil || idx == nil {
		return
	}
	defer idx.Close()
	idx.IndexPage(page)
}

// RemoveNoteFromIndex removes a note from the search index.
func (a *App) RemoveNoteFromIndex(id string) {
	idx, err := search.TryOpen(a.Vault.IndexDir())
	if err != nil || idx == nil {
		return
	}
	defer idx.Close()
	idx.RemoveNote(id)
}

// RemovePageFromIndex removes a page from the search index.
func (a *App) RemovePageFromIndex(id string) {
	idx, err := search.TryOpen(a.Vault.IndexDir())
	if err != nil || idx == nil {
		return
	}
	defer idx.Close()
	idx.RemovePage(id)
}
