package notes

import (
	"testing"
	"time"
)

func TestNewDailyNote(t *testing.T) {
	date := "2026-01-21"
	note := NewDailyNote(date)

	if note.Type != NoteTypeDaily {
		t.Errorf("expected type %s, got %s", NoteTypeDaily, note.Type)
	}
	if note.Date != date {
		t.Errorf("expected date %s, got %s", date, note.Date)
	}
	if note.ID != "" {
		t.Errorf("daily note should not have ID")
	}
}

func TestNewFloatingNote(t *testing.T) {
	id := "abc123de"
	note := NewFloatingNote(id)

	if note.Type != NoteTypeFloating {
		t.Errorf("expected type %s, got %s", NoteTypeFloating, note.Type)
	}
	if note.ID != id {
		t.Errorf("expected ID %s, got %s", id, note.ID)
	}
	if note.Date != "" {
		t.Errorf("floating note should not have date")
	}
}

func TestGenerateID(t *testing.T) {
	id, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID failed: %v", err)
	}

	if len(id) != 8 {
		t.Errorf("expected ID length 8, got %d", len(id))
	}

	// Check that it's lowercase hex
	for _, r := range id {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			t.Errorf("ID contains non-hex character: %c", r)
		}
	}
}

func TestAddTags(t *testing.T) {
	note := NewDailyNote("2026-01-21")
	note.AddTags("work", "meeting")

	if len(note.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(note.Tags))
	}

	// Adding duplicate should not increase count
	note.AddTags("work")
	if len(note.Tags) != 2 {
		t.Errorf("expected 2 tags after adding duplicate, got %d", len(note.Tags))
	}
}

func TestRemoveTags(t *testing.T) {
	note := NewDailyNote("2026-01-21")
	note.AddTags("work", "meeting", "urgent")

	note.RemoveTags("meeting")
	if len(note.Tags) != 2 {
		t.Errorf("expected 2 tags after removal, got %d", len(note.Tags))
	}

	if note.HasTag("meeting") {
		t.Errorf("tag 'meeting' should have been removed")
	}
}

func TestTouch(t *testing.T) {
	note := NewDailyNote("2026-01-21")
	before := time.Now()
	time.Sleep(10 * time.Millisecond)
	note.Touch()

	if note.Updated.Before(before) {
		t.Errorf("Touch() should update timestamp")
	}
}
