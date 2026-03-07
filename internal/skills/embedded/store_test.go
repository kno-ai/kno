package embedded

import (
	"testing"
)

func TestEmbeddedSkillsPresent(t *testing.T) {
	store := New()

	required := []string{
		"save.md",
		"distill.md",
		"load.md",
		"page.md",
		"status.md",
	}

	for _, name := range required {
		content, err := store.Get(name)
		if err != nil {
			t.Errorf("missing embedded skill %q: %v", name, err)
			continue
		}
		if len(content) < 50 {
			t.Errorf("skill %q suspiciously short (%d bytes)", name, len(content))
		}
	}

	all, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != len(required) {
		t.Errorf("expected %d skills, found %d: %v", len(required), len(all), all)
	}
}
