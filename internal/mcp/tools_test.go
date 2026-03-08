package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kno-ai/kno/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

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

func TestSetOptionHandler_NoGitContext(t *testing.T) {
	sc := &SessionContext{}
	handler := setOptionHandler(sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "auto_load_on_confirm",
		"value": true,
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return error result about no git context
	if !result.IsError {
		t.Error("expected error result for no git context")
	}
}

func TestSetOptionHandler_UnknownKey(t *testing.T) {
	dir := t.TempDir()
	sc := &SessionContext{
		Git: &GitContext{RepoRoot: dir, RepoName: "test-repo"},
	}
	handler := setOptionHandler(sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "unknown_key",
		"value": true,
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for unknown key")
	}
}

func TestSetOptionHandler_SetAutoLoad(t *testing.T) {
	dir := t.TempDir()
	sc := &SessionContext{
		Git: &GitContext{RepoRoot: dir, RepoName: "test-repo"},
	}
	handler := setOptionHandler(sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "auto_load_on_confirm",
		"value": true,
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %v", result)
	}

	// Verify .kno file was written
	knoPath := filepath.Join(dir, ".kno")
	if _, err := os.Stat(knoPath); os.IsNotExist(err) {
		t.Fatal(".kno file not created")
	}

	// Verify in-memory update
	if sc.RepoConfig == nil {
		t.Fatal("sc.RepoConfig not updated")
	}
	if sc.RepoConfig.Skill.AutoLoadOnConfirm == nil || !*sc.RepoConfig.Skill.AutoLoadOnConfirm {
		t.Error("in-memory auto_load_on_confirm not set to true")
	}

	// Verify file can be loaded back
	loaded, err := config.LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Skill.AutoLoadOnConfirm == nil || !*loaded.Skill.AutoLoadOnConfirm {
		t.Error("round-trip: auto_load_on_confirm not true")
	}
}

func TestSetOptionHandler_SetAutoLoadFalse(t *testing.T) {
	dir := t.TempDir()
	sc := &SessionContext{
		Git: &GitContext{RepoRoot: dir, RepoName: "test-repo"},
	}
	handler := setOptionHandler(sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "auto_load_on_confirm",
		"value": false,
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %v", result)
	}

	// Verify file round-trip
	loaded, err := config.LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Skill.AutoLoadOnConfirm == nil || *loaded.Skill.AutoLoadOnConfirm {
		t.Error("round-trip: auto_load_on_confirm should be false")
	}
}

func TestSetOptionHandler_NonBoolValue(t *testing.T) {
	dir := t.TempDir()
	sc := &SessionContext{
		Git: &GitContext{RepoRoot: dir, RepoName: "test-repo"},
	}
	handler := setOptionHandler(sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "auto_load_on_confirm",
		"value": "not-a-bool",
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for non-bool value")
	}
}

func TestSetOptionHandler_PreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()

	// Write an existing .kno with nudge_level
	level := "active"
	existing := &config.RepoConfig{
		Skill: config.RepoSkillConfig{NudgeLevel: &level},
	}
	if err := config.SaveRepoConfig(dir, existing); err != nil {
		t.Fatal(err)
	}

	sc := &SessionContext{
		Git:        &GitContext{RepoRoot: dir, RepoName: "test-repo"},
		RepoConfig: existing,
	}
	handler := setOptionHandler(sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"key":   "auto_load_on_confirm",
		"value": true,
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %v", result)
	}

	// Verify nudge_level was preserved
	loaded, err := config.LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Skill.NudgeLevel == nil || *loaded.Skill.NudgeLevel != "active" {
		t.Error("existing nudge_level should be preserved")
	}
	if loaded.Skill.AutoLoadOnConfirm == nil || !*loaded.Skill.AutoLoadOnConfirm {
		t.Error("new auto_load_on_confirm should be set")
	}
}

func TestSetOptionHandler_MissingKey(t *testing.T) {
	dir := t.TempDir()
	sc := &SessionContext{
		Git: &GitContext{RepoRoot: dir, RepoName: "test-repo"},
	}
	handler := setOptionHandler(sc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"value": true,
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing key")
	}
}
