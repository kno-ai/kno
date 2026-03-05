package sanitize

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"SQS Throughput Tuning", "sqs-throughput-tuning"},
		{"  leading/trailing spaces  ", "leading-trailing-spaces"},
		{"special!@#chars$%^", "special-chars"},
		{"", "capture"},
		{"   ", "capture"},
		{"already-slugged", "already-slugged"},
		{"UPPER CASE", "upper-case"},
		{"multiple---dashes", "multiple-dashes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSafeJoin(t *testing.T) {
	t.Run("valid path", func(t *testing.T) {
		result, err := SafeJoin("/base/dir", "sub/file.md")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "/base/dir/sub/file.md" {
			t.Errorf("got %q, want /base/dir/sub/file.md", result)
		}
	})

	t.Run("path traversal blocked", func(t *testing.T) {
		_, err := SafeJoin("/base/dir", "../../etc/passwd")
		if err == nil {
			t.Fatal("expected error for path traversal, got nil")
		}
	})

	t.Run("dot-dot in middle blocked", func(t *testing.T) {
		_, err := SafeJoin("/base/dir", "sub/../../other")
		if err == nil {
			t.Fatal("expected error for path traversal, got nil")
		}
	})
}
