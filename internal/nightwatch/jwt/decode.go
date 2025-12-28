// Package jwt implements JWT utilities for nightwatch.
package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// ErrorType represents the type of JWT error for exit code mapping.
type ErrorType int

const (
	ErrInvalidFormat ErrorType = iota
	ErrVerificationFailed
	ErrExpired
	ErrNotYetValid
	ErrUnsupportedAlgorithm
	ErrKeyLoad
)

// JWTError represents an error with an associated exit code.
type JWTError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *JWTError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *JWTError) Unwrap() error {
	return e.Err
}

// ExitCode returns the appropriate exit code per specification.
func (e *JWTError) ExitCode() int {
	switch e.Type {
	case ErrVerificationFailed:
		return 2
	case ErrExpired:
		return 3
	case ErrNotYetValid:
		return 4
	default:
		return 1
	}
}

// NewError creates a new JWTError.
func NewError(t ErrorType, msg string, err error) *JWTError {
	return &JWTError{Type: t, Message: msg, Err: err}
}

// DecodedToken represents a parsed JWT.
type DecodedToken struct {
	Header    map[string]interface{} `json:"header"`
	Payload   map[string]interface{} `json:"payload"`
	Signature string                 `json:"signature"`
	Verified  bool                   `json:"verified"`
	Raw       string                 `json:"raw"`
}

// DecodeWithoutVerification parses a JWT without verifying the signature.
func DecodeWithoutVerification(token string) (*DecodedToken, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, NewError(ErrInvalidFormat, "invalid token format: expected 3 parts separated by dots", nil)
	}

	header, err := decodeSegment(parts[0], "header")
	if err != nil {
		return nil, err
	}

	payload, err := decodeSegment(parts[1], "payload")
	if err != nil {
		return nil, err
	}

	return &DecodedToken{
		Header:    header,
		Payload:   payload,
		Signature: parts[2],
		Verified:  false,
		Raw:       token,
	}, nil
}

// decodeSegment decodes a base64url JWT segment into a map.
func decodeSegment(segment, name string) (map[string]interface{}, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return nil, NewError(ErrInvalidFormat, fmt.Sprintf("invalid base64url encoding in %s", name), err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, NewError(ErrInvalidFormat, fmt.Sprintf("invalid JSON in %s", name), err)
	}

	return result, nil
}

// ParseClaimValue parses a claim value string, attempting JSON parsing first.
func ParseClaimValue(value string) interface{} {
	var result interface{}
	if err := json.Unmarshal([]byte(value), &result); err == nil {
		return result
	}
	return value
}

// ValidateAlgorithm checks if the algorithm is supported.
func ValidateAlgorithm(alg string) error {
	switch alg {
	case "HS256", "HS384", "HS512", "RS256", "RS384", "RS512", "ES256", "ES384", "ES512":
		return nil
	default:
		return NewError(ErrUnsupportedAlgorithm,
			fmt.Sprintf("unsupported algorithm: %s (supported: HS256, HS384, HS512, RS256, RS384, RS512, ES256, ES384, ES512)", alg), nil)
	}
}

// GetAlgorithmType returns "HMAC", "RSA", or "ECDSA" based on the algorithm prefix.
func GetAlgorithmType(alg string) string {
	if len(alg) < 2 {
		return "UNKNOWN"
	}
	switch alg[:2] {
	case "HS":
		return "HMAC"
	case "RS":
		return "RSA"
	case "ES":
		return "ECDSA"
	default:
		return "UNKNOWN"
	}
}
