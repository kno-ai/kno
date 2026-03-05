package internal_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kno-ai/kno/internal/capture"
	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/vault/fs"
)

func TestIntegration_SetupAndCapture(t *testing.T) {
	vaultDir := t.TempDir()

	v := fs.New(vaultDir, ".")
	if err := v.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout: %v", err)
	}

	// Verify directories were created.
	for _, sub := range []string{"captures", "knowledge"} {
		dir := filepath.Join(vaultDir, sub)
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("expected dir %s to exist: %v", sub, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", sub)
		}
	}

	// Create a capture via the service.
	svc := &capture.Service{Vault: v, MaxBodyBytes: 60000}

	result, err := svc.Create(capture.CreateParams{
		Title:      "Test Capture",
		BodyMD:     "## TL;DR\n\nThis is a test.\n\n## Next steps\n\n- Nothing",
		SourceKind: "stdin",
		SourceTool: "kno_cli",
		Meta:       map[string]string{"topic": "testing", "project": "kno"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if !strings.HasPrefix(result.ID, "cap_") {
		t.Errorf("unexpected ID: %s", result.ID)
	}

	// Verify capture directory structure.
	mdPath := filepath.Join(result.Path, "capture.md")
	metaPath := filepath.Join(result.Path, "meta.json")

	// Check capture.md
	mdContent, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("reading capture.md: %v", err)
	}
	for _, check := range []string{"# Test Capture", "## TL;DR", "This is a test."} {
		if !strings.Contains(string(mdContent), check) {
			t.Errorf("capture.md missing %q", check)
		}
	}
	// Should not have frontmatter.
	if strings.HasPrefix(string(mdContent), "---") {
		t.Error("capture.md should not contain frontmatter")
	}

	// Check meta.json
	metaContent, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("reading meta.json: %v", err)
	}
	var meta model.CaptureMeta
	if err := json.Unmarshal(metaContent, &meta); err != nil {
		t.Fatalf("parsing meta.json: %v", err)
	}
	if meta.ID != result.ID {
		t.Errorf("meta.ID = %q, want %q", meta.ID, result.ID)
	}
	if meta.Title != "Test Capture" {
		t.Errorf("meta.Title = %q", meta.Title)
	}
	if meta.Status != "raw" {
		t.Errorf("meta.Status = %q", meta.Status)
	}
	if meta.Meta["topic"] != "testing" {
		t.Errorf("meta.Meta[topic] = %q", meta.Meta["topic"])
	}
	if meta.Meta["project"] != "kno" {
		t.Errorf("meta.Meta[project] = %q", meta.Meta["project"])
	}
}

func TestIntegration_CaptureWithTruncation(t *testing.T) {
	vaultDir := t.TempDir()
	v := fs.New(vaultDir, ".")
	if err := v.EnsureLayout(); err != nil {
		t.Fatal(err)
	}

	svc := &capture.Service{Vault: v, MaxBodyBytes: 100}

	bigBody := strings.Repeat("x", 200)
	result, err := svc.Create(capture.CreateParams{
		BodyMD:     bigBody,
		SourceKind: "stdin",
		SourceTool: "kno_cli",
	})
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(result.Path, "capture.md"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "[truncated: exceeded max body size]") {
		t.Error("expected truncation notice")
	}
}

func TestIntegration_ListCaptures(t *testing.T) {
	vaultDir := t.TempDir()
	v := fs.New(vaultDir, ".")
	if err := v.EnsureLayout(); err != nil {
		t.Fatal(err)
	}

	svc := &capture.Service{Vault: v, MaxBodyBytes: 60000}

	for _, title := range []string{"First", "Second", "Third"} {
		_, err := svc.Create(capture.CreateParams{
			Title:      title,
			BodyMD:     "## TL;DR\n\ntest\n\n## Next steps\n\n- none",
			SourceKind: "stdin",
			SourceTool: "kno_cli",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	names, err := v.ListCaptures(2)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 2 {
		t.Errorf("expected 2 captures, got %d", len(names))
	}

	// Should be newest first (Third before Second).
	if len(names) >= 2 && !strings.Contains(names[0], "third") {
		t.Errorf("expected newest first, got %q", names[0])
	}
}

func TestIntegration_VaultWriteFileSandboxing(t *testing.T) {
	vaultDir := t.TempDir()
	v := fs.New(vaultDir, ".")
	if err := v.EnsureLayout(); err != nil {
		t.Fatal(err)
	}

	err := v.WriteFile("../../escape.txt", []byte("bad"))
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}
