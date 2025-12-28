// Package regimen implements the CLI commands for regimen.
package regimen

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
	Use:   "regimen",
	Short: "Goal management for your Vim wiki",
	Long: `Regimen - Goal management for your Vim wiki.

Manage goals directly from the terminal with full CRUD operations,
visualization, and organization by topic.

Examples:
    regimen goals add "Buy groceries"
    regimen goals add "Finish report" --topic work --priority high --due 2025-01-15
    regimen goals list
    regimen goals done a1b2c3
    regimen goals view tree`,
}

func init() {
	rootCmd.Version = fmt.Sprintf("%s (built %s)", Version, BuildTime)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
