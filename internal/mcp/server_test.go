package mcp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/config"
)

// stubSkillStore implements skills.Store for testing.
type stubSkillStore struct {
	skills map[string]string
}

func (s *stubSkillStore) Get(name string) (string, error) {
	content, ok := s.skills[name]
	if !ok {
		return "", fmt.Errorf("skill %q not found", name)
	}
	return content, nil
}

func (s *stubSkillStore) List() ([]string, error) {
	var names []string
	for k := range s.skills {
		names = append(names, k)
	}
	return names, nil
}

func TestAwarenessInstructions_Off(t *testing.T) {
	a := &app.App{
		Config: config.Config{
			Skill: config.SkillConfig{NudgeLevel: "off"},
		},
		Skills: &stubSkillStore{},
	}
	got := awarenessInstructions(a, &SessionContext{})
	if got != "" {
		t.Errorf("expected empty string for off level, got %q", got)
	}
}

func TestAwarenessInstructions_Active(t *testing.T) {
	awareness := "# Awareness\n\nYou are kno."
	a := &app.App{
		Config: config.Config{
			Skill: config.SkillConfig{NudgeLevel: "active"},
		},
		Skills: &stubSkillStore{skills: map[string]string{"awareness.md": awareness}},
	}
	got := awarenessInstructions(a, &SessionContext{})
	if got != awareness {
		t.Errorf("expected raw awareness skill for active level\ngot:  %q\nwant: %q", got, awareness)
	}
}

func TestAwarenessInstructions_Light(t *testing.T) {
	awareness := "# Awareness\n\nYou are kno."
	a := &app.App{
		Config: config.Config{
			Skill: config.SkillConfig{NudgeLevel: "light"},
		},
		Skills: &stubSkillStore{skills: map[string]string{"awareness.md": awareness}},
	}
	got := awarenessInstructions(a, &SessionContext{})
	if !strings.HasPrefix(got, awareness) {
		t.Errorf("expected light output to start with awareness skill\ngot prefix: %q", got[:min(len(got), 50)])
	}
	if !strings.Contains(got, "Nudge level: light") {
		t.Error("expected light output to contain 'Nudge level: light' restraint section")
	}
}

func TestAwarenessInstructions_SkillMissing(t *testing.T) {
	a := &app.App{
		Config: config.Config{
			Skill: config.SkillConfig{NudgeLevel: "active"},
		},
		Skills: &stubSkillStore{skills: map[string]string{}},
	}
	got := awarenessInstructions(a, &SessionContext{})
	if got != "" {
		t.Errorf("expected empty string when skill is missing, got %q", got)
	}
}

func TestAwarenessInstructions_RepoOverrideToOff(t *testing.T) {
	awareness := "# Awareness\n\nYou are kno."
	a := &app.App{
		Config: config.Config{
			Skill: config.SkillConfig{NudgeLevel: "active"},
		},
		Skills: &stubSkillStore{skills: map[string]string{"awareness.md": awareness}},
	}
	off := "off"
	sc := &SessionContext{
		RepoConfig: &config.RepoConfig{
			Skill: config.RepoSkillConfig{NudgeLevel: &off},
		},
	}
	got := awarenessInstructions(a, sc)
	if got != "" {
		t.Errorf("repo override to off should return empty, got %q", got)
	}
}

func TestAwarenessInstructions_RepoOverrideToActive(t *testing.T) {
	awareness := "# Awareness\n\nYou are kno."
	a := &app.App{
		Config: config.Config{
			Skill: config.SkillConfig{NudgeLevel: "light"},
		},
		Skills: &stubSkillStore{skills: map[string]string{"awareness.md": awareness}},
	}
	active := "active"
	sc := &SessionContext{
		RepoConfig: &config.RepoConfig{
			Skill: config.RepoSkillConfig{NudgeLevel: &active},
		},
	}
	got := awarenessInstructions(a, sc)
	// Active level returns raw skill without restraint note
	if got != awareness {
		t.Errorf("repo override to active should return raw awareness\ngot:  %q\nwant: %q", got, awareness)
	}
	if strings.Contains(got, "Nudge level: light") {
		t.Error("should not contain light restraint when overridden to active")
	}
}
