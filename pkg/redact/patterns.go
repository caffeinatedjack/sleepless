// Package redact provides PII and secret redaction utilities.
package redact

import (
	"regexp"
)

// PatternType identifies a category of sensitive data.
type PatternType string

const (
	Email      PatternType = "EMAIL"
	Phone      PatternType = "PHONE"
	IP         PatternType = "IP"
	CreditCard PatternType = "CREDIT_CARD"
	UUID       PatternType = "UUID"
	Name       PatternType = "NAME"
)

// MatchFunc is an optional function to validate regex matches.
// If provided, only matches where this returns true are redacted.
type MatchFunc func(match string) bool

// Pattern defines a regex pattern for detecting sensitive data.
type Pattern struct {
	Type    PatternType
	Regex   *regexp.Regexp
	Replace string
	Matcher MatchFunc // Optional: if set, only redact when Matcher returns true
}

// DefaultPatterns are enabled by default.
var DefaultPatterns = []Pattern{
	{
		Type:    Email,
		Regex:   regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
		Replace: "[EMAIL]",
	},
	{
		Type:    Phone,
		Regex:   regexp.MustCompile(`(?:\+?1[-.\s]?)?(?:\(?\d{3}\)?[-.\s]?)?\d{3}[-.\s]?\d{4}`),
		Replace: "[PHONE]",
	},
	{
		Type: IP,
		Regex: regexp.MustCompile(`(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)|` +
			`(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|` +
			`(?:[0-9a-fA-F]{1,4}:){1,7}:|` +
			`::(?:[0-9a-fA-F]{1,4}:){0,6}[0-9a-fA-F]{1,4}`),
		Replace: "[IP]",
	},
	{
		Type:    CreditCard,
		Regex:   regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`),
		Replace: "[CREDIT_CARD]",
	},
	{
		Type:    UUID,
		Regex:   regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`),
		Replace: "[UUID]",
	},
}

// RegisterPattern appends p to the default pattern list.
//
// Nightwatch uses this to add the NAME matcher at startup.
func RegisterPattern(p Pattern) {
	DefaultPatterns = append(DefaultPatterns, p)
}

// AllPatternTypes lists the built-in pattern types.
func AllPatternTypes() []PatternType {
	return []PatternType{
		Email, Phone, IP, CreditCard, UUID, Name,
	}
}

// PatternsByType builds a lookup table from DefaultPatterns.
//
// If multiple patterns share a type, the last one wins.
func PatternsByType() map[PatternType]Pattern {
	m := make(map[PatternType]Pattern)
	for _, p := range DefaultPatterns {
		m[p.Type] = p
	}
	return m
}
