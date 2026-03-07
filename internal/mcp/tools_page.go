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

func registerPageTools(s *server.MCPServer, a *app.App) {
	s.AddTool(mcp.NewTool("kno_page_create",
		mcp.WithDescription("Create a new page."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Page name.")),
		mcp.WithString("content", mcp.Description("Initial page content (markdown).")),
		mcp.WithObject("meta", mcp.Description("Optional metadata.")),
	), pageCreateHandler(a))

	s.AddTool(mcp.NewTool("kno_page_list",
		mcp.WithDescription("List all pages."),
		mcp.WithObject("filter", mcp.Description("Filter criteria.")),
	), pageListHandler(a))

	s.AddTool(mcp.NewTool("kno_page_show",
		mcp.WithDescription("Read a page by ID."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Page ID.")),
	), pageShowHandler(a))

	s.AddTool(mcp.NewTool("kno_page_update",
		mcp.WithDescription("Update a page's content or metadata."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Page ID.")),
		mcp.WithString("content", mcp.Description("New content (replaces existing).")),
		mcp.WithObject("meta", mcp.Description("Metadata to set/update.")),
	), pageUpdateHandler(a))

	s.AddTool(mcp.NewTool("kno_page_rename",
		mcp.WithDescription("Rename a page. Updates files, search index, and note references."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Current page ID.")),
		mcp.WithString("name", mcp.Required(), mcp.Description("New page name.")),
	), pageRenameHandler(a))

	s.AddTool(mcp.NewTool("kno_page_search",
		mcp.WithDescription("Search pages by text query."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query.")),
		mcp.WithNumber("limit", mcp.Description("Maximum results.")),
		mcp.WithObject("filter", mcp.Description("Filter criteria.")),
	), pageSearchHandler(a))
}

func pageCreateHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}

		content := req.GetString("content", "")
		meta := extractMeta(req.GetArguments(), "meta")

		page := model.Page{
			Name:      name,
			CreatedAt: time.Now(),
			Content:   content,
			Metadata:  meta,
		}

		id, err := a.Vault.WritePage(page)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("create failed: %v", err)), nil
		}
		page.ID = id
		a.IndexPage(page)

		data, _ := json.MarshalIndent(map[string]any{
			"id":         id,
			"name":       name,
			"created_at": page.CreatedAt.Format(time.RFC3339),
		}, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func pageListHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filters := extractFilter(req.GetArguments())

		metas, err := a.Vault.ListPages()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing: %v", err)), nil
		}

		type result struct {
			ID        string         `json:"id"`
			Name      string         `json:"name"`
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
				Name:      m.Name,
				Metadata:  metaMapForMCP(m.Metadata),
				CreatedAt: m.CreatedAt,
			})
		}
		if filtered == nil {
			filtered = []result{}
		}

		data, _ := json.MarshalIndent(filtered, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func pageShowHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}

		page, err := a.Vault.ReadPage(id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Not found: page %q", id)), nil
		}

		out := map[string]any{
			"id":         page.ID,
			"name":       page.Name,
			"content":    page.Content,
			"metadata":   metaMapForMCP(page.Metadata),
			"created_at": page.CreatedAt.Format(time.RFC3339),
		}

		data, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func pageUpdateHandler(a *app.App) server.ToolHandlerFunc {
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

		if err := a.Vault.UpdatePage(id, content, meta); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("update failed: %v", err)), nil
		}

		if p, err := a.Vault.ReadPage(id); err == nil {
			a.IndexPage(p)
		}

		data, _ := json.MarshalIndent(map[string]any{
			"id":         id,
			"updated_at": time.Now().Format(time.RFC3339),
		}, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func pageRenameHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}

		newID, err := a.Vault.RenamePage(id, name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("rename failed: %v", err)), nil
		}

		// Update search index.
		a.RemovePageFromIndex(id)
		if page, err := a.Vault.ReadPage(newID); err == nil {
			a.IndexPage(page)
		}

		// Update distilled_into references on notes.
		if newID != id {
			allNotes, _ := a.Vault.ListNotes(0)
			for _, nm := range allNotes {
				if nm.Metadata == nil || !nm.Metadata.Has("distilled_into") {
					continue
				}
				vals := nm.Metadata["distilled_into"]
				changed := false
				for i, val := range vals {
					if val == id {
						vals[i] = newID
						changed = true
					}
				}
				if changed {
					updateMeta := make(model.MetaMap)
					updateMeta["distilled_into"] = vals
					a.Vault.UpdateNote(nm.ID, nil, updateMeta)
				}
			}
		}

		data, _ := json.MarshalIndent(map[string]any{
			"old_id": id,
			"new_id": newID,
			"name":   name,
		}, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func pageSearchHandler(a *app.App) server.ToolHandlerFunc {
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

		results, err := idx.SearchPages(query, limit*2)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		type result struct {
			ID      string  `json:"id"`
			Name    string  `json:"name"`
			Score   float64 `json:"score"`
			Excerpt string  `json:"excerpt,omitempty"`
		}

		var output []result
		for _, r := range results {
			if len(output) >= limit {
				break
			}
			pageMeta, err := a.Vault.ReadPageMeta(r.ID)
			if err != nil {
				continue
			}
			if filters != nil && !pageMeta.Metadata.MatchesFilter(filters) {
				continue
			}
			sr := result{ID: r.ID, Name: pageMeta.Name, Score: r.Score}
			if p, err := a.Vault.ReadPage(r.ID); err == nil && p.Content != "" {
				excerpt := p.Content
				if len(excerpt) > 200 {
					excerpt = excerpt[:200] + "..."
				}
				sr.Excerpt = excerpt
			}
			output = append(output, sr)
		}
		if output == nil {
			output = []result{}
		}

		data, _ := json.MarshalIndent(output, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}
