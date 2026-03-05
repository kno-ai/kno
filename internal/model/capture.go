package model

import "time"

type CaptureNote struct {
	ID        string
	CreatedAt time.Time
	Source    CaptureSource
	Title     string
	BodyMD    string
	Meta      map[string]string
	Status    string // raw
}

type CaptureSource struct {
	Kind string `json:"kind"` // claude_desktop, clipboard, stdin, file
	Tool string `json:"tool"` // claude, kno_cli
}

type CaptureMeta struct {
	ID      string            `json:"id"`
	Created string            `json:"created"`
	Title   string            `json:"title,omitempty"`
	Source  CaptureSource     `json:"source"`
	Status  string            `json:"status"`
	Meta    map[string]string `json:"meta,omitempty"`
}

type CaptureWriteResult struct {
	Path    string
	ID      string
	Created time.Time
}
