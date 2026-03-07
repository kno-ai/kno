package model

import "time"

// Page is the in-memory representation of a page.
type Page struct {
	ID        string
	Name      string
	CreatedAt time.Time
	Content   string
	Metadata  MetaMap
}

// PageMeta is the on-disk JSON representation of page metadata.
type PageMeta struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	CreatedAt string  `json:"created_at"`
	Metadata  MetaMap `json:"metadata,omitempty"`
}
