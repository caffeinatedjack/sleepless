// Package doze implements the CLI commands for doze.
package doze

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
	Use:   "doze",
	Short: "Configuration and dotfile management",
	Long: `Doze - Configuration and dotfile management.

Manage dotfiles, environment variables, and configuration profiles from the terminal.

Examples:
    doze dotfiles link ~/dotfiles
    doze env set EDITOR=nvim
    doze profile switch work
    doze env export --shell bash`,
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
