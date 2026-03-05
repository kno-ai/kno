package capture

import (
	"strings"
	"testing"
	"time"

	"github.com/kno-ai/kno/internal/model"
)

func TestNewID(t *testing.T) {
	ts := time.Date(2026, 3, 5, 10, 12, 22, 0, time.UTC)
	id := NewID(ts)

	if !strings.HasPrefix(id, "cap_20260305T101222Z_") {
		t.Errorf("unexpected ID format: %q", id)
	}

	parts := strings.SplitN(id, "_", 3)
	if len(parts) != 3 || len(parts[2]) != 8 {
		t.Errorf("expected 8-char random suffix, got %q", id)
	}
}

func TestDirName(t *testing.T) {
	ts := time.Date(2026, 3, 5, 10, 12, 22, 0, time.FixedZone("EST", -5*3600))

	t.Run("with title", func(t *testing.T) {
		note := model.CaptureNote{Title: "SQS Throughput", CreatedAt: ts}
		dn := DirName(note)
		if dn != "20260305T101222-0500_sqs-throughput" {
			t.Errorf("got %q", dn)
		}
	})

	t.Run("without title", func(t *testing.T) {
		note := model.CaptureNote{CreatedAt: ts}
		dn := DirName(note)
		if dn != "20260305T101222-0500_capture" {
			t.Errorf("got %q", dn)
		}
	})
}
