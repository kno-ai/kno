package model

import "time"

// Note is the in-memory representation of a saved session note.
type Note struct {
	ID        string
	CreatedAt time.Time
	Title     string
	Content   string
	Metadata  MetaMap
}

// NoteMeta is the on-disk JSON representation of note metadata.
type NoteMeta struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	CreatedAt string  `json:"created_at"`
	Metadata  MetaMap `json:"metadata,omitempty"`
}
