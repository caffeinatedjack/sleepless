package password

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"gitlab.com/caffeinatedjack/sleepless/pkg/decide"
)

// WordList holds a list of words for passphrase generation.
type WordList struct {
	words []string
}

// NewWordList creates a WordList from a slice of words.
func NewWordList(words []string) *WordList {
	return &WordList{words: words}
}

// Size returns the number of words in the list.
func (wl *WordList) Size() int {
	return len(wl.words)
}

// ErrEmptyWordList is returned when a word list has no words.
var ErrEmptyWordList = errors.New("word list is empty")

// ErrWordListTooSmall is a warning threshold for low-entropy word lists.
const WordListMinSize = 100

// LoadWordList loads a word list from a file.
// Lines starting with # are treated as comments and skipped.
// Empty lines are skipped.
func LoadWordList(path string) (*WordList, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open word list: %w", err)
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		words = append(words, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read word list: %w", err)
	}

	if len(words) == 0 {
		return nil, ErrEmptyWordList
	}

	return &WordList{words: words}, nil
}

// LoadWordListFromString parses a word list from a string (one word per line).
func LoadWordListFromString(data string) (*WordList, error) {
	var words []string
	for _, line := range strings.Split(data, "\n") {
		word := strings.TrimSpace(line)
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}
		words = append(words, word)
	}

	if len(words) == 0 {
		return nil, ErrEmptyWordList
	}

	return &WordList{words: words}, nil
}

// SelectWords randomly selects n words from the word list.
func (wl *WordList) SelectWords(rng decide.RNG, n int) ([]string, error) {
	if len(wl.words) == 0 {
		return nil, ErrEmptyWordList
	}
	if n < 1 {
		return nil, errors.New("word count must be at least 1")
	}

	words := make([]string, n)
	for i := 0; i < n; i++ {
		idx, err := rng.Intn(len(wl.words))
		if err != nil {
			return nil, fmt.Errorf("failed to select word: %w", err)
		}
		words[i] = wl.words[idx]
	}

	return words, nil
}

// PhraseOptions configures passphrase formatting.
type PhraseOptions struct {
	Separator  string
	Capitalize bool
	Numbers    bool // Append a random digit to each word
}

// DefaultPhraseOptions returns sensible defaults for passphrase generation.
func DefaultPhraseOptions() PhraseOptions {
	return PhraseOptions{
		Separator:  "-",
		Capitalize: false,
		Numbers:    false,
	}
}

// FormatPassphrase formats a list of words into a passphrase.
func FormatPassphrase(words []string, opts PhraseOptions, rng decide.RNG) (string, error) {
	formatted := make([]string, len(words))

	for i, word := range words {
		w := word

		if opts.Capitalize && len(w) > 0 {
			w = strings.ToUpper(string(w[0])) + w[1:]
		}

		if opts.Numbers {
			digit, err := rng.Intn(10)
			if err != nil {
				return "", fmt.Errorf("failed to generate digit: %w", err)
			}
			w = fmt.Sprintf("%s%d", w, digit)
		}

		formatted[i] = w
	}

	return strings.Join(formatted, opts.Separator), nil
}

// GeneratePassphrase generates a complete passphrase.
func GeneratePassphrase(wl *WordList, wordCount int, opts PhraseOptions, rng decide.RNG) (string, error) {
	words, err := wl.SelectWords(rng, wordCount)
	if err != nil {
		return "", err
	}

	return FormatPassphrase(words, opts, rng)
}
