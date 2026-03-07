package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/search"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerNoteTools(s *server.MCPServer, a *app.App) {
	// kno_note_create
	s.AddTool(mcp.NewTool("kno_note_create",
		mcp.WithDescription("Create a new note in the vault."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title for the note.")),
		mcp.WithString("content", mcp.Required(), mcp.Description("The note content (markdown).")),
		mcp.WithObject("meta", mcp.Description("Optional metadata key-value pairs.")),
	), noteCreateHandler(a))

	// kno_note_list
	s.AddTool(mcp.NewTool("kno_note_list",
		mcp.WithDescription("List notes from the vault."),
		mcp.WithNumber("limit", mcp.Description("Maximum number of notes to return.")),
		mcp.WithObject("filter", mcp.Description("Filter criteria as key-value pairs. Use null value for absent keys.")),
	), noteListHandler(a))

	// kno_note_show
	s.AddTool(mcp.NewTool("kno_note_show",
		mcp.WithDescription("Read one or more notes by ID."),
		mcp.WithString("id", mcp.Description("Single note ID to read.")),
		mcp.WithArray("ids", mcp.Description("Multiple note IDs to read.")),
	), noteShowHandler(a))

	// kno_note_update
	s.AddTool(mcp.NewTool("kno_note_update",
		mcp.WithDescription("Update a note's content or metadata."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Note ID to update.")),
		mcp.WithString("content", mcp.Description("New content (replaces existing).")),
		mcp.WithObject("meta", mcp.Description("Metadata key-value pairs to set/update.")),
	), noteUpdateHandler(a))

	// kno_note_delete
	s.AddTool(mcp.NewTool("kno_note_delete",
		mcp.WithDescription("Permanently delete a saved session by ID."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Note ID to delete.")),
	), noteDeleteHandler(a))

	// kno_note_search
	s.AddTool(mcp.NewTool("kno_note_search",
		mcp.WithDescription("Search notes by text query."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query.")),
		mcp.WithNumber("limit", mcp.Description("Maximum results.")),
		mcp.WithObject("filter", mcp.Description("Filter criteria.")),
	), noteSearchHandler(a))
}

func noteCreateHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError("title is required"), nil
		}
		content, err := req.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError("content is required"), nil
		}

		meta := extractMeta(req.GetArguments(), "meta")

		// Check capacity
		count, err := a.Vault.CountNotes()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("counting notes: %v", err)), nil
		}

		var autoRemoved string
		var autoRemovedUncurated bool
		if count >= a.Config.Notes.MaxCount {
			oldest, err := a.Vault.OldestCuratedNoteID()
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("finding removable note: %v", err)), nil
			}
			if oldest == "" {
				// No curated notes — fall back to oldest note overall
				oldest, err = a.Vault.OldestNoteID()
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("finding oldest note: %v", err)), nil
				}
				if oldest == "" {
					return mcp.NewToolResultError(fmt.Sprintf("vault at capacity (%d) with nothing to remove", a.Config.Notes.MaxCount)), nil
				}
				autoRemovedUncurated = true
			}
			if err := a.Vault.DeleteNote(oldest); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("auto-removing: %v", err)), nil
			}
			autoRemoved = oldest
		}

		note := model.Note{
			CreatedAt: time.Now(),
			Title:     title,
			Content:   content,
			Metadata:  meta,
		}

		id, err := a.Vault.WriteNote(note)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("save failed: %v", err)), nil
		}
		note.ID = id
		a.IndexNote(note)

		if autoRemoved != "" {
			a.RemoveNoteFromIndex(autoRemoved)
		}

		result := map[string]any{
			"id":           id,
			"title":        title,
			"created_at":   note.CreatedAt.Format(time.RFC3339),
			"auto_removed": nil,
		}
		if autoRemoved != "" {
			result["auto_removed"] = autoRemoved
			if autoRemovedUncurated {
				result["auto_removed_uncurated"] = true
			}
		}

		data, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func noteListHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := a.Config.Notes.DefaultListLimit
		if l, ok := req.GetArguments()["limit"]; ok {
			if f, ok := l.(float64); ok && f > 0 {
				limit = int(f)
			}
		}

		filters := extractFilter(req.GetArguments())

		metas, err := a.Vault.ListNotes(0)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing: %v", err)), nil
		}

		type result struct {
			ID        string         `json:"id"`
			Title     string         `json:"title"`
			Metadata  map[string]any `json:"metadata"`
			CreatedAt string         `json:"created_at"`
		}

		var filtered []result
		for _, m := range metas {
			if filters != nil && !m.Metadata.MatchesFilter(filters) {
				continue
			}
			filtered = append(filtered, result{
				ID:        m.ID,
				Title:     m.Title,
				Metadata:  noteMetaForMCP(m.Metadata),
				CreatedAt: m.CreatedAt,
			})
			if len(filtered) >= limit {
				break
			}
		}

		if filtered == nil {
			filtered = []result{}
		}
		data, _ := json.MarshalIndent(filtered, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func noteShowHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ids := resolveIDs(req.GetArguments())
		if len(ids) == 0 {
			return mcp.NewToolResultError("id or ids is required"), nil
		}

		type result struct {
			ID        string         `json:"id"`
			Title     string         `json:"title"`
			Content   string         `json:"content"`
			Metadata  map[string]any `json:"metadata"`
			CreatedAt string         `json:"created_at"`
		}

		var results []result
		for _, id := range ids {
			note, err := a.Vault.ReadNote(id)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Not found: note %q", id)), nil
			}
			results = append(results, result{
				ID:        note.ID,
				Title:     note.Title,
				Content:   note.Content,
				Metadata:  noteMetaForMCP(note.Metadata),
				CreatedAt: note.CreatedAt.Format(time.RFC3339),
			})
		}

		data, _ := json.MarshalIndent(results, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func noteUpdateHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}

		meta := extractMeta(req.GetArguments(), "meta")
		var content *string
		if c := req.GetString("content", ""); c != "" {
			content = &c
		}

		if content == nil && meta == nil {
			return mcp.NewToolResultError("nothing to update"), nil
		}

		if err := a.Vault.UpdateNote(id, content, meta); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("update failed: %v", err)), nil
		}

		if note, err := a.Vault.ReadNote(id); err == nil {
			a.IndexNote(note)
		}

		data, _ := json.MarshalIndent(map[string]any{
			"id":         id,
			"updated_at": time.Now().Format(time.RFC3339),
		}, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func noteDeleteHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}

		noteMeta, err := a.Vault.ReadNoteMeta(id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Not found: note %q", id)), nil
		}

		if err := a.Vault.DeleteNote(id); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("delete failed: %v", err)), nil
		}
		a.RemoveNoteFromIndex(id)

		data, _ := json.MarshalIndent(map[string]any{
			"id":      id,
			"title":   noteMeta.Title,
			"deleted": true,
		}, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func noteSearchHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("query is required"), nil
		}

		limit := a.Config.Search.DefaultLimit
		if l, ok := req.GetArguments()["limit"]; ok {
			if f, ok := l.(float64); ok && f > 0 {
				limit = int(f)
			}
		}

		filters := extractFilter(req.GetArguments())

		idx, err := search.Open(a.Vault.IndexDir())
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search index not available: %v", err)), nil
		}
		defer idx.Close()

		results, err := idx.SearchNotes(query, limit*2)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		type result struct {
			ID        string         `json:"id"`
			Title     string         `json:"title"`
			Score     float64        `json:"score"`
			Metadata  map[string]any `json:"metadata"`
			CreatedAt string         `json:"created_at"`
		}

		var output []result
		for _, r := range results {
			if len(output) >= limit {
				break
			}
			noteMeta, err := a.Vault.ReadNoteMeta(r.ID)
			if err != nil {
				continue
			}
			if filters != nil && !noteMeta.Metadata.MatchesFilter(filters) {
				continue
			}
			output = append(output, result{
				ID:        r.ID,
				Title:     noteMeta.Title,
				Score:     r.Score,
				Metadata:  noteMetaForMCP(noteMeta.Metadata),
				CreatedAt: noteMeta.CreatedAt,
			})
		}
		if output == nil {
			output = []result{}
		}

		data, _ := json.MarshalIndent(output, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

// --- helpers ---

// metaMapForMCP converts MetaMap for MCP JSON output.
// Single-value keys become scalars; multi-value keys become arrays.
func metaMapForMCP(m model.MetaMap) map[string]any {
	out := make(map[string]any)
	for k, vs := range m {
		if len(vs) == 1 {
			out[k] = vs[0]
		} else {
			out[k] = vs
		}
	}
	return out
}

// noteMetaForMCP converts MetaMap for MCP note JSON output, ensuring
// curated_at and curated_into are always present (null if absent).
func noteMetaForMCP(m model.MetaMap) map[string]any {
	out := metaMapForMCP(m)
	if _, ok := out["curated_at"]; !ok {
		out["curated_at"] = nil
	}
	if _, ok := out["curated_into"]; !ok {
		out["curated_into"] = nil
	}
	return out
}

func extractMeta(args map[string]any, key string) model.MetaMap {
	raw, ok := args[key]
	if !ok {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	meta := make(model.MetaMap, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case string:
			meta[k] = []string{val}
		case []any:
			var strs []string
			for _, item := range val {
				if s, ok := item.(string); ok {
					strs = append(strs, s)
				}
			}
			meta[k] = strs
		}
	}
	return meta
}

func extractFilter(args map[string]any) map[string]string {
	raw, ok := args["filter"]
	if !ok {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	filters := make(map[string]string, len(m))
	for k, v := range m {
		if v == nil {
			filters[k] = "null"
		} else if s, ok := v.(string); ok {
			filters[k] = s
		}
	}
	return filters
}

func resolveIDs(args map[string]any) []string {
	// Single ID
	if id, ok := args["id"].(string); ok && id != "" {
		return []string{id}
	}
	// Array of IDs
	if ids, ok := args["ids"].([]any); ok {
		var result []string
		for _, id := range ids {
			if s, ok := id.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}
