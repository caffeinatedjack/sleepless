package regimen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/notes"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var (
	showDate      string
	listLimit     int
	listNoteTags  string
	listDate      string
	listRange     []string
	outputJSON    bool
	outputCompact bool
)

// today command
var noteTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's daily note",
	Long:  `Display today's daily note.`,
	RunE:  runNoteToday,
}

// show command
var noteShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Display a specific note",
	Long: `Display a specific note.

For floating notes: provide the ID or unique prefix.
For daily notes: use --date flag.

Examples:
    regimen note show abc123
    regimen note show --date 2026-01-20`,
	RunE: runNoteShow,
}

// list command
var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent notes",
	Long: `List recent notes with timestamps and previews.

Examples:
    regimen note list
    regimen note list --limit 10
    regimen note list --tags work,meeting`,
	RunE: runNoteList,
}

// week command
var noteWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show notes from the past 7 days",
	Long:  `Display a list of daily notes from the past 7 days with previews.`,
	RunE:  runNoteWeek,
}

// month command
var noteMonthCmd = &cobra.Command{
	Use:   "month",
	Short: "Show notes from the current month",
	Long:  `Display a list of daily notes from the current month with previews.`,
	RunE:  runNoteMonth,
}

func init() {
	noteCmd.AddCommand(noteTodayCmd)
	noteCmd.AddCommand(noteShowCmd)
	noteCmd.AddCommand(noteListCmd)
	noteCmd.AddCommand(noteWeekCmd)
	noteCmd.AddCommand(noteMonthCmd)

	// show flags
	noteShowCmd.Flags().StringVar(&showDate, "date", "", "Date for daily note (YYYY-MM-DD)")

	// list flags
	noteListCmd.Flags().IntVar(&listLimit, "limit", 20, "Maximum entries to show")
	noteListCmd.Flags().StringVar(&listNoteTags, "tags", "", "Filter by tags (comma-separated)")
	noteListCmd.Flags().StringVar(&listDate, "date", "", "Show specific date")
	noteListCmd.Flags().StringSliceVar(&listRange, "range", nil, "Date range (start end)")

	// output format flags
	noteShowCmd.Flags().BoolVar(&outputJSON, "json", false, "Output in JSON format")
	noteListCmd.Flags().BoolVar(&outputJSON, "json", false, "Output in JSON format")
	noteListCmd.Flags().BoolVar(&outputCompact, "compact", false, "Single-line summaries")
}

func runNoteToday(cmd *cobra.Command, args []string) error {
	today := time.Now().Format("2006-01-02")
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	note, err := store.LoadDaily(today)
	if err != nil {
		// Note doesn't exist yet
		ui.PrintDim(fmt.Sprintf("No entries for %s", today))
		return nil
	}

	displayNote(note)
	return nil
}

func runNoteShow(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	if showDate != "" {
		note, err := store.LoadDaily(showDate)
		if err != nil {
			return err
		}
		displayNote(note)
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("provide note ID or use --date flag")
	}

	idOrPrefix := args[0]
	note, err := store.LoadFloating(idOrPrefix)
	if err != nil {
		return err
	}

	displayNote(note)
	return nil
}

func runNoteList(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	// TODO: Implement proper filtering and listing
	// For now, just list files in notes directory

	entries, err := os.ReadDir(store.NotesDir)
	if err != nil {
		return err
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		if count >= listLimit {
			break
		}

		path := filepath.Join(store.NotesDir, entry.Name())
		displayFileSummary(path, entry.Name())
		count++
	}

	if count == 0 {
		ui.PrintDim("No notes found")
	}

	return nil
}

func runNoteWeek(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	now := time.Now()
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		showDailySummary(store, date)
	}

	return nil
}

func runNoteMonth(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	now := time.Now()
	year, month, _ := now.Date()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())

	for d := firstDay; d.Month() == month; d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		showDailySummary(store, date)
	}

	return nil
}

func displayNote(note *notes.Note) {
	if note.Type == notes.NoteTypeDaily {
		fmt.Printf("# Daily Note: %s\n\n", note.Date)
	} else {
		fmt.Printf("# Floating Note: %s\n\n", note.ID)
	}

	if len(note.Tags) > 0 {
		fmt.Printf("Tags: %s\n\n", strings.Join(note.Tags, ", "))
	}

	fmt.Println(note.Body)
}

func displayFileSummary(path, filename string) {
	base := strings.TrimSuffix(filename, ".md")

	// Try to read first line of body for preview
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("%s\n", ui.DimStyle.Render(base))
		return
	}

	// Extract preview (skip frontmatter, get first content line)
	lines := strings.Split(string(content), "\n")
	var preview string
	inFrontmatter := false
	fmCount := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			fmCount++
			if fmCount == 1 {
				inFrontmatter = true
			} else if fmCount == 2 {
				inFrontmatter = false
			}
			continue
		}

		if !inFrontmatter && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
			preview = line
			if len(preview) > 60 {
				preview = preview[:57] + "..."
			}
			break
		}
	}

	fmt.Printf("%s  %s\n", ui.DimStyle.Render(base), preview)
}

func showDailySummary(store *notes.Store, date string) {
	note, err := store.LoadDaily(date)
	if err != nil {
		return // Skip if doesn't exist
	}

	// Get preview from body
	preview := getPreview(note.Body)
	fmt.Printf("%s  %s\n", date, preview)
}

func getPreview(body string) string {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			if len(line) > 70 {
				return line[:67] + "..."
			}
			return line
		}
	}
	return ""
}
