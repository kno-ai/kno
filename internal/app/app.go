package app

import (
	"fmt"

	"github.com/kno-ai/kno/internal/capture"
	"github.com/kno-ai/kno/internal/config"
	"github.com/kno-ai/kno/internal/skills"
	"github.com/kno-ai/kno/internal/skills/embedded"
	"github.com/kno-ai/kno/internal/vault"
	"github.com/kno-ai/kno/internal/vault/fs"
)

// App wires together the core services. Both CLI and MCP use this.
type App struct {
	Config  config.Config
	Vault   vault.Vault
	Capture *capture.Service
	Skills  skills.Store
}

// FromConfig builds an App from a loaded config.
func FromConfig(cfg config.Config) (*App, error) {
	if cfg.VaultPath == "" {
		return nil, fmt.Errorf("vault_path not configured; run 'kno setup' first")
	}

	v := fs.New(cfg.VaultPath, cfg.KnoSubdir)
	capSvc := &capture.Service{
		Vault:        v,
		MaxBodyBytes: cfg.Capture.MaxBodyBytes,
	}
	skillStore := embedded.New()

	return &App{
		Config:  cfg,
		Vault:   v,
		Capture: capSvc,
		Skills:  skillStore,
	}, nil
}
