package update

import "testing"

func TestIsNewer(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"0.3.0", "0.2.0", true},
		{"1.0.0", "0.9.9", true},
		{"0.2.1", "0.2.0", true},
		{"0.2.0", "0.2.0", false},
		{"0.1.0", "0.2.0", false},
		{"0.2.0", "0.3.0", false},
	}
	for _, tt := range tests {
		if got := isNewer(tt.a, tt.b); got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	if msg := compareVersions("0.2.0", "0.2.0"); msg != "" {
		t.Errorf("same version should return empty, got %q", msg)
	}
	if msg := compareVersions("0.3.0", "0.2.0"); msg != "" {
		t.Errorf("newer local version should return empty, got %q", msg)
	}
	if msg := compareVersions("0.1.0", "0.2.0"); msg == "" {
		t.Error("older local version should return update message")
	}
	if msg := compareVersions("v0.1.0", "v0.2.0"); msg == "" {
		t.Error("should handle v-prefixed versions")
	}
}
