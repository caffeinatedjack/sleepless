package doze

import (
	"os"
	"testing"

	"gitlab.com/caffeinatedjack/sleepless/pkg/config"
)

func setupTestConfig(t *testing.T) func() {
	tmpDir := t.TempDir()
	os.Setenv("DOZE_CONFIG_DIR", tmpDir)

	// Create initial config
	cfg := config.DefaultConfig()
	cfg.CreateProfile("work")
	cfg.CreateProfile("personal")
	cfg.Profiles["work"].Variables["AWS_PROFILE"] = "work-account"
	cfg.Profiles["personal"].Variables["EDITOR"] = "vim"
	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	return func() {
		os.Unsetenv("DOZE_CONFIG_DIR")
		// Reset global flags
		jsonOutput = false
	}
}

func TestProfileList(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Test that list runs without error
	err := runProfileList(profileListCmd, []string{})
	if err != nil {
		t.Fatalf("runProfileList() error = %v", err)
	}
}

func TestProfileListJSON(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	jsonOutput = true
	err := runProfileList(profileListCmd, []string{})
	if err != nil {
		t.Fatalf("runProfileList() with JSON error = %v", err)
	}
}

func TestProfileCurrent(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := runProfileCurrent(profileCurrentCmd, []string{})
	if err != nil {
		t.Fatalf("runProfileCurrent() error = %v", err)
	}
}

func TestProfileSwitch(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := runProfileSwitch(profileSwitchCmd, []string{"work"})
	if err != nil {
		t.Fatalf("runProfileSwitch() error = %v", err)
	}

	// Verify switch was persisted
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.CurrentProfile != "work" {
		t.Errorf("expected current profile 'work', got '%s'", cfg.CurrentProfile)
	}
}

func TestProfileSwitchNonExistent(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := runProfileSwitch(profileSwitchCmd, []string{"nonexistent"})
	if err == nil {
		t.Error("runProfileSwitch() should fail for non-existent profile")
	}
}

func TestProfileCreate(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := runProfileCreate(profileCreateCmd, []string{"staging"})
	if err != nil {
		t.Fatalf("runProfileCreate() error = %v", err)
	}

	// Verify profile was created
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if _, exists := cfg.Profiles["staging"]; !exists {
		t.Error("staging profile should exist after creation")
	}
}

func TestProfileCreateDuplicate(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := runProfileCreate(profileCreateCmd, []string{"work"})
	if err == nil {
		t.Error("runProfileCreate() should fail for duplicate profile")
	}
}

func TestProfileDelete(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := runProfileDelete(profileDeleteCmd, []string{"personal"})
	if err != nil {
		t.Fatalf("runProfileDelete() error = %v", err)
	}

	// Verify profile was deleted
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if _, exists := cfg.Profiles["personal"]; exists {
		t.Error("personal profile should not exist after deletion")
	}
}

func TestProfileDeleteDefault(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := runProfileDelete(profileDeleteCmd, []string{"default"})
	if err == nil {
		t.Error("runProfileDelete() should fail for default profile")
	}
}

func TestProfileDeleteCurrent(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Switch to work profile first
	runProfileSwitch(profileSwitchCmd, []string{"work"})

	err := runProfileDelete(profileDeleteCmd, []string{"work"})
	if err == nil {
		t.Error("runProfileDelete() should fail for current profile")
	}
}

func TestProfileCopy(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := runProfileCopy(profileCopyCmd, []string{"work", "work-copy"})
	if err != nil {
		t.Fatalf("runProfileCopy() error = %v", err)
	}

	// Verify profile was copied
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	copied, exists := cfg.Profiles["work-copy"]
	if !exists {
		t.Fatal("work-copy profile should exist after copy")
	}

	if copied.Variables["AWS_PROFILE"] != "work-account" {
		t.Error("variables should be copied from source profile")
	}
}

func TestProfileCopyNonExistentSource(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := runProfileCopy(profileCopyCmd, []string{"nonexistent", "dest"})
	if err == nil {
		t.Error("runProfileCopy() should fail for non-existent source")
	}
}

func TestProfileCopyDuplicateDest(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := runProfileCopy(profileCopyCmd, []string{"work", "personal"})
	if err == nil {
		t.Error("runProfileCopy() should fail for existing destination")
	}
}
