// Package nightwatch implements the CLI commands for the nightwatch security tool.
package nightwatch

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags
	Version   = "dev"
	BuildTime = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "nightwatch",
	Short: "Security utilities",
	Long: `Nightwatch - Security utilities.

Security-focused tools for development and operations:
  - PII/secret redaction
  - Password and passphrase generation
  - JWT token operations
  - Fake data generation for testing

Examples:
    nightwatch redact "Contact john@example.com"
    nightwatch password generate --length 24
    nightwatch password phrase --words 6
    nightwatch jwt decode <token>
    nightwatch fake email --count 5`,
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s (built %s)", Version, BuildTime)
}

// Execute is the entry point for the nightwatch CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
