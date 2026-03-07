package model

import (
	"encoding/json"
	"strings"
)

// MetaMap stores metadata where each key maps to one or more string values.
// Single-value keys serialize as JSON scalars; multi-value keys as arrays.
type MetaMap map[string][]string

// Get returns the first value for a key, or empty string if absent.
func (m MetaMap) Get(key string) string {
	if vals, ok := m[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// Set sets a key to a single value, replacing any existing values.
func (m MetaMap) Set(key, value string) {
	m[key] = []string{value}
}

// Add appends a value to a key (creating the key if needed).
func (m MetaMap) Add(key, value string) {
	m[key] = append(m[key], value)
}

// AddUnique appends a value to a key only if it's not already present.
func (m MetaMap) AddUnique(key, value string) {
	for _, v := range m[key] {
		if v == value {
			return
		}
	}
	m[key] = append(m[key], value)
}

// Deduplicate removes duplicate values from all multi-value keys.
func (m MetaMap) Deduplicate() {
	for k, vs := range m {
		if len(vs) <= 1 {
			continue
		}
		seen := make(map[string]bool, len(vs))
		unique := vs[:0]
		for _, v := range vs {
			if !seen[v] {
				seen[v] = true
				unique = append(unique, v)
			}
		}
		m[k] = unique
	}
}

// Has returns true if the key exists.
func (m MetaMap) Has(key string) bool {
	_, ok := m[key]
	return ok
}

// MarshalJSON serializes MetaMap. Single-value keys become JSON strings;
// multi-value keys become JSON arrays.
func (m MetaMap) MarshalJSON() ([]byte, error) {
	out := make(map[string]any, len(m))
	for k, vs := range m {
		if len(vs) == 1 {
			out[k] = vs[0]
		} else {
			out[k] = vs
		}
	}
	return json.Marshal(out)
}

// UnmarshalJSON deserializes MetaMap from JSON. Accepts both string and
// array-of-string values.
func (m *MetaMap) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*m = make(MetaMap, len(raw))
	for k, v := range raw {
		switch val := v.(type) {
		case string:
			(*m)[k] = []string{val}
		case []any:
			strs := make([]string, 0, len(val))
			for _, item := range val {
				if s, ok := item.(string); ok {
					strs = append(strs, s)
				}
			}
			(*m)[k] = strs
		case nil:
			// JSON null: store the key with nil value to preserve the null semantics.
			// MatchesFilter treats absent keys and nil-value keys equivalently for "null" filter.
			// We skip storing nil to keep MetaMap clean (absent == null).
		}
	}
	return nil
}

// MatchesFilter returns true if this MetaMap matches all the given filters.
// A filter value of "null" matches keys that are absent.
// For array values, matches if any element equals the filter value.
func (m MetaMap) MatchesFilter(filters map[string]string) bool {
	for k, want := range filters {
		vals, exists := m[k]
		if want == "null" {
			if exists {
				return false
			}
			continue
		}
		if !exists {
			return false
		}
		found := false
		for _, v := range vals {
			if v == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ParseMetaFlags parses repeated --meta key=value flags into a MetaMap.
// Duplicate keys produce multi-value entries.
func ParseMetaFlags(pairs []string) (MetaMap, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	m := make(MetaMap, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			return nil, &InvalidFlagError{Flag: "--meta", Value: p, Reason: "expected key=value"}
		}
		m.Add(k, v)
	}
	return m, nil
}

// ParseFilterFlags parses repeated --filter key=value flags into a simple map.
func ParseFilterFlags(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			return nil, &InvalidFlagError{Flag: "--filter", Value: p, Reason: "expected key=value"}
		}
		m[k] = v
	}
	return m, nil
}

// Merge returns a new MetaMap with values from other overlaid on m.
// Keys in other replace same keys in m; keys not in other are preserved.
// A nil value in other removes the key entirely.
func (m MetaMap) Merge(other MetaMap) MetaMap {
	result := make(MetaMap, len(m)+len(other))
	for k, v := range m {
		result[k] = v
	}
	for k, v := range other {
		if v == nil {
			delete(result, k)
		} else {
			result[k] = v
		}
	}
	return result
}

type InvalidFlagError struct {
	Flag   string
	Value  string
	Reason string
}

func (e *InvalidFlagError) Error() string {
	return "invalid " + e.Flag + " value " + `"` + e.Value + `": ` + e.Reason
}
