package publish

import (
	"strings"
	"testing"
	"time"

	"github.com/kno-ai/kno/internal/model"
)

func TestStripGuidance_WithComment(t *testing.T) {
	content := `<!-- Guidance -->
Focus on operational lessons. Skip theory.

## RDS

- Pin parameter groups.
`
	got := stripGuidance(content)
	if strings.Contains(got, "Guidance") {
		t.Errorf("expected guidance stripped, got: %s", got)
	}
	if strings.Contains(got, "operational lessons") {
		t.Errorf("expected guidance prose stripped, got: %s", got)
	}
	if !strings.Contains(got, "## RDS") {
		t.Errorf("expected content preserved, got: %s", got)
	}
}

func TestStripGuidance_NoComment(t *testing.T) {
	content := "## RDS\n\n- Pin parameter groups.\n"
	got := stripGuidance(content)
	if got != content {
		t.Errorf("expected no change, got: %s", got)
	}
}

func TestStripGuidance_EmptyContent(t *testing.T) {
	got := stripGuidance("")
	if got != "" {
		t.Errorf("expected empty, got: %q", got)
	}
}

func TestRenderFrontmatter_Frontmatter(t *testing.T) {
	page := model.Page{
		ID:        "aws-infrastructure",
		Name:      "AWS Infrastructure",
		CreatedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Content:   "## RDS\n\n- Pin parameter groups.\n",
		Metadata: model.MetaMap{
			"tags":            {"aws", "ecs", "rds"},
			"summary":         {"Operational lessons for AWS."},
			"last_curated_at": {"2026-03-07T10:00:00Z"},
		},
	}

	got := renderFrontmatter(page, page.Content, nil)

	checks := []string{
		"title: AWS Infrastructure",
		"aliases: [AWS Infrastructure]",
		"tags: [aws, ecs, rds]",
		"summary: Operational lessons for AWS.",
		"created: 2026-01-15",
		"updated: 2026-03-07",
		"## RDS",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderFrontmatter_NoTags(t *testing.T) {
	page := model.Page{
		ID:        "test",
		Name:      "Test Page",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Content:   "content\n",
	}

	got := renderFrontmatter(page, page.Content, nil)
	if strings.Contains(got, "tags:") {
		t.Errorf("expected no tags field, got:\n%s", got)
	}
}

func TestRenderFrontmatter_YAMLQuoting(t *testing.T) {
	page := model.Page{
		ID:        "test",
		Name:      "RDS: Parameter Tuning",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Content:   "content\n",
		Metadata: model.MetaMap{
			"summary": {"RDS tuning: connection pools & params."},
		},
	}

	got := renderFrontmatter(page, page.Content, nil)
	if !strings.Contains(got, `title: "RDS: Parameter Tuning"`) {
		t.Errorf("expected quoted title, got:\n%s", got)
	}
}

func TestInsertWikilinks_MultiWord(t *testing.T) {
	content := "See the AWS Infrastructure page for details."
	names := []string{"AWS Infrastructure", "Go"}
	got := insertWikilinks(content, "Other Page", names)
	if !strings.Contains(got, "[[AWS Infrastructure]]") {
		t.Errorf("expected wikilink inserted, got: %s", got)
	}
	// Single-word "Go" should NOT be wikilinked.
	if strings.Contains(got, "[[Go]]") {
		t.Errorf("single-word name should not be wikilinked, got: %s", got)
	}
}

func TestInsertWikilinks_SkipsSelf(t *testing.T) {
	content := "This is the AWS Infrastructure page."
	names := []string{"AWS Infrastructure"}
	got := insertWikilinks(content, "AWS Infrastructure", names)
	if strings.Contains(got, "[[AWS Infrastructure]]") {
		t.Errorf("should not wikilink self-references, got: %s", got)
	}
}

func TestInsertWikilinks_SkipsCodeBlocks(t *testing.T) {
	content := "Reference AWS Infrastructure here.\n\n```\nAWS Infrastructure in code\n```\n"
	names := []string{"AWS Infrastructure"}
	got := insertWikilinks(content, "Other", names)
	// Should wikilink in prose.
	if !strings.Contains(got, "[[AWS Infrastructure]] here") {
		t.Errorf("expected wikilink in prose, got: %s", got)
	}
	// Should NOT wikilink inside code block.
	lines := strings.Split(got, "\n")
	inCode := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCode = !inCode
			continue
		}
		if inCode && strings.Contains(line, "[[AWS Infrastructure]]") {
			t.Errorf("should not wikilink inside code blocks, got: %s", got)
		}
	}
}

func TestInsertWikilinks_SkipsExistingWikilinks(t *testing.T) {
	content := "See [[AWS Infrastructure]] for details about AWS Infrastructure."
	names := []string{"AWS Infrastructure"}
	got := insertWikilinks(content, "Other", names)
	// The existing wikilink should not be double-wrapped.
	if strings.Contains(got, "[[[[") {
		t.Errorf("should not double-wrap existing wikilinks, got: %s", got)
	}
	// The second bare occurrence should be wikilinked.
	if !strings.Contains(got, "about [[AWS Infrastructure]]") {
		t.Errorf("expected second occurrence wikilinked, got: %s", got)
	}
}

func TestRenderFrontmatter_TagsWithSpecialChars(t *testing.T) {
	page := model.Page{
		ID:        "test",
		Name:      "Test Page",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Content:   "content\n",
		Metadata: model.MetaMap{
			"tags": {"go", "c++", "key: value"},
		},
	}

	got := renderFrontmatter(page, page.Content, nil)
	// Tags with special chars should be quoted.
	if !strings.Contains(got, `"c++"`) && !strings.Contains(got, "c++") {
		t.Errorf("expected c++ tag in output, got:\n%s", got)
	}
}

func TestStripGuidance_AllGuidance(t *testing.T) {
	content := "<!-- This is all guidance -->\nJust guidance prose."
	got := stripGuidance(content)
	if got != "" {
		t.Errorf("expected empty for all-guidance content, got: %q", got)
	}
}

func TestRenderPage_PlainStripsGuidance(t *testing.T) {
	page := model.Page{
		Content: "<!-- Guidance -->\nSkip theory.\n\n## RDS\n\nContent.\n",
	}
	got := renderPage(page, "markdown", nil)
	if strings.Contains(got, "Guidance") {
		t.Error("plain format should strip guidance")
	}
	if !strings.Contains(got, "## RDS") {
		t.Error("plain format should preserve content")
	}
}
