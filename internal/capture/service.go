package capture

import (
	"fmt"
	"time"

	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/vault"
)

// Service handles capture creation and writing.
type Service struct {
	Vault        vault.Vault
	MaxBodyBytes int
}

// CreateParams are the inputs for creating a capture.
type CreateParams struct {
	Title      string
	BodyMD     string
	SourceKind string
	SourceTool string
	Meta       map[string]string
}

// Create builds a CaptureNote from params and writes it to the vault.
func (s *Service) Create(p CreateParams) (model.CaptureWriteResult, error) {
	now := time.Now()

	body := p.BodyMD
	if s.MaxBodyBytes > 0 && len(body) > s.MaxBodyBytes {
		body = body[:s.MaxBodyBytes] + "\n\n> [truncated: exceeded max body size]\n"
	}

	note := model.CaptureNote{
		ID:        NewID(now),
		CreatedAt: now,
		Source: model.CaptureSource{
			Kind: p.SourceKind,
			Tool: p.SourceTool,
		},
		Title:  p.Title,
		BodyMD: body,
		Meta:   p.Meta,
		Status: "raw",
	}

	result, err := s.Vault.WriteCapture(note)
	if err != nil {
		return model.CaptureWriteResult{}, fmt.Errorf("writing capture: %w", err)
	}

	return result, nil
}
