package guard

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gitlab.com/caffeinatedjack/sleepless/pkg/redact"
)

// Baseline represents a set of known findings to suppress.
type Baseline struct {
	Version      int       `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	Fingerprints []string  `json:"fingerprints"`
}

// LoadBaseline reads a baseline file from disk.
func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline: %w", err)
	}

	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, fmt.Errorf("failed to parse baseline: %w", err)
	}

	if baseline.Version != 1 {
		return nil, fmt.Errorf("unsupported baseline version: %d", baseline.Version)
	}

	return &baseline, nil
}

// SaveBaseline writes a baseline to disk or stdout.
func SaveBaseline(baseline *Baseline, path string) error {
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %w", err)
	}

	if path == "" || path == "-" {
		// Write to stdout
		_, err = os.Stdout.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
		fmt.Println() // Add newline
		return nil
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write baseline: %w", err)
	}

	return nil
}

// CreateBaseline creates a baseline from scan results.
func CreateBaseline(results *ScanResult) *Baseline {
	baseline := &Baseline{
		Version:      1,
		CreatedAt:    time.Now(),
		Fingerprints: make([]string, 0, len(results.Findings)),
	}

	// Generate fingerprints for all findings
	seen := make(map[string]bool)
	for _, f := range results.Findings {
		fp := Fingerprint(f)
		if !seen[fp] {
			baseline.Fingerprints = append(baseline.Fingerprints, fp)
			seen[fp] = true
		}
	}

	return baseline
}

// ApplyBaseline filters findings based on a baseline, returning only new findings.
func ApplyBaseline(results *ScanResult, baseline *Baseline) *ScanResult {
	if baseline == nil {
		return results
	}

	// Build lookup map
	allowed := make(map[string]bool)
	for _, fp := range baseline.Fingerprints {
		allowed[fp] = true
	}

	// Filter findings
	filtered := &ScanResult{
		Findings: make([]Finding, 0),
		Counts:   make(map[redact.PatternType]int),
		Total:    0,
	}

	for _, f := range results.Findings {
		fp := Fingerprint(f)
		if !allowed[fp] {
			// This is a new finding (not in baseline)
			filtered.Findings = append(filtered.Findings, f)
			filtered.Counts[f.Type]++
			filtered.Total++
		}
	}

	return filtered
}
