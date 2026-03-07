package fs

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/vault/sanitize"
)

// Vault is a filesystem-based vault implementation.
type Vault struct {
	root string
}

func New(root string) *Vault {
	return &Vault{root: root}
}

func (v *Vault) Root() string     { return v.root }
func (v *Vault) NotesDir() string { return filepath.Join(v.root, "notes") }
func (v *Vault) PagesDir() string { return filepath.Join(v.root, "pages") }
func (v *Vault) IndexDir() string { return filepath.Join(v.root, "index") }

func (v *Vault) EnsureLayout() error {
	for _, d := range []string{v.NotesDir(), v.PagesDir()} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}
	return nil
}

// --- Note operations ---

func (v *Vault) WriteNote(note model.Note) (string, error) {
	if note.ID == "" {
		note.ID = newNoteID(note.CreatedAt)
	}

	dir, err := sanitize.SafeJoin(v.NotesDir(), note.ID)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating note dir: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "content.md"), []byte(note.Content), 0o644); err != nil {
		return "", fmt.Errorf("writing content.md: %w", err)
	}

	meta := model.NoteMeta{
		ID:        note.ID,
		Title:     note.Title,
		CreatedAt: note.CreatedAt.Format(time.RFC3339),
		Metadata:  note.Metadata,
	}
	if err := writeJSON(filepath.Join(dir, "meta.json"), meta); err != nil {
		return "", err
	}

	return note.ID, nil
}

func (v *Vault) ReadNote(id string) (model.Note, error) {
	dir, err := sanitize.SafeJoin(v.NotesDir(), id)
	if err != nil {
		return model.Note{}, err
	}

	meta, err := readNoteMeta(dir)
	if err != nil {
		return model.Note{}, fmt.Errorf("note %q: %w", id, err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "content.md"))
	if err != nil {
		return model.Note{}, fmt.Errorf("note %q content: %w", id, err)
	}

	createdAt, _ := time.Parse(time.RFC3339, meta.CreatedAt)

	return model.Note{
		ID:        meta.ID,
		CreatedAt: createdAt,
		Title:     meta.Title,
		Content:   string(content),
		Metadata:  meta.Metadata,
	}, nil
}

func (v *Vault) ReadNoteMeta(id string) (model.NoteMeta, error) {
	dir, err := sanitize.SafeJoin(v.NotesDir(), id)
	if err != nil {
		return model.NoteMeta{}, err
	}
	return readNoteMeta(dir)
}

func (v *Vault) UpdateNote(id string, content *string, meta model.MetaMap) error {
	dir, err := sanitize.SafeJoin(v.NotesDir(), id)
	if err != nil {
		return err
	}

	existing, err := readNoteMeta(dir)
	if err != nil {
		return fmt.Errorf("note %q: %w", id, err)
	}

	if content != nil {
		if err := os.WriteFile(filepath.Join(dir, "content.md"), []byte(*content), 0o644); err != nil {
			return fmt.Errorf("writing content: %w", err)
		}
	}

	if meta != nil {
		if existing.Metadata == nil {
			existing.Metadata = make(model.MetaMap)
		}
		existing.Metadata = existing.Metadata.Merge(meta)
	}

	return writeJSON(filepath.Join(dir, "meta.json"), existing)
}

func (v *Vault) ListNotes(limit int) ([]model.NoteMeta, error) {
	entries, err := os.ReadDir(v.NotesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading notes dir: %w", err)
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))

	if limit > 0 && len(dirs) > limit {
		dirs = dirs[:limit]
	}

	metas := make([]model.NoteMeta, 0, len(dirs))
	for _, d := range dirs {
		m, err := readNoteMeta(filepath.Join(v.NotesDir(), d))
		if err != nil {
			continue
		}
		metas = append(metas, m)
	}
	return metas, nil
}

func (v *Vault) DeleteNote(id string) error {
	dir, err := sanitize.SafeJoin(v.NotesDir(), id)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("note %q not found", id)
	}
	return os.RemoveAll(dir)
}

func (v *Vault) CountNotes() (int, error) {
	entries, err := os.ReadDir(v.NotesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			count++
		}
	}
	return count, nil
}

func (v *Vault) OldestDistilledNoteID() (string, error) {
	entries, err := os.ReadDir(v.NotesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs) // oldest first

	for _, d := range dirs {
		m, err := readNoteMeta(filepath.Join(v.NotesDir(), d))
		if err != nil {
			continue
		}
		if m.Metadata.Has("distilled_at") {
			return m.ID, nil
		}
	}
	return "", nil
}

func (v *Vault) OldestNoteID() (string, error) {
	entries, err := os.ReadDir(v.NotesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			return e.Name(), nil // dir names sort oldest first
		}
	}
	return "", nil
}

// --- Page operations ---

// pageContentPath returns the path to a page's content file.
func (v *Vault) pageContentPath(id string) (string, error) {
	return sanitize.SafeJoin(v.PagesDir(), id+".md")
}

// pageMetaPath returns the path to a page's metadata sidecar file.
func (v *Vault) pageMetaPath(id string) (string, error) {
	return sanitize.SafeJoin(v.PagesDir(), id+".meta.json")
}

// isLegacyPage checks if a page exists in the old folder-based layout.
func (v *Vault) isLegacyPage(id string) bool {
	dir, err := sanitize.SafeJoin(v.PagesDir(), id)
	if err != nil {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// migrateLegacyPage converts a page from the old folder layout to the new
// flat layout. Returns true if migration occurred.
func (v *Vault) migrateLegacyPage(id string) bool {
	dir, err := sanitize.SafeJoin(v.PagesDir(), id)
	if err != nil {
		return false
	}

	// Read from old layout
	oldContent := filepath.Join(dir, "content.md")
	oldMeta := filepath.Join(dir, "meta.json")

	content, errC := os.ReadFile(oldContent)
	metaData, errM := os.ReadFile(oldMeta)
	if errC != nil || errM != nil {
		return false
	}

	// Write to new layout
	contentPath, err := v.pageContentPath(id)
	if err != nil {
		return false
	}
	metaPath, err := v.pageMetaPath(id)
	if err != nil {
		return false
	}

	if err := os.WriteFile(contentPath, content, 0o644); err != nil {
		return false
	}
	if err := os.WriteFile(metaPath, metaData, 0o644); err != nil {
		os.Remove(contentPath) // rollback
		return false
	}

	// Remove old directory
	os.RemoveAll(dir)
	return true
}

func (v *Vault) WritePage(page model.Page) (string, error) {
	if page.ID == "" {
		page.ID = sanitize.Slugify(page.Name)
	}

	contentPath, err := v.pageContentPath(page.ID)
	if err != nil {
		return "", err
	}
	metaPath, err := v.pageMetaPath(page.ID)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(contentPath, []byte(page.Content), 0o644); err != nil {
		return "", fmt.Errorf("writing page content: %w", err)
	}

	meta := model.PageMeta{
		ID:        page.ID,
		Name:      page.Name,
		CreatedAt: page.CreatedAt.Format(time.RFC3339),
		Metadata:  page.Metadata,
	}
	if err := writeJSON(metaPath, meta); err != nil {
		return "", err
	}

	return page.ID, nil
}

func (v *Vault) ReadPage(id string) (model.Page, error) {
	// Migrate legacy layout if needed.
	if v.isLegacyPage(id) {
		v.migrateLegacyPage(id)
	}

	metaPath, err := v.pageMetaPath(id)
	if err != nil {
		return model.Page{}, err
	}
	meta, err := readPageMetaFile(metaPath)
	if err != nil {
		return model.Page{}, fmt.Errorf("page %q: %w", id, err)
	}

	contentPath, err := v.pageContentPath(id)
	if err != nil {
		return model.Page{}, err
	}
	content, err := os.ReadFile(contentPath)
	if err != nil {
		return model.Page{}, fmt.Errorf("page %q content: %w", id, err)
	}

	createdAt, _ := time.Parse(time.RFC3339, meta.CreatedAt)

	return model.Page{
		ID:        meta.ID,
		Name:      meta.Name,
		CreatedAt: createdAt,
		Content:   string(content),
		Metadata:  meta.Metadata,
	}, nil
}

func (v *Vault) ReadPageMeta(id string) (model.PageMeta, error) {
	// Migrate legacy layout if needed.
	if v.isLegacyPage(id) {
		v.migrateLegacyPage(id)
	}

	metaPath, err := v.pageMetaPath(id)
	if err != nil {
		return model.PageMeta{}, err
	}
	return readPageMetaFile(metaPath)
}

func (v *Vault) UpdatePage(id string, content *string, meta model.MetaMap) error {
	// Migrate legacy layout if needed.
	if v.isLegacyPage(id) {
		v.migrateLegacyPage(id)
	}

	metaPath, err := v.pageMetaPath(id)
	if err != nil {
		return err
	}
	existing, err := readPageMetaFile(metaPath)
	if err != nil {
		return fmt.Errorf("page %q: %w", id, err)
	}

	if content != nil {
		contentPath, err := v.pageContentPath(id)
		if err != nil {
			return err
		}
		if err := os.WriteFile(contentPath, []byte(*content), 0o644); err != nil {
			return fmt.Errorf("writing content: %w", err)
		}
	}

	if meta != nil {
		if existing.Metadata == nil {
			existing.Metadata = make(model.MetaMap)
		}
		existing.Metadata = existing.Metadata.Merge(meta)
	}

	return writeJSON(metaPath, existing)
}

func (v *Vault) ListPages() ([]model.PageMeta, error) {
	entries, err := os.ReadDir(v.PagesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading pages dir: %w", err)
	}

	// Migrate any legacy pages first.
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			v.migrateLegacyPage(e.Name())
		}
	}

	// Re-read after migration.
	entries, err = os.ReadDir(v.PagesDir())
	if err != nil {
		return nil, fmt.Errorf("reading pages dir: %w", err)
	}

	var metas []model.PageMeta
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".meta.json") {
			continue
		}
		metaPath := filepath.Join(v.PagesDir(), name)
		m, err := readPageMetaFile(metaPath)
		if err != nil {
			continue
		}
		metas = append(metas, m)
	}
	return metas, nil
}

func (v *Vault) DeletePage(id string) error {
	// Handle legacy layout.
	if v.isLegacyPage(id) {
		dir, err := sanitize.SafeJoin(v.PagesDir(), id)
		if err != nil {
			return err
		}
		return os.RemoveAll(dir)
	}

	contentPath, err := v.pageContentPath(id)
	if err != nil {
		return err
	}
	metaPath, err := v.pageMetaPath(id)
	if err != nil {
		return err
	}

	if _, err := os.Stat(contentPath); os.IsNotExist(err) {
		if _, err := os.Stat(metaPath); os.IsNotExist(err) {
			return fmt.Errorf("page %q not found", id)
		}
	}

	os.Remove(contentPath)
	os.Remove(metaPath)
	return nil
}

// --- helpers ---

func newNoteID(t time.Time) string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		b = []byte{0, 0, 0}
	}
	return fmt.Sprintf("%s-%s", t.Format("20060102T150405Z0700"), hex.EncodeToString(b))
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling JSON: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func readNoteMeta(dir string) (model.NoteMeta, error) {
	data, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		return model.NoteMeta{}, err
	}
	var meta model.NoteMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return model.NoteMeta{}, err
	}
	return meta, nil
}

func readPageMetaFile(path string) (model.PageMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.PageMeta{}, err
	}
	var meta model.PageMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return model.PageMeta{}, err
	}
	return meta, nil
}
