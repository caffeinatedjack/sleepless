// Package when provides timezone utilities for world clock and time conversion.
package when

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the when command configuration.
type Config struct {
	// Aliases maps user-defined names to IANA zone strings.
	Aliases map[string]string `json:"aliases"`
	// Configured is the list of zones to display in the world clock.
	Configured []string `json:"configured"`
}

// NewConfig creates an empty configuration.
func NewConfig() *Config {
	return &Config{
		Aliases:    make(map[string]string),
		Configured: []string{},
	}
}

// configPath returns the path to the config file.
func configPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}

	return filepath.Join(configDir, "regimen", "when.json"), nil
}

// LoadConfig loads the configuration from disk.
// Returns an empty config if the file doesn't exist.
func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Ensure maps are initialized
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}
	if cfg.Configured == nil {
		cfg.Configured = []string{}
	}

	return &cfg, nil
}

// Save writes the configuration to disk with user-only permissions.
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with user-only permissions (0600)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddAlias adds or updates an alias and adds the zone to configured zones.
func (c *Config) AddAlias(alias, zone string) {
	c.Aliases[alias] = zone

	// Add to configured if not already present
	for _, z := range c.Configured {
		if z == alias {
			return
		}
	}
	c.Configured = append(c.Configured, alias)
}

// RemoveAlias removes an alias and its entry from configured zones.
func (c *Config) RemoveAlias(alias string) bool {
	if _, ok := c.Aliases[alias]; !ok {
		return false
	}

	delete(c.Aliases, alias)

	// Remove from configured
	for i, z := range c.Configured {
		if z == alias {
			c.Configured = append(c.Configured[:i], c.Configured[i+1:]...)
			break
		}
	}

	return true
}
