package regimen

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage notes in your wiki",
	Long: `Manage notes in your wiki.

Commands:
    add           Add a new note or entry
    edit          Edit an existing note
    delete        Delete a note
    today         Show today's daily note
    week          Show notes from the past 7 days
    month         Show notes from the current month
    show          Display a specific note
    list          List recent notes
    search        Search notes by text and tags
    tags          Manage and view tags
    random        Display random notes
    report        Generate a report from recent notes
    stats         Display statistics
    template      Manage note templates

Examples:
    regimen note add "Had an idea for improving the login flow"
    regimen note add "Sprint planning meeting" --tags "meeting,work"
    regimen note add --floating "Research: Distributed caching"
    regimen note today
    regimen note search "API design"`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if wiki is encrypted
		wikiDir := getWikiDir()
		if err := checkWikiEncrypted(wikiDir); err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		// Ensure notes directory exists
		notesDir := filepath.Join(wikiDir, "notes")
		if err := os.MkdirAll(notesDir, 0755); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(noteCmd)
}
