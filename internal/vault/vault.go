package vault

import "github.com/kno-ai/kno/internal/model"

// Vault is the interface for reading/writing notes and pages.
type Vault interface {
	// Root returns the absolute path to the vault root.
	Root() string

	// EnsureLayout creates the vault directory structure if missing.
	EnsureLayout() error

	// Note operations

	WriteNote(note model.Note) (string, error) // returns ID
	ReadNote(id string) (model.Note, error)
	ReadNoteMeta(id string) (model.NoteMeta, error)
	UpdateNote(id string, content *string, meta model.MetaMap) error
	ListNotes(limit int) ([]model.NoteMeta, error)
	DeleteNote(id string) error
	CountNotes() (total int, err error)
	OldestCuratedNoteID() (string, error) // for auto-removal (curated first)
	OldestNoteID() (string, error)        // for auto-removal fallback (any note)

	// Page operations

	WritePage(page model.Page) (string, error) // returns ID
	ReadPage(id string) (model.Page, error)
	ReadPageMeta(id string) (model.PageMeta, error)
	UpdatePage(id string, content *string, meta model.MetaMap) error
	RenamePage(oldID, newName string) (newID string, err error)
	ListPages() ([]model.PageMeta, error)
	DeletePage(id string) error

	// Paths
	NotesDir() string
	PagesDir() string
	IndexDir() string
}
