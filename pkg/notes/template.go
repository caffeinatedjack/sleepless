// Package notes provides note storage and management functionality.
package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bufio"
)

// Template represents a note template.
type Template struct {
	Name    string
	Content string
}

// ListTemplates returns all available templates.
func (s *Store) ListTemplates() ([]Template, error) {
	if err := os.MkdirAll(s.TemplatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	entries, err := os.ReadDir(s.TemplatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []Template
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		path := filepath.Join(s.TemplatesDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue // Skip templates we can't read
		}

		templates = append(templates, Template{
			Name:    name,
			Content: string(content),
		})
	}

	return templates, nil
}

// GetTemplate retrieves a template by name.
func (s *Store) GetTemplate(name string) (*Template, error) {
	path := filepath.Join(s.TemplatesDir, name+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template %q not found", name)
		}
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	return &Template{
		Name:    name,
		Content: string(content),
	}, nil
}

// CreateTemplate creates a new template.
func (s *Store) CreateTemplate(name, content string) error {
	if err := os.MkdirAll(s.TemplatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	path := filepath.Join(s.TemplatesDir, name+".md")

	// Check if template already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("template %q already exists", name)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	return nil
}

// DeleteTemplate deletes a template.
func (s *Store) DeleteTemplate(name string) error {
	path := filepath.Join(s.TemplatesDir, name+".md")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("template %q not found", name)
		}
		return fmt.Errorf("failed to delete template: %w", err)
	}
	return nil
}

// ApplyTemplate processes a template and applies it to create note content.
// Supports placeholders:
// - {{DATE}} - current date in YYYY-MM-DD format
// - {{TIME}} - current time in HH:MM format
// - {{DATETIME}} - current datetime
// - {{PROMPT:question}} - prompts user for input
func ApplyTemplate(template string, promptReader *bufio.Reader) (string, error) {
	result := template
	now := time.Now()

	// Replace simple placeholders
	result = strings.ReplaceAll(result, "{{DATE}}", now.Format("2006-01-02"))
	result = strings.ReplaceAll(result, "{{TIME}}", now.Format("15:04"))
	result = strings.ReplaceAll(result, "{{DATETIME}}", now.Format("2006-01-02 15:04"))

	// Process PROMPT placeholders
	for {
		start := strings.Index(result, "{{PROMPT:")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}}")
		if end == -1 {
			return "", fmt.Errorf("unclosed PROMPT placeholder at position %d", start)
		}
		end += start

		// Extract prompt question
		placeholder := result[start : end+2]
		question := strings.TrimSpace(result[start+9 : end])

		// Get user input
		var answer string
		if promptReader != nil {
			fmt.Fprintf(os.Stderr, "%s: ", question)
			line, err := promptReader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("failed to read prompt answer: %w", err)
			}
			answer = strings.TrimSpace(line)
		} else {
			answer = "" // No prompt reader, use empty string
		}

		// Replace placeholder with answer
		result = strings.Replace(result, placeholder, answer, 1)
	}

	return result, nil
}

// Built-in templates
var BuiltInTemplates = map[string]string{
	"meeting": `# Meeting: {{PROMPT:Meeting title}}

**Date:** {{DATE}}
**Time:** {{TIME}}
**Attendees:** {{PROMPT:Attendees (comma-separated)}}

## Agenda
{{PROMPT:Agenda items}}

## Notes


## Action Items
- [ ] 

## Next Steps
`,

	"reflection": `# Reflection: {{DATE}}

## What went well today?
{{PROMPT:What went well}}

## What could be improved?
{{PROMPT:What could be improved}}

## Key learnings


## Tomorrow's focus
`,

	"idea": `# Idea: {{PROMPT:Idea title}}

**Date:** {{DATE}}

## The Idea
{{PROMPT:Describe the idea}}

## Why it matters


## Next steps
- [ ] 

## Related notes
`,

	"report": `# {{PROMPT:Report title}}

**Date:** {{DATE}}
**Author:** {{PROMPT:Author name}}

## Summary


## Details


## Metrics


## Recommendations


## Next Steps
- [ ] 
`,
}

// EnsureBuiltInTemplates creates built-in templates if they don't exist.
func (s *Store) EnsureBuiltInTemplates() error {
	if err := os.MkdirAll(s.TemplatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	for name, content := range BuiltInTemplates {
		path := filepath.Join(s.TemplatesDir, name+".md")

		// Only create if it doesn't exist
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create built-in template %q: %w", name, err)
			}
		}
	}

	return nil
}
