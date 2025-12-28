package password

import "math"

// CalculatePasswordEntropy calculates entropy in bits for a character-based password.
// Formula: log2(charsetSize^length) = length * log2(charsetSize)
func CalculatePasswordEntropy(charsetSize, length int) float64 {
	if charsetSize <= 0 || length <= 0 {
		return 0
	}
	return float64(length) * math.Log2(float64(charsetSize))
}

// CalculatePhraseEntropy calculates entropy in bits for a word-based passphrase.
// Formula: log2(wordlistSize^wordCount) = wordCount * log2(wordlistSize)
func CalculatePhraseEntropy(wordlistSize, wordCount int) float64 {
	if wordlistSize <= 0 || wordCount <= 0 {
		return 0
	}
	return float64(wordCount) * math.Log2(float64(wordlistSize))
}

// EntropyToStrength maps entropy bits to a strength rating.
func EntropyToStrength(entropy float64) string {
	switch {
	case entropy < 40:
		return "weak"
	case entropy < 60:
		return "fair"
	case entropy < 80:
		return "good"
	case entropy < 100:
		return "strong"
	default:
		return "excellent"
	}
}
