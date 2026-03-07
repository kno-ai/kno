package search

import (
	"testing"
	"time"

	"github.com/kno-ai/kno/internal/model"
)

func TestSearchNotes(t *testing.T) {
	dir := t.TempDir()
	idx, err := Open(dir + "/test.bleve")
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	meta := make(model.MetaMap)
	meta.Set("tags", "aws")

	err = idx.IndexNote(model.Note{
		ID:        "note-001",
		CreatedAt: time.Now(),
		Title:     "SQS retry strategy",
		Content:   "We decided to use exponential backoff for SQS message retries.",
		Metadata:  meta,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = idx.IndexNote(model.Note{
		ID:        "note-002",
		CreatedAt: time.Now(),
		Title:     "React auth flow",
		Content:   "Implemented OAuth2 with refresh tokens.",
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := idx.SearchNotes("SQS retry", 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].ID != "note-001" {
		t.Errorf("expected note-001 first, got %s", results[0].ID)
	}
}

func TestSearchPages(t *testing.T) {
	dir := t.TempDir()
	idx, err := Open(dir + "/test.bleve")
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	err = idx.IndexPage(model.Page{
		ID:      "aws-infrastructure",
		Name:    "AWS Infrastructure",
		Content: "Focus on VPC, subnets, and security groups.",
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := idx.SearchPages("VPC subnet", 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].ID != "aws-infrastructure" {
		t.Errorf("expected aws-infrastructure, got %s", results[0].ID)
	}
}

func TestSearchRemove(t *testing.T) {
	dir := t.TempDir()
	idx, err := Open(dir + "/test.bleve")
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	err = idx.IndexNote(model.Note{
		ID:      "note-remove",
		Title:   "To Remove",
		Content: "unique-searchable-content",
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := idx.SearchNotes("unique-searchable-content", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("expected result before removal")
	}

	if err := idx.RemoveNote("note-remove"); err != nil {
		t.Fatal(err)
	}

	results, err = idx.SearchNotes("unique-searchable-content", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results after removal, got %d", len(results))
	}
}
