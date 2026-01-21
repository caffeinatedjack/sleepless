package regimen

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCommandsBlockWhenEncrypted verifies that commands fail when wiki is encrypted.
func TestCommandsBlockWhenEncrypted(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .encrypted marker
	markerPath := filepath.Join(tmpDir, ".encrypted")
	markerContent := `{"salt":"abc123","argon2_time":3,"argon2_memory":131072,"argon2_threads":4,"encrypted_files":0}`
	if err := os.WriteFile(markerPath, []byte(markerContent), 0644); err != nil {
		t.Fatalf("Failed to create .encrypted marker: %v", err)
	}

	// Test checkWikiEncrypted function
	err := checkWikiEncrypted(tmpDir)
	if err == nil {
		t.Error("Expected checkWikiEncrypted to return error for encrypted wiki")
	}

	if err != ErrWikiEncrypted {
		t.Errorf("Expected ErrWikiEncrypted, got: %v", err)
	}
}

// TestCommandsAllowedWhenNotEncrypted verifies that commands work when wiki is not encrypted.
func TestCommandsAllowedWhenNotEncrypted(t *testing.T) {
	tmpDir := t.TempDir()

	// Test checkWikiEncrypted function (no marker file)
	err := checkWikiEncrypted(tmpDir)
	if err != nil {
		t.Errorf("Expected no error for non-encrypted wiki, got: %v", err)
	}
}
