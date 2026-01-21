// Package notes provides note storage and management functionality.
package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/caffeinatedjack/sleepless/pkg/markdown"
	"gopkg.in/yaml.v3"
)

// Store provides persistence for notes.
type Store struct {
	WikiDir      string
	NotesDir     string
	TemplatesDir string
}

// NewStore creates a new note store for the given wiki directory.
func NewStore(wikiDir string) *Store {
	return &Store{
		WikiDir:      wikiDir,
		NotesDir:     filepath.Join(wikiDir, "notes"),
		TemplatesDir: filepath.Join(wikiDir, ".templates"),
	}
}

// EnsureStructure creates the notes directory if it doesn't exist.
func (s *Store) EnsureStructure() error {
	return os.MkdirAll(s.NotesDir, 0755)
}

// IsEncrypted returns true if the wiki is encrypted.
func (s *Store) IsEncrypted() bool {
	markerPath := filepath.Join(s.WikiDir, ".encrypted")
	_, err := os.Stat(markerPath)
	return err == nil
}

// DailyPath returns the file path for a daily note.
func (s *Store) DailyPath(date string) (string, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", fmt.Errorf("invalid date format: %w", err)
	}
	return filepath.Join(s.NotesDir, date+".md"), nil
}

// FloatingPath returns the file path for a floating note.
func (s *Store) FloatingPath(id string) (string, error) {
	// Validate ID format (8-char lowercase hex)
	if len(id) < 6 || len(id) > 8 {
		return "", fmt.Errorf("invalid ID length: must be 6-8 characters")
	}
	if !isHexString(id) {
		return "", fmt.Errorf("invalid ID format: must be lowercase hex")
	}
	return filepath.Join(s.NotesDir, id+".md"), nil
}

// LoadDaily loads a daily note for the given date.
func (s *Store) LoadDaily(date string) (*Note, error) {
	path, err := s.DailyPath(date)
	if err != nil {
		return nil, err
	}

	return s.loadNote(path)
}

// LoadFloating loads a floating note by ID or unique prefix.
func (s *Store) LoadFloating(idOrPrefix string) (*Note, error) {
	// Try exact match first
	if len(idOrPrefix) == 8 {
		path, err := s.FloatingPath(idOrPrefix)
		if err == nil {
			if _, err := os.Stat(path); err == nil {
				return s.loadNote(path)
			}
		}
	}

	// Try prefix matching
	matches, err := s.findFloatingByPrefix(idOrPrefix)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no note found matching %q", idOrPrefix)
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf("ambiguous ID prefix %q: matches %d notes", idOrPrefix, len(matches))
	}

	return s.loadNote(matches[0])
}

// SaveDaily saves a daily note.
func (s *Store) SaveDaily(note *Note) error {
	if note.Type != NoteTypeDaily {
		return fmt.Errorf("cannot save non-daily note as daily")
	}

	path, err := s.DailyPath(note.Date)
	if err != nil {
		return err
	}

	return s.saveNote(path, note)
}

// SaveFloating saves a floating note.
func (s *Store) SaveFloating(note *Note) error {
	if note.Type != NoteTypeFloating {
		return fmt.Errorf("cannot save non-floating note as floating")
	}

	path, err := s.FloatingPath(note.ID)
	if err != nil {
		return err
	}

	return s.saveNote(path, note)
}

// DeleteDaily deletes a daily note.
func (s *Store) DeleteDaily(date string) error {
	path, err := s.DailyPath(date)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("note not found for date %s", date)
	}

	return os.Remove(path)
}

// DeleteFloating deletes a floating note by ID or prefix.
func (s *Store) DeleteFloating(idOrPrefix string) error {
	note, err := s.LoadFloating(idOrPrefix)
	if err != nil {
		return err
	}

	path, _ := s.FloatingPath(note.ID)
	return os.Remove(path)
}

// loadNote loads a note from a file path.
func (s *Store) loadNote(path string) (*Note, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("note not found")
		}
		return nil, err
	}

	// Parse frontmatter
	fm, body, err := markdown.ParseFrontmatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Unmarshal frontmatter into Note struct
	var note Note
	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	if err := yaml.Unmarshal(fmBytes, &note); err != nil {
		return nil, fmt.Errorf("failed to unmarshal frontmatter: %w", err)
	}

	note.Body = strings.TrimSpace(body)

	// Extract inline tags from body and merge with frontmatter tags
	inlineTags := markdown.ExtractInlineTags(note.Body)
	note.AddTags(inlineTags...)

	return &note, nil
}

// saveNote saves a note to a file path.
func (s *Store) saveNote(path string, note *Note) error {
	if err := s.EnsureStructure(); err != nil {
		return err
	}

	// Update timestamp
	note.Touch()

	// Serialize frontmatter and body
	content, err := markdown.SerializeFrontmatter(note, note.Body)
	if err != nil {
		return fmt.Errorf("failed to serialize note: %w", err)
	}

	// Write atomically (temp file + rename)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

// findFloatingByPrefix finds all floating notes matching the given ID prefix.
func (s *Store) findFloatingByPrefix(prefix string) ([]string, error) {
	entries, err := os.ReadDir(s.NotesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		// Check if it's a floating note (not a date)
		base := strings.TrimSuffix(name, ".md")
		if _, err := time.Parse("2006-01-02", base); err == nil {
			// It's a date, skip
			continue
		}

		// Check if it matches the prefix
		if strings.HasPrefix(base, prefix) {
			matches = append(matches, filepath.Join(s.NotesDir, name))
		}
	}

	return matches, nil
}

// isHexString returns true if s contains only lowercase hex characters.
func isHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}
