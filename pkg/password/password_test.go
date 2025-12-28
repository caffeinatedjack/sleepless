package password

import (
	"math"
	"strings"
	"testing"

	"gitlab.com/caffeinatedjack/sleepless/pkg/decide"
)

func TestBuildCharSet(t *testing.T) {
	tests := []struct {
		name        string
		opts        CharSetOptions
		wantErr     bool
		wantMinLen  int
		wantSets    int
		containsAll bool // check that result contains expected chars
	}{
		{
			name:       "default options",
			opts:       DefaultCharSetOptions(),
			wantErr:    false,
			wantMinLen: 50, // at least 50 chars after ambiguous removal
			wantSets:   3,
		},
		{
			name: "all sets enabled",
			opts: CharSetOptions{
				Lowercase:      true,
				Uppercase:      true,
				Digits:         true,
				Symbols:        true,
				AllowAmbiguous: true,
			},
			wantErr:    false,
			wantMinLen: 90, // 26+26+10+28 (symbols charset has 28 chars)
			wantSets:   4,
		},
		{
			name: "lowercase only",
			opts: CharSetOptions{
				Lowercase: true,
			},
			wantErr:    false,
			wantMinLen: 25, // 26 minus ambiguous 'l'
			wantSets:   1,
		},
		{
			name:    "no sets enabled",
			opts:    CharSetOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charset, sets, err := BuildCharSet(tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(charset) < tt.wantMinLen {
				t.Errorf("charset length = %d, want >= %d", len(charset), tt.wantMinLen)
			}
			if len(sets) != tt.wantSets {
				t.Errorf("got %d sets, want %d", len(sets), tt.wantSets)
			}
		})
	}
}

func TestCalculatePasswordEntropy(t *testing.T) {
	tests := []struct {
		name        string
		charsetSize int
		length      int
		wantApprox  float64
		tolerance   float64
	}{
		{
			name:        "alphanumeric 16 chars",
			charsetSize: 62,
			length:      16,
			wantApprox:  95.3,
			tolerance:   0.5,
		},
		{
			name:        "full charset 16 chars",
			charsetSize: 94,
			length:      16,
			wantApprox:  104.9,
			tolerance:   0.5,
		},
		{
			name:        "digits only 6 chars",
			charsetSize: 10,
			length:      6,
			wantApprox:  19.9,
			tolerance:   0.5,
		},
		{
			name:        "zero length",
			charsetSize: 62,
			length:      0,
			wantApprox:  0,
			tolerance:   0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculatePasswordEntropy(tt.charsetSize, tt.length)
			if math.Abs(got-tt.wantApprox) > tt.tolerance {
				t.Errorf("CalculatePasswordEntropy(%d, %d) = %.2f, want ~%.2f", tt.charsetSize, tt.length, got, tt.wantApprox)
			}
		})
	}
}

func TestCalculatePhraseEntropy(t *testing.T) {
	tests := []struct {
		name         string
		wordlistSize int
		wordCount    int
		wantApprox   float64
		tolerance    float64
	}{
		{
			name:         "diceware 6 words",
			wordlistSize: 7776,
			wordCount:    6,
			wantApprox:   77.5,
			tolerance:    0.5,
		},
		{
			name:         "small list 4 words",
			wordlistSize: 1000,
			wordCount:    4,
			wantApprox:   39.9,
			tolerance:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculatePhraseEntropy(tt.wordlistSize, tt.wordCount)
			if math.Abs(got-tt.wantApprox) > tt.tolerance {
				t.Errorf("CalculatePhraseEntropy(%d, %d) = %.2f, want ~%.2f", tt.wordlistSize, tt.wordCount, got, tt.wantApprox)
			}
		})
	}
}

func TestEntropyToStrength(t *testing.T) {
	tests := []struct {
		entropy float64
		want    string
	}{
		{20, "weak"},
		{39, "weak"},
		{40, "fair"},
		{59, "fair"},
		{60, "good"},
		{79, "good"},
		{80, "strong"},
		{99, "strong"},
		{100, "excellent"},
		{150, "excellent"},
	}

	for _, tt := range tests {
		got := EntropyToStrength(tt.entropy)
		if got != tt.want {
			t.Errorf("EntropyToStrength(%.0f) = %q, want %q", tt.entropy, got, tt.want)
		}
	}
}

func TestGeneratePassword(t *testing.T) {
	seed := int64(42)
	rng := decide.NewRNG(&seed)
	gen := NewGenerator(rng)

	tests := []struct {
		name    string
		opts    PasswordOptions
		wantLen int
		wantErr bool
	}{
		{
			name:    "default options",
			opts:    DefaultPasswordOptions(),
			wantLen: 16,
			wantErr: false,
		},
		{
			name: "custom length",
			opts: PasswordOptions{
				Length:     24,
				CharSet:    DefaultCharSetOptions(),
				RequireAll: true,
			},
			wantLen: 24,
			wantErr: false,
		},
		{
			name: "invalid length",
			opts: PasswordOptions{
				Length:  0,
				CharSet: DefaultCharSetOptions(),
			},
			wantErr: true,
		},
		{
			name: "length too long",
			opts: PasswordOptions{
				Length:  MaxPasswordLength + 1,
				CharSet: DefaultCharSetOptions(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pwd, err := gen.GeneratePassword(tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(pwd) != tt.wantLen {
				t.Errorf("password length = %d, want %d", len(pwd), tt.wantLen)
			}
		})
	}
}

func TestGeneratePasswordDeterministic(t *testing.T) {
	// Same seed should produce same password
	seed := int64(12345)

	rng1 := decide.NewRNG(&seed)
	gen1 := NewGenerator(rng1)
	pwd1, _ := gen1.GeneratePassword(DefaultPasswordOptions())

	seed2 := int64(12345) // Same seed
	rng2 := decide.NewRNG(&seed2)
	gen2 := NewGenerator(rng2)
	pwd2, _ := gen2.GeneratePassword(DefaultPasswordOptions())

	if pwd1 != pwd2 {
		t.Errorf("same seed produced different passwords: %q vs %q", pwd1, pwd2)
	}
}

func TestGeneratePasswordRequireAll(t *testing.T) {
	seed := int64(42)
	rng := decide.NewRNG(&seed)
	gen := NewGenerator(rng)

	opts := PasswordOptions{
		Length: 20,
		CharSet: CharSetOptions{
			Lowercase:      true,
			Uppercase:      true,
			Digits:         true,
			Symbols:        true,
			AllowAmbiguous: true,
		},
		RequireAll: true,
	}

	// Generate multiple passwords and verify each contains all sets
	for i := 0; i < 10; i++ {
		pwd, err := gen.GeneratePassword(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hasLower := strings.ContainsAny(pwd, Lowercase.Chars)
		hasUpper := strings.ContainsAny(pwd, Uppercase.Chars)
		hasDigit := strings.ContainsAny(pwd, Digits.Chars)
		hasSymbol := strings.ContainsAny(pwd, Symbols.Chars)

		if !hasLower {
			t.Errorf("password %q missing lowercase", pwd)
		}
		if !hasUpper {
			t.Errorf("password %q missing uppercase", pwd)
		}
		if !hasDigit {
			t.Errorf("password %q missing digit", pwd)
		}
		if !hasSymbol {
			t.Errorf("password %q missing symbol", pwd)
		}
	}
}

func TestParsePattern(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		wantLen   int
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "simple pattern",
			pattern: "XXX-999",
			wantLen: 7, // 3 upper + literal - + 3 digit
		},
		{
			name:    "all special chars",
			pattern: "Xx9#?*",
			wantLen: 6,
		},
		{
			name:    "escaped chars",
			pattern: `\X\x\9`,
			wantLen: 3,
		},
		{
			name:      "trailing backslash",
			pattern:   `abc\`,
			wantErr:   true,
			errSubstr: "trailing backslash",
		},
		{
			name:      "invalid escape",
			pattern:   `\z`,
			wantErr:   true,
			errSubstr: "invalid escape",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := ParsePattern(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(tokens) != tt.wantLen {
				t.Errorf("got %d tokens, want %d", len(tokens), tt.wantLen)
			}
		})
	}
}

func TestGenerateFromPattern(t *testing.T) {
	seed := int64(42)
	rng := decide.NewRNG(&seed)

	tests := []struct {
		name    string
		pattern string
		check   func(result string) bool
	}{
		{
			name:    "license key format",
			pattern: "XXXX-XXXX",
			check: func(result string) bool {
				parts := strings.Split(result, "-")
				if len(parts) != 2 {
					return false
				}
				for _, part := range parts {
					if len(part) != 4 {
						return false
					}
					for _, c := range part {
						if c < 'A' || c > 'Z' {
							return false
						}
					}
				}
				return true
			},
		},
		{
			name:    "mixed pattern",
			pattern: "XXX-999-xxx",
			check: func(result string) bool {
				return len(result) == 11 && result[3] == '-' && result[7] == '-'
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := ParsePattern(tt.pattern)
			if err != nil {
				t.Fatalf("ParsePattern error: %v", err)
			}

			result, err := GenerateFromPattern(tokens, rng)
			if err != nil {
				t.Fatalf("GenerateFromPattern error: %v", err)
			}

			if !tt.check(result) {
				t.Errorf("result %q did not pass check for pattern %q", result, tt.pattern)
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		wantStrong string
		wantCommon bool
	}{
		{
			name:       "weak common password",
			password:   "password123",
			wantStrong: "weak",
			wantCommon: true,
		},
		{
			name:       "keyboard walk",
			password:   "asdfghjkl123",
			wantStrong: "weak",
			wantCommon: false, // but HasKeyboard should be true
		},
		{
			name:       "reasonable password",
			password:   "Xk9#mP2@qR5!tY7",
			wantStrong: "strong",
			wantCommon: false,
		},
		{
			name:       "long random",
			password:   "aB3$kL9#mP2@qR5!tY7^uZ1&wX4%vN6*",
			wantStrong: "excellent",
			wantCommon: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := CheckPassword(tt.password)

			if report.Length != len(tt.password) {
				t.Errorf("length = %d, want %d", report.Length, len(tt.password))
			}

			if report.Strength != tt.wantStrong {
				t.Errorf("strength = %q, want %q (entropy: %.1f)", report.Strength, tt.wantStrong, report.Entropy)
			}

			if report.HasCommon != tt.wantCommon {
				t.Errorf("HasCommon = %v, want %v", report.HasCommon, tt.wantCommon)
			}
		})
	}
}

func TestLoadWordListFromString(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "simple list",
			data:    "apple\nbanana\ncherry",
			wantLen: 3,
		},
		{
			name:    "with comments and blanks",
			data:    "# comment\napple\n\nbanana\n# another comment\ncherry\n",
			wantLen: 3,
		},
		{
			name:    "empty list",
			data:    "# only comments\n\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wl, err := LoadWordListFromString(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if wl.Size() != tt.wantLen {
				t.Errorf("word list size = %d, want %d", wl.Size(), tt.wantLen)
			}
		})
	}
}

func TestGeneratePassphrase(t *testing.T) {
	seed := int64(42)
	rng := decide.NewRNG(&seed)

	wl, err := LoadWordListFromString("apple\nbanana\ncherry\ndate\nelderberry\nfig")
	if err != nil {
		t.Fatalf("failed to load word list: %v", err)
	}

	tests := []struct {
		name      string
		wordCount int
		opts      PhraseOptions
		checkFunc func(string) bool
	}{
		{
			name:      "basic passphrase",
			wordCount: 4,
			opts:      DefaultPhraseOptions(),
			checkFunc: func(s string) bool {
				parts := strings.Split(s, "-")
				return len(parts) == 4
			},
		},
		{
			name:      "capitalized",
			wordCount: 3,
			opts: PhraseOptions{
				Separator:  "-",
				Capitalize: true,
			},
			checkFunc: func(s string) bool {
				parts := strings.Split(s, "-")
				for _, p := range parts {
					if len(p) == 0 || p[0] < 'A' || p[0] > 'Z' {
						return false
					}
				}
				return true
			},
		},
		{
			name:      "space separator",
			wordCount: 3,
			opts: PhraseOptions{
				Separator: " ",
			},
			checkFunc: func(s string) bool {
				return strings.Contains(s, " ") && !strings.Contains(s, "-")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phrase, err := GeneratePassphrase(wl, tt.wordCount, tt.opts, rng)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.checkFunc(phrase) {
				t.Errorf("passphrase %q did not pass check", phrase)
			}
		})
	}
}

func TestGenerateAPIKey(t *testing.T) {
	seed := int64(42)
	rng := decide.NewRNG(&seed)

	tests := []struct {
		name      string
		opts      APIKeyOptions
		wantLen   int
		checkFunc func(string) bool
	}{
		{
			name:    "default hex",
			opts:    DefaultAPIKeyOptions(),
			wantLen: 32,
			checkFunc: func(s string) bool {
				for _, c := range s {
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
						return false
					}
				}
				return true
			},
		},
		{
			name: "with prefix",
			opts: APIKeyOptions{
				Prefix: "sk_test",
				Length: 16,
				Format: FormatHex,
			},
			wantLen: 8 + 16, // "sk_test_" + 16 chars
			checkFunc: func(s string) bool {
				return strings.HasPrefix(s, "sk_test_")
			},
		},
		{
			name: "base64url",
			opts: APIKeyOptions{
				Length: 20,
				Format: FormatBase64URL,
			},
			wantLen: 20,
			checkFunc: func(s string) bool {
				for _, c := range s {
					valid := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
						(c >= '0' && c <= '9') || c == '-' || c == '_'
					if !valid {
						return false
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateAPIKey(tt.opts, rng)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(key) != tt.wantLen {
				t.Errorf("key length = %d, want %d (key: %q)", len(key), tt.wantLen, key)
			}
			if !tt.checkFunc(key) {
				t.Errorf("key %q did not pass format check", key)
			}
		})
	}
}
