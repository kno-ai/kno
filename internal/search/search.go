package search

import (
	"fmt"
	"os"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/vault"
)

type Result struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
	Kind  string  `json:"kind"` // "note" or "page"
}

type Index struct {
	index bleve.Index
	path  string
}

// Open opens or creates the bleve index at the given path.
func Open(path string) (*Index, error) {
	idx, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		mapping := bleve.NewIndexMapping()
		idx, err = bleve.New(path, mapping)
		if err != nil {
			return nil, fmt.Errorf("creating index: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("opening index: %w", err)
	}
	return &Index{index: idx, path: path}, nil
}

// TryOpen opens an existing index or creates one if it doesn't exist.
// Returns nil (no error) only on unexpected failures, so callers can
// treat a nil result as "indexing unavailable" without crashing.
func TryOpen(path string) (*Index, error) {
	return Open(path)
}

// Close closes the index.
func (idx *Index) Close() error {
	return idx.index.Close()
}

type indexDoc struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Meta    string `json:"meta"` // flattened metadata for search
}

// IndexNote adds or updates a note in the search index.
func (idx *Index) IndexNote(note model.Note) error {
	doc := indexDoc{
		ID:      note.ID,
		Kind:    "note",
		Title:   note.Title,
		Content: note.Content,
		Meta:    flattenMeta(note.Metadata),
	}
	return idx.index.Index("note:"+note.ID, doc)
}

// IndexPage adds or updates a page in the search index.
func (idx *Index) IndexPage(page model.Page) error {
	doc := indexDoc{
		ID:      page.ID,
		Kind:    "page",
		Title:   page.Name,
		Content: page.Content,
		Meta:    flattenMeta(page.Metadata),
	}
	return idx.index.Index("page:"+page.ID, doc)
}

// RemoveNote removes a note from the index.
func (idx *Index) RemoveNote(id string) error {
	return idx.index.Delete("note:" + id)
}

// RemovePage removes a page from the index.
func (idx *Index) RemovePage(id string) error {
	return idx.index.Delete("page:" + id)
}

// SearchNotes searches for notes matching the query.
func (idx *Index) SearchNotes(query string, limit int) ([]Result, error) {
	return idx.search(query, "note", limit)
}

// SearchPages searches for pages matching the query.
func (idx *Index) SearchPages(query string, limit int) ([]Result, error) {
	return idx.search(query, "page", limit)
}

func (idx *Index) search(queryStr, kind string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 10
	}

	q := bleve.NewQueryStringQuery(queryStr)
	req := bleve.NewSearchRequestOptions(q, limit, 0, false)

	sr, err := idx.index.Search(req)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var results []Result
	for _, hit := range sr.Hits {
		var hitKind, hitID string
		if strings.HasPrefix(hit.ID, "note:") {
			hitKind = "note"
			hitID = hit.ID[5:]
		} else if strings.HasPrefix(hit.ID, "page:") {
			hitKind = "page"
			hitID = hit.ID[5:]
		} else {
			continue
		}

		if kind != "" && hitKind != kind {
			continue
		}

		results = append(results, Result{
			ID:    hitID,
			Score: hit.Score,
			Kind:  hitKind,
		})
	}

	return results, nil
}

// Rebuild drops and recreates the index from vault contents.
func Rebuild(v vault.Vault) (*Index, error) {
	indexPath := v.IndexDir()

	// Remove existing index
	os.RemoveAll(indexPath)

	idx, err := Open(indexPath)
	if err != nil {
		return nil, err
	}

	// Index all notes
	notes, err := v.ListNotes(0)
	if err != nil {
		idx.Close()
		return nil, fmt.Errorf("listing notes: %w", err)
	}
	for _, nm := range notes {
		note, err := v.ReadNote(nm.ID)
		if err != nil {
			continue
		}
		if err := idx.IndexNote(note); err != nil {
			continue
		}
	}

	// Index all pages
	pages, err := v.ListPages()
	if err != nil {
		idx.Close()
		return nil, fmt.Errorf("listing pages: %w", err)
	}
	for _, pm := range pages {
		page, err := v.ReadPage(pm.ID)
		if err != nil {
			continue
		}
		if err := idx.IndexPage(page); err != nil {
			continue
		}
	}

	return idx, nil
}

func flattenMeta(m model.MetaMap) string {
	if m == nil {
		return ""
	}
	var parts []string
	for k, vs := range m {
		for _, v := range vs {
			parts = append(parts, k+":"+v)
		}
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}
