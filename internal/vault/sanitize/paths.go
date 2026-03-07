package sanitize

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	leadingTrailing = regexp.MustCompile(`^-+|-+$`)
)

// Slugify converts a string into a URL/filename-safe slug.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = leadingTrailing.ReplaceAllString(s, "")
	if len(s) > 80 {
		s = s[:80]
		s = leadingTrailing.ReplaceAllString(s, "")
	}
	if s == "" {
		return "note"
	}
	return s
}

// SafeJoin joins a base directory with a relative path and ensures the result
// is within the base directory. Returns an error if path traversal is detected.
func SafeJoin(base, rel string) (string, error) {
	full := filepath.Join(base, rel)
	abs, err := filepath.Abs(full)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("resolving base: %w", err)
	}
	if !strings.HasPrefix(abs, absBase+string(filepath.Separator)) && abs != absBase {
		return "", fmt.Errorf("path traversal detected: %q escapes %q", rel, base)
	}
	return abs, nil
}
