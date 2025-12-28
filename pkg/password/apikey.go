package password

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/caffeinatedjack/sleepless/pkg/decide"
)

// APIKeyFormat represents the encoding format for API keys.
type APIKeyFormat string

const (
	FormatHex       APIKeyFormat = "hex"
	FormatBase32    APIKeyFormat = "base32"
	FormatBase64    APIKeyFormat = "base64"
	FormatBase64URL APIKeyFormat = "base64url"
)

// ValidFormats lists all valid API key formats.
var ValidFormats = []APIKeyFormat{FormatHex, FormatBase32, FormatBase64, FormatBase64URL}

// ErrInvalidFormat is returned for unknown formats.
var ErrInvalidFormat = errors.New("invalid format")

// ParseFormat parses a format string into an APIKeyFormat.
func ParseFormat(s string) (APIKeyFormat, error) {
	switch strings.ToLower(s) {
	case "hex":
		return FormatHex, nil
	case "base32":
		return FormatBase32, nil
	case "base64":
		return FormatBase64, nil
	case "base64url":
		return FormatBase64URL, nil
	default:
		return "", fmt.Errorf("%w: %q (valid formats: hex, base32, base64, base64url)", ErrInvalidFormat, s)
	}
}

// APIKeyOptions configures API key generation.
type APIKeyOptions struct {
	Prefix string
	Length int // Length of random part (in output characters)
	Format APIKeyFormat
}

// DefaultAPIKeyOptions returns sensible defaults for API key generation.
func DefaultAPIKeyOptions() APIKeyOptions {
	return APIKeyOptions{
		Prefix: "",
		Length: 32,
		Format: FormatHex,
	}
}

// GenerateAPIKey generates an API key with the given options.
func GenerateAPIKey(opts APIKeyOptions, rng decide.RNG) (string, error) {
	if opts.Length < 1 {
		return "", errors.New("API key length must be at least 1")
	}

	// Calculate how many random bytes we need based on format
	var randomBytes []byte
	var encoded string

	switch opts.Format {
	case FormatHex:
		// Hex: 2 chars per byte, so we need length/2 bytes (round up)
		numBytes := (opts.Length + 1) / 2
		randomBytes = make([]byte, numBytes)
		for i := range randomBytes {
			b, err := rng.Intn(256)
			if err != nil {
				return "", fmt.Errorf("failed to generate random byte: %w", err)
			}
			randomBytes[i] = byte(b)
		}
		encoded = hex.EncodeToString(randomBytes)
		if len(encoded) > opts.Length {
			encoded = encoded[:opts.Length]
		}

	case FormatBase32:
		// Base32: 8 chars per 5 bytes
		numBytes := (opts.Length*5 + 7) / 8
		randomBytes = make([]byte, numBytes)
		for i := range randomBytes {
			b, err := rng.Intn(256)
			if err != nil {
				return "", fmt.Errorf("failed to generate random byte: %w", err)
			}
			randomBytes[i] = byte(b)
		}
		encoded = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
		if len(encoded) > opts.Length {
			encoded = encoded[:opts.Length]
		}

	case FormatBase64:
		// Base64: 4 chars per 3 bytes
		numBytes := (opts.Length*3 + 3) / 4
		randomBytes = make([]byte, numBytes)
		for i := range randomBytes {
			b, err := rng.Intn(256)
			if err != nil {
				return "", fmt.Errorf("failed to generate random byte: %w", err)
			}
			randomBytes[i] = byte(b)
		}
		encoded = base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(randomBytes)
		if len(encoded) > opts.Length {
			encoded = encoded[:opts.Length]
		}

	case FormatBase64URL:
		// Base64URL: 4 chars per 3 bytes
		numBytes := (opts.Length*3 + 3) / 4
		randomBytes = make([]byte, numBytes)
		for i := range randomBytes {
			b, err := rng.Intn(256)
			if err != nil {
				return "", fmt.Errorf("failed to generate random byte: %w", err)
			}
			randomBytes[i] = byte(b)
		}
		encoded = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(randomBytes)
		if len(encoded) > opts.Length {
			encoded = encoded[:opts.Length]
		}

	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidFormat, opts.Format)
	}

	// Add prefix if specified
	if opts.Prefix != "" {
		return opts.Prefix + "_" + encoded, nil
	}

	return encoded, nil
}

// APIKeyEntropyBits returns the entropy in bits for an API key of given format and length.
func APIKeyEntropyBits(format APIKeyFormat, length int) float64 {
	var charsetSize int
	switch format {
	case FormatHex:
		charsetSize = 16
	case FormatBase32:
		charsetSize = 32
	case FormatBase64, FormatBase64URL:
		charsetSize = 64
	default:
		return 0
	}
	return CalculatePasswordEntropy(charsetSize, length)
}
