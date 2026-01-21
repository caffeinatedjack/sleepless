package regimen

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/notes"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var (
	editDate   string
	deleteDate string
)

var noteEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit an existing note",
	Long: `Edit an existing note.

For floating notes: provide the ID or unique prefix.
For daily notes: use --date flag.

Examples:
    regimen note edit abc123
    regimen note edit --date 2026-01-20`,
	RunE: runNoteEdit,
}

var noteDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a note",
	Long: `Delete a note.

For floating notes: provide the ID or unique prefix.
For daily notes: use --date flag.

Examples:
    regimen note delete abc123
    regimen note delete --date 2026-01-20`,
	RunE: runNoteDelete,
}

func init() {
	noteCmd.AddCommand(noteEditCmd)
	noteCmd.AddCommand(noteDeleteCmd)

	noteEditCmd.Flags().StringVar(&editDate, "date", "", "Date for daily note (YYYY-MM-DD)")
	noteDeleteCmd.Flags().StringVar(&deleteDate, "date", "", "Date for daily note (YYYY-MM-DD)")
}

func runNoteEdit(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	if editDate != "" {
		// Edit daily note
		path, err := store.DailyPath(editDate)
		if err != nil {
			return err
		}
		return openEditor(path)
	}

	if len(args) == 0 {
		return fmt.Errorf("provide note ID or use --date flag")
	}

	// Edit floating note
	idOrPrefix := args[0]
	note, err := store.LoadFloating(idOrPrefix)
	if err != nil {
		return err
	}

	path, _ := store.FloatingPath(note.ID)
	return openEditor(path)
}

func runNoteDelete(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	if deleteDate != "" {
		// Delete daily note
		if err := store.DeleteDaily(deleteDate); err != nil {
			return err
		}
		ui.Success(fmt.Sprintf("Deleted daily note for %s", deleteDate))
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("provide note ID or use --date flag")
	}

	// Delete floating note
	idOrPrefix := args[0]
	note, err := store.LoadFloating(idOrPrefix)
	if err != nil {
		return err
	}

	if err := store.DeleteFloating(note.ID); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("Deleted floating note %s", note.ID))
	return nil
}
