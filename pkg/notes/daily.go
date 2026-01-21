package notes

import (
	"fmt"
	"strings"
	"time"
)

// AppendDailyEntry appends a new timestamped section to a daily note.
// Returns the updated note.
func AppendDailyEntry(note *Note, text string, tags []string) *Note {
	// Get current time for timestamp heading
	now := time.Now()
	heading := fmt.Sprintf("## %02d:%02d\n\n", now.Hour(), now.Minute())

	// Build entry content
	var entry strings.Builder
	entry.WriteString(heading)
	entry.WriteString(text)
	entry.WriteString("\n")

	// Add inline tags if provided
	if len(tags) > 0 {
		entry.WriteString("\n")
		for _, tag := range tags {
			entry.WriteString("#")
			entry.WriteString(tag)
			entry.WriteString(" ")
		}
		entry.WriteString("\n")
	}

	// Append to body
	if note.Body != "" {
		note.Body += "\n\n" + entry.String()
	} else {
		// First entry - add title heading
		note.Body = fmt.Sprintf("# %s\n\n", note.Date) + entry.String()
	}

	// Merge tags into frontmatter
	note.AddTags(tags...)

	return note
}

// GetOrCreateDaily loads or creates a daily note for the given date.
func (s *Store) GetOrCreateDaily(date string) (*Note, error) {
	note, err := s.LoadDaily(date)
	if err != nil {
		// Note doesn't exist, create it
		note = NewDailyNote(date)
		note.Body = fmt.Sprintf("# %s\n\n", date)
	}
	return note, nil
}
