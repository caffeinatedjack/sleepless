// Package markdown provides frontmatter parsing and serialization.
package markdown

import (
	"bytes"
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	frontmatterDelimiter = "---"
)

// ErrNoFrontmatter is returned when a document has no frontmatter.
var ErrNoFrontmatter = errors.New("no frontmatter found")

// ParseFrontmatter extracts YAML frontmatter from markdown content.
// Returns the frontmatter metadata and the remaining body content.
func ParseFrontmatter(content string) (frontmatter map[string]interface{}, body string, err error) {
	lines := strings.Split(content, "\n")

	// Check for frontmatter delimiter at start
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != frontmatterDelimiter {
		return nil, content, ErrNoFrontmatter
	}

	// Find closing delimiter
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == frontmatterDelimiter {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return nil, content, ErrNoFrontmatter
	}

	// Extract frontmatter YAML
	frontmatterLines := lines[1:endIdx]
	frontmatterYAML := strings.Join(frontmatterLines, "\n")

	var fm map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &fm); err != nil {
		return nil, content, err
	}

	// Extract body (everything after closing delimiter)
	bodyLines := lines[endIdx+1:]
	body = strings.Join(bodyLines, "\n")

	return fm, body, nil
}

// SerializeFrontmatter converts frontmatter metadata and body into a complete markdown document.
func SerializeFrontmatter(frontmatter interface{}, body string) (string, error) {
	var buf bytes.Buffer

	// Write opening delimiter
	buf.WriteString(frontmatterDelimiter)
	buf.WriteString("\n")

	// Marshal frontmatter to YAML
	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", err
	}
	buf.Write(yamlData)

	// Write closing delimiter
	buf.WriteString(frontmatterDelimiter)
	buf.WriteString("\n")

	// Write body
	buf.WriteString(body)

	return buf.String(), nil
}

// ExtractInlineTags extracts inline #tags from markdown content.
// Returns tags in lowercase, deduplicated.
func ExtractInlineTags(content string) []string {
	tagMap := make(map[string]bool)
	words := strings.Fields(content)

	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			tag := strings.TrimPrefix(word, "#")
			tag = strings.ToLower(tag)
			// Remove trailing punctuation
			tag = strings.TrimRight(tag, ".,;:!?")
			if tag != "" {
				tagMap[tag] = true
			}
		}
	}

	var tags []string
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	return tags
}
