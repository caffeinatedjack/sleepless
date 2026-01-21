// Package guard provides secret and PII scanning utilities for nightwatch guard.
package guard

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"gitlab.com/caffeinatedjack/sleepless/pkg/redact"
)

// Finding represents a detected match of sensitive data in a file.
type Finding struct {
	File     string             `json:"file"`
	Line     int                `json:"line"`
	Column   int                `json:"column"`
	Type     redact.PatternType `json:"type"`
	Excerpt  string             `json:"excerpt"` // Redacted excerpt for display
	rawMatch string             // Not exported - used for fingerprinting only
}

// ScanResult contains the results of scanning one or more files.
type ScanResult struct {
	Findings []Finding                  `json:"findings"`
	Counts   map[redact.PatternType]int `json:"counts"`
	Total    int                        `json:"total"`
}

// Scanner scans files for sensitive data patterns.
type Scanner struct {
	redactor *redact.Redactor
}

// NewScanner creates a Scanner with default patterns.
func NewScanner() (*Scanner, error) {
	// Use default patterns from redact package
	redactor, err := redact.NewRedactor(redact.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create redactor: %w", err)
	}
	return &Scanner{redactor: redactor}, nil
}

// ScanFile scans a single file and returns findings.
func (s *Scanner) ScanFile(path string) (*ScanResult, error) {
	// Check if file is likely UTF-8 text
	if !isTextFile(path) {
		return nil, fmt.Errorf("skipping binary file: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Use the redactor's Check method
	report, err := s.redactor.Check(f)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Convert redact.Finding to guard.Finding with file info
	result := &ScanResult{
		Findings: make([]Finding, 0, len(report.Findings)),
		Counts:   report.Counts,
		Total:    len(report.Findings),
	}

	// Re-read file to get excerpts for findings
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file for excerpts: %w", err)
	}
	lines := strings.Split(string(content), "\n")

	for _, rf := range report.Findings {
		// Get raw match for fingerprinting
		var rawMatch string
		if rf.Line > 0 && rf.Line <= len(lines) {
			line := lines[rf.Line-1]
			if rf.Column > 0 && rf.Column <= len(line) {
				// Extract the matched portion from the line
				// We need to find what was matched - use the pattern regex
				for _, p := range s.redactor.Patterns() {
					if p.Type == rf.Type {
						matches := p.Regex.FindStringIndex(line[rf.Column-1:])
						if matches != nil {
							rawMatch = line[rf.Column-1+matches[0] : rf.Column-1+matches[1]]
							break
						}
					}
				}
			}
		}

		// Create redacted excerpt
		excerpt := ""
		if rf.Line > 0 && rf.Line <= len(lines) {
			excerpt = s.redactor.Redact(lines[rf.Line-1])
			// Truncate if too long
			if len(excerpt) > 100 {
				excerpt = excerpt[:97] + "..."
			}
		}

		result.Findings = append(result.Findings, Finding{
			File:     path,
			Line:     rf.Line,
			Column:   rf.Column,
			Type:     rf.Type,
			Excerpt:  excerpt,
			rawMatch: rawMatch,
		})
	}

	return result, nil
}

// ScanDirectory recursively scans all text files in a directory.
func (s *Scanner) ScanDirectory(root string) (*ScanResult, error) {
	result := &ScanResult{
		Findings: make([]Finding, 0),
		Counts:   make(map[redact.PatternType]int),
		Total:    0,
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() && strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Try to scan the file
		fileResult, err := s.ScanFile(path)
		if err != nil {
			// Log warning but continue scanning
			// TODO: collect warnings
			return nil
		}

		// Merge results
		result.Findings = append(result.Findings, fileResult.Findings...)
		result.Total += fileResult.Total
		for t, c := range fileResult.Counts {
			result.Counts[t] += c
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("directory walk failed: %w", err)
	}

	return result, nil
}

// isTextFile checks if a file is likely UTF-8 text by reading the first few bytes.
func isTextFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	buf = buf[:n]

	// Check for null bytes (common in binary files)
	for _, b := range buf {
		if b == 0 {
			return false
		}
	}

	// Check if valid UTF-8
	return utf8.Valid(buf)
}

// Fingerprint computes a stable identifier for a finding.
// This is used for baseline suppression.
func Fingerprint(f Finding) string {
	// Hash the type and raw match value
	data := fmt.Sprintf("%s:%s", f.Type, f.rawMatch)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("sha256:%x", hash)
}
