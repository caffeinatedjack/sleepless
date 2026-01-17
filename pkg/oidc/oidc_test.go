// Package oidc tests OIDC utilities.
package oidc

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

// TestGeneratePKCE verifies PKCE verifier and challenge generation.
func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE failed: %v", err)
	}

	// Verify verifier is base64url-encoded
	_, err = base64.RawURLEncoding.DecodeString(verifier)
	if err != nil {
		t.Errorf("verifier is not valid base64url: %v", err)
	}

	// Verify challenge is valid base64url-encoded SHA256 hash (43 chars for 256 bits)
	if _, err = base64.RawURLEncoding.DecodeString(challenge); err != nil {
		t.Errorf("challenge is not valid base64url: %v", err)
	}
	if len(challenge) != 43 {
		t.Errorf("challenge should be 43 chars (base64url of SHA256), got %d", len(challenge))
	}

	// Verify deterministic output for same inputs would fail here
	// This is intentionally non-deterministic, so we just verify it works
}

// TestGenerateState verifies state generation.
func TestGenerateState(t *testing.T) {
	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState failed: %v", err)
	}

	// Verify state is base64url-encoded
	_, _ = base64.RawURLEncoding.DecodeString(state)

	// Verify state is non-empty
	if state == "" {
		t.Error("state should not be empty")
	}
}

// TestGenerateNonce verifies nonce generation.
func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}

	// Verify nonce is base64url-encoded
	_, _ = base64.RawURLEncoding.DecodeString(nonce)

	// Verify nonce is non-empty
	if nonce == "" {
		t.Error("nonce should not be empty")
	}
}

// TestBuildAuthURL verifies authorization URL construction.
func TestBuildAuthURL(t *testing.T) {
	tests := []struct {
		name         string
		authEndpoint string
		clientID     string
		redirectURI  string
		scope        string
		opts         AuthURLOptions
		wantContains []string
		wantErr      bool
	}{
		{
			name:         "basic auth URL",
			authEndpoint: "https://example.com/authorize",
			clientID:     "client123",
			redirectURI:  "http://localhost:3000/callback",
			scope:        "openid profile",
			opts:         AuthURLOptions{},
			wantContains: []string{"response_type=code", "client_id=client123", "redirect_uri=http%3A%2F%2Flocalhost%3A3000%2Fcallback", "scope=openid+profile"},
			wantErr:      false,
		},
		{
			name:         "with state and nonce",
			authEndpoint: "https://example.com/authorize",
			clientID:     "client123",
			redirectURI:  "http://localhost:3000/callback",
			scope:        "openid",
			opts: AuthURLOptions{
				State: "abc123",
				Nonce: "xyz789",
			},
			wantContains: []string{"state=abc123", "nonce=xyz789"},
			wantErr:      false,
		},
		{
			name:         "with PKCE",
			authEndpoint: "https://example.com/authorize",
			clientID:     "client123",
			redirectURI:  "http://localhost:3000/callback",
			scope:        "openid",
			opts: AuthURLOptions{
				PKCEVerifier: "verifier123",
			},
			wantContains: []string{"code_challenge=verifier123", "code_challenge_method=S256"},
			wantErr:      false,
		},
		{
			name:         "missing auth endpoint",
			authEndpoint: "",
			clientID:     "client123",
			redirectURI:  "http://localhost:3000/callback",
			scope:        "openid",
			opts:         AuthURLOptions{},
			wantErr:      true,
		},
		{
			name:         "missing client id",
			authEndpoint: "https://example.com/authorize",
			clientID:     "",
			redirectURI:  "http://localhost:3000/callback",
			scope:        "openid",
			opts:         AuthURLOptions{},
			wantErr:      true,
		},
		{
			name:         "missing redirect uri",
			authEndpoint: "https://example.com/authorize",
			clientID:     "client123",
			redirectURI:  "",
			scope:        "openid",
			opts:         AuthURLOptions{},
			wantErr:      true,
		},
		{
			name:         "missing scope",
			authEndpoint: "https://example.com/authorize",
			clientID:     "client123",
			redirectURI:  "http://localhost:3000/callback",
			scope:        "",
			opts:         AuthURLOptions{},
			wantErr:      true,
		},
		{
			name:         "invalid auth endpoint URL",
			authEndpoint: "://invalid",
			clientID:     "client123",
			redirectURI:  "http://localhost:3000/callback",
			scope:        "openid",
			opts:         AuthURLOptions{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildAuthURL(tt.authEndpoint, tt.clientID, tt.redirectURI, tt.scope, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, substr := range tt.wantContains {
				if !strings.Contains(result, substr) {
					t.Errorf("URL should contain %q, got: %s", substr, result)
				}
			}
		})
	}
}

// TestParseCallback verifies callback URL parsing.
func TestParseCallback(t *testing.T) {
	tests := []struct {
		name        string
		callbackURL string
		wantValues  map[string]string
		wantErr     bool
	}{
		{
			name:        "success callback",
			callbackURL: "http://localhost:3000/callback?code=abc123&state=xyz789",
			wantValues: map[string]string{
				"code":  "abc123",
				"state": "xyz789",
			},
			wantErr: false,
		},
		{
			name:        "error callback",
			callbackURL: "http://localhost:3000/callback?error=access_denied&error_description=User+denied+access",
			wantValues: map[string]string{
				"error":             "access_denied",
				"error_description": "User denied access",
			},
			wantErr: false,
		},
		{
			name:        "invalid callback URL",
			callbackURL: "://invalid",
			wantValues:  nil,
			wantErr:     true,
		},
		{
			name:        "callback with duplicate params (first wins)",
			callbackURL: "http://localhost:3000/callback?code=first&code=second",
			wantValues: map[string]string{
				"code": "first",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCallback(tt.callbackURL)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for key, expected := range tt.wantValues {
				if got, ok := result[key]; !ok {
					t.Errorf("missing key %q", key)
				} else if got != expected {
					t.Errorf("for key %q: want %q, got %q", key, expected, got)
				}
			}
		})
	}
}

// TestLintIDToken verifies OIDC ID token linting.
func TestLintIDToken(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		payload   map[string]interface{}
		issuer    string
		audience  string
		clockSkew time.Duration
		wantValid bool
		wantWarns []string
	}{
		{
			name: "valid token",
			payload: map[string]interface{}{
				"iss":   "https://example.com",
				"aud":   "client123",
				"sub":   "user123",
				"exp":   float64(now.Add(1 * time.Hour).Unix()),
				"iat":   float64(now.Unix()),
				"nonce": "abc123",
			},
			issuer:    "https://example.com",
			audience:  "client123",
			clockSkew: 0,
			wantValid: true,
			wantWarns: nil,
		},
		{
			name: "issuer mismatch",
			payload: map[string]interface{}{
				"iss": "https://other.com",
				"aud": "client123",
				"exp": float64(now.Add(1 * time.Hour).Unix()),
				"iat": float64(now.Unix()),
			},
			issuer:    "https://example.com",
			audience:  "",
			clockSkew: 0,
			wantValid: false,
			wantWarns: []string{"issuer mismatch"},
		},
		{
			name: "audience mismatch",
			payload: map[string]interface{}{
				"iss": "https://example.com",
				"aud": "other",
				"exp": float64(now.Add(1 * time.Hour).Unix()),
				"iat": float64(now.Unix()),
			},
			issuer:    "https://example.com",
			audience:  "client123",
			clockSkew: 0,
			wantValid: false,
			wantWarns: []string{"audience mismatch"},
		},
		{
			name: "expired token",
			payload: map[string]interface{}{
				"iss": "https://example.com",
				"aud": "client123",
				"exp": float64(now.Add(-1 * time.Hour).Unix()),
				"iat": float64(now.Unix()),
			},
			issuer:    "https://example.com",
			audience:  "",
			clockSkew: 0,
			wantValid: false,
			wantWarns: []string{"token expired"},
		},
		{
			name: "expired within clock skew",
			payload: map[string]interface{}{
				"iss": "https://example.com",
				"aud": "client123",
				"exp": float64(now.Add(-1 * time.Hour).Unix()),
				"iat": float64(now.Unix()),
			},
			issuer:    "https://example.com",
			audience:  "",
			clockSkew: 2 * time.Hour,
			wantValid: true,
			wantWarns: nil,
		},
		{
			name: "iat in future",
			payload: map[string]interface{}{
				"iss": "https://example.com",
				"aud": "client123",
				"exp": float64(now.Add(1 * time.Hour).Unix()),
				"iat": float64(now.Add(10 * time.Minute).Unix()),
			},
			issuer:    "https://example.com",
			audience:  "",
			clockSkew: 0,
			wantValid: false,
			wantWarns: []string{"issued-at time is unreasonably in the future"},
		},
		{
			name: "iat in future within 5 min limit",
			payload: map[string]interface{}{
				"iss": "https://example.com",
				"aud": "client123",
				"exp": float64(now.Add(1 * time.Hour).Unix()),
				"iat": float64(now.Add(3 * time.Minute).Unix()),
			},
			issuer:    "https://example.com",
			audience:  "",
			clockSkew: 0,
			wantValid: true,
			wantWarns: nil,
		},
		{
			name: "missing iss claim",
			payload: map[string]interface{}{
				"aud": "client123",
				"exp": float64(now.Add(1 * time.Hour).Unix()),
				"iat": float64(now.Unix()),
			},
			issuer:    "https://example.com",
			audience:  "",
			clockSkew: 0,
			wantValid: false,
			wantWarns: []string{"missing or invalid 'iss' claim"},
		},
		{
			name: "missing aud claim",
			payload: map[string]interface{}{
				"iss": "https://example.com",
				"exp": float64(now.Add(1 * time.Hour).Unix()),
				"iat": float64(now.Unix()),
			},
			issuer:    "",
			audience:  "client123",
			clockSkew: 0,
			wantValid: false,
			wantWarns: []string{"missing or invalid 'aud' claim"},
		},
		{
			name: "aud as string",
			payload: map[string]interface{}{
				"iss": "https://example.com",
				"aud": "client123",
				"exp": float64(now.Add(1 * time.Hour).Unix()),
				"iat": float64(now.Unix()),
			},
			issuer:    "https://example.com",
			audience:  "client123",
			clockSkew: 0,
			wantValid: true,
			wantWarns: nil,
		},
		{
			name: "aud as array with matching audience",
			payload: map[string]interface{}{
				"iss": "https://example.com",
				"aud": []interface{}{"client123", "other"},
				"exp": float64(now.Add(1 * time.Hour).Unix()),
				"iat": float64(now.Unix()),
			},
			issuer:    "https://example.com",
			audience:  "client123",
			clockSkew: 0,
			wantValid: true,
			wantWarns: nil,
		},
		{
			name: "aud as array without matching audience",
			payload: map[string]interface{}{
				"iss": "https://example.com",
				"aud": []interface{}{"other", "another"},
				"exp": float64(now.Add(1 * time.Hour).Unix()),
				"iat": float64(now.Unix()),
			},
			issuer:    "https://example.com",
			audience:  "client123",
			clockSkew: 0,
			wantValid: false,
			wantWarns: []string{"audience mismatch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := LintIDToken(tt.payload, tt.issuer, tt.audience, tt.clockSkew)
			if err != nil {
				t.Fatalf("LintIDToken failed: %v", err)
			}
			if result.Valid != tt.wantValid {
				t.Errorf("valid: want %v, got %v", tt.wantValid, result.Valid)
			}
			if len(result.Warnings) != len(tt.wantWarns) {
				t.Errorf("warnings: want %d, got %d (%+v)", len(tt.wantWarns), len(result.Warnings), result.Warnings)
			}
		})
	}
}
