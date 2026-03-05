package vault

import "github.com/kno-ai/kno/internal/model"

// Vault is the interface for reading/writing kno-managed files in a vault.
type Vault interface {
	// Root returns the absolute path to the vault root.
	Root() string

	// KnoDir returns the absolute path to the kno subdirectory.
	KnoDir() string

	// EnsureLayout creates the kno directory structure if missing.
	EnsureLayout() error

	// WriteCapture writes a capture note and returns the result.
	WriteCapture(note model.CaptureNote) (model.CaptureWriteResult, error)

	// WriteFile writes content to a path relative to the kno subdirectory.
	// The path is validated to prevent traversal outside the kno dir.
	WriteFile(relPath string, content []byte) error

	// ReadFile reads a file at a path relative to the kno subdirectory.
	ReadFile(relPath string) ([]byte, error)

	// ListCaptures returns capture directory names sorted newest-first.
	ListCaptures(limit int) ([]string, error)

	// ReadCapture reads a capture's meta.json and capture.md by directory name.
	ReadCapture(dirName string) (model.CaptureMeta, string, error)
}
