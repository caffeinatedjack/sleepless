package markdown

import (
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	content := `---
title: Test Note
tags:
  - work
  - meeting
---

# Test Note

This is the body content.`

	fm, body, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseFrontmatter failed: %v", err)
	}

	if fm == nil {
		t.Fatal("frontmatter should not be nil")
	}

	if title, ok := fm["title"].(string); !ok || title != "Test Note" {
		t.Errorf("expected title 'Test Note', got %v", fm["title"])
	}

	if !strings.Contains(body, "# Test Note") {
		t.Errorf("body should contain title heading")
	}
}

func TestParseFrontmatterNoFrontmatter(t *testing.T) {
	content := "# Just a heading\n\nNo frontmatter here."

	_, body, err := ParseFrontmatter(content)
	if err != ErrNoFrontmatter {
		t.Errorf("expected ErrNoFrontmatter, got %v", err)
	}

	if body != content {
		t.Errorf("body should be unchanged when no frontmatter")
	}
}

func TestSerializeFrontmatter(t *testing.T) {
	fm := map[string]interface{}{
		"title": "Test",
		"tags":  []string{"work"},
	}
	body := "# Test\n\nContent here."

	result, err := SerializeFrontmatter(fm, body)
	if err != nil {
		t.Fatalf("SerializeFrontmatter failed: %v", err)
	}

	if !strings.HasPrefix(result, "---\n") {
		t.Errorf("result should start with frontmatter delimiter")
	}

	if !strings.Contains(result, body) {
		t.Errorf("result should contain body")
	}
}

func TestExtractInlineTags(t *testing.T) {
	content := "This is a note about #work and #meeting. Also #urgent!"

	tags := ExtractInlineTags(content)
	expectedTags := map[string]bool{
		"work":    true,
		"meeting": true,
		"urgent":  true,
	}

	if len(tags) != len(expectedTags) {
		t.Errorf("expected %d tags, got %d", len(expectedTags), len(tags))
	}

	for _, tag := range tags {
		if !expectedTags[tag] {
			t.Errorf("unexpected tag: %s", tag)
		}
	}
}

func TestExtractInlineTagsNoTags(t *testing.T) {
	content := "This note has no tags."

	tags := ExtractInlineTags(content)
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}
}
