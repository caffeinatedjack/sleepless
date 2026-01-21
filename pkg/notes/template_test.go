package notes

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestApplyTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		prompts  map[string]string // prompt question -> answer
		check    func(result string) error
	}{
		{
			name:     "simple placeholders",
			template: "Date: {{DATE}}, Time: {{TIME}}",
			prompts:  map[string]string{},
			check: func(result string) error {
				if !strings.Contains(result, "Date: ") {
					return errorf("missing Date")
				}
				if !strings.Contains(result, "Time: ") {
					return errorf("missing Time")
				}
				return nil
			},
		},
		{
			name:     "datetime placeholder",
			template: "Created: {{DATETIME}}",
			prompts:  map[string]string{},
			check: func(result string) error {
				if !strings.Contains(result, "Created: ") {
					return errorf("missing Created")
				}
				return nil
			},
		},
		{
			name:     "single prompt",
			template: "Title: {{PROMPT:Enter title}}",
			prompts: map[string]string{
				"Enter title": "My Title",
			},
			check: func(result string) error {
				if !strings.Contains(result, "Title: My Title") {
					return errorf("prompt not replaced: got %q", result)
				}
				return nil
			},
		},
		{
			name:     "multiple prompts",
			template: "Meeting: {{PROMPT:Title}}\nAttendees: {{PROMPT:Who attended?}}",
			prompts: map[string]string{
				"Title":         "Sprint Planning",
				"Who attended?": "Alice, Bob",
			},
			check: func(result string) error {
				if !strings.Contains(result, "Meeting: Sprint Planning") {
					return errorf("first prompt not replaced")
				}
				if !strings.Contains(result, "Attendees: Alice, Bob") {
					return errorf("second prompt not replaced")
				}
				return nil
			},
		},
		{
			name:     "mixed placeholders and prompts",
			template: "# {{PROMPT:Title}}\n\nDate: {{DATE}}\nAuthor: {{PROMPT:Author}}",
			prompts: map[string]string{
				"Title":  "Weekly Report",
				"Author": "John Doe",
			},
			check: func(result string) error {
				if !strings.Contains(result, "# Weekly Report") {
					return errorf("title prompt not replaced")
				}
				if !strings.Contains(result, "Date: ") {
					return errorf("date placeholder not replaced")
				}
				if !strings.Contains(result, "Author: John Doe") {
					return errorf("author prompt not replaced")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock reader that returns prompt answers
			var inputLines []string
			for question, answer := range tt.prompts {
				// We need to match prompts in order they appear
				if strings.Contains(tt.template, "{{PROMPT:"+question+"}}") {
					inputLines = append(inputLines, answer)
				}
			}
			input := strings.Join(inputLines, "\n") + "\n"
			reader := bufio.NewReader(strings.NewReader(input))

			result, err := ApplyTemplate(tt.template, reader)
			if err != nil {
				t.Fatalf("ApplyTemplate failed: %v", err)
			}

			if err := tt.check(result); err != nil {
				t.Errorf("Check failed: %v\nGot: %s", err, result)
			}
		})
	}
}

func TestApplyTemplateNoReader(t *testing.T) {
	// When no reader is provided, prompts should be replaced with empty strings
	template := "Title: {{PROMPT:Enter title}}"
	result, err := ApplyTemplate(template, nil)
	if err != nil {
		t.Fatalf("ApplyTemplate failed: %v", err)
	}

	if result != "Title: " {
		t.Errorf("Expected 'Title: ', got %q", result)
	}
}

func TestApplyTemplateUnclosedPrompt(t *testing.T) {
	template := "Title: {{PROMPT:Enter title"
	_, err := ApplyTemplate(template, nil)
	if err == nil {
		t.Error("Expected error for unclosed prompt")
	}
}

func TestBuiltInTemplates(t *testing.T) {
	expectedTemplates := []string{"meeting", "reflection", "idea", "report"}

	for _, name := range expectedTemplates {
		t.Run(name, func(t *testing.T) {
			tmpl, ok := BuiltInTemplates[name]
			if !ok {
				t.Errorf("Built-in template %q not found", name)
			}

			if tmpl == "" {
				t.Errorf("Built-in template %q is empty", name)
			}

			// Verify it contains some expected structure
			if !strings.Contains(tmpl, "{{") {
				t.Errorf("Built-in template %q should contain placeholders", name)
			}
		})
	}
}

func TestStoreTemplateOperations(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Test CreateTemplate
	tmplName := "test-template"
	tmplContent := "# Test\n\n{{PROMPT:Question}}"

	err := store.CreateTemplate(tmplName, tmplContent)
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	// Test GetTemplate
	tmpl, err := store.GetTemplate(tmplName)
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}

	if tmpl.Name != tmplName {
		t.Errorf("Expected name %q, got %q", tmplName, tmpl.Name)
	}

	if tmpl.Content != tmplContent {
		t.Errorf("Expected content %q, got %q", tmplContent, tmpl.Content)
	}

	// Test ListTemplates
	templates, err := store.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}

	found := false
	for _, t := range templates {
		if t.Name == tmplName {
			found = true
			break
		}
	}

	if !found {
		t.Error("Created template not found in list")
	}

	// Test DeleteTemplate
	err = store.DeleteTemplate(tmplName)
	if err != nil {
		t.Fatalf("DeleteTemplate failed: %v", err)
	}

	// Verify deletion
	_, err = store.GetTemplate(tmplName)
	if err == nil {
		t.Error("Expected error when getting deleted template")
	}
}

func TestStoreEnsureBuiltInTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Ensure built-in templates
	err := store.EnsureBuiltInTemplates()
	if err != nil {
		t.Fatalf("EnsureBuiltInTemplates failed: %v", err)
	}

	// Verify all built-in templates exist
	for name := range BuiltInTemplates {
		tmpl, err := store.GetTemplate(name)
		if err != nil {
			t.Errorf("Built-in template %q not created: %v", name, err)
			continue
		}

		if tmpl.Content == "" {
			t.Errorf("Built-in template %q has empty content", name)
		}
	}

	// Run again - should be idempotent (not overwrite existing)
	err = store.EnsureBuiltInTemplates()
	if err != nil {
		t.Errorf("EnsureBuiltInTemplates should be idempotent: %v", err)
	}
}

func TestCreateTemplateDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	tmplName := "duplicate"
	tmplContent := "test content"

	// First creation should succeed
	err := store.CreateTemplate(tmplName, tmplContent)
	if err != nil {
		t.Fatalf("First CreateTemplate failed: %v", err)
	}

	// Second creation should fail
	err = store.CreateTemplate(tmplName, tmplContent)
	if err == nil {
		t.Error("Expected error when creating duplicate template")
	}
}

func TestGetTemplateNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	_, err := store.GetTemplate("nonexistent")
	if err == nil {
		t.Error("Expected error when getting nonexistent template")
	}
}

func TestDeleteTemplateNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	err := store.DeleteTemplate("nonexistent")
	if err == nil {
		t.Error("Expected error when deleting nonexistent template")
	}
}

// Helper function for creating errors in tests
func errorf(format string, args ...interface{}) error {
	return &testError{msg: strings.TrimSpace(fmt.Sprintf(format, args...))}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
