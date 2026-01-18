// Package oidc implements OpenID Connect (OIDC) utilities for nightwatch.
// It includes PKCE generation, state/nonce generation, authorization URL building,
// callback parsing, and OIDC ID token linting.
package oidc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// LintResult represents the result of an ID token lint.
type LintResult struct {
	Valid     bool       `json:"valid"`
	Warnings  []string   `json:"warnings,omitempty"`
	Issuer    *string    `json:"issuer,omitempty"`
	Audience  *string    `json:"audience,omitempty"`
	Expiry    *time.Time `json:"expiry,omitempty"`
	NotBefore *time.Time `json:"not_before,omitempty"`
	IssuedAt  *time.Time `json:"issued_at,omitempty"`
	Nonce     *string    `json:"nonce,omitempty"`
}

// GeneratePKCE generates a PKCE verifier and S256 challenge per RFC 7636.
// The verifier is a URL-safe random string, and the challenge is
// base64url-encoded(sha256(verifier || .)).
func GeneratePKCE() (verifier, challenge string, err error) {
	// Generate a cryptographically random verifier (43-128 characters per RFC 7636)
	// Use 32 bytes = 256 bits -> ~43 base64url chars
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	verifier = base64.RawURLEncoding.EncodeToString(b)

	// Compute S256 challenge: BASE64URL-ENCODE(SHA256(ASCII(verifier)))
	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge, nil
}

// GenerateState generates a URL-safe random state value.
func GenerateState() (string, error) {
	return generateRandomString(32)
}

// GenerateNonce generates a URL-safe random nonce value.
func GenerateNonce() (string, error) {
	return generateRandomString(32)
}

// generateRandomString generates a URL-safe random string of length n using crypto/rand.
func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// BuildAuthURL constructs an authorization URL from the provided parameters.
// It does NOT modify or reorder user-provided values.
func BuildAuthURL(authEndpoint, clientID, redirectURI, scope string, opts AuthURLOptions) (string, error) {
	// Validate required parameters
	if authEndpoint == "" {
		return "", fmt.Errorf("auth-endpoint is required")
	}
	if clientID == "" {
		return "", fmt.Errorf("client-id is required")
	}
	if redirectURI == "" {
		return "", fmt.Errorf("redirect-uri is required")
	}
	if scope == "" {
		return "", fmt.Errorf("scope is required")
	}

	// Parse the base auth endpoint URL
	u, err := url.Parse(authEndpoint)
	if err != nil {
		return "", fmt.Errorf("invalid auth-endpoint URL: %w", err)
	}

	// Build query parameters
	query := url.Values{}

	// Required OIDC parameters
	query.Set("response_type", "code")
	query.Set("client_id", clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("scope", scope)

	// Optional parameters
	if opts.State != "" {
		query.Set("state", opts.State)
	}
	if opts.Nonce != "" {
		query.Set("nonce", opts.Nonce)
	}
	if opts.PKCEChallenge != "" {
		query.Set("code_challenge", opts.PKCEChallenge)
		query.Set("code_challenge_method", "S256")
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

// AuthURLOptions represents optional parameters for BuildAuthURL.
type AuthURLOptions struct {
	State         string
	Nonce         string
	PKCEChallenge string
}

// ParseCallback parses an authorization callback URL and extracts relevant parameters.
func ParseCallback(callbackURL string) (map[string]string, error) {
	u, err := url.Parse(callbackURL)
	if err != nil {
		return nil, fmt.Errorf("invalid callback URL: %w", err)
	}

	query := u.Query()
	result := make(map[string]string)
	for key, values := range query {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result, nil
}

// LintIDToken performs OIDC-specific checks on an ID token payload.
// It accepts the parsed JWT payload and optional expectations.
func LintIDToken(payload map[string]interface{}, issuer, audience string, clockSkew time.Duration) (*LintResult, error) {
	result := &LintResult{Valid: true}

	// Extract and validate issuer
	if iss, ok := payload["iss"].(string); ok {
		result.Issuer = &iss
		if issuer != "" && iss != issuer {
			result.Warnings = append(result.Warnings, fmt.Sprintf("issuer mismatch: expected %q, got %q", issuer, iss))
			result.Valid = false
		}
	} else if issuer != "" {
		result.Warnings = append(result.Warnings, "missing or invalid 'iss' claim")
		result.Valid = false
	}

	// Extract and validate audience
	if audClaim, ok := payload["aud"]; ok {
		audienceMatched := false
		switch v := audClaim.(type) {
		case string:
			if audience != "" && v == audience {
				audienceMatched = true
			}
			result.Audience = &v
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok && (audience == "" || s == audience) {
					if audience == "" && result.Audience == nil {
						result.Audience = &s
					}
					if s == audience {
						audienceMatched = true
					}
				}
			}
		}
		if audience != "" && !audienceMatched {
			result.Warnings = append(result.Warnings, fmt.Sprintf("audience mismatch: expected %q not in aud claim", audience))
			result.Valid = false
		}
	} else if audience != "" {
		result.Warnings = append(result.Warnings, "missing or invalid 'aud' claim")
		result.Valid = false
	}

	// Extract subject (no validation needed)
	if sub, ok := payload["sub"].(string); ok {
		// Just store for reference; validation is provider-specific
		_ = sub
	}

	// Extract and validate time-based claims
	now := time.Now().UTC()
	allowanceBefore := now.Add(-clockSkew)

	// Validate exp (expiration)
	if expClaim, ok := payload["exp"]; ok {
		var expTime time.Time
		switch v := expClaim.(type) {
		case float64:
			expTime = time.Unix(int64(v), 0).UTC()
		case string:
			if i, err := strconv.ParseFloat(v, 64); err == nil {
				expTime = time.Unix(int64(i), 0).UTC()
			}
		}

		if !expTime.IsZero() {
			result.Expiry = &expTime
			if expTime.Before(allowanceBefore) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("token expired at %s", expTime))
				result.Valid = false
			}
		}
	}

	// Validate iat (issued at)
	if iatClaim, ok := payload["iat"]; ok {
		var iatTime time.Time
		switch v := iatClaim.(type) {
		case float64:
			iatTime = time.Unix(int64(v), 0).UTC()
		case string:
			if i, err := strconv.ParseFloat(v, 64); err == nil {
				iatTime = time.Unix(int64(i), 0).UTC()
			}
		}

		if !iatTime.IsZero() {
			result.IssuedAt = &iatTime
			// Check if iat is unreasonably in the future (e.g., > 5 minutes)
			if iatTime.After(now.Add(5 * time.Minute)) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("issued-at time %s is unreasonably in the future (>5 minutes)", iatTime))
				result.Valid = false
			}
		}
	}

	// Extract nonce (for information only, no validation since it's opaque to us)
	if nonce, ok := payload["nonce"].(string); ok {
		result.Nonce = &nonce
	}

	return result, nil
}
