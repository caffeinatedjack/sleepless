package notes

import (
	"os"
	"path/filepath"
	"strings"
)

// SearchResult represents a search match.
type SearchResult struct {
	Note    *Note
	Path    string
	Context string
}

// Search searches for notes matching the query.
// TODO: Implement proper indexing for better performance.
func (s *Store) Search(query string, tags []string, useOR bool) ([]*SearchResult, error) {
	var results []*SearchResult

	entries, err := os.ReadDir(s.NotesDir)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(s.NotesDir, entry.Name())
		note, err := s.loadNote(path)
		if err != nil {
			continue
		}

		// Check tag filter
		if len(tags) > 0 {
			hasTag := false
			for _, tag := range tags {
				if note.HasTag(tag) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Check if content matches query
		bodyLower := strings.ToLower(note.Body)
		if strings.Contains(bodyLower, queryLower) {
			context := extractContext(note.Body, query)
			results = append(results, &SearchResult{
				Note:    note,
				Path:    path,
				Context: context,
			})
		}
	}

	return results, nil
}

// extractContext extracts a snippet around the matched query.
func extractContext(body, query string) string {
	bodyLower := strings.ToLower(body)
	queryLower := strings.ToLower(query)

	idx := strings.Index(bodyLower, queryLower)
	if idx == -1 {
		// Return first 100 chars as fallback
		if len(body) > 100 {
			return body[:97] + "..."
		}
		return body
	}

	// Extract context around match (±50 chars)
	start := idx - 50
	if start < 0 {
		start = 0
	}

	end := idx + len(query) + 50
	if end > len(body) {
		end = len(body)
	}

	context := body[start:end]
	if start > 0 {
		context = "..." + context
	}
	if end < len(body) {
		context = context + "..."
	}

	return strings.TrimSpace(context)
}

// ListAllTags returns all tags with their usage counts.
func (s *Store) ListAllTags() (map[string]int, error) {
	tagCounts := make(map[string]int)

	entries, err := os.ReadDir(s.NotesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(s.NotesDir, entry.Name())
		note, err := s.loadNote(path)
		if err != nil {
			continue
		}

		for _, tag := range note.Tags {
			tagCounts[tag]++
		}
	}

	return tagCounts, nil
}

// FindByTag returns all notes with the specified tag.
func (s *Store) FindByTag(tag string) ([]*Note, error) {
	var notes []*Note

	entries, err := os.ReadDir(s.NotesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(s.NotesDir, entry.Name())
		note, err := s.loadNote(path)
		if err != nil {
			continue
		}

		if note.HasTag(tag) {
			notes = append(notes, note)
		}
	}

	return notes, nil
}
