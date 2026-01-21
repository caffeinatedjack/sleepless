// Package notes provides note storage and management functionality.
package notes

import (
	"time"
)

// NoteType represents the type of note.
type NoteType string

const (
	// NoteTypeDaily represents a daily note (date-addressed).
	NoteTypeDaily NoteType = "daily"
	// NoteTypeFloating represents a floating note (ID-addressed).
	NoteTypeFloating NoteType = "floating"
)

// Note represents a single note entry.
type Note struct {
	// Type is the note type (daily or floating).
	Type NoteType `yaml:"type"`

	// ID is the unique identifier for floating notes (8-char lowercase hex).
	ID string `yaml:"id,omitempty"`

	// Date is the date for daily notes (YYYY-MM-DD format).
	Date string `yaml:"date,omitempty"`

	// Created is the creation timestamp.
	Created time.Time `yaml:"created"`

	// Updated is the last modification timestamp.
	Updated time.Time `yaml:"updated,omitempty"`

	// Tags are labels for organization and filtering.
	Tags []string `yaml:"tags,omitempty"`

	// Body is the markdown content (without frontmatter).
	Body string `yaml:"-"`
}

// NewDailyNote creates a new daily note for the given date.
func NewDailyNote(date string) *Note {
	now := time.Now()
	return &Note{
		Type:    NoteTypeDaily,
		Date:    date,
		Created: now,
		Tags:    []string{},
	}
}

// NewFloatingNote creates a new floating note with the given ID.
func NewFloatingNote(id string) *Note {
	now := time.Now()
	return &Note{
		Type:    NoteTypeFloating,
		ID:      id,
		Created: now,
		Tags:    []string{},
	}
}

// Touch updates the Updated timestamp to the current time.
func (n *Note) Touch() {
	n.Updated = time.Now()
}

// AddTags adds tags to the note, avoiding duplicates.
func (n *Note) AddTags(tags ...string) {
	existing := make(map[string]bool)
	for _, t := range n.Tags {
		existing[t] = true
	}

	for _, tag := range tags {
		if !existing[tag] {
			n.Tags = append(n.Tags, tag)
			existing[tag] = true
		}
	}
}

// RemoveTags removes tags from the note.
func (n *Note) RemoveTags(tags ...string) {
	toRemove := make(map[string]bool)
	for _, tag := range tags {
		toRemove[tag] = true
	}

	filtered := n.Tags[:0]
	for _, tag := range n.Tags {
		if !toRemove[tag] {
			filtered = append(filtered, tag)
		}
	}
	n.Tags = filtered
}

// HasTag returns true if the note has the specified tag.
func (n *Note) HasTag(tag string) bool {
	for _, t := range n.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
