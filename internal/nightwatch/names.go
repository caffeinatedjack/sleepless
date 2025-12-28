package nightwatch

import (
	_ "embed"
	"regexp"
	"strings"

	"gitlab.com/caffeinatedjack/sleepless/pkg/redact"
)

//go:embed data/firstnames.txt
var firstnamesData string

//go:embed data/surnames.txt
var surnamesData string

var (
	firstnames map[string]bool
	surnames   map[string]bool
	nameRegex  *regexp.Regexp
)

func init() {
	firstnames = loadNames(firstnamesData)
	surnames = loadNames(surnamesData)

	// Build regex to match words that could be names (case-insensitive)
	// Matches: John, john, Mary, mary, O'Brien, McDonald, etc.
	nameRegex = regexp.MustCompile(`(?i)\b[a-z]+(?:'[a-z]+)?\b`)

	// Register the NAME pattern with the redact package
	redact.RegisterPattern(namePattern())
}

func loadNames(data string) map[string]bool {
	names := make(map[string]bool)
	for _, line := range strings.Split(data, "\n") {
		name := strings.TrimSpace(line)
		if name != "" && !strings.HasPrefix(name, "#") {
			// Store lowercase for case-insensitive matching
			names[strings.ToLower(name)] = true
		}
	}
	return names
}

func namePattern() redact.Pattern {
	return redact.Pattern{
		Type:    redact.Name,
		Regex:   nameRegex,
		Replace: "[NAME]",
		// Custom matcher - only replace if it's actually a known name
		Matcher: func(match string) bool {
			lower := strings.ToLower(match)
			return firstnames[lower] || surnames[lower]
		},
	}
}
