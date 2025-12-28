package redact

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Mode controls how detected values are rewritten.
//
// The default is ModeReplace, which swaps matches for a bracketed token
// like [EMAIL] or [UUID].
type Mode int

const (
	ModeReplace Mode = iota // Replace with [TYPE]
	ModeMask                // Partial masking
	ModeHash                // Replace with [TYPE:hash]
)

// Options describes what to look for and how to replace it.
//
// Use Only/Except to narrow the built-in set, and CustomRegex to add a one-off
// pattern.
type Options struct {
	Mode        Mode
	Only        []PatternType
	Except      []PatternType
	CustomRegex *regexp.Regexp
	CustomName  string
}

// Redactor applies one or more patterns to text and rewrites matches.
//
// Create one with NewRedactor and reuse it; it keeps the compiled regexes and
// options together.
type Redactor struct {
	patterns []Pattern
	mode     Mode
}

// NewRedactor builds a Redactor from Options.
//
// It starts from the default pattern set, then applies Only/Except and adds the
// optional CustomRegex.
func NewRedactor(opts Options) (*Redactor, error) {
	patterns := selectPatterns(opts)

	// Add custom pattern if provided
	if opts.CustomRegex != nil {
		name := opts.CustomName
		if name == "" {
			name = "CUSTOM"
		}
		patterns = append(patterns, Pattern{
			Type:    PatternType(name),
			Regex:   opts.CustomRegex,
			Replace: fmt.Sprintf("[%s]", name),
		})
	}

	if len(patterns) == 0 {
		return nil, fmt.Errorf("no patterns selected")
	}

	return &Redactor{
		patterns: patterns,
		mode:     opts.Mode,
	}, nil
}

func selectPatterns(opts Options) []Pattern {
	// Start with defaults
	patterns := make([]Pattern, len(DefaultPatterns))
	copy(patterns, DefaultPatterns)

	// If --only is specified, filter to just those types
	if len(opts.Only) > 0 {
		only := make(map[PatternType]bool)
		for _, t := range opts.Only {
			only[t] = true
		}
		filtered := patterns[:0]
		for _, p := range patterns {
			if only[p.Type] {
				filtered = append(filtered, p)
			}
		}
		patterns = filtered
	}

	// Remove --except types
	if len(opts.Except) > 0 {
		except := make(map[PatternType]bool)
		for _, t := range opts.Except {
			except[t] = true
		}
		filtered := patterns[:0]
		for _, p := range patterns {
			if !except[p.Type] {
				filtered = append(filtered, p)
			}
		}
		patterns = filtered
	}

	return patterns
}

func lineScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	const maxLineSize = 1024 * 1024 // 1MB
	scanner.Buffer(make([]byte, 64*1024), maxLineSize)
	return scanner
}

// Redact returns input with any matches replaced according to the Redactor mode.
func (r *Redactor) Redact(input string) string {
	result := input
	for _, p := range r.patterns {
		pattern := p // capture for closure
		result = pattern.Regex.ReplaceAllStringFunc(result, func(match string) string {
			// If pattern has a Matcher, only redact if it returns true
			if pattern.Matcher != nil && !pattern.Matcher(match) {
				return match
			}
			return r.replace(pattern.Type, match)
		})
	}
	return result
}

func (r *Redactor) replace(ptype PatternType, match string) string {
	switch r.mode {
	case ModeMask:
		return mask(ptype, match)
	case ModeHash:
		return hash(ptype, match)
	default:
		return fmt.Sprintf("[%s]", ptype)
	}
}

func mask(ptype PatternType, match string) string {
	switch ptype {
	case Email:
		parts := strings.Split(match, "@")
		if len(parts) == 2 {
			local := parts[0]
			domain := parts[1]
			if len(local) > 1 {
				local = string(local[0]) + strings.Repeat("*", len(local)-1)
			}
			domParts := strings.Split(domain, ".")
			if len(domParts) > 0 {
				domParts[0] = strings.Repeat("*", len(domParts[0]))
			}
			return local + "@" + strings.Join(domParts, ".")
		}
		return strings.Repeat("*", len(match))
	case Phone:
		// Keep last 4 digits
		if len(match) >= 4 {
			return strings.Repeat("*", len(match)-4) + match[len(match)-4:]
		}
		return strings.Repeat("*", len(match))
	case CreditCard:
		// Keep last 4 digits
		cleaned := strings.ReplaceAll(strings.ReplaceAll(match, "-", ""), " ", "")
		if len(cleaned) >= 4 {
			return strings.Repeat("*", len(cleaned)-4) + cleaned[len(cleaned)-4:]
		}
		return strings.Repeat("*", len(match))
	case IP:
		return "[***.***.***.***]"
	default:
		// Generic masking: show first and last char
		if len(match) > 2 {
			return string(match[0]) + strings.Repeat("*", len(match)-2) + string(match[len(match)-1])
		}
		return strings.Repeat("*", len(match))
	}
}

func hash(ptype PatternType, match string) string {
	h := sha256.Sum256([]byte(match))
	return fmt.Sprintf("[%s:%x]", ptype, h[:3])
}

// Patterns returns the patterns this Redactor will apply, in order.
func (r *Redactor) Patterns() []Pattern {
	return r.patterns
}

// RedactStream redacts input as it is read and writes the result to w.
//
// It processes the stream line-by-line to keep memory use predictable.
func (rd *Redactor) RedactStream(r io.Reader, w io.Writer) error {
	scanner := lineScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		redacted := rd.Redact(line)
		if _, err := fmt.Fprintln(w, redacted); err != nil {
			return fmt.Errorf("write failed: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read failed: %w", err)
	}

	return nil
}

// Finding represents a detected match of sensitive data.
type Finding struct {
	Type   PatternType
	Line   int
	Column int
}

// Report contains the results of a check operation.
type Report struct {
	Findings []Finding
	Counts   map[PatternType]int
}

// Check scans the input and returns findings without modifying content.
func (rd *Redactor) Check(r io.Reader) (*Report, error) {
	report := &Report{
		Counts: make(map[PatternType]int),
	}

	scanner := lineScanner(r)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, p := range rd.patterns {
			matches := p.Regex.FindAllStringSubmatchIndex(line, -1)
			for _, match := range matches {
				matchStr := line[match[0]:match[1]]
				// If pattern has a Matcher, only count if it returns true
				if p.Matcher != nil && !p.Matcher(matchStr) {
					continue
				}
				report.Findings = append(report.Findings, Finding{
					Type:   p.Type,
					Line:   lineNum,
					Column: match[0] + 1, // 1-indexed
				})
				report.Counts[p.Type]++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return report, nil
}

// TotalFindings returns the total number of findings.
func (r *Report) TotalFindings() int {
	return len(r.Findings)
}
