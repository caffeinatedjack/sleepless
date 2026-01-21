package notes

import (
	"reflect"
	"testing"
)

func TestParseTaskReferences(t *testing.T) {
	content := `Had a meeting about @task:abc123 and also discussed @task:def456.
Follow up on @task:abc123 again later.`

	refs := ParseTaskReferences(content)

	expected := []string{"abc123", "def456"}

	if len(refs) != len(expected) {
		t.Errorf("expected %d refs, got %d", len(expected), len(refs))
	}

	refMap := make(map[string]bool)
	for _, ref := range refs {
		refMap[ref] = true
	}

	for _, exp := range expected {
		if !refMap[exp] {
			t.Errorf("expected reference %s not found", exp)
		}
	}
}

func TestParseTaskReferencesNoRefs(t *testing.T) {
	content := "This note has no task references."

	refs := ParseTaskReferences(content)

	if len(refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(refs))
	}
}

func TestParseTaskReferencesDeduplicate(t *testing.T) {
	content := "@task:abc123 and @task:abc123 again."

	refs := ParseTaskReferences(content)

	if len(refs) != 1 {
		t.Errorf("expected 1 unique ref, got %d", len(refs))
	}

	if !reflect.DeepEqual(refs, []string{"abc123"}) {
		t.Errorf("expected [abc123], got %v", refs)
	}
}
