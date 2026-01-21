package crypto

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestEncryptDecryptIntegration tests the full encrypt/decrypt workflow.
func TestEncryptDecryptIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	// Setup a realistic wiki structure
	setupWikiStructure(t, tmpDir)

	passphrase := "integration-test-passphrase"

	// Phase 1: Encrypt the wiki
	t.Log("Phase 1: Encrypting wiki...")
	encReport, err := EncryptWiki(tmpDir, passphrase)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	t.Logf("Encrypted %d files, skipped %d, failed %d",
		len(encReport.Encrypted), len(encReport.Skipped), len(encReport.Failed))

	// Verify encryption worked
	verifyEncrypted(t, tmpDir)

	// Phase 2: Decrypt the wiki
	t.Log("Phase 2: Decrypting wiki...")
	decReport, err := DecryptWiki(tmpDir, passphrase)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	t.Logf("Decrypted %d files, skipped %d, failed %d",
		len(decReport.Decrypted), len(decReport.Skipped), len(decReport.Failed))

	// Verify decryption worked
	verifyDecrypted(t, tmpDir)

	// Phase 3: Verify content integrity
	verifyContentIntegrity(t, tmpDir)
}

// TestPartialEncryptionRecovery tests recovery when encryption fails partway through.
func TestPartialEncryptionRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Create test files
	for i := 1; i <= 5; i++ {
		path := filepath.Join(notesDir, fmt.Sprintf("note%d.md", i))
		if err := os.WriteFile(path, []byte(fmt.Sprintf("content %d", i)), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	// Create a read-only file that will fail to encrypt
	readonlyFile := filepath.Join(notesDir, "readonly.md")
	if err := os.WriteFile(readonlyFile, []byte("readonly content"), 0444); err != nil {
		t.Fatalf("Failed to write readonly file: %v", err)
	}

	// Try to encrypt (should succeed for most files, fail for readonly)
	report, err := EncryptWiki(tmpDir, "test-pass")
	if err != nil {
		// This is expected if partial encryption occurs
		t.Logf("Encryption error (expected for readonly file): %v", err)
	}

	// Should have some successful encryptions
	if len(report.Encrypted) == 0 {
		t.Error("Expected some successful encryptions")
	}

	t.Logf("Partial encryption: %d succeeded, %d failed",
		len(report.Encrypted), len(report.Failed))
}

// TestMultipleEncryptDecryptCycles tests multiple encryption/decryption cycles.
func TestMultipleEncryptDecryptCycles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Create test file
	testFile := filepath.Join(notesDir, "test.md")
	originalContent := "original content for multiple cycles"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	passphrase := "cycle-test-pass"

	// Run 3 encrypt/decrypt cycles
	for cycle := 1; cycle <= 3; cycle++ {
		t.Logf("Cycle %d: Encrypting...", cycle)
		if _, err := EncryptWiki(tmpDir, passphrase); err != nil {
			t.Fatalf("Cycle %d encryption failed: %v", cycle, err)
		}

		t.Logf("Cycle %d: Decrypting...", cycle)
		if _, err := DecryptWiki(tmpDir, passphrase); err != nil {
			t.Fatalf("Cycle %d decryption failed: %v", cycle, err)
		}

		// Verify content is still correct
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Cycle %d: failed to read file: %v", cycle, err)
		}
		if string(content) != originalContent {
			t.Errorf("Cycle %d: content mismatch", cycle)
		}
	}
}

// Helper functions

func setupWikiStructure(t *testing.T, wikiDir string) {
	t.Helper()

	// Create directories
	dirs := []string{
		"notes",
		"recipes",
		"tasks",
		".git/objects", // Should be skipped
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(wikiDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create files
	files := map[string]string{
		"notes/2026-01-15.md":  "# Daily Note\n\n## 10:00\nMeeting notes",
		"notes/2026-01-16.md":  "# Daily Note\n\n## 11:30\nProject update",
		"notes/abc12345.md":    "# Floating Note\n\nSome ideas",
		"recipes/pasta.md":     "# Pasta Recipe\n\nIngredients...",
		"tasks/tasks.json":     `{"tasks":[]}`,
		"config.txt":           "# Config (should not be encrypted)",
		".git/objects/test.md": "# Git object (should not be encrypted)",
	}

	for path, content := range files {
		fullPath := filepath.Join(wikiDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}
}

func verifyEncrypted(t *testing.T, wikiDir string) {
	t.Helper()

	// .encrypted marker should exist
	markerPath := filepath.Join(wikiDir, ".encrypted")
	if _, err := os.Stat(markerPath); err != nil {
		t.Errorf("Expected .encrypted marker to exist: %v", err)
	}

	// .md and .json files should be replaced with .enc
	notesDir := filepath.Join(wikiDir, "notes")
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		t.Fatalf("Failed to read notes dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".md") && !strings.Contains(name, ".enc") {
			t.Errorf("Found unencrypted .md file: %s", name)
		}
	}

	// config.txt should still exist (not encrypted)
	configPath := filepath.Join(wikiDir, "config.txt")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("config.txt should not be encrypted: %v", err)
	}

	// .git files should still exist (not encrypted)
	gitFile := filepath.Join(wikiDir, ".git/objects/test.md")
	if _, err := os.Stat(gitFile); err != nil {
		t.Errorf(".git files should not be encrypted: %v", err)
	}
}

func verifyDecrypted(t *testing.T, wikiDir string) {
	t.Helper()

	// .encrypted marker should NOT exist
	markerPath := filepath.Join(wikiDir, ".encrypted")
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error(".encrypted marker should be removed after decryption")
	}

	// .enc files should be removed
	err := filepath.Walk(wikiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".enc") {
			t.Errorf("Found .enc file after decryption: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}
}

func verifyContentIntegrity(t *testing.T, wikiDir string) {
	t.Helper()

	// Verify specific file contents
	testCases := []struct {
		path     string
		contains string
	}{
		{"notes/2026-01-15.md", "Meeting notes"},
		{"notes/2026-01-16.md", "Project update"},
		{"notes/abc12345.md", "Some ideas"},
		{"recipes/pasta.md", "Ingredients"},
		{"tasks/tasks.json", `"tasks"`},
	}

	for _, tc := range testCases {
		fullPath := filepath.Join(wikiDir, tc.path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", tc.path, err)
			continue
		}
		if !strings.Contains(string(content), tc.contains) {
			t.Errorf("File %s missing expected content %q", tc.path, tc.contains)
		}
	}
}

// TestCLIEncryptDecrypt tests encrypt/decrypt via CLI (if regimen binary exists).
func TestCLIEncryptDecrypt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// Check if regimen binary exists
	_, err := exec.LookPath("regimen")
	if err != nil {
		t.Skip("regimen binary not found in PATH, skipping CLI test")
	}

	tmpDir := t.TempDir()
	setupWikiStructure(t, tmpDir)

	passphrase := "cli-test-pass"

	// Test encrypt command
	t.Log("Testing 'regimen encrypt' command...")
	encryptCmd := exec.Command("regimen", "--wiki-dir", tmpDir, "encrypt", "--passphrase-stdin")
	encryptCmd.Stdin = strings.NewReader(passphrase + "\n")
	if output, err := encryptCmd.CombinedOutput(); err != nil {
		t.Fatalf("regimen encrypt failed: %v\nOutput: %s", err, output)
	}

	// Verify encryption worked
	markerPath := filepath.Join(tmpDir, ".encrypted")
	if _, err := os.Stat(markerPath); err != nil {
		t.Error(".encrypted marker not created by CLI")
	}

	// Test decrypt command
	t.Log("Testing 'regimen decrypt' command...")
	decryptCmd := exec.Command("regimen", "--wiki-dir", tmpDir, "decrypt", "--passphrase-stdin")
	decryptCmd.Stdin = strings.NewReader(passphrase + "\n")
	if output, err := decryptCmd.CombinedOutput(); err != nil {
		t.Fatalf("regimen decrypt failed: %v\nOutput: %s", err, output)
	}

	// Verify decryption worked
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error(".encrypted marker still exists after CLI decrypt")
	}
}
