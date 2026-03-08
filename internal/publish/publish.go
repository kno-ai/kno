package publish

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kno-ai/kno/internal/config"
	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/vault"
)

// Result describes what happened for a single page publish.
type Result struct {
	PageID   string
	PageName string
	Target   string
	Path     string
	Err      error
}

// PublishPages renders the given pages to all targets in the given format.
// If pageIDs is nil, all pages are published.
func PublishPages(v vault.Vault, targets []config.PublishTarget, pageIDs []string) ([]Result, error) {
	pages, err := loadPages(v, pageIDs)
	if err != nil {
		return nil, err
	}

	// Collect ALL page names for wikilink cross-referencing, not just the
	// pages being published. This ensures single-page publishes still get
	// wikilinks to other pages.
	allPageNames, err := allPageNames(v)
	if err != nil {
		// Fall back to just the pages being published.
		allPageNames = make([]string, len(pages))
		for i, p := range pages {
			allPageNames[i] = p.Name
		}
	}

	var results []Result
	for _, target := range targets {
		expandedPath := expandHome(target.Path)
		if err := os.MkdirAll(expandedPath, 0o755); err != nil {
			for _, p := range pages {
				results = append(results, Result{
					PageID:   p.ID,
					PageName: p.Name,
					Target:   target.Path,
					Err:      fmt.Errorf("creating directory: %w", err),
				})
			}
			continue
		}

		for _, page := range pages {
			content := renderPage(page, target.Format, allPageNames)
			outPath := filepath.Join(expandedPath, page.ID+".md")
			err := os.WriteFile(outPath, []byte(content), 0o644)
			results = append(results, Result{
				PageID:   page.ID,
				PageName: page.Name,
				Target:   target.Path,
				Path:     outPath,
				Err:      err,
			})
		}
	}

	return results, nil
}

func loadPages(v vault.Vault, pageIDs []string) ([]model.Page, error) {
	if pageIDs != nil {
		var pages []model.Page
		for _, id := range pageIDs {
			p, err := v.ReadPage(id)
			if err != nil {
				return nil, fmt.Errorf("page %q not found", id)
			}
			pages = append(pages, p)
		}
		return pages, nil
	}

	metas, err := v.ListPages()
	if err != nil {
		return nil, err
	}
	var pages []model.Page
	for _, m := range metas {
		p, err := v.ReadPage(m.ID)
		if err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}
	return pages, nil
}

// allPageNames returns names of all pages in the vault for wikilink matching.
func allPageNames(v vault.Vault) ([]string, error) {
	metas, err := v.ListPages()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(metas))
	for i, m := range metas {
		names[i] = m.Name
	}
	return names, nil
}

func renderPage(page model.Page, format string, allPageNames []string) string {
	content := stripGuidance(page.Content)

	switch format {
	case "frontmatter":
		return renderFrontmatter(page, content, allPageNames)
	default:
		return content
	}
}

// stripGuidance removes an HTML comment block at the top of the content.
// Matches <!-- ... --> at the start (with optional leading whitespace),
// followed by any guidance text until the next section or blank line.
func stripGuidance(content string) string {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "<!--") {
		return content
	}

	// Find end of comment block.
	endIdx := strings.Index(trimmed, "-->")
	if endIdx < 0 {
		return content
	}

	after := trimmed[endIdx+3:]

	// Skip any non-heading prose that follows the comment (the guidance text).
	// Stop at the first heading, horizontal rule, or after consuming all
	// non-heading text.
	lines := strings.Split(after, "\n")
	startLine := 0
	for i, line := range lines {
		stripped := strings.TrimSpace(line)
		if stripped == "" {
			continue
		}
		if strings.HasPrefix(stripped, "#") || strings.HasPrefix(stripped, "---") {
			startLine = i
			break
		}
		// This line is guidance prose — skip it.
		startLine = i + 1
	}

	remaining := strings.TrimSpace(strings.Join(lines[startLine:], "\n"))
	if remaining == "" {
		return ""
	}
	return remaining + "\n"
}

func renderFrontmatter(page model.Page, content string, allPageNames []string) string {
	// Tags are stored on the page metadata by the curate skill.
	tags := page.Metadata["tags"]
	var fm strings.Builder
	fm.WriteString("---\n")
	fm.WriteString(fmt.Sprintf("title: %s\n", yamlQuote(page.Name)))
	fm.WriteString(fmt.Sprintf("aliases: [%s]\n", yamlQuote(page.Name)))

	if len(tags) > 0 {
		fm.WriteString("tags: [")
		for i, t := range tags {
			if i > 0 {
				fm.WriteString(", ")
			}
			fm.WriteString(yamlQuote(t))
		}
		fm.WriteString("]\n")
	}

	if summary := page.Metadata.Get("summary"); summary != "" {
		fm.WriteString(fmt.Sprintf("summary: %s\n", yamlQuote(summary)))
	}

	fm.WriteString(fmt.Sprintf("created: %s\n", page.CreatedAt.Format("2006-01-02")))

	if lastCurated := page.Metadata.Get("last_curated_at"); lastCurated != "" {
		if t, err := time.Parse(time.RFC3339, lastCurated); err == nil {
			fm.WriteString(fmt.Sprintf("updated: %s\n", t.Format("2006-01-02")))
		}
	}

	fm.WriteString("---\n\n")

	body := insertWikilinks(content, page.Name, allPageNames)
	return fm.String() + body
}

// insertWikilinks replaces references to other page names with [[wikilinks]].
// Conservative: only multi-word names (2+ words), case-sensitive exact match,
// word-boundary, outside fenced code blocks, skip existing wikilinks.
func insertWikilinks(content, selfName string, allPageNames []string) string {
	// Precompile regexes for eligible page names. Each pattern matches the
	// name at a word boundary but only when NOT already inside [[ ]].
	var patterns []*regexp.Regexp
	for _, name := range allPageNames {
		if name == selfName {
			continue
		}
		if len(strings.Fields(name)) < 2 {
			continue
		}
		escaped := regexp.QuoteMeta(name)
		// Negative-lookbehind isn't available in Go regex, so we match an
		// optional [[ prefix and decide in the replacement function.
		re := regexp.MustCompile(`(\[\[)?\b` + escaped + `\b`)
		patterns = append(patterns, re)
	}

	if len(patterns) == 0 {
		return content
	}

	// Split into code-fenced and non-code sections.
	parts := splitCodeBlocks(content)
	for i, part := range parts {
		if part.isCode {
			continue
		}
		text := part.text
		for _, re := range patterns {
			text = re.ReplaceAllStringFunc(text, func(match string) string {
				if strings.HasPrefix(match, "[[") {
					return match // already wikilinked
				}
				return "[[" + match + "]]"
			})
		}
		parts[i].text = text
	}

	var out strings.Builder
	for _, p := range parts {
		out.WriteString(p.text)
	}
	return out.String()
}

type textBlock struct {
	text   string
	isCode bool
}

// splitCodeBlocks splits content into alternating prose/code sections.
func splitCodeBlocks(content string) []textBlock {
	var blocks []textBlock
	lines := strings.Split(content, "\n")
	var current strings.Builder
	inCode := false

	for i, line := range lines {
		if i > 0 {
			current.WriteString("\n")
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inCode {
				// End of code block — include this line in the code block.
				current.WriteString(line)
				blocks = append(blocks, textBlock{text: current.String(), isCode: true})
				current.Reset()
				inCode = false
				continue
			}
			// Start of code block — flush prose.
			if current.Len() > 0 {
				blocks = append(blocks, textBlock{text: current.String(), isCode: false})
				current.Reset()
			}
			current.WriteString(line)
			inCode = true
			continue
		}
		current.WriteString(line)
	}

	if current.Len() > 0 {
		blocks = append(blocks, textBlock{text: current.String(), isCode: inCode})
	}

	return blocks
}

// yamlQuote wraps a string in double quotes if it contains YAML special characters.
func yamlQuote(s string) string {
	if strings.ContainsAny(s, ":{}[]#&*!|>'\"%@`,?") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
