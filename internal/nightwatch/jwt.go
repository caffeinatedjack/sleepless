package nightwatch

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/internal/nightwatch/jwt"
)

var jwtJSON bool

var jwtCmd = &cobra.Command{
	Use:   "jwt",
	Short: "JWT utilities",
	Long: `JSON Web Token (JWT) utilities for decoding, verifying, creating, and inspecting tokens.

Examples:
    nightwatch jwt decode <token>
    nightwatch jwt verify <token> --secret <secret>
    nightwatch jwt create --secret <secret> --claim sub=user123`,
}

func init() {
	rootCmd.AddCommand(jwtCmd)
	jwtCmd.PersistentFlags().BoolVar(&jwtJSON, "json", false, "Output in JSON format")
}

// readTokenArg reads a token from args or stdin.
func readTokenArg(args []string) (string, error) {
	if len(args) > 0 {
		return strings.TrimSpace(args[0]), nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return strings.TrimSpace(scanner.Text()), nil
		}
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("error reading stdin: %w", err)
		}
	}

	return "", fmt.Errorf("no token provided")
}

// --- decode command ---

var decodeCmd = &cobra.Command{
	Use:   "decode [token]",
	Short: "Decode a JWT without verifying the signature",
	Long: `Decode a JWT and display header, payload, and signature without verifying.

Warning: This command does NOT verify the token signature.

Examples:
    nightwatch jwt decode eyJhbGciOiJIUzI1NiIs...
    echo "<token>" | nightwatch jwt decode`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := readTokenArg(args)
		if err != nil {
			return err
		}

		decoded, err := jwt.DecodeWithoutVerification(token)
		if err != nil {
			return err
		}

		fmt.Print(jwt.FormatDecode(decoded, jwt.OutputOption{JSON: jwtJSON}))
		return nil
	},
}

// --- header command ---

var headerCmd = &cobra.Command{
	Use:   "header [token]",
	Short: "Decode and display only the JWT header",
	Long: `Decode and display only the JWT header as JSON.

Examples:
    nightwatch jwt header eyJhbGciOiJIUzI1NiIs...
    echo "<token>" | nightwatch jwt header | jq .alg`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := readTokenArg(args)
		if err != nil {
			return err
		}

		decoded, err := jwt.DecodeWithoutVerification(token)
		if err != nil {
			return err
		}

		fmt.Print(jwt.FormatHeader(decoded.Header))
		return nil
	},
}

// --- payload command ---

var payloadCmd = &cobra.Command{
	Use:   "payload [token]",
	Short: "Decode and display only the JWT payload",
	Long: `Decode and display only the JWT payload as JSON.

Examples:
    nightwatch jwt payload eyJhbGciOiJIUzI1NiIs...
    echo "<token>" | nightwatch jwt payload | jq .sub`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := readTokenArg(args)
		if err != nil {
			return err
		}

		decoded, err := jwt.DecodeWithoutVerification(token)
		if err != nil {
			return err
		}

		fmt.Print(jwt.FormatPayload(decoded.Payload))
		return nil
	},
}

// --- create command ---

var (
	createSecret     string
	createSecretFile string
	createKey        string
	createAlg        string
	createClaims     []string
	createPayload    string
	createIss        string
	createSub        string
	createAud        string
	createExp        string
	createNbf        string
	createIat        string
	createJti        string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new JWT",
	Long: `Create a new JWT with custom claims and sign it.

Use --secret for HMAC signing or --key for RSA/ECDSA signing.

Examples:
    nightwatch jwt create --secret mykey --claim sub=user123 --exp 1h
    nightwatch jwt create --secret mykey --payload '{"sub":"123","role":"admin"}'
    nightwatch jwt create --key private.pem --alg RS256 --sub user123 --exp 24h`,
	RunE: func(cmd *cobra.Command, args []string) error {
		secret, err := resolveSecret(createSecret, createSecretFile)
		if err != nil {
			return err
		}

		claims, err := parseClaims(createClaims)
		if err != nil {
			return err
		}

		result, err := jwt.CreateToken(createAlg, secret, createKey, claims, createPayload,
			createIss, createSub, createAud, createExp, createNbf, createIat, createJti)
		if err != nil {
			return err
		}

		fmt.Print(jwt.FormatCreate(result.Token, result.Header, result.Payload, jwt.OutputOption{JSON: jwtJSON}))
		return nil
	},
}

// --- verify command ---

var (
	verifySecret     string
	verifySecretFile string
	verifyKey        string
)

var verifyCmd = &cobra.Command{
	Use:   "verify [token]",
	Short: "Verify a JWT signature and claims",
	Long: `Verify a JWT signature and check expiration and nbf claims.

Use --secret for HMAC verification or --key for RSA/ECDSA verification.

Examples:
    nightwatch jwt verify <token> --secret mykey
    nightwatch jwt verify <token> --key public.pem`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		secret, err := resolveSecret(verifySecret, verifySecretFile)
		if err != nil {
			return err
		}

		token, err := readTokenArg(args)
		if err != nil {
			return err
		}

		valid, errMsg, header, payload, err := jwt.VerifyToken(token, secret, verifyKey)
		if err != nil {
			if jwtErr, ok := err.(*jwt.JWTError); ok {
				fmt.Fprintln(os.Stderr, jwtErr.Message)
				os.Exit(jwtErr.ExitCode())
			}
			return err
		}

		alg, _ := header["alg"].(string)
		fmt.Print(jwt.FormatVerify(valid, alg, header, payload, errMsg, jwt.OutputOption{JSON: jwtJSON}))

		if !valid {
			os.Exit(2)
		}
		return nil
	},
}

// --- exp command ---

var expCmd = &cobra.Command{
	Use:   "exp [token]",
	Short: "Check JWT expiration status",
	Long: `Check the expiration status of a JWT without verifying the signature.

Examples:
    nightwatch jwt exp <token>
    echo "<token>" | nightwatch jwt exp`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := readTokenArg(args)
		if err != nil {
			return err
		}

		result, err := jwt.CheckExpiration(token)
		if err != nil {
			return err
		}

		fmt.Print(jwt.FormatExp(result, jwt.OutputOption{JSON: jwtJSON}))
		return nil
	},
}

// --- helpers ---

func resolveSecret(secret, secretFile string) (string, error) {
	if secretFile != "" {
		if secret != "" {
			return "", fmt.Errorf("--secret and --secret-file are mutually exclusive")
		}
		data, err := os.ReadFile(secretFile)
		if err != nil {
			return "", fmt.Errorf("failed to read secret file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	if secret != "" {
		fmt.Fprintln(os.Stderr, "Warning: secrets via CLI args may be visible in process lists. Consider --secret-file.")
	}
	return secret, nil
}

func parseClaims(claimStrs []string) (map[string]interface{}, error) {
	claims := make(map[string]interface{})
	for _, s := range claimStrs {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid claim format: %s (expected key=value)", s)
		}
		claims[strings.TrimSpace(parts[0])] = jwt.ParseClaimValue(strings.TrimSpace(parts[1]))
	}
	return claims, nil
}

func init() {
	jwtCmd.AddCommand(decodeCmd)
	jwtCmd.AddCommand(headerCmd)
	jwtCmd.AddCommand(payloadCmd)
	jwtCmd.AddCommand(createCmd)
	jwtCmd.AddCommand(verifyCmd)
	jwtCmd.AddCommand(expCmd)

	// create flags
	createCmd.Flags().StringVar(&createSecret, "secret", "", "HMAC secret key")
	createCmd.Flags().StringVar(&createSecretFile, "secret-file", "", "Read HMAC secret from file")
	createCmd.Flags().StringVar(&createKey, "key", "", "Path to PEM-encoded private key")
	createCmd.Flags().StringVar(&createAlg, "alg", "", "Signing algorithm (HS256, RS256, ES256, etc.)")
	createCmd.Flags().StringArrayVar(&createClaims, "claim", nil, "Custom claim (key=value, repeatable)")
	createCmd.Flags().StringVar(&createPayload, "payload", "", "Complete payload as JSON")
	createCmd.Flags().StringVar(&createIss, "iss", "", "Issuer claim")
	createCmd.Flags().StringVar(&createSub, "sub", "", "Subject claim")
	createCmd.Flags().StringVar(&createAud, "aud", "", "Audience claim")
	createCmd.Flags().StringVar(&createExp, "exp", "", "Expiration duration (e.g., 1h, 30m, 7d)")
	createCmd.Flags().StringVar(&createNbf, "nbf", "", "Not-before duration (e.g., 5m)")
	createCmd.Flags().StringVar(&createIat, "iat", "", "Issued-at time (Unix timestamp or RFC3339)")
	createCmd.Flags().StringVar(&createJti, "jti", "", "JWT ID claim")

	// verify flags
	verifyCmd.Flags().StringVar(&verifySecret, "secret", "", "HMAC secret key")
	verifyCmd.Flags().StringVar(&verifySecretFile, "secret-file", "", "Read HMAC secret from file")
	verifyCmd.Flags().StringVar(&verifyKey, "key", "", "Path to PEM-encoded public key")
}
