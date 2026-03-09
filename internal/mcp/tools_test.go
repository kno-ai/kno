package mcp

import (
	"context"
	"testing"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// testApp creates a minimal App with a temp vault for testing.
func testApp(t *testing.T) *app.App {
	t.Helper()
	vaultDir := t.TempDir()
	return &app.App{
		VaultPath: vaultDir,
		Config:    config.DefaultConfig(),
	}
}

func TestExtractMeta_LowercaseKeys(t *testing.T) {
	args := map[string]any{
		"meta": map[string]any{
			"Type":   "decision",
			"STATUS": "open",
			"tags":   []any{"aws", "rds"},
			"Repo":   "cloud-infra",
		},
	}
	meta := extractMeta(args, "meta")

	tests := []struct {
		key  string
		want string
	}{
		{"type", "decision"},
		{"status", "open"},
		{"repo", "cloud-infra"},
	}
	for _, tt := range tests {
		if meta.Get(tt.key) != tt.want {
			t.Errorf("meta[%q] = %q, want %q", tt.key, meta.Get(tt.key), tt.want)
		}
	}

	// Uppercase keys should not exist
	for _, upper := range []string{"Type", "STATUS", "Repo"} {
		if _, ok := meta[upper]; ok {
			t.Errorf("uppercase key %q should not exist", upper)
		}
	}

	// Array values should pass through
	if tags := meta["tags"]; len(tags) != 2 || tags[0] != "aws" || tags[1] != "rds" {
		t.Errorf("tags = %v, want [aws rds]", tags)
	}
}

func TestExtractMeta_Nil(t *testing.T) {
	meta := extractMeta(map[string]any{}, "meta")
	if meta != nil {
		t.Error("expected nil for missing key")
	}
}

func TestSetOptionHandler_NoProjectVault(t *testing.T) {
	a := testApp(t)
	sc := &SessionContext{}
	handler := setOptionHandler(a, sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "page",
		"value": "test-page",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when setting page without a project vault")
	}
}

func TestSetOptionHandler_UnknownKey(t *testing.T) {
	a := testApp(t)
	sc := &SessionContext{}
	handler := setOptionHandler(a, sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "unknown_key",
		"value": "whatever",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for unknown key")
	}
}

func TestSetOptionHandler_SetPage_ProjectVault(t *testing.T) {
	a := testApp(t)
	sc := &SessionContext{
		IsProjectVault:   true,
		ProjectVaultPath: a.VaultPath,
	}
	handler := setOptionHandler(a, sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "page",
		"value": "cloud-infra",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %v", result)
	}

	// Verify in-memory update.
	if a.Config.Page != "cloud-infra" {
		t.Errorf("in-memory page = %q, want cloud-infra", a.Config.Page)
	}

	// Verify config was written to vault.
	loaded, err := config.Load(a.VaultPath)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Page != "cloud-infra" {
		t.Errorf("round-trip: page = %q, want cloud-infra", loaded.Page)
	}
}

func TestSetOptionHandler_SetNudgeLevel_ProjectVault(t *testing.T) {
	a := testApp(t)
	sc := &SessionContext{
		IsProjectVault:   true,
		ProjectVaultPath: a.VaultPath,
	}
	handler := setOptionHandler(a, sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "nudge_level",
		"value": "light",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %v", result)
	}

	// Verify in-memory update.
	if a.Config.Skill.NudgeLevel != "light" {
		t.Errorf("in-memory nudge_level = %q, want light", a.Config.Skill.NudgeLevel)
	}

	// Verify config was written to vault.
	loaded, err := config.Load(a.VaultPath)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Skill.NudgeLevel != "light" {
		t.Errorf("round-trip: nudge_level = %q, want light", loaded.Skill.NudgeLevel)
	}
}

func TestSetOptionHandler_InvalidNudgeLevel(t *testing.T) {
	a := testApp(t)
	sc := &SessionContext{
		IsProjectVault:   true,
		ProjectVaultPath: a.VaultPath,
	}
	handler := setOptionHandler(a, sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "nudge_level",
		"value": "bogus",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for invalid nudge_level")
	}
}

func TestSetOptionHandler_PromptProjectSetup(t *testing.T) {
	a := testApp(t)
	sc := &SessionContext{}
	handler := setOptionHandler(a, sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "prompt_project_setup",
		"value": "false",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %v", result)
	}

	// Verify in-memory update.
	if a.Config.Skill.PromptProjectSetup == nil || *a.Config.Skill.PromptProjectSetup {
		t.Error("expected prompt_project_setup = false in memory")
	}

	// Verify vault config was written.
	loaded, err := config.Load(a.VaultPath)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Skill.PromptProjectSetup == nil || *loaded.Skill.PromptProjectSetup {
		t.Error("expected prompt_project_setup = false on disk")
	}
}

func TestSetOptionHandler_PromptProjectSetup_InvalidValue(t *testing.T) {
	a := testApp(t)
	sc := &SessionContext{}
	handler := setOptionHandler(a, sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "prompt_project_setup",
		"value": "maybe",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for non-bool value")
	}
}

func TestSetOptionHandler_MissingKey(t *testing.T) {
	a := testApp(t)
	sc := &SessionContext{}
	handler := setOptionHandler(a, sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"value": "something",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing key")
	}
}
