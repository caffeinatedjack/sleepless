// Package nightwatch implements OIDC CLI commands.
package nightwatch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/internal/nightwatch/jwt"
	"gitlab.com/caffeinatedjack/sleepless/pkg/oidc"
)

var oidcJSON bool

func init() {
	initOIDC()
	initAuthURL()
	initIDTokenLint()
}

var oidcCmd = &cobra.Command{
	Use:   "oidc",
	Short: "OIDC utilities",
	Long: `OpenID Connect (OIDC) utilities for PKCE generation, authorization URL building, callback parsing, and ID token inspection.

Examples:
    nightwatch oidc pkce
    nightwatch oidc state
    nightwatch oidc nonce
    nightwatch oidc auth-url --auth-endpoint ... --client-id ...
    nightwatch oidc callback "http://localhost:3000/callback?code=..."
    nightwatch oidc idtoken decode <jwt>
    nightwatch oidc idtoken lint <jwt>`,
}

func initOIDC() {
	rootCmd.AddCommand(oidcCmd)
	oidcCmd.PersistentFlags().BoolVar(&oidcJSON, "json", false, "Output in JSON format")

	// pkce command
	oidcCmd.AddCommand(pkceCmd)

	// state command
	oidcCmd.AddCommand(stateCmd)

	// nonce command
	oidcCmd.AddCommand(nonceCmd)

	// auth-url command
	oidcCmd.AddCommand(authURLCmd)

	// callback command
	oidcCmd.AddCommand(callbackCmd)

	// idtoken commands
	oidcCmd.AddCommand(idtokenCmd)
	idtokenCmd.AddCommand(idtokenDecodeCmd)
	idtokenCmd.AddCommand(idtokenLintCmd)
}

// readOIDCTokenArg reads a token from args or stdin (local copy for oidc).
func readOIDCTokenArg(args []string) (string, error) {
	if len(args) > 0 {
		return strings.TrimSpace(args[0]), nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if scanner.Text() != "" {
				return strings.TrimSpace(scanner.Text()), nil
			}
		}
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("error reading stdin: %w", err)
		}
	}

	return "", fmt.Errorf("no token provided")
}

// --- pkce command ---

var pkceCmd = &cobra.Command{
	Use:   "pkce",
	Short: "Generate PKCE verifier and S256 challenge",
	Long: `Generate a PKCE (Proof Key for Code Exchange) verifier and S256 challenge per RFC 7636.

The verifier is a URL-safe random string, and the challenge is base64url-encoded(sha256(verifier || .)).

Examples:
    nightwatch oidc pkce
    nightwatch oidc pkce --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verifier, challenge, err := oidc.GeneratePKCE()
		if err != nil {
			return fmt.Errorf("failed to generate PKCE: %w", err)
		}

		if oidcJSON {
			output := map[string]string{
				"verifier":  verifier,
				"challenge": challenge,
			}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			fmt.Printf("Verifier: %s\n", verifier)
			fmt.Printf("Challenge: %s\n", challenge)
		}
		return nil
	},
}

// --- state command ---

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Generate a state value",
	Long: `Generate a URL-safe state value suitable for CSRF protection in OAuth 2.0 flows.

The state value is generated using cryptographically secure randomness.

Examples:
    nightwatch oidc state
    nightwatch oidc state --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := oidc.GenerateState()
		if err != nil {
			return fmt.Errorf("failed to generate state: %w", err)
		}

		if oidcJSON {
			output := map[string]string{"state": state}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			fmt.Println(state)
		}
		return nil
	},
}

// --- nonce command ---

var nonceCmd = &cobra.Command{
	Use:   "nonce",
	Short: "Generate a nonce value",
	Long: `Generate a URL-safe nonce value suitable for OIDC ID token binding.

The nonce value is generated using cryptographically secure randomness.

Examples:
    nightwatch oidc nonce
    nightwatch oidc nonce --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nonce, err := oidc.GenerateNonce()
		if err != nil {
			return fmt.Errorf("failed to generate nonce: %w", err)
		}

		if oidcJSON {
			output := map[string]string{"nonce": nonce}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			fmt.Println(nonce)
		}
		return nil
	},
}

// --- auth-url command ---

var (
	authEndpoint string
	clientID     string
	redirectURI  string
	scope        string
	authState    string
	authNonce    string
	pkceVerifier string
)

var authURLCmd = &cobra.Command{
	Use:   "auth-url",
	Short: "Build an authorization URL",
	Long: `Construct an OAuth 2.0 / OIDC authorization URL from provided parameters.

Required parameters:
  --auth-endpoint <url>   The provider's authorization endpoint
  --client-id <string>     Client identifier
  --redirect-uri <url>     Redirect URI after authorization
  --scope <string>        Requested scopes

Optional parameters:
  --state <string>        State value for CSRF protection
  --nonce <string>        Nonce value for ID token binding
  --pkce-challenge <string> PKCE challenge (use 'nightwatch oidc pkce' to generate)

The URL is constructed with proper parameter encoding.

Examples:
    nightwatch oidc auth-url \\
      --auth-endpoint "https://issuer.example.com/oauth2/v2.0/authorize" \\
      --client-id "my-client" \\
      --redirect-uri "http://localhost:3000/callback" \\
      --scope "openid profile email" \\
      --state "..." \\
      --nonce "..."`,
	RunE: func(cmd *cobra.Command, args []string) error {
		url, err := oidc.BuildAuthURL(authEndpoint, clientID, redirectURI, scope, oidc.AuthURLOptions{
			State:         authState,
			Nonce:         authNonce,
			PKCEChallenge: pkceVerifier,
		})
		if err != nil {
			return err
		}
		fmt.Println(url)
		return nil
	},
}

func initAuthURL() {
	authURLCmd.Flags().StringVar(&authEndpoint, "auth-endpoint", "", "Authorization endpoint URL (required)")
	authURLCmd.Flags().StringVar(&clientID, "client-id", "", "Client identifier (required)")
	authURLCmd.Flags().StringVar(&redirectURI, "redirect-uri", "", "Redirect URI (required)")
	authURLCmd.Flags().StringVar(&scope, "scope", "", "Requested scopes (required)")
	authURLCmd.Flags().StringVar(&authState, "state", "", "State value for CSRF protection")
	authURLCmd.Flags().StringVar(&authNonce, "nonce", "", "Nonce value for ID token binding")
	authURLCmd.Flags().StringVar(&pkceVerifier, "pkce-challenge", "", "PKCE challenge (use 'nightwatch oidc pkce' to generate)")
}

// --- callback command ---

var callbackCmd = &cobra.Command{
	Use:   "callback <url>",
	Short: "Parse a callback URL",
	Long: `Parse an authorization callback URL and extract relevant parameters such as code, state, error, and error_description.

Examples:
    nightwatch oidc callback "http://localhost:3000/callback?code=abc123&state=xyz789"
    nightwatch oidc callback "http://localhost:3000/callback?error=access_denied&error_description=User+denied+access"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		params, err := oidc.ParseCallback(args[0])
		if err != nil {
			return err
		}

		if oidcJSON {
			data, err := json.MarshalIndent(params, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			fmt.Println("Callback parameters:")
			for key, value := range params {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
		return nil
	},
}

// --- idtoken commands ---

var idtokenCmd = &cobra.Command{
	Use:   "idtoken",
	Short: "ID token utilities",
	Long:  `Utilities for inspecting and linting OIDC ID tokens.`,
}

// idtoken decode command
var idtokenDecodeCmd = &cobra.Command{
	Use:   "decode <jwt>",
	Short: "Decode an ID token",
	Long: `Decode an OIDC ID token and display its claims.

This command does NOT verify the token signature.

Examples:
    nightwatch oidc idtoken decode <jwt>
    echo "<jwt>" | nightwatch oidc idtoken decode`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := readOIDCTokenArg(args)
		if err != nil {
			return err
		}

		decoded, err := jwt.DecodeWithoutVerification(token)
		if err != nil {
			return err
		}

		note := "WARNING: Token signature was NOT verified."
		fmt.Println(jwt.FormatDecode(decoded, jwt.OutputOption{JSON: oidcJSON}))
		if !oidcJSON {
			fmt.Println(note)
		}
		return nil
	},
}

// idtoken lint command
var (
	idtokenIssuer    string
	idtokenAudience  string
	idtokenClockSkew string
)

var idtokenLintCmd = &cobra.Command{
	Use:   "lint <jwt>",
	Short: "Lint an ID token",
	Long: `Perform OIDC-specific checks on an ID token and report warnings.

Optional parameters:
  --issuer <url>       Expected issuer URL (will warn on mismatch)
  --audience <string>   Expected audience (will warn on mismatch)
  --clock-skew <duration>  Clock skew allowance (default: 0s)

Examples:
    nightwatch oidc idtoken lint <jwt>
    nightwatch oidc idtoken lint <jwt> --issuer https://issuer.example.com --audience my-client
    nightwatch oidc idtoken lint <jwt> --clock-skew 5m`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := readOIDCTokenArg(args)
		if err != nil {
			return err
		}

		decoded, err := jwt.DecodeWithoutVerification(token)
		if err != nil {
			return err
		}

		// Parse optional flags
		issuer := ""
		audience := ""
		skew := 0 * time.Second
		if cmd.Flags().Changed("issuer") {
			issuer, _ = cmd.Flags().GetString("issuer")
		}
		if cmd.Flags().Changed("audience") {
			audience, _ = cmd.Flags().GetString("audience")
		}
		if cmd.Flags().Changed("clock-skew") {
			skewStr, _ := cmd.Flags().GetString("clock-skew")
			if skewStr != "" {
				var d time.Duration
				if d, err = time.ParseDuration(skewStr); err != nil {
					return fmt.Errorf("invalid clock-skew duration: %w", err)
				}
				skew = d
			}
		}

		// Run OIDC lint
		result, err := oidc.LintIDToken(decoded.Payload, issuer, audience, skew)
		if err != nil {
			return err
		}

		if oidcJSON {
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			// Human-readable output
			fmt.Printf("Valid: %v\n", result.Valid)
			if result.Issuer != nil {
				fmt.Printf("Issuer: %s\n", *result.Issuer)
			}
			if result.Audience != nil {
				fmt.Printf("Audience: %s\n", *result.Audience)
			}
			if result.Expiry != nil {
				fmt.Printf("Expiry: %s\n", result.Expiry.Format(time.RFC3339))
			}
			if result.NotBefore != nil {
				fmt.Printf("Not Before: %s\n", result.NotBefore.Format(time.RFC3339))
			}
			if result.IssuedAt != nil {
				fmt.Printf("Issued At: %s\n", result.IssuedAt.Format(time.RFC3339))
			}
			if result.Nonce != nil {
				fmt.Printf("Nonce: %s\n", *result.Nonce)
			}
			for _, warning := range result.Warnings {
				fmt.Printf("Warning: %s\n", warning)
			}

			// Non-zero exit if invalid
			if !result.Valid {
				os.Exit(1)
			}
		}
		return nil
	},
}

func initIDTokenLint() {
	idtokenLintCmd.Flags().StringVar(&idtokenIssuer, "issuer", "", "Expected issuer URL")
	idtokenLintCmd.Flags().StringVar(&idtokenAudience, "audience", "", "Expected audience")
	idtokenLintCmd.Flags().StringVar(&idtokenClockSkew, "clock-skew", "", "Clock skew allowance")
}
