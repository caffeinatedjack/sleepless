package nightwatch

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/decide"
	"gitlab.com/caffeinatedjack/sleepless/pkg/password"
)

//go:embed data/wordlist.txt
var wordlistData string

var (
	// Common flags
	pwdCount       int
	pwdSeed        *int64
	pwdJSON        bool
	pwdShowEntropy bool

	// Generate flags
	pwdLength         int
	pwdSymbols        bool
	pwdNoUppercase    bool
	pwdNoLowercase    bool
	pwdNoDigits       bool
	pwdAllowAmbiguous bool
	pwdNoRequireAll   bool

	// Phrase flags
	phraseWords      int
	phraseSeparator  string
	phraseCapitalize bool
	phraseNumbers    bool
	phraseWordlist   string

	// API flags
	apiPrefix string
	apiLength int
	apiFormat string
)

func init() {
	rootCmd.AddCommand(passwordCmd)

	// Common persistent flags
	passwordCmd.PersistentFlags().IntVar(&pwdCount, "count", 1, "Number of passwords to generate")
	passwordCmd.PersistentFlags().BoolVar(&pwdJSON, "json", false, "Output in JSON format")
	passwordCmd.PersistentFlags().BoolVar(&pwdShowEntropy, "show-entropy", false, "Display entropy information")
	passwordCmd.PersistentFlags().Int64("seed", 0, "Random seed for reproducible output (testing only)")
	passwordCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("seed") {
			seed, _ := cmd.Flags().GetInt64("seed")
			pwdSeed = &seed
		} else {
			pwdSeed = nil
		}
	}

	// Add subcommands
	passwordCmd.AddCommand(passwordGenerateCmd)
	passwordCmd.AddCommand(passwordPhraseCmd)
	passwordCmd.AddCommand(passwordAPICmd)
	passwordCmd.AddCommand(passwordPatternCmd)
	passwordCmd.AddCommand(passwordCheckCmd)

	// Generate flags
	passwordGenerateCmd.Flags().IntVar(&pwdLength, "length", 16, "Password length")
	passwordGenerateCmd.Flags().BoolVar(&pwdSymbols, "symbols", false, "Include symbol characters")
	passwordGenerateCmd.Flags().BoolVar(&pwdNoUppercase, "no-uppercase", false, "Exclude uppercase letters")
	passwordGenerateCmd.Flags().BoolVar(&pwdNoLowercase, "no-lowercase", false, "Exclude lowercase letters")
	passwordGenerateCmd.Flags().BoolVar(&pwdNoDigits, "no-digits", false, "Exclude digits")
	passwordGenerateCmd.Flags().BoolVar(&pwdAllowAmbiguous, "allow-ambiguous", false, "Include ambiguous characters (0O, 1lI, etc.)")
	passwordGenerateCmd.Flags().BoolVar(&pwdNoRequireAll, "no-require-all", false, "Don't require one character from each enabled set")

	// Phrase flags
	passwordPhraseCmd.Flags().IntVar(&phraseWords, "words", 6, "Number of words")
	passwordPhraseCmd.Flags().StringVar(&phraseSeparator, "separator", "-", "Word separator")
	passwordPhraseCmd.Flags().BoolVar(&phraseCapitalize, "capitalize", false, "Capitalize first letter of each word")
	passwordPhraseCmd.Flags().BoolVar(&phraseNumbers, "numbers", false, "Append a random digit to each word")
	passwordPhraseCmd.Flags().StringVar(&phraseWordlist, "wordlist", "", "Path to custom word list file")

	// API flags
	passwordAPICmd.Flags().StringVar(&apiPrefix, "prefix", "", "Key prefix (followed by _)")
	passwordAPICmd.Flags().IntVar(&apiLength, "length", 32, "Length of random part")
	passwordAPICmd.Flags().StringVar(&apiFormat, "format", "hex", "Output format: hex, base32, base64, base64url")
}

var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Generate passwords, passphrases, and API keys",
	Long: `Generate cryptographically secure passwords, passphrases, and secrets.

Subcommands:
  generate  Character-based password with configurable sets
  phrase    Word-based passphrase using word lists
  api       API key with optional prefix and encoding
  pattern   Pattern-based generation (XXX-999-xxx)
  check     Analyze password strength

Examples:
  nightwatch password generate
  nightwatch password generate --length 24 --symbols
  nightwatch password phrase --words 8
  nightwatch password api --prefix "sk_prod"
  nightwatch password pattern "XXXX-XXXX-XXXX"
  nightwatch password check "MyP@ssw0rd"`,
}

var passwordGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a character-based password",
	Long: `Generate a password using configurable character sets.

By default, passwords include lowercase, uppercase, and digits.
Use --symbols to add special characters.
Use --no-uppercase, --no-lowercase, --no-digits to exclude sets.

Examples:
  nightwatch password generate
  nightwatch password generate --length 24
  nightwatch password generate --length 20 --symbols
  nightwatch password generate --count 5
  nightwatch password generate --no-uppercase --no-digits`,
	RunE: runPasswordGenerate,
}

var passwordPhraseCmd = &cobra.Command{
	Use:   "phrase",
	Short: "Generate a word-based passphrase",
	Long: `Generate a passphrase using random words from a word list.

Uses a built-in word list by default (~1000 words).
Use --wordlist to provide a custom word list (one word per line).

Examples:
  nightwatch password phrase
  nightwatch password phrase --words 8
  nightwatch password phrase --separator " "
  nightwatch password phrase --capitalize --numbers
  nightwatch password phrase --wordlist custom.txt`,
	RunE: runPasswordPhrase,
}

var passwordAPICmd = &cobra.Command{
	Use:   "api",
	Short: "Generate an API key",
	Long: `Generate an API key with optional prefix and encoding format.

Formats:
  hex       Hexadecimal (0-9, a-f) - default
  base32    Base32 (A-Z, 2-7)
  base64    Base64 (A-Za-z0-9+/)
  base64url Base64 URL-safe (A-Za-z0-9-_)

Examples:
  nightwatch password api
  nightwatch password api --prefix "sk_prod"
  nightwatch password api --length 40 --format base64url
  nightwatch password api --prefix "ghp" --format base64url --length 36`,
	RunE: runPasswordAPI,
}

var passwordPatternCmd = &cobra.Command{
	Use:   "pattern <pattern>",
	Short: "Generate a password matching a pattern",
	Long: `Generate a password based on a pattern template.

Pattern characters:
  X  Random uppercase letter (A-Z)
  x  Random lowercase letter (a-z)
  9  Random digit (0-9)
  #  Random alphanumeric (a-zA-Z0-9)
  ?  Random symbol (!@#$%...)
  *  Random printable character
  \X Literal X (escape special chars)
  
Any other character is literal.

Examples:
  nightwatch password pattern "XXX-999-xxx"
  nightwatch password pattern "XXXX-XXXX-XXXX-XXXX"
  nightwatch password pattern "SN-####-####"
  nightwatch password pattern "USER-999-xxx"`,
	Args: cobra.ExactArgs(1),
	RunE: runPasswordPattern,
}

var passwordCheckCmd = &cobra.Command{
	Use:   "check <password>",
	Short: "Check password strength",
	Long: `Analyze a password and report its strength.

Reports:
  - Length
  - Character sets used
  - Entropy (bits)
  - Strength rating (weak/fair/good/strong/excellent)
  - Common pattern detection

Examples:
  nightwatch password check "password123"
  nightwatch password check "Tr0ub4dor&3"
  nightwatch password check "correct-horse-battery-staple"`,
	Args: cobra.ExactArgs(1),
	RunE: runPasswordCheck,
}

func runPasswordGenerate(cmd *cobra.Command, args []string) error {
	// Emit warning for seeded mode
	if pwdSeed != nil {
		fmt.Fprintln(os.Stderr, "WARNING: Seeded passwords are for testing only. DO NOT use for production.")
	}

	// Build options
	opts := password.PasswordOptions{
		Length: pwdLength,
		CharSet: password.CharSetOptions{
			Lowercase:      !pwdNoLowercase,
			Uppercase:      !pwdNoUppercase,
			Digits:         !pwdNoDigits,
			Symbols:        pwdSymbols,
			AllowAmbiguous: pwdAllowAmbiguous,
		},
		RequireAll: !pwdNoRequireAll,
	}

	// Create RNG and generator
	rng := decide.NewRNG(pwdSeed)
	gen := password.NewGenerator(rng)

	// Generate passwords
	results := make([]string, pwdCount)
	for i := 0; i < pwdCount; i++ {
		pwd, err := gen.GeneratePassword(opts)
		if err != nil {
			return err
		}
		results[i] = pwd
	}

	// Calculate entropy
	charsetSize, err := password.GetCharSetSize(opts.CharSet)
	if err != nil {
		return err
	}
	entropy := password.CalculatePasswordEntropy(charsetSize, opts.Length)

	// Warn if low entropy
	if entropy < 40 {
		fmt.Fprintln(os.Stderr, "WARNING: Low entropy password. Consider increasing length or enabling more character sets.")
	}

	// Output
	if pwdJSON {
		return outputPasswordJSON("password", results, entropy)
	}
	return outputPasswordText(results, entropy)
}

func runPasswordPhrase(cmd *cobra.Command, args []string) error {
	// Emit warning for seeded mode
	if pwdSeed != nil {
		fmt.Fprintln(os.Stderr, "WARNING: Seeded passwords are for testing only. DO NOT use for production.")
	}

	// Load word list
	var wl *password.WordList
	var err error
	if phraseWordlist != "" {
		wl, err = password.LoadWordList(phraseWordlist)
		if err != nil {
			return fmt.Errorf("failed to load word list: %w", err)
		}
	} else {
		wl, err = password.LoadWordListFromString(wordlistData)
		if err != nil {
			return fmt.Errorf("failed to load built-in word list: %w", err)
		}
	}

	// Warn if word list is small
	if wl.Size() < password.WordListMinSize {
		fmt.Fprintf(os.Stderr, "WARNING: Word list has only %d words. Consider using a larger list for better entropy.\n", wl.Size())
	}

	// Build options
	opts := password.PhraseOptions{
		Separator:  phraseSeparator,
		Capitalize: phraseCapitalize,
		Numbers:    phraseNumbers,
	}

	// Create RNG
	rng := decide.NewRNG(pwdSeed)

	// Generate passphrases
	results := make([]string, pwdCount)
	for i := 0; i < pwdCount; i++ {
		phrase, err := password.GeneratePassphrase(wl, phraseWords, opts, rng)
		if err != nil {
			return err
		}
		results[i] = phrase
	}

	// Calculate entropy
	entropy := password.CalculatePhraseEntropy(wl.Size(), phraseWords)

	// Warn if low entropy
	if entropy < 40 {
		fmt.Fprintln(os.Stderr, "WARNING: Low entropy passphrase. Consider using more words.")
	}

	// Output
	if pwdJSON {
		return outputPasswordJSON("phrase", results, entropy)
	}
	return outputPasswordText(results, entropy)
}

func runPasswordAPI(cmd *cobra.Command, args []string) error {
	// Emit warning for seeded mode
	if pwdSeed != nil {
		fmt.Fprintln(os.Stderr, "WARNING: Seeded passwords are for testing only. DO NOT use for production.")
	}

	// Parse format
	format, err := password.ParseFormat(apiFormat)
	if err != nil {
		return err
	}

	// Build options
	opts := password.APIKeyOptions{
		Prefix: apiPrefix,
		Length: apiLength,
		Format: format,
	}

	// Create RNG
	rng := decide.NewRNG(pwdSeed)

	// Generate API keys
	results := make([]string, pwdCount)
	for i := 0; i < pwdCount; i++ {
		key, err := password.GenerateAPIKey(opts, rng)
		if err != nil {
			return err
		}
		results[i] = key
	}

	// Calculate entropy
	entropy := password.APIKeyEntropyBits(format, apiLength)

	// Output
	if pwdJSON {
		return outputPasswordJSON("api", results, entropy)
	}
	return outputPasswordText(results, entropy)
}

func runPasswordPattern(cmd *cobra.Command, args []string) error {
	// Emit warning for seeded mode
	if pwdSeed != nil {
		fmt.Fprintln(os.Stderr, "WARNING: Seeded passwords are for testing only. DO NOT use for production.")
	}

	patternStr := args[0]

	// Parse pattern
	tokens, err := password.ParsePattern(patternStr)
	if err != nil {
		return err
	}

	// Create RNG
	rng := decide.NewRNG(pwdSeed)

	// Generate passwords
	results := make([]string, pwdCount)
	for i := 0; i < pwdCount; i++ {
		pwd, err := password.GenerateFromPattern(tokens, rng)
		if err != nil {
			return err
		}
		results[i] = pwd
	}

	// Calculate entropy
	entropy := password.CalculatePatternEntropy(tokens)

	// Warn if low entropy
	if entropy < 40 {
		fmt.Fprintln(os.Stderr, "WARNING: Low entropy pattern. Consider adding more random characters.")
	}

	// Output
	if pwdJSON {
		return outputPatternJSON(patternStr, results, entropy)
	}
	return outputPasswordText(results, entropy)
}

func runPasswordCheck(cmd *cobra.Command, args []string) error {
	pwd := args[0]
	report := password.CheckPassword(pwd)

	if pwdJSON {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Human-readable output
	fmt.Printf("Length: %d characters\n", report.Length)
	fmt.Printf("Character sets: %s\n", strings.Join(report.CharSets, ", "))
	fmt.Printf("Entropy: %.1f bits\n", report.Entropy)
	fmt.Printf("Strength: %s\n", strings.ToUpper(report.Strength))

	if report.HasCommon {
		fmt.Println("Warning: Contains common password pattern")
	}
	if report.HasKeyboard {
		fmt.Println("Warning: Contains keyboard walk pattern")
	}

	return nil
}

// Output helpers

type passwordOutput struct {
	Mode    string   `json:"mode"`
	Seed    *int64   `json:"seed"`
	Count   int      `json:"count"`
	Entropy float64  `json:"entropy,omitempty"`
	Results []string `json:"results"`
	Warning string   `json:"warning,omitempty"`
}

type patternOutput struct {
	Mode    string   `json:"mode"`
	Pattern string   `json:"pattern"`
	Seed    *int64   `json:"seed"`
	Count   int      `json:"count"`
	Entropy float64  `json:"entropy,omitempty"`
	Results []string `json:"results"`
	Warning string   `json:"warning,omitempty"`
}

func outputPasswordJSON(mode string, results []string, entropy float64) error {
	output := passwordOutput{
		Mode:    mode,
		Seed:    pwdSeed,
		Count:   pwdCount,
		Entropy: entropy,
		Results: results,
	}
	if pwdSeed != nil {
		output.Warning = "Seeded passwords are for testing only. DO NOT use for production."
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputPatternJSON(pattern string, results []string, entropy float64) error {
	output := patternOutput{
		Mode:    "pattern",
		Pattern: pattern,
		Seed:    pwdSeed,
		Count:   pwdCount,
		Entropy: entropy,
		Results: results,
	}
	if pwdSeed != nil {
		output.Warning = "Seeded passwords are for testing only. DO NOT use for production."
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputPasswordText(results []string, entropy float64) error {
	for _, r := range results {
		fmt.Println(r)
	}
	if pwdShowEntropy {
		strength := password.EntropyToStrength(entropy)
		fmt.Fprintf(os.Stderr, "Entropy: %.1f bits (%s)\n", entropy, strength)
	}
	return nil
}
