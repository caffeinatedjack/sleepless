package password

import (
	"strings"
	"unicode"
)

// StrengthReport contains the results of password strength analysis.
type StrengthReport struct {
	Length      int      `json:"password_length"`
	CharSets    []string `json:"character_sets"`
	Entropy     float64  `json:"entropy"`
	Strength    string   `json:"strength"`
	HasCommon   bool     `json:"has_common,omitempty"`
	HasKeyboard bool     `json:"has_keyboard,omitempty"`
}

// Common weak passwords to check against.
var commonPasswords = []string{
	"password", "123456", "12345678", "qwerty", "abc123",
	"monkey", "1234567", "letmein", "trustno1", "dragon",
	"baseball", "iloveyou", "master", "sunshine", "ashley",
	"football", "shadow", "123123", "654321", "superman",
	"qazwsx", "michael", "password1", "password123", "welcome",
	"login", "admin", "princess", "starwars", "passw0rd",
}

// Common keyboard walks to detect.
var keyboardWalks = []string{
	"qwerty", "qwertyuiop", "asdfgh", "asdfghjkl", "zxcvbn",
	"qazwsx", "1qaz2wsx", "qweasd", "123qwe", "1q2w3e",
}

// CheckPassword analyzes a password and returns a strength report.
func CheckPassword(password string) *StrengthReport {
	report := &StrengthReport{
		Length: len(password),
	}

	// Detect character sets
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSymbol := false
	charsetSize := 0

	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			if !hasLower {
				hasLower = true
				charsetSize += 26
			}
		case unicode.IsUpper(r):
			if !hasUpper {
				hasUpper = true
				charsetSize += 26
			}
		case unicode.IsDigit(r):
			if !hasDigit {
				hasDigit = true
				charsetSize += 10
			}
		default:
			// Assume symbol/special character
			if !hasSymbol {
				hasSymbol = true
				charsetSize += 32 // Approximate symbol count
			}
		}
	}

	// Build character set list
	if hasLower {
		report.CharSets = append(report.CharSets, "lowercase")
	}
	if hasUpper {
		report.CharSets = append(report.CharSets, "uppercase")
	}
	if hasDigit {
		report.CharSets = append(report.CharSets, "digits")
	}
	if hasSymbol {
		report.CharSets = append(report.CharSets, "symbols")
	}

	// Calculate entropy
	if charsetSize > 0 && len(password) > 0 {
		report.Entropy = CalculatePasswordEntropy(charsetSize, len(password))
	}

	// Check for common passwords
	lowerPwd := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPwd == common || strings.Contains(lowerPwd, common) {
			report.HasCommon = true
			break
		}
	}

	// Check for keyboard walks
	for _, walk := range keyboardWalks {
		if strings.Contains(lowerPwd, walk) {
			report.HasKeyboard = true
			break
		}
	}

	// Determine strength rating
	// Penalize if common pattern detected
	effectiveEntropy := report.Entropy
	if report.HasCommon || report.HasKeyboard {
		// Significantly reduce effective entropy for common patterns
		effectiveEntropy = effectiveEntropy * 0.3
	}

	report.Strength = EntropyToStrength(effectiveEntropy)

	return report
}
