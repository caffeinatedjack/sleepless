package regimen

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/notes"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var (
	addDate     string
	addNoteTags string
	addFloating bool
	addDaily    bool
	addTemplate string
)

var noteAddCmd = &cobra.Command{
	Use:   "add [text]",
	Short: "Add a new note",
	Long: `Add a new note.

With text argument: appends entry to daily note with timestamp heading.
Without text: opens editor for detailed entry.

Examples:
    regimen note add "Had an idea for improving the login flow"
    regimen note add "Sprint planning meeting" --tags "meeting,work"
    regimen note add --floating "Research: Distributed caching"
    regimen note add --date 2026-01-20
    regimen note add`,
	RunE: runNoteAdd,
}

func init() {
	noteCmd.AddCommand(noteAddCmd)
	noteAddCmd.Flags().StringVar(&addDate, "date", "", "Date for daily note (YYYY-MM-DD)")
	noteAddCmd.Flags().StringVar(&addNoteTags, "tags", "", "Comma-separated tags")
	noteAddCmd.Flags().BoolVar(&addFloating, "floating", false, "Create floating note")
	noteAddCmd.Flags().BoolVar(&addDaily, "daily", false, "Create daily note (default)")
	noteAddCmd.Flags().StringVar(&addTemplate, "template", "", "Use template")
}

func runNoteAdd(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	if err := store.EnsureStructure(); err != nil {
		return err
	}

	// Ensure built-in templates exist
	if err := store.EnsureBuiltInTemplates(); err != nil {
		return fmt.Errorf("failed to ensure templates: %w", err)
	}

	// Handle template mode
	if addTemplate != "" {
		return runAddFromTemplate(store, args)
	}

	// Parse tags
	var tagList []string
	if addNoteTags != "" {
		for _, tag := range strings.Split(addNoteTags, ",") {
			tag = strings.TrimSpace(strings.ToLower(tag))
			if tag != "" {
				tagList = append(tagList, tag)
			}
		}
	}

	// Determine target date for daily notes
	targetDate := addDate
	if targetDate == "" && !addFloating {
		targetDate = time.Now().Format("2006-01-02")
	}

	if addFloating {
		return runAddFloating(store, args, tagList)
	}

	return runAddDaily(store, targetDate, args, tagList)
}

func runAddDaily(store *notes.Store, date string, args []string, tags []string) error {
	note, err := store.GetOrCreateDaily(date)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		// Editor mode
		return openEditorForDaily(store, note)
	}

	// Inline text mode - append with timestamp
	text := strings.Join(args, " ")
	notes.AppendDailyEntry(note, text, tags)

	if err := store.SaveDaily(note); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("Added entry to %s", date))
	return nil
}

func runAddFloating(store *notes.Store, args []string, tags []string) error {
	text := strings.Join(args, " ")

	note, err := store.CreateFloating("", text, tags)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		// Editor mode
		return openEditorForFloating(store, note)
	}

	if err := store.SaveFloating(note); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("Created floating note %s", note.ID))
	return nil
}

func openEditorForDaily(store *notes.Store, note *notes.Note) error {
	path, _ := store.DailyPath(note.Date)
	return openEditor(path)
}

func openEditorForFloating(store *notes.Store, note *notes.Note) error {
	// Save first to create file
	if err := store.SaveFloating(note); err != nil {
		return err
	}

	path, _ := store.FloatingPath(note.ID)
	if err := openEditor(path); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("Created floating note %s", note.ID))
	return nil
}

func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Default to vim
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func runAddFromTemplate(store *notes.Store, args []string) error {
	// Get template
	tmpl, err := store.GetTemplate(addTemplate)
	if err != nil {
		return err
	}

	// Apply template with prompts
	reader := bufio.NewReader(os.Stdin)
	content, err := notes.ApplyTemplate(tmpl.Content, reader)
	if err != nil {
		return fmt.Errorf("failed to apply template: %w", err)
	}

	// Determine if this should be a daily or floating note
	// If --date or --daily is specified, create daily note
	// Otherwise create floating note
	if addDate != "" || addDaily {
		targetDate := addDate
		if targetDate == "" {
			targetDate = time.Now().Format("2006-01-02")
		}

		note, err := store.GetOrCreateDaily(targetDate)
		if err != nil {
			return err
		}

		// Parse tags from --tags flag
		var tagList []string
		if addNoteTags != "" {
			for _, tag := range strings.Split(addNoteTags, ",") {
				tag = strings.TrimSpace(strings.ToLower(tag))
				if tag != "" {
					tagList = append(tagList, tag)
				}
			}
		}

		// Append content to daily note
		notes.AppendDailyEntry(note, content, tagList)

		if err := store.SaveDaily(note); err != nil {
			return err
		}

		ui.Success(fmt.Sprintf("Added templated entry to %s", targetDate))
		return nil
	}

	// Create floating note
	var tagList []string
	if addNoteTags != "" {
		for _, tag := range strings.Split(addNoteTags, ",") {
			tag = strings.TrimSpace(strings.ToLower(tag))
			if tag != "" {
				tagList = append(tagList, tag)
			}
		}
	}

	note, err := store.CreateFloating("", content, tagList)
	if err != nil {
		return err
	}

	if err := store.SaveFloating(note); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("Created floating note %s from template %q", note.ID, addTemplate))
	return nil
}
