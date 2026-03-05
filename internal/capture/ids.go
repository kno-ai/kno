package capture

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kno-ai/kno/internal/model"
	"github.com/kno-ai/kno/internal/vault/sanitize"
)

// NewID generates a capture ID like cap_20260305T101222-0500_a1b2c3.
func NewID(t time.Time) string {
	randBytes := make([]byte, 4)
	if _, err := rand.Read(randBytes); err != nil {
		// Fallback to timestamp-based uniqueness if entropy unavailable.
		randBytes = []byte(fmt.Sprintf("%04x", t.UnixNano()%0xFFFF))
	}
	return fmt.Sprintf("cap_%s_%s",
		t.Format("20060102T150405Z0700"),
		hex.EncodeToString(randBytes),
	)
}

// DirName generates the capture directory name from a CaptureNote.
func DirName(note model.CaptureNote) string {
	slug := "capture"
	if note.Title != "" {
		slug = sanitize.Slugify(note.Title)
	}
	ts := note.CreatedAt.Format("20060102T150405Z0700")
	return fmt.Sprintf("%s_%s", ts, slug)
}
