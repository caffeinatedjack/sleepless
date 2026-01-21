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

	// Global wiki directory flag
	wikiDir string
)

const (
	defaultWikiDir = "~/wiki"
	envWikiDir     = "REGIMEN_WIKI_DIR"
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

	// Global persistent flag for wiki directory
	rootCmd.PersistentFlags().StringVar(&wikiDir, "wiki-dir", "",
		fmt.Sprintf("Wiki directory (default %s, env: %s)", defaultWikiDir, envWikiDir))
}

// getWikiDir returns the wiki directory to use.
// Priority: --wiki-dir flag > REGIMEN_WIKI_DIR env > default ~/wiki
func getWikiDir() string {
	if wikiDir != "" {
		return expandPath(wikiDir)
	}
	if env := os.Getenv(envWikiDir); env != "" {
		return expandPath(env)
	}
	return expandPath(defaultWikiDir)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
