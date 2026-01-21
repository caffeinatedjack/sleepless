package notes

import (
	"fmt"
)

// CreateFloating creates a new floating note with generated ID.
func (s *Store) CreateFloating(title string, body string, tags []string) (*Note, error) {
	id, err := GenerateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	note := NewFloatingNote(id)

	// Build body with title
	if title != "" {
		note.Body = fmt.Sprintf("# %s\n\n%s", title, body)
	} else {
		note.Body = body
	}

	note.AddTags(tags...)

	return note, nil
}

// GetOrCreateFloating loads a floating note by ID/prefix or creates a new one.
func (s *Store) GetOrCreateFloating(idOrPrefix string) (*Note, error) {
	note, err := s.LoadFloating(idOrPrefix)
	if err == nil {
		return note, nil
	}

	// Note doesn't exist, create it
	return s.CreateFloating("", "", nil)
}
