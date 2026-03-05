package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kno-ai/kno/internal/capture"
	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/vault/sanitize"
)

// Vault is a filesystem-based vault adapter.
type Vault struct {
	root      string
	knoSubdir string
}

// New creates a new filesystem Vault adapter.
func New(root, knoSubdir string) *Vault {
	return &Vault{
		root:      root,
		knoSubdir: knoSubdir,
	}
}

func (v *Vault) Root() string {
	return v.root
}

func (v *Vault) KnoDir() string {
	return filepath.Join(v.root, v.knoSubdir)
}

func (v *Vault) capturesDir() string {
	return filepath.Join(v.KnoDir(), "captures")
}

// EnsureLayout creates the kno directory structure if missing.
func (v *Vault) EnsureLayout() error {
	dirs := []string{
		filepath.Join(v.KnoDir(), "captures"),
		filepath.Join(v.KnoDir(), "knowledge"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}
	return nil
}

// WriteCapture creates a capture directory with capture.md and meta.json.
func (v *Vault) WriteCapture(note model.CaptureNote) (model.CaptureWriteResult, error) {
	dirName := capture.DirName(note)

	// Verify the path stays within captures dir.
	dirPath, err := sanitize.SafeJoin(v.capturesDir(), dirName)
	if err != nil {
		return model.CaptureWriteResult{}, err
	}

	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return model.CaptureWriteResult{}, fmt.Errorf("creating capture dir: %w", err)
	}

	// Write capture.md
	md := capture.RenderMarkdown(note)
	mdPath := filepath.Join(dirPath, "capture.md")
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		return model.CaptureWriteResult{}, fmt.Errorf("writing capture.md: %w", err)
	}

	// Write meta.json
	meta := capture.RenderMeta(note)
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return model.CaptureWriteResult{}, fmt.Errorf("marshalling meta: %w", err)
	}
	metaPath := filepath.Join(dirPath, "meta.json")
	if err := os.WriteFile(metaPath, metaData, 0o644); err != nil {
		return model.CaptureWriteResult{}, fmt.Errorf("writing meta.json: %w", err)
	}

	return model.CaptureWriteResult{
		Path:    dirPath,
		ID:      note.ID,
		Created: note.CreatedAt,
	}, nil
}

// WriteFile writes content to a path relative to the kno subdirectory.
func (v *Vault) WriteFile(relPath string, content []byte) error {
	dest, err := sanitize.SafeJoin(v.KnoDir(), relPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("creating parent dirs: %w", err)
	}
	return os.WriteFile(dest, content, 0o644)
}

// ReadFile reads a file at a path relative to the kno subdirectory.
func (v *Vault) ReadFile(relPath string) ([]byte, error) {
	src, err := sanitize.SafeJoin(v.KnoDir(), relPath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(src)
}

// ListCaptures returns capture directory names sorted newest-first.
// If limit <= 0, all captures are returned.
func (v *Vault) ListCaptures(limit int) ([]string, error) {
	entries, err := os.ReadDir(v.capturesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading captures dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}

	// Directory names are timestamp-prefixed, so lexicographic = chronological.
	sort.Sort(sort.Reverse(sort.StringSlice(names)))

	if limit > 0 && len(names) > limit {
		names = names[:limit]
	}

	return names, nil
}

// ReadCapture reads a capture's meta.json and capture.md by directory name.
func (v *Vault) ReadCapture(dirName string) (model.CaptureMeta, string, error) {
	dirPath, err := sanitize.SafeJoin(v.capturesDir(), dirName)
	if err != nil {
		return model.CaptureMeta{}, "", err
	}

	metaData, err := os.ReadFile(filepath.Join(dirPath, "meta.json"))
	if err != nil {
		return model.CaptureMeta{}, "", fmt.Errorf("reading meta.json: %w", err)
	}

	var meta model.CaptureMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return model.CaptureMeta{}, "", fmt.Errorf("parsing meta.json: %w", err)
	}

	mdData, err := os.ReadFile(filepath.Join(dirPath, "capture.md"))
	if err != nil {
		return model.CaptureMeta{}, "", fmt.Errorf("reading capture.md: %w", err)
	}

	return meta, string(mdData), nil
}
