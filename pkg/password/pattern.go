package password

import (
	"errors"
	"fmt"
	"strings"

	"gitlab.com/caffeinatedjack/sleepless/pkg/decide"
)

// PatternToken represents a token in a parsed pattern.
type PatternToken struct {
	Type  string // "literal", "upper", "lower", "digit", "any", "alnum", "symbol"
	Value string // For literal tokens, the character(s)
}

// Pattern character meanings:
// X = random uppercase letter
// x = random lowercase letter
// 9 = random digit
// * = random printable ASCII character
// # = random alphanumeric (a-zA-Z0-9)
// ? = random symbol
// \X, \x, \9, \*, \#, \? = literal character (escaped)
// Any other character = literal

// Character sets for pattern generation.
const (
	patternUpper     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	patternLower     = "abcdefghijklmnopqrstuvwxyz"
	patternDigit     = "0123456789"
	patternAlnum     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	patternSymbol    = "!@#$%^&*()-_=+[]{}|;:,.<>?/~"
	patternPrintable = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+[]{}|;:,.<>?/~ "
)

// Special pattern characters that have meaning.
var specialPatternChars = map[rune]string{
	'X': "upper",
	'x': "lower",
	'9': "digit",
	'*': "any",
	'#': "alnum",
	'?': "symbol",
}

// ErrInvalidEscape is returned for invalid escape sequences.
var ErrInvalidEscape = errors.New("invalid escape sequence")

// ParsePattern parses a pattern string into tokens.
func ParsePattern(pattern string) ([]PatternToken, error) {
	var tokens []PatternToken
	runes := []rune(pattern)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '\\' {
			// Escape sequence
			if i+1 >= len(runes) {
				return nil, fmt.Errorf("%w: trailing backslash at position %d", ErrInvalidEscape, i)
			}
			next := runes[i+1]
			// Only allow escaping special characters
			if _, isSpecial := specialPatternChars[next]; isSpecial || next == '\\' {
				tokens = append(tokens, PatternToken{Type: "literal", Value: string(next)})
				i++ // Skip the escaped character
			} else {
				return nil, fmt.Errorf("%w: \\%c at position %d", ErrInvalidEscape, next, i)
			}
		} else if tokenType, isSpecial := specialPatternChars[r]; isSpecial {
			tokens = append(tokens, PatternToken{Type: tokenType})
		} else {
			tokens = append(tokens, PatternToken{Type: "literal", Value: string(r)})
		}
	}

	return tokens, nil
}

// GenerateFromPattern generates a string from parsed pattern tokens.
func GenerateFromPattern(tokens []PatternToken, rng decide.RNG) (string, error) {
	var result strings.Builder

	for _, token := range tokens {
		switch token.Type {
		case "literal":
			result.WriteString(token.Value)
		case "upper":
			c, err := randomChar(patternUpper, rng)
			if err != nil {
				return "", err
			}
			result.WriteByte(c)
		case "lower":
			c, err := randomChar(patternLower, rng)
			if err != nil {
				return "", err
			}
			result.WriteByte(c)
		case "digit":
			c, err := randomChar(patternDigit, rng)
			if err != nil {
				return "", err
			}
			result.WriteByte(c)
		case "alnum":
			c, err := randomChar(patternAlnum, rng)
			if err != nil {
				return "", err
			}
			result.WriteByte(c)
		case "symbol":
			c, err := randomChar(patternSymbol, rng)
			if err != nil {
				return "", err
			}
			result.WriteByte(c)
		case "any":
			c, err := randomChar(patternPrintable, rng)
			if err != nil {
				return "", err
			}
			result.WriteByte(c)
		default:
			return "", fmt.Errorf("unknown token type: %s", token.Type)
		}
	}

	return result.String(), nil
}

// randomChar selects a random character from the given charset.
func randomChar(charset string, rng decide.RNG) (byte, error) {
	idx, err := rng.Intn(len(charset))
	if err != nil {
		return 0, err
	}
	return charset[idx], nil
}

// CalculatePatternEntropy calculates the entropy of a pattern.
func CalculatePatternEntropy(tokens []PatternToken) float64 {
	var totalEntropy float64

	for _, token := range tokens {
		var charsetSize int
		switch token.Type {
		case "literal":
			charsetSize = 1 // No entropy for literals
		case "upper":
			charsetSize = len(patternUpper)
		case "lower":
			charsetSize = len(patternLower)
		case "digit":
			charsetSize = len(patternDigit)
		case "alnum":
			charsetSize = len(patternAlnum)
		case "symbol":
			charsetSize = len(patternSymbol)
		case "any":
			charsetSize = len(patternPrintable)
		}

		if charsetSize > 1 {
			totalEntropy += CalculatePasswordEntropy(charsetSize, 1)
		}
	}

	return totalEntropy
}
