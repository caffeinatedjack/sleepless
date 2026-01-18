package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}

	if cfg.CurrentProfile != "default" {
		t.Errorf("expected current_profile 'default', got '%s'", cfg.CurrentProfile)
	}

	if len(cfg.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(cfg.Profiles))
	}

	if _, exists := cfg.Profiles["default"]; !exists {
		t.Error("default profile should exist")
	}

	if cfg.Dotfiles == nil {
		t.Error("dotfiles should be initialized")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid version",
			cfg: &Config{
				Version:        99,
				CurrentProfile: "default",
				Profiles: map[string]Profile{
					"default": {Variables: make(map[string]string)},
				},
			},
			wantErr: true,
		},
		{
			name: "empty current profile",
			cfg: &Config{
				Version:        1,
				CurrentProfile: "",
				Profiles: map[string]Profile{
					"default": {Variables: make(map[string]string)},
				},
			},
			wantErr: true,
		},
		{
			name: "no profiles",
			cfg: &Config{
				Version:        1,
				CurrentProfile: "default",
				Profiles:       map[string]Profile{},
			},
			wantErr: true,
		},
		{
			name: "current profile does not exist",
			cfg: &Config{
				Version:        1,
				CurrentProfile: "nonexistent",
				Profiles: map[string]Profile{
					"default": {Variables: make(map[string]string)},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateProfile(t *testing.T) {
	cfg := DefaultConfig()

	err := cfg.CreateProfile("work")
	if err != nil {
		t.Fatalf("CreateProfile() error = %v", err)
	}

	if _, exists := cfg.Profiles["work"]; !exists {
		t.Error("work profile should exist after creation")
	}

	// Try to create duplicate
	err = cfg.CreateProfile("work")
	if err == nil {
		t.Error("CreateProfile() should fail for duplicate profile")
	}
}

func TestDeleteProfile(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CreateProfile("temp")

	// Should not delete default
	err := cfg.DeleteProfile("default")
	if err == nil {
		t.Error("DeleteProfile() should fail for default profile")
	}

	// Should not delete current profile
	cfg.CurrentProfile = "temp"
	err = cfg.DeleteProfile("temp")
	if err == nil {
		t.Error("DeleteProfile() should fail for current profile")
	}

	// Should delete non-current profile
	cfg.CurrentProfile = "default"
	err = cfg.DeleteProfile("temp")
	if err != nil {
		t.Errorf("DeleteProfile() error = %v", err)
	}

	if _, exists := cfg.Profiles["temp"]; exists {
		t.Error("temp profile should not exist after deletion")
	}
}

func TestCopyProfile(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Profiles["source"] = Profile{
		Variables: map[string]string{
			"VAR1": "value1",
			"VAR2": "value2",
		},
	}

	err := cfg.CopyProfile("source", "dest")
	if err != nil {
		t.Fatalf("CopyProfile() error = %v", err)
	}

	dest, exists := cfg.Profiles["dest"]
	if !exists {
		t.Fatal("dest profile should exist after copy")
	}

	if dest.Variables["VAR1"] != "value1" {
		t.Error("variables not copied correctly")
	}

	// Modify dest to ensure deep copy
	dest.Variables["VAR1"] = "modified"
	cfg.Profiles["dest"] = dest

	if cfg.Profiles["source"].Variables["VAR1"] != "value1" {
		t.Error("source profile was affected by dest modification")
	}
}

func TestSetCurrentProfile(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CreateProfile("work")

	err := cfg.SetCurrentProfile("work")
	if err != nil {
		t.Fatalf("SetCurrentProfile() error = %v", err)
	}

	if cfg.CurrentProfile != "work" {
		t.Errorf("expected current_profile 'work', got '%s'", cfg.CurrentProfile)
	}

	// Try to set non-existent profile
	err = cfg.SetCurrentProfile("nonexistent")
	if err == nil {
		t.Error("SetCurrentProfile() should fail for non-existent profile")
	}
}

func TestGetCurrentProfile(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Profiles["default"].Variables["TEST"] = "value"

	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		t.Fatalf("GetCurrentProfile() error = %v", err)
	}

	if profile.Variables["TEST"] != "value" {
		t.Error("GetCurrentProfile() returned wrong profile")
	}

	// Test with invalid current profile
	cfg.CurrentProfile = "nonexistent"
	_, err = cfg.GetCurrentProfile()
	if err == nil {
		t.Error("GetCurrentProfile() should fail for non-existent profile")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	os.Setenv("DOZE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("DOZE_CONFIG_DIR")

	// Create and save config
	cfg := DefaultConfig()
	cfg.CreateProfile("work")
	cfg.Profiles["work"].Variables["TEST"] = "value"
	cfg.CurrentProfile = "work"

	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Check file exists
	configPath := filepath.Join(tmpDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load config
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.CurrentProfile != "work" {
		t.Errorf("expected current_profile 'work', got '%s'", loaded.CurrentProfile)
	}

	if loaded.Profiles["work"].Variables["TEST"] != "value" {
		t.Error("variables not loaded correctly")
	}
}

func TestLoadNonExistentConfig(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	os.Setenv("DOZE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("DOZE_CONFIG_DIR")

	// Load should return default config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CurrentProfile != "default" {
		t.Errorf("expected default config, got current_profile '%s'", cfg.CurrentProfile)
	}
}
