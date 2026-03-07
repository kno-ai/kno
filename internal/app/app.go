package app

import (
	"fmt"

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

// --- Token estimation ---

// EstimateTokens returns a rough token count (~4 chars per token).
func EstimateTokens(s string) int {
	return (len(s) + 3) / 4
}

// --- Content validation ---

// ValidateNoteContent checks note content against notes.max_content_tokens.
func (a *App) ValidateNoteContent(content string) error {
	max := a.Config.Notes.MaxContentTokens
	if max <= 0 {
		return nil
	}
	est := EstimateTokens(content)
	if est > max {
		return fmt.Errorf("content too large: ~%d tokens exceeds notes.max_content_tokens (%d)", est, max)
	}
	return nil
}

// ValidatePageContent checks page content against pages.max_content_tokens.
func (a *App) ValidatePageContent(content string) error {
	max := a.Config.Pages.MaxContentTokens
	if max <= 0 {
		return nil
	}
	est := EstimateTokens(content)
	if est > max {
		return fmt.Errorf("content too large: ~%d tokens exceeds pages.max_content_tokens (%d)", est, max)
	}
	return nil
}

// --- Auto-removal ---

// AutoRemoveResult describes a note that was auto-removed to make room.
type AutoRemoveResult struct {
	ID        string
	Title     string
	Uncurated bool
}

// AutoRemoveOldestNote removes the oldest note if the vault is at capacity.
// Returns nil if no removal was needed.
func (a *App) AutoRemoveOldestNote() (*AutoRemoveResult, error) {
	count, err := a.Vault.CountNotes()
	if err != nil {
		return nil, err
	}
	if count < a.Config.Notes.MaxCount {
		return nil, nil
	}

	oldest, err := a.Vault.OldestCuratedNoteID()
	if err != nil {
		return nil, err
	}
	uncurated := false
	if oldest == "" {
		oldest, err = a.Vault.OldestNoteID()
		if err != nil {
			return nil, err
		}
		if oldest == "" {
			return nil, fmt.Errorf("vault at capacity (%d notes) with nothing to remove", a.Config.Notes.MaxCount)
		}
		uncurated = true
	}

	var title string
	if rm, err := a.Vault.ReadNoteMeta(oldest); err == nil {
		title = rm.Title
	}

	if err := a.Vault.DeleteNote(oldest); err != nil {
		return nil, fmt.Errorf("auto-removing note: %w", err)
	}
	a.RemoveNoteFromIndex(oldest)

	return &AutoRemoveResult{ID: oldest, Title: title, Uncurated: uncurated}, nil
}

// --- Search index ---

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
