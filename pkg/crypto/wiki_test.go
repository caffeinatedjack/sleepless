package crypto

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	// Create temporary wiki directory
	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Create test files
	testFiles := map[string]string{
		"notes/test1.md":  "# Test Note 1\n\nThis is a test note.",
		"notes/test2.md":  "# Test Note 2\n\nAnother test note with some content.",
		"tasks.json":      `{"tasks": [{"id": "abc", "title": "Test task"}]}`,
		"config.txt":      "This should not be encrypted",
		"notes/nested.md": "Nested note content",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", path, err)
		}
	}

	passphrase := "test-passphrase-123"

	// Test encryption
	encReport, err := EncryptWiki(tmpDir, passphrase)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Verify encryption report
	if len(encReport.Encrypted) != 4 { // test1.md, test2.md, nested.md, tasks.json
		t.Errorf("Expected 4 encrypted files, got %d", len(encReport.Encrypted))
	}

	// Verify .encrypted marker exists
	markerPath := filepath.Join(tmpDir, ".encrypted")
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error(".encrypted marker file not created")
	}

	// Verify original files are replaced with .enc files
	for path := range testFiles {
		if filepath.Ext(path) == ".md" || filepath.Ext(path) == ".json" {
			fullPath := filepath.Join(tmpDir, path)
			if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
				t.Errorf("Original file %s still exists after encryption", path)
			}
			encPath := fullPath + ".enc"
			if _, err := os.Stat(encPath); os.IsNotExist(err) {
				t.Errorf("Encrypted file %s not created", encPath)
			}
		}
	}

	// Verify config.txt was not encrypted
	configPath := filepath.Join(tmpDir, "config.txt")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.txt should not be encrypted")
	}

	// Test decryption
	decReport, err := DecryptWiki(tmpDir, passphrase)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Verify decryption report
	if len(decReport.Decrypted) != 4 {
		t.Errorf("Expected 4 decrypted files, got %d", len(decReport.Decrypted))
	}

	// Verify original files are restored
	for path, expectedContent := range testFiles {
		if filepath.Ext(path) == ".md" || filepath.Ext(path) == ".json" {
			fullPath := filepath.Join(tmpDir, path)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				t.Errorf("Failed to read decrypted file %s: %v", path, err)
				continue
			}
			if string(content) != expectedContent {
				t.Errorf("Content mismatch for %s:\nExpected: %q\nGot: %q", path, expectedContent, string(content))
			}

			// Verify .enc file is removed
			encPath := fullPath + ".enc"
			if _, err := os.Stat(encPath); !os.IsNotExist(err) {
				t.Errorf("Encrypted file %s still exists after decryption", encPath)
			}
		}
	}

	// Verify .encrypted marker is removed
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error(".encrypted marker file still exists after decryption")
	}
}

func TestEncryptAlreadyEncrypted(t *testing.T) {
	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(notesDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	passphrase := "test-passphrase"

	// First encryption should succeed
	if _, err := EncryptWiki(tmpDir, passphrase); err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	// Second encryption should fail
	if _, err := EncryptWiki(tmpDir, passphrase); err == nil {
		t.Error("Expected error when encrypting already encrypted wiki, got nil")
	}
}

func TestDecryptNotEncrypted(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to decrypt without .encrypted marker
	_, err := DecryptWiki(tmpDir, "test-passphrase")
	if err == nil {
		t.Error("Expected error when decrypting non-encrypted wiki, got nil")
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(notesDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Encrypt with one passphrase
	if _, err := EncryptWiki(tmpDir, "correct-passphrase"); err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Try to decrypt with wrong passphrase
	report, err := DecryptWiki(tmpDir, "wrong-passphrase")
	if err == nil {
		t.Error("Expected error when decrypting with wrong passphrase, got nil")
	}

	// Verify at least one file failed to decrypt
	if len(report.Failed) == 0 {
		t.Error("Expected failed decryptions with wrong passphrase")
	}
}

func TestEncryptedFileFormat(t *testing.T) {
	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(notesDir, "test.md")
	testContent := "test content for format check"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Encrypt
	if _, err := EncryptWiki(tmpDir, "test-passphrase"); err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Read encrypted file
	encFile := testFile + ".enc"
	encData, err := os.ReadFile(encFile)
	if err != nil {
		t.Fatalf("Failed to read encrypted file: %v", err)
	}

	// Verify format: magic header (10) + version (1) + nonce (12) + ciphertext (>0)
	if len(encData) < 23 {
		t.Errorf("Encrypted file too short: %d bytes", len(encData))
	}

	// Verify magic header
	if string(encData[:10]) != magicHeader {
		t.Errorf("Invalid magic header: %q", string(encData[:10]))
	}

	// Verify version
	if encData[10] != formatVersion {
		t.Errorf("Invalid format version: %d", encData[10])
	}
}

func TestAADPreventsFileSwapping(t *testing.T) {
	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Create two test files
	file1 := filepath.Join(notesDir, "test1.md")
	file2 := filepath.Join(notesDir, "test2.md")
	if err := os.WriteFile(file1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("Failed to write test file 1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("Failed to write test file 2: %v", err)
	}

	passphrase := "test-passphrase"

	// Encrypt
	if _, err := EncryptWiki(tmpDir, passphrase); err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Swap encrypted files
	enc1 := file1 + ".enc"
	enc2 := file2 + ".enc"
	data1, err := os.ReadFile(enc1)
	if err != nil {
		t.Fatalf("Failed to read enc1: %v", err)
	}
	data2, err := os.ReadFile(enc2)
	if err != nil {
		t.Fatalf("Failed to read enc2: %v", err)
	}

	// Swap the encrypted contents
	if err := os.WriteFile(enc1, data2, 0644); err != nil {
		t.Fatalf("Failed to write swapped enc1: %v", err)
	}
	if err := os.WriteFile(enc2, data1, 0644); err != nil {
		t.Fatalf("Failed to write swapped enc2: %v", err)
	}

	// Try to decrypt - should fail due to AAD mismatch
	report, err := DecryptWiki(tmpDir, passphrase)
	if err == nil {
		t.Error("Expected decryption to fail after file swapping, got nil error")
	}

	// Verify files failed to decrypt
	if len(report.Failed) == 0 {
		t.Error("Expected failed decryptions after file swapping")
	}
}

func TestEncryptSkipsGitDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	objectsDir := filepath.Join(gitDir, "objects")
	if err := os.MkdirAll(objectsDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create a .md file in .git directory
	gitFile := filepath.Join(objectsDir, "test.md")
	if err := os.WriteFile(gitFile, []byte("git content"), 0644); err != nil {
		t.Fatalf("Failed to write git file: %v", err)
	}

	// Create a regular .md file
	regularFile := filepath.Join(tmpDir, "regular.md")
	if err := os.WriteFile(regularFile, []byte("regular content"), 0644); err != nil {
		t.Fatalf("Failed to write regular file: %v", err)
	}

	// Encrypt
	report, err := EncryptWiki(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Verify only regular file was encrypted
	if len(report.Encrypted) != 1 {
		t.Errorf("Expected 1 encrypted file, got %d", len(report.Encrypted))
	}

	// Verify .git file was not encrypted
	if _, err := os.Stat(gitFile); os.IsNotExist(err) {
		t.Error(".git file should not be encrypted")
	}
	if _, err := os.Stat(gitFile + ".enc"); !os.IsNotExist(err) {
		t.Error(".git file should not have .enc version")
	}
}

func TestArgon2Parameters(t *testing.T) {
	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(notesDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Encrypt
	if _, err := EncryptWiki(tmpDir, "test-passphrase"); err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Read marker file
	markerPath := filepath.Join(tmpDir, ".encrypted")
	markerData, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("Failed to read marker: %v", err)
	}

	// Parse marker
	var marker Marker
	if err := json.Unmarshal(markerData, &marker); err != nil {
		t.Fatalf("Failed to parse marker: %v", err)
	}

	// Verify Argon2 parameters
	if marker.Argon2Time != argonTime {
		t.Errorf("Expected Argon2Time=%d, got %d", argonTime, marker.Argon2Time)
	}
	if marker.Argon2Memory != argonMemory {
		t.Errorf("Expected Argon2Memory=%d, got %d", argonMemory, marker.Argon2Memory)
	}
	if marker.Argon2Threads != argonThreads {
		t.Errorf("Expected Argon2Threads=%d, got %d", argonThreads, marker.Argon2Threads)
	}

	// Verify salt is valid hex
	if len(marker.Salt) != saltSize*2 { // hex encoding doubles length
		t.Errorf("Expected salt length %d, got %d", saltSize*2, len(marker.Salt))
	}
}
