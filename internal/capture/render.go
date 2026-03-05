package capture

import (
	"strings"

	"github.com/kno-ai/kno/internal/model"
)

// RenderMarkdown produces the capture markdown content (no frontmatter).
func RenderMarkdown(note model.CaptureNote) string {
	var b strings.Builder

	if note.Title != "" {
		b.WriteString("# ")
		b.WriteString(note.Title)
		b.WriteString("\n\n")
	}

	body := note.BodyMD
	if hasRequiredHeadings(body) {
		b.WriteString(body)
	} else {
		b.WriteString(wrapInStandardHeadings(body))
	}

	b.WriteString("\n")
	return b.String()
}

// RenderMeta produces the CaptureMeta for JSON serialization.
func RenderMeta(note model.CaptureNote) model.CaptureMeta {
	return model.CaptureMeta{
		ID:      note.ID,
		Created: note.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		Title:   note.Title,
		Source:  note.Source,
		Status:  note.Status,
		Meta:    note.Meta,
	}
}

func hasRequiredHeadings(body string) bool {
	return strings.Contains(body, "## TL;DR") && strings.Contains(body, "## Next steps")
}

func wrapInStandardHeadings(body string) string {
	var b strings.Builder
	b.WriteString("## TL;DR\n\n")
	b.WriteString(strings.TrimSpace(body))
	b.WriteString("\n\n## Decisions\n\n")
	b.WriteString("\n\n## Key points\n\n")
	b.WriteString("\n\n## Next steps\n\n")
	b.WriteString("\n\n## Snippets\n\n")
	return b.String()
}
