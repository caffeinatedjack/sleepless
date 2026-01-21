package guard

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/caffeinatedjack/sleepless/pkg/redact"
)

func TestFingerprint(t *testing.T) {
	// Create two findings with the same type and rawMatch
	f1 := Finding{
		File:     "file1.txt",
		Line:     10,
		Column:   5,
		Type:     redact.Email,
		Excerpt:  "test@example.com",
		rawMatch: "test@example.com",
	}

	f2 := Finding{
		File:     "file2.txt", // Different file
		Line:     20,          // Different line
		Column:   10,          // Different column
		Type:     redact.Email,
		Excerpt:  "test@example.com",
		rawMatch: "test@example.com",
	}

	fp1 := Fingerprint(f1)
	fp2 := Fingerprint(f2)

	// Same type and rawMatch should produce same fingerprint regardless of file/line/column
	if fp1 != fp2 {
		t.Errorf("Expected same fingerprint for same type/match, got %s != %s", fp1, fp2)
	}

	// Different rawMatch should produce different fingerprint
	f3 := Finding{
		Type:     redact.Email,
		rawMatch: "different@example.com",
	}
	fp3 := Fingerprint(f3)
	if fp1 == fp3 {
		t.Errorf("Expected different fingerprints for different matches")
	}
}

func TestCreateBaseline(t *testing.T) {
	results := &ScanResult{
		Findings: []Finding{
			{Type: redact.Email, rawMatch: "test1@example.com"},
			{Type: redact.Email, rawMatch: "test2@example.com"},
			{Type: redact.Phone, rawMatch: "555-1234"},
			// Duplicate
			{Type: redact.Email, rawMatch: "test1@example.com"},
		},
		Counts: map[redact.PatternType]int{
			redact.Email: 3,
			redact.Phone: 1,
		},
		Total: 4,
	}

	baseline := CreateBaseline(results)

	if baseline.Version != 1 {
		t.Errorf("Expected version 1, got %d", baseline.Version)
	}

	// Should have 3 unique fingerprints (duplicate should be deduplicated)
	if len(baseline.Fingerprints) != 3 {
		t.Errorf("Expected 3 unique fingerprints, got %d", len(baseline.Fingerprints))
	}
}

func TestApplyBaseline(t *testing.T) {
	results := &ScanResult{
		Findings: []Finding{
			{Type: redact.Email, rawMatch: "test1@example.com", File: "a.txt", Line: 1},
			{Type: redact.Email, rawMatch: "test2@example.com", File: "b.txt", Line: 2},
			{Type: redact.Phone, rawMatch: "555-1234", File: "c.txt", Line: 3},
		},
		Counts: map[redact.PatternType]int{
			redact.Email: 2,
			redact.Phone: 1,
		},
		Total: 3,
	}

	// Create baseline with first two findings
	baseline := &Baseline{
		Version: 1,
		Fingerprints: []string{
			Fingerprint(results.Findings[0]),
			Fingerprint(results.Findings[1]),
		},
	}

	filtered := ApplyBaseline(results, baseline)

	// Only the phone number should remain (not in baseline)
	if len(filtered.Findings) != 1 {
		t.Errorf("Expected 1 finding after baseline, got %d", len(filtered.Findings))
	}

	if filtered.Total != 1 {
		t.Errorf("Expected total 1, got %d", filtered.Total)
	}

	if filtered.Findings[0].Type != redact.Phone {
		t.Errorf("Expected Phone finding, got %s", filtered.Findings[0].Type)
	}
}

func TestSaveAndLoadBaseline(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, "baseline.json")

	original := &Baseline{
		Version: 1,
		Fingerprints: []string{
			"sha256:abc123",
			"sha256:def456",
		},
	}

	// Save baseline
	err := SaveBaseline(original, baselinePath)
	if err != nil {
		t.Fatalf("Failed to save baseline: %v", err)
	}

	// Load baseline
	loaded, err := LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	// Verify
	if loaded.Version != original.Version {
		t.Errorf("Version mismatch: expected %d, got %d", original.Version, loaded.Version)
	}

	if len(loaded.Fingerprints) != len(original.Fingerprints) {
		t.Errorf("Fingerprint count mismatch: expected %d, got %d", len(original.Fingerprints), len(loaded.Fingerprints))
	}

	for i, fp := range original.Fingerprints {
		if loaded.Fingerprints[i] != fp {
			t.Errorf("Fingerprint %d mismatch: expected %s, got %s", i, fp, loaded.Fingerprints[i])
		}
	}
}

func TestScanFile(t *testing.T) {
	// Create a temporary file with test content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := `This is a test file.
Email: test@example.com
Phone: 555-1234-5678
Another email: alice@test.org
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	scanner, err := NewScanner()
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	result, err := scanner.ScanFile(testFile)
	if err != nil {
		t.Fatalf("Failed to scan file: %v", err)
	}

	// Should find at least 2 emails and 1 phone
	if result.Total < 3 {
		t.Errorf("Expected at least 3 findings, got %d", result.Total)
	}

	// Check that findings have proper structure
	for _, f := range result.Findings {
		if f.File != testFile {
			t.Errorf("Expected file %s, got %s", testFile, f.File)
		}
		if f.Line == 0 {
			t.Errorf("Line should be non-zero")
		}
		if f.Column == 0 {
			t.Errorf("Column should be non-zero")
		}
		if f.Excerpt == "" {
			t.Errorf("Excerpt should not be empty")
		}
	}
}
