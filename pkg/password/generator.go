package password

import (
	"errors"
	"fmt"
	"strings"

	"gitlab.com/caffeinatedjack/sleepless/pkg/decide"
)

// Generator handles password generation using a provided RNG.
type Generator struct {
	rng decide.RNG
}

// NewGenerator creates a new password generator with the given RNG.
func NewGenerator(rng decide.RNG) *Generator {
	return &Generator{rng: rng}
}

// PasswordOptions configures character-based password generation.
type PasswordOptions struct {
	Length     int
	CharSet    CharSetOptions
	RequireAll bool // Require at least one character from each enabled set
}

// DefaultPasswordOptions returns sensible defaults for password generation.
func DefaultPasswordOptions() PasswordOptions {
	return PasswordOptions{
		Length:     16,
		CharSet:    DefaultCharSetOptions(),
		RequireAll: true,
	}
}

// ErrInvalidLength is returned when password length is invalid.
var ErrInvalidLength = errors.New("password length must be at least 1")

// MaxPasswordLength is the maximum allowed password length.
const MaxPasswordLength = 1024

// ErrLengthTooLong is returned when password length exceeds maximum.
var ErrLengthTooLong = fmt.Errorf("password length cannot exceed %d", MaxPasswordLength)

// GeneratePassword generates a password with the given options.
func (g *Generator) GeneratePassword(opts PasswordOptions) (string, error) {
	if opts.Length < 1 {
		return "", ErrInvalidLength
	}
	if opts.Length > MaxPasswordLength {
		return "", ErrLengthTooLong
	}

	charset, sets, err := BuildCharSet(opts.CharSet)
	if err != nil {
		return "", err
	}

	// If RequireAll is true and we have multiple sets, we need to ensure
	// at least one character from each set is included.
	if opts.RequireAll && len(sets) > 1 {
		// Check if length is sufficient for all required sets
		if opts.Length < len(sets) {
			return "", fmt.Errorf("password length %d is too short to include characters from all %d enabled sets", opts.Length, len(sets))
		}
		return g.generateWithRequirement(opts.Length, charset, sets)
	}

	return g.generateSimple(opts.Length, charset)
}

// generateSimple generates a password without charset representation requirements.
func (g *Generator) generateSimple(length int, charset string) (string, error) {
	var result strings.Builder
	result.Grow(length)

	for i := 0; i < length; i++ {
		idx, err := g.rng.Intn(len(charset))
		if err != nil {
			return "", fmt.Errorf("failed to generate random index: %w", err)
		}
		result.WriteByte(charset[idx])
	}

	return result.String(), nil
}

// generateWithRequirement generates a password ensuring at least one char from each set.
// Strategy: place one character from each required set at random positions,
// then fill remaining positions with random characters from the full charset.
func (g *Generator) generateWithRequirement(length int, charset string, sets []CharSet) (string, error) {
	result := make([]byte, length)

	// Track which positions are already filled
	filled := make([]bool, length)

	// For each required set, place one character at a random unfilled position
	for _, set := range sets {
		// Find an unfilled position
		pos, err := g.findUnfilledPosition(filled)
		if err != nil {
			return "", err
		}

		// Pick a random character from this set
		// Need to filter charset to only include chars from this set
		setChars := filterCharset(charset, set.Chars)
		if len(setChars) == 0 {
			// This set's characters were filtered out (e.g., by ambiguous removal)
			// Skip this requirement
			continue
		}

		idx, err := g.rng.Intn(len(setChars))
		if err != nil {
			return "", fmt.Errorf("failed to generate random index: %w", err)
		}

		result[pos] = setChars[idx]
		filled[pos] = true
	}

	// Fill remaining positions with random characters from full charset
	for i := 0; i < length; i++ {
		if !filled[i] {
			idx, err := g.rng.Intn(len(charset))
			if err != nil {
				return "", fmt.Errorf("failed to generate random index: %w", err)
			}
			result[i] = charset[idx]
		}
	}

	return string(result), nil
}

// findUnfilledPosition finds a random unfilled position in the filled slice.
func (g *Generator) findUnfilledPosition(filled []bool) (int, error) {
	// Count unfilled positions
	var unfilled []int
	for i, f := range filled {
		if !f {
			unfilled = append(unfilled, i)
		}
	}

	if len(unfilled) == 0 {
		return 0, errors.New("no unfilled positions available")
	}

	idx, err := g.rng.Intn(len(unfilled))
	if err != nil {
		return 0, err
	}

	return unfilled[idx], nil
}

// filterCharset returns only characters that appear in both charset and allowed.
func filterCharset(charset, allowed string) string {
	var result strings.Builder
	for _, c := range charset {
		if strings.ContainsRune(allowed, c) {
			result.WriteRune(c)
		}
	}
	return result.String()
}

// GetCharSetSize returns the size of the charset for the given options.
// This is useful for entropy calculation.
func GetCharSetSize(opts CharSetOptions) (int, error) {
	charset, _, err := BuildCharSet(opts)
	if err != nil {
		return 0, err
	}
	return len(charset), nil
}
