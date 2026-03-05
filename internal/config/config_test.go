package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := DefaultConfig()
	cfg.VaultPath = "/tmp/test-vault"

	// Write directly to temp path.
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Read back.
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var loaded Config
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatal(err)
	}

	if loaded.VaultPath != "/tmp/test-vault" {
		t.Errorf("VaultPath = %q, want /tmp/test-vault", loaded.VaultPath)
	}
	if loaded.KnoSubdir != "." {
		t.Errorf("KnoSubdir = %q, want .", loaded.KnoSubdir)
	}
}
