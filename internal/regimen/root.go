// Package regimen implements the CLI commands for regimen.
package regimen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

	encryptedMarkerFile = ".encrypted"
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

// ErrWikiEncrypted is returned when attempting to access an encrypted wiki.
var ErrWikiEncrypted = errors.New("wiki is encrypted. Run 'regimen decrypt' first")

// checkWikiEncrypted returns an error if the wiki directory is encrypted.
// Commands that read/write wiki data MUST call this function.
func checkWikiEncrypted(wikiDir string) error {
	markerPath := filepath.Join(wikiDir, encryptedMarkerFile)
	if _, err := os.Stat(markerPath); err == nil {
		return ErrWikiEncrypted
	}
	return nil
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
