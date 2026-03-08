package model

import (
	"encoding/json"
	"testing"
)

func TestParseMetaFlags(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantErr bool
		check   func(MetaMap) bool
	}{
		{"nil input", nil, false, func(m MetaMap) bool { return m == nil }},
		{"empty input", []string{}, false, func(m MetaMap) bool { return m == nil }},
		{"single value", []string{"key=value"}, false, func(m MetaMap) bool {
			return m.Get("key") == "value" && len(m["key"]) == 1
		}},
		{"duplicate keys produce array", []string{"tags=aws", "tags=sqs"}, false, func(m MetaMap) bool {
			return len(m["tags"]) == 2 && m["tags"][0] == "aws" && m["tags"][1] == "sqs"
		}},
		{"value with equals", []string{"url=https://example.com?a=1"}, false, func(m MetaMap) bool {
			return m.Get("url") == "https://example.com?a=1"
		}},
		{"missing equals", []string{"noequals"}, true, nil},
		{"empty value", []string{"key="}, false, func(m MetaMap) bool {
			return m.Get("key") == ""
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := ParseMetaFlags(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.check != nil && !tt.check(m) {
				t.Errorf("check failed for %v, got %v", tt.input, m)
			}
		})
	}
}

func TestParseFilterFlags(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantErr bool
		check   func(map[string]string) bool
	}{
		{"nil input", nil, false, func(m map[string]string) bool { return m == nil }},
		{"null filter", []string{"curated_at=null"}, false, func(m map[string]string) bool {
			return m["curated_at"] == "null"
		}},
		{"missing equals", []string{"bad"}, true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := ParseFilterFlags(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.check != nil && !tt.check(m) {
				t.Errorf("check failed")
			}
		})
	}
}

func TestMetaMapJSONNullHandling(t *testing.T) {
	// JSON with null values should deserialize without error
	input := `{"a": "hello", "b": null, "c": ["x", "y"]}`
	var m MetaMap
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		t.Fatal(err)
	}
	if m.Get("a") != "hello" {
		t.Errorf("a = %q", m.Get("a"))
	}
	if m.Has("b") {
		t.Error("b should not be present (JSON null)")
	}
	if len(m["c"]) != 2 {
		t.Errorf("c = %v", m["c"])
	}
}

func TestMetaMapMergePreservesUntouched(t *testing.T) {
	m := make(MetaMap)
	m.Set("keep", "original")
	m.Set("change", "old")

	other := make(MetaMap)
	other.Set("change", "new")

	result := m.Merge(other)
	if result.Get("keep") != "original" {
		t.Errorf("keep = %q, want original", result.Get("keep"))
	}
	if result.Get("change") != "new" {
		t.Errorf("change = %q, want new", result.Get("change"))
	}
}

func TestParseMetaFlagsLowercaseKeys(t *testing.T) {
	m, err := ParseMetaFlags([]string{"Type=decision", "STATUS=open"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m["type"]; !ok {
		t.Error("expected lowercase key 'type'")
	}
	if _, ok := m["status"]; !ok {
		t.Error("expected lowercase key 'status'")
	}
	if m.Get("type") != "decision" {
		t.Errorf("type = %q, want 'decision'", m.Get("type"))
	}
	if m.Get("status") != "open" {
		t.Errorf("status = %q, want 'open'", m.Get("status"))
	}
}

func TestMetaMapMatchesFilterCaseInsensitive(t *testing.T) {
	m := make(MetaMap)
	m.Set("status", "active")

	// Filter with uppercase value should match lowercase stored value
	if !m.MatchesFilter(map[string]string{"status": "Active"}) {
		t.Error("should match case-insensitively: stored 'active' vs filter 'Active'")
	}

	// Reverse: uppercase stored, lowercase filter
	m2 := make(MetaMap)
	m2.Set("status", "Active")
	if !m2.MatchesFilter(map[string]string{"status": "active"}) {
		t.Error("should match case-insensitively: stored 'Active' vs filter 'active'")
	}

	// "null" sentinel must remain case-sensitive (absent key check)
	m3 := make(MetaMap)
	// key is absent, so filtering for "null" should match
	if !m3.MatchesFilter(map[string]string{"missing_key": "null"}) {
		t.Error("null sentinel should match absent key")
	}
	// key is present, so filtering for "null" should NOT match
	m3.Set("missing_key", "something")
	if m3.MatchesFilter(map[string]string{"missing_key": "null"}) {
		t.Error("null sentinel should not match present key")
	}
}

func TestMetaMapMatchesFilterMultiple(t *testing.T) {
	m := make(MetaMap)
	m.Set("status", "active")
	m.Add("tags", "aws")

	// Both filters must match
	if !m.MatchesFilter(map[string]string{"status": "active", "tags": "aws"}) {
		t.Error("should match both")
	}
	if m.MatchesFilter(map[string]string{"status": "active", "tags": "gcp"}) {
		t.Error("should not match when one filter misses")
	}
}
