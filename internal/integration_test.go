package internal_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kno-ai/kno/internal/config"
	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/vault/fs"
)

func setupVault(t *testing.T) *fs.Vault {
	t.Helper()
	dir := t.TempDir()
	v := fs.New(dir)
	if err := v.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout: %v", err)
	}
	return v
}

func TestIntegration_VaultLayout(t *testing.T) {
	v := setupVault(t)

	for _, sub := range []string{"notes", "pages"} {
		dir := filepath.Join(v.Root(), sub)
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("expected dir %s to exist: %v", sub, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", sub)
		}
	}
}

func TestIntegration_NoteCreateAndRead(t *testing.T) {
	v := setupVault(t)

	meta := make(model.MetaMap)
	meta.Add("tags", "aws")
	meta.Add("tags", "sqs")
	meta.Set("summary", "test note")

	note := model.Note{
		CreatedAt: time.Now(),
		Title:     "Test Note",
		Content:   "## TL;DR\n\nThis is a test.\n",
		Metadata:  meta,
	}

	id, err := v.WriteNote(note)
	if err != nil {
		t.Fatalf("WriteNote: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	// Read it back
	got, err := v.ReadNote(id)
	if err != nil {
		t.Fatalf("ReadNote: %v", err)
	}
	if got.Title != "Test Note" {
		t.Errorf("Title = %q, want %q", got.Title, "Test Note")
	}
	if got.Content != "## TL;DR\n\nThis is a test.\n" {
		t.Errorf("Content = %q", got.Content)
	}
	if len(got.Metadata["tags"]) != 2 {
		t.Errorf("expected 2 tags, got %v", got.Metadata["tags"])
	}

	// Read meta only
	nm, err := v.ReadNoteMeta(id)
	if err != nil {
		t.Fatalf("ReadNoteMeta: %v", err)
	}
	if nm.ID != id {
		t.Errorf("meta ID = %q, want %q", nm.ID, id)
	}
}

func TestIntegration_NoteListOrder(t *testing.T) {
	v := setupVault(t)

	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for i, title := range []string{"First", "Second", "Third"} {
		_, err := v.WriteNote(model.Note{
			CreatedAt: base.Add(time.Duration(i) * time.Hour),
			Title:     title,
			Content:   "test",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	metas, err := v.ListNotes(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 2 {
		t.Fatalf("expected 2, got %d", len(metas))
	}
	if metas[0].Title != "Third" {
		t.Errorf("expected newest first, got %q", metas[0].Title)
	}
}

func TestIntegration_NoteUpdate(t *testing.T) {
	v := setupVault(t)

	id, err := v.WriteNote(model.Note{
		CreatedAt: time.Now(),
		Title:     "Original",
		Content:   "original content",
	})
	if err != nil {
		t.Fatal(err)
	}

	newContent := "updated content"
	newMeta := make(model.MetaMap)
	newMeta.Set("distilled_at", "2026-03-06T12:00:00Z")

	if err := v.UpdateNote(id, &newContent, newMeta); err != nil {
		t.Fatal(err)
	}

	got, err := v.ReadNote(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != "updated content" {
		t.Errorf("Content = %q", got.Content)
	}
	if got.Metadata.Get("distilled_at") != "2026-03-06T12:00:00Z" {
		t.Errorf("distilled_at = %q", got.Metadata.Get("distilled_at"))
	}
}

func TestIntegration_NoteDelete(t *testing.T) {
	v := setupVault(t)

	id, err := v.WriteNote(model.Note{
		CreatedAt: time.Now(),
		Title:     "To Delete",
		Content:   "bye",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := v.DeleteNote(id); err != nil {
		t.Fatal(err)
	}

	_, err = v.ReadNote(id)
	if err == nil {
		t.Fatal("expected error reading deleted note")
	}
}

func TestIntegration_NoteCapacity(t *testing.T) {
	v := setupVault(t)

	// Create note, mark it distilled
	id, err := v.WriteNote(model.Note{
		CreatedAt: time.Now(),
		Title:     "Old Distilled",
		Content:   "old",
	})
	if err != nil {
		t.Fatal(err)
	}

	distMeta := make(model.MetaMap)
	distMeta.Set("distilled_at", "2026-01-01T00:00:00Z")
	if err := v.UpdateNote(id, nil, distMeta); err != nil {
		t.Fatal(err)
	}

	// Verify OldestDistilledNoteID finds it
	oldest, err := v.OldestDistilledNoteID()
	if err != nil {
		t.Fatal(err)
	}
	if oldest != id {
		t.Errorf("OldestDistilledNoteID = %q, want %q", oldest, id)
	}
}

func TestIntegration_PageCreateAndRead(t *testing.T) {
	v := setupVault(t)

	page := model.Page{
		Name:      "AWS Infrastructure",
		CreatedAt: time.Now(),
		Content:   "# AWS Infrastructure\n\nGuidance: focus on networking.\n",
	}

	id, err := v.WritePage(page)
	if err != nil {
		t.Fatalf("WritePage: %v", err)
	}
	if id != "aws-infrastructure" {
		t.Errorf("ID = %q, want %q", id, "aws-infrastructure")
	}

	got, err := v.ReadPage(id)
	if err != nil {
		t.Fatalf("ReadPage: %v", err)
	}
	if got.Name != "AWS Infrastructure" {
		t.Errorf("Name = %q", got.Name)
	}
	if !strings.Contains(got.Content, "focus on networking") {
		t.Errorf("Content = %q", got.Content)
	}
}

func TestIntegration_PageUpdate(t *testing.T) {
	v := setupVault(t)

	id, err := v.WritePage(model.Page{
		Name:      "Auth System",
		CreatedAt: time.Now(),
		Content:   "original",
	})
	if err != nil {
		t.Fatal(err)
	}

	newContent := "updated auth content"
	if err := v.UpdatePage(id, &newContent, nil); err != nil {
		t.Fatal(err)
	}

	got, err := v.ReadPage(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != "updated auth content" {
		t.Errorf("Content = %q", got.Content)
	}
}

func TestIntegration_PageList(t *testing.T) {
	v := setupVault(t)

	for _, name := range []string{"Page A", "Page B"} {
		_, err := v.WritePage(model.Page{
			Name:      name,
			CreatedAt: time.Now(),
			Content:   "content",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	metas, err := v.ListPages()
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 2 {
		t.Errorf("expected 2 pages, got %d", len(metas))
	}
}

func TestIntegration_PageDelete(t *testing.T) {
	v := setupVault(t)

	id, err := v.WritePage(model.Page{
		Name:      "To Delete",
		CreatedAt: time.Now(),
		Content:   "bye",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := v.DeletePage(id); err != nil {
		t.Fatal(err)
	}

	_, err = v.ReadPage(id)
	if err == nil {
		t.Fatal("expected error reading deleted page")
	}
}

func TestIntegration_MetaMapJSON(t *testing.T) {
	m := make(model.MetaMap)
	m.Set("summary", "one value")
	m.Add("tags", "aws")
	m.Add("tags", "sqs")

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	// summary should be scalar, tags should be array
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	if _, ok := raw["summary"].(string); !ok {
		t.Errorf("summary should be string, got %T", raw["summary"])
	}
	if _, ok := raw["tags"].([]any); !ok {
		t.Errorf("tags should be array, got %T", raw["tags"])
	}

	// Round-trip
	var m2 model.MetaMap
	if err := json.Unmarshal(data, &m2); err != nil {
		t.Fatal(err)
	}
	if m2.Get("summary") != "one value" {
		t.Errorf("summary = %q", m2.Get("summary"))
	}
	if len(m2["tags"]) != 2 {
		t.Errorf("tags = %v", m2["tags"])
	}
}

func TestIntegration_MetaMapFilter(t *testing.T) {
	m := make(model.MetaMap)
	m.Set("status", "active")
	m.Add("tags", "aws")
	m.Add("tags", "sqs")

	// Match scalar
	if !m.MatchesFilter(map[string]string{"status": "active"}) {
		t.Error("should match status=active")
	}
	if m.MatchesFilter(map[string]string{"status": "inactive"}) {
		t.Error("should not match status=inactive")
	}

	// Match array (contains)
	if !m.MatchesFilter(map[string]string{"tags": "aws"}) {
		t.Error("should match tags=aws")
	}

	// Match null (absent)
	if !m.MatchesFilter(map[string]string{"distilled_at": "null"}) {
		t.Error("should match distilled_at=null when absent")
	}
	if m.MatchesFilter(map[string]string{"status": "null"}) {
		t.Error("should not match status=null when present")
	}
}

func TestIntegration_MetaMapMerge(t *testing.T) {
	m := make(model.MetaMap)
	m.Set("a", "1")
	m.Set("b", "2")

	other := make(model.MetaMap)
	other.Set("b", "3")
	other["a"] = nil // remove

	result := m.Merge(other)
	if result.Has("a") {
		t.Error("a should be removed")
	}
	if result.Get("b") != "3" {
		t.Errorf("b = %q, want 3", result.Get("b"))
	}
}

func TestIntegration_Config(t *testing.T) {
	dir := t.TempDir()

	cfg := config.DefaultConfig()
	if err := config.Save(dir, cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Notes.MaxCount != 200 {
		t.Errorf("MaxCount = %d", loaded.Notes.MaxCount)
	}
	if loaded.Pages.MaxContentTokens != 8000 {
		t.Errorf("MaxContentTokens = %d", loaded.Pages.MaxContentTokens)
	}
}

func TestIntegration_PathTraversal(t *testing.T) {
	v := setupVault(t)
	_, err := v.ReadNote("../../etc/passwd")
	if err == nil {
		t.Fatal("expected path traversal error")
	}
}

func TestIntegration_CapacityAutoRemoval(t *testing.T) {
	dir := t.TempDir()
	v := fs.New(dir)
	if err := v.EnsureLayout(); err != nil {
		t.Fatal(err)
	}

	// Write config with max_count=3
	cfg := config.DefaultConfig()
	cfg.Notes.MaxCount = 3
	if err := config.Save(dir, cfg); err != nil {
		t.Fatal(err)
	}

	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create 3 notes, mark first as distilled
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		id, err := v.WriteNote(model.Note{
			CreatedAt: base.Add(time.Duration(i) * time.Hour),
			Title:     fmt.Sprintf("Note %d", i),
			Content:   "test",
		})
		if err != nil {
			t.Fatal(err)
		}
		ids[i] = id
	}

	// Mark first as distilled
	distMeta := make(model.MetaMap)
	distMeta.Set("distilled_at", "2026-01-01T00:00:00Z")
	if err := v.UpdateNote(ids[0], nil, distMeta); err != nil {
		t.Fatal(err)
	}

	// Verify at capacity
	count, _ := v.CountNotes()
	if count != 3 {
		t.Fatalf("expected 3 notes, got %d", count)
	}

	// OldestDistilledNoteID should return first note
	oldest, err := v.OldestDistilledNoteID()
	if err != nil {
		t.Fatal(err)
	}
	if oldest != ids[0] {
		t.Errorf("oldest distilled = %q, want %q", oldest, ids[0])
	}

	// Delete it (simulating auto-removal)
	if err := v.DeleteNote(oldest); err != nil {
		t.Fatal(err)
	}

	count, _ = v.CountNotes()
	if count != 2 {
		t.Fatalf("expected 2 notes after removal, got %d", count)
	}

	// No more distilled notes
	oldest, err = v.OldestDistilledNoteID()
	if err != nil {
		t.Fatal(err)
	}
	if oldest != "" {
		t.Errorf("expected no distilled notes, got %q", oldest)
	}
}

func TestIntegration_PageDeleteNoteCleanup(t *testing.T) {
	v := setupVault(t)

	// Create two pages
	pageA, err := v.WritePage(model.Page{Name: "Page A", CreatedAt: time.Now(), Content: "a"})
	if err != nil {
		t.Fatal(err)
	}
	pageB, err := v.WritePage(model.Page{Name: "Page B", CreatedAt: time.Now(), Content: "b"})
	if err != nil {
		t.Fatal(err)
	}

	// Create notes with various distilled_into states
	noteSingle, err := v.WriteNote(model.Note{
		CreatedAt: time.Now(), Title: "Single ref", Content: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	noteMulti, err := v.WriteNote(model.Note{
		CreatedAt: time.Now(), Title: "Multi ref", Content: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	noteUnrelated, err := v.WriteNote(model.Note{
		CreatedAt: time.Now(), Title: "Unrelated", Content: "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Mark noteSingle as distilled into pageA only
	singleMeta := make(model.MetaMap)
	singleMeta.Set("distilled_at", "2026-01-01T00:00:00Z")
	singleMeta.Set("distilled_into", pageA)
	if err := v.UpdateNote(noteSingle, nil, singleMeta); err != nil {
		t.Fatal(err)
	}

	// Mark noteMulti as distilled into both pages
	multiMeta := make(model.MetaMap)
	multiMeta.Set("distilled_at", "2026-01-01T00:00:00Z")
	multiMeta.Add("distilled_into", pageA)
	multiMeta.Add("distilled_into", pageB)
	if err := v.UpdateNote(noteMulti, nil, multiMeta); err != nil {
		t.Fatal(err)
	}

	// Delete pageA and simulate the cleanup
	if err := v.DeletePage(pageA); err != nil {
		t.Fatal(err)
	}

	// Simulate admin page delete cleanup
	allNotes, _ := v.ListNotes(0)
	for _, nm := range allNotes {
		if nm.Metadata == nil || !nm.Metadata.Has("distilled_into") {
			continue
		}
		vals := nm.Metadata["distilled_into"]
		var remaining []string
		for _, val := range vals {
			if val != pageA {
				remaining = append(remaining, val)
			}
		}
		if len(remaining) == len(vals) {
			continue
		}
		updateMeta := make(model.MetaMap)
		if len(remaining) == 0 {
			updateMeta["distilled_into"] = nil
			updateMeta["distilled_at"] = nil
		} else {
			updateMeta["distilled_into"] = remaining
		}
		v.UpdateNote(nm.ID, nil, updateMeta)
	}

	// noteSingle should have distilled_into and distilled_at removed (re-eligible)
	single, err := v.ReadNote(noteSingle)
	if err != nil {
		t.Fatal(err)
	}
	if single.Metadata.Has("distilled_into") {
		t.Error("noteSingle should have distilled_into removed")
	}
	if single.Metadata.Has("distilled_at") {
		t.Error("noteSingle should have distilled_at removed")
	}

	// noteMulti should still have distilled_into=[pageB]
	multi, err := v.ReadNote(noteMulti)
	if err != nil {
		t.Fatal(err)
	}
	if !multi.Metadata.Has("distilled_into") {
		t.Fatal("noteMulti should still have distilled_into")
	}
	if len(multi.Metadata["distilled_into"]) != 1 || multi.Metadata["distilled_into"][0] != pageB {
		t.Errorf("noteMulti distilled_into = %v, want [%s]", multi.Metadata["distilled_into"], pageB)
	}

	// noteUnrelated should be unchanged
	unrelated, err := v.ReadNote(noteUnrelated)
	if err != nil {
		t.Fatal(err)
	}
	if unrelated.Metadata.Has("distilled_into") {
		t.Error("noteUnrelated should not have distilled_into")
	}

}

func TestIntegration_NoteUpdatePreservesExistingMeta(t *testing.T) {
	v := setupVault(t)

	meta := make(model.MetaMap)
	meta.Set("original_key", "original_value")
	meta.Set("to_change", "old")

	id, err := v.WriteNote(model.Note{
		CreatedAt: time.Now(),
		Title:     "Test",
		Content:   "test",
		Metadata:  meta,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Update only one key
	updateMeta := make(model.MetaMap)
	updateMeta.Set("to_change", "new")
	updateMeta.Set("added_key", "added")

	if err := v.UpdateNote(id, nil, updateMeta); err != nil {
		t.Fatal(err)
	}

	got, err := v.ReadNote(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Get("original_key") != "original_value" {
		t.Errorf("original_key = %q, want original_value", got.Metadata.Get("original_key"))
	}
	if got.Metadata.Get("to_change") != "new" {
		t.Errorf("to_change = %q, want new", got.Metadata.Get("to_change"))
	}
	if got.Metadata.Get("added_key") != "added" {
		t.Errorf("added_key = %q, want added", got.Metadata.Get("added_key"))
	}
}

func TestIntegration_VaultNonexistentOperations(t *testing.T) {
	v := setupVault(t)

	// Read nonexistent note
	_, err := v.ReadNote("nonexistent")
	if err == nil {
		t.Error("expected error reading nonexistent note")
	}

	// Update nonexistent note
	content := "new"
	err = v.UpdateNote("nonexistent", &content, nil)
	if err == nil {
		t.Error("expected error updating nonexistent note")
	}

	// Delete nonexistent note
	err = v.DeleteNote("nonexistent")
	if err == nil {
		t.Error("expected error deleting nonexistent note")
	}

	// Read nonexistent page
	_, err = v.ReadPage("nonexistent")
	if err == nil {
		t.Error("expected error reading nonexistent page")
	}

	// Delete nonexistent page
	err = v.DeletePage("nonexistent")
	if err == nil {
		t.Error("expected error deleting nonexistent page")
	}
}
