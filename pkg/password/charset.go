// Package password provides password generation and strength checking utilities.
package password

import (
	"errors"
	"strings"
)

// CharSet represents a named set of characters for password generation.
type CharSet struct {
	Name  string
	Chars string
}

// Predefined character sets.
var (
	Lowercase = CharSet{Name: "lowercase", Chars: "abcdefghijklmnopqrstuvwxyz"}
	Uppercase = CharSet{Name: "uppercase", Chars: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"}
	Digits    = CharSet{Name: "digits", Chars: "0123456789"}
	Symbols   = CharSet{Name: "symbols", Chars: "!@#$%^&*()-_=+[]{}|;:,.<>?/~"}

	// Ambiguous characters that can be confused in certain fonts.
	AmbiguousChars = "0O1lI5S"
)

// CharSetOptions configures which character sets to include.
type CharSetOptions struct {
	Lowercase      bool
	Uppercase      bool
	Digits         bool
	Symbols        bool
	AllowAmbiguous bool
}

// DefaultCharSetOptions returns the default options (lowercase, uppercase, digits, no symbols, no ambiguous).
func DefaultCharSetOptions() CharSetOptions {
	return CharSetOptions{
		Lowercase:      true,
		Uppercase:      true,
		Digits:         true,
		Symbols:        false,
		AllowAmbiguous: false,
	}
}

// ErrEmptyCharSet is returned when no character sets are enabled.
var ErrEmptyCharSet = errors.New("no character sets enabled; enable at least one of: lowercase, uppercase, digits, symbols")

// BuildCharSet composes a character set from the given options.
// Returns the combined character string, the list of enabled CharSets, and any error.
func BuildCharSet(opts CharSetOptions) (string, []CharSet, error) {
	var chars strings.Builder
	var sets []CharSet

	if opts.Lowercase {
		chars.WriteString(Lowercase.Chars)
		sets = append(sets, Lowercase)
	}
	if opts.Uppercase {
		chars.WriteString(Uppercase.Chars)
		sets = append(sets, Uppercase)
	}
	if opts.Digits {
		chars.WriteString(Digits.Chars)
		sets = append(sets, Digits)
	}
	if opts.Symbols {
		chars.WriteString(Symbols.Chars)
		sets = append(sets, Symbols)
	}

	if chars.Len() == 0 {
		return "", nil, ErrEmptyCharSet
	}

	result := chars.String()

	// Remove ambiguous characters if not allowed
	if !opts.AllowAmbiguous {
		result = removeChars(result, AmbiguousChars)
	}

	if len(result) == 0 {
		return "", nil, ErrEmptyCharSet
	}

	return result, sets, nil
}

// removeChars removes all characters in 'remove' from 's'.
func removeChars(s, remove string) string {
	var result strings.Builder
	for _, c := range s {
		if !strings.ContainsRune(remove, c) {
			result.WriteRune(c)
		}
	}
	return result.String()
}

// ContainsCharFromSet checks if the string contains at least one character from the given set.
func ContainsCharFromSet(s string, set CharSet) bool {
	for _, c := range s {
		if strings.ContainsRune(set.Chars, c) {
			return true
		}
	}
	return false
}
