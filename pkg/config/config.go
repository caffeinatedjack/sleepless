// Package config handles doze configuration file loading and saving.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the doze configuration structure.
type Config struct {
	Version        int                `json:"version"`
	CurrentProfile string             `json:"current_profile"`
	Profiles       map[string]Profile `json:"profiles"`
	Dotfiles       []DotfileRepo      `json:"dotfiles,omitempty"`
}

// Profile represents a named set of environment variables.
type Profile struct {
	Variables map[string]string `json:"variables"`
}

// DotfileRepo represents a dotfile repository configuration.
type DotfileRepo struct {
	Repo     string     `json:"repo"`
	Links    []LinkSpec `json:"links,omitempty"`
	LinkedAt string     `json:"linked_at,omitempty"`
}

// LinkSpec represents a single symlink specification.
type LinkSpec struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// GetConfigDir returns the configuration directory path.
// It checks DOZE_CONFIG_DIR environment variable, otherwise uses ~/.config/doze.
func GetConfigDir() (string, error) {
	if dir := os.Getenv("DOZE_CONFIG_DIR"); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}

	return filepath.Join(home, ".config", "doze"), nil
}

// GetConfigPath returns the full path to the config.json file.
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the configuration file from disk.
func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the loaded configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfig returns a new configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Version:        1,
		CurrentProfile: "default",
		Profiles: map[string]Profile{
			"default": {
				Variables: make(map[string]string),
			},
		},
		Dotfiles: []DotfileRepo{},
	}
}

// GetCurrentProfile returns the current active profile.
func (c *Config) GetCurrentProfile() (*Profile, error) {
	profile, exists := c.Profiles[c.CurrentProfile]
	if !exists {
		return nil, fmt.Errorf("current profile '%s' does not exist", c.CurrentProfile)
	}
	return &profile, nil
}

// SetCurrentProfile changes the active profile.
func (c *Config) SetCurrentProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}
	c.CurrentProfile = name
	return nil
}

// CreateProfile creates a new empty profile.
func (c *Config) CreateProfile(name string) error {
	if _, exists := c.Profiles[name]; exists {
		return fmt.Errorf("profile '%s' already exists", name)
	}
	c.Profiles[name] = Profile{
		Variables: make(map[string]string),
	}
	return nil
}

// DeleteProfile removes a profile.
func (c *Config) DeleteProfile(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete the default profile")
	}
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}
	if c.CurrentProfile == name {
		return fmt.Errorf("cannot delete the currently active profile")
	}
	delete(c.Profiles, name)
	return nil
}

// CopyProfile duplicates a profile to a new name.
func (c *Config) CopyProfile(source, dest string) error {
	srcProfile, exists := c.Profiles[source]
	if !exists {
		return fmt.Errorf("source profile '%s' does not exist", source)
	}
	if _, exists := c.Profiles[dest]; exists {
		return fmt.Errorf("destination profile '%s' already exists", dest)
	}

	// Deep copy variables
	newVars := make(map[string]string, len(srcProfile.Variables))
	for k, v := range srcProfile.Variables {
		newVars[k] = v
	}

	c.Profiles[dest] = Profile{
		Variables: newVars,
	}
	return nil
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version: %d", c.Version)
	}

	if c.CurrentProfile == "" {
		return fmt.Errorf("current_profile must be set")
	}

	if len(c.Profiles) == 0 {
		return fmt.Errorf("at least one profile must be defined")
	}

	if _, exists := c.Profiles[c.CurrentProfile]; !exists {
		return fmt.Errorf("current profile '%s' does not exist in profiles map", c.CurrentProfile)
	}

	// Validate profile names are not empty
	for name := range c.Profiles {
		if name == "" {
			return fmt.Errorf("profile name cannot be empty")
		}
	}

	return nil
}
