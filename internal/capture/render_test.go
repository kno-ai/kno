package capture

import (
	"strings"
	"testing"
	"time"

	"github.com/kno-ai/kno/internal/model"
)

func TestRenderMarkdown(t *testing.T) {
	note := model.CaptureNote{
		Title:  "SQS Throughput",
		BodyMD: "## TL;DR\n\nSQS stuff.\n\n## Next steps\n\n- Do things",
	}

	result := RenderMarkdown(note)

	checks := []string{
		"# SQS Throughput",
		"## TL;DR",
		"SQS stuff.",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("rendered output missing %q", check)
		}
	}

	// Should NOT contain frontmatter.
	if strings.Contains(result, "---") {
		t.Error("markdown should not contain frontmatter delimiters")
	}
}

func TestRenderMeta(t *testing.T) {
	ts := time.Date(2026, 3, 5, 10, 12, 22, 0, time.FixedZone("EST", -5*3600))

	note := model.CaptureNote{
		ID:        "cap_20260305T101222-0500_abcd1234",
		CreatedAt: ts,
		Source:    model.CaptureSource{Kind: "clipboard", Tool: "kno_cli"},
		Title:     "SQS Throughput",
		Status:    "raw",
		Meta:      map[string]string{"topic": "aws/sqs"},
	}

	meta := RenderMeta(note)

	if meta.ID != note.ID {
		t.Errorf("ID = %q, want %q", meta.ID, note.ID)
	}
	if meta.Created != "2026-03-05T10:12:22-05:00" {
		t.Errorf("Created = %q", meta.Created)
	}
	if meta.Title != "SQS Throughput" {
		t.Errorf("Title = %q", meta.Title)
	}
	if meta.Source.Kind != "clipboard" {
		t.Errorf("Source.Kind = %q", meta.Source.Kind)
	}
	if meta.Meta["topic"] != "aws/sqs" {
		t.Errorf("Meta[topic] = %q", meta.Meta["topic"])
	}
}

func TestRenderMarkdownWrapsBodyWithoutRequiredHeadings(t *testing.T) {
	note := model.CaptureNote{
		BodyMD: "Just some plain text content.",
	}

	result := RenderMarkdown(note)

	if !strings.Contains(result, "## TL;DR") {
		t.Error("expected TL;DR heading to be added")
	}
	if !strings.Contains(result, "## Next steps") {
		t.Error("expected Next steps heading to be added")
	}
	if !strings.Contains(result, "Just some plain text content.") {
		t.Error("expected original body to be preserved")
	}
}
