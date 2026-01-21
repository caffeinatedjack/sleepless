package regimen

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/notes"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var (
	randomTag   string
	randomCount int
	reportDays  int
)

// random command
var noteRandomCmd = &cobra.Command{
	Use:   "random",
	Short: "Display random notes",
	Long: `Display one or more random notes.

Examples:
    regimen note random
    regimen note random --tag idea --count 3`,
	RunE: runNoteRandom,
}

// report command
var noteReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a report from recent notes",
	Long: `Generate a report from recent daily notes.

By default, includes entries from the past day tagged with #work or #progress.

Examples:
    regimen note report
    regimen note report --days 7`,
	RunE: runNoteReport,
}

// stats command
var noteStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display statistics",
	Long: `Display statistics about your notes.

Shows entry count, tag distribution, and note streaks.`,
	RunE: runNoteStats,
}

func init() {
	noteCmd.AddCommand(noteRandomCmd)
	noteCmd.AddCommand(noteReportCmd)
	noteCmd.AddCommand(noteStatsCmd)

	// random flags
	noteRandomCmd.Flags().StringVar(&randomTag, "tag", "", "Filter by tag")
	noteRandomCmd.Flags().IntVar(&randomCount, "count", 1, "Number of random notes")

	// report flags
	noteReportCmd.Flags().IntVar(&reportDays, "days", 1, "Number of days to include")
}

func runNoteRandom(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	// Get all notes
	var noteList []*notes.Note

	if randomTag != "" {
		// Filter by tag
		filtered, err := store.FindByTag(randomTag)
		if err != nil {
			return err
		}
		noteList = filtered
	} else {
		// Get all notes
		entries, err := os.ReadDir(store.NotesDir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			path := filepath.Join(store.NotesDir, entry.Name())
			note, err := store.LoadNote(path)
			if err != nil {
				continue
			}
			noteList = append(noteList, note)
		}
	}

	if len(noteList) == 0 {
		ui.PrintDim("No notes found")
		return nil
	}

	// Select random notes
	count := randomCount
	if count > len(noteList) {
		count = len(noteList)
	}

	// Shuffle and select
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(noteList), func(i, j int) {
		noteList[i], noteList[j] = noteList[j], noteList[i]
	})

	for i := 0; i < count; i++ {
		if i > 0 {
			fmt.Println("\n---")
		}
		displayNote(noteList[i])
	}

	return nil
}

func runNoteReport(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	now := time.Now()

	fmt.Printf("Report (Last %d day(s))\n\n", reportDays)

	for i := reportDays - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		note, err := store.LoadDaily(date)
		if err != nil {
			continue
		}

		// Filter by work tags
		hasWorkTag := note.HasTag("work") || note.HasTag("progress") || note.HasTag("meeting")
		if !hasWorkTag {
			continue
		}

		// Show day header
		dayName := now.AddDate(0, 0, -i).Format("Monday")
		if i == 0 {
			fmt.Printf("## Today (%s)\n\n", date)
		} else if i == 1 {
			fmt.Printf("## Yesterday (%s)\n\n", date)
		} else {
			fmt.Printf("## %s (%s)\n\n", dayName, date)
		}

		// Extract relevant sections
		lines := strings.Split(note.Body, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") {
				fmt.Printf("- %s\n", line)
			}
		}
		fmt.Println()
	}

	return nil
}

func runNoteStats(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	entries, err := os.ReadDir(store.NotesDir)
	if err != nil {
		return err
	}

	totalNotes := 0
	dailyCount := 0
	floatingCount := 0
	tagCounts := make(map[string]int)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(store.NotesDir, entry.Name())
		note, err := store.LoadNote(path)
		if err != nil {
			continue
		}

		totalNotes++
		if note.Type == notes.NoteTypeDaily {
			dailyCount++
		} else {
			floatingCount++
		}

		for _, tag := range note.Tags {
			tagCounts[tag]++
		}
	}

	fmt.Println("Note Statistics")
	fmt.Println("===============")
	fmt.Printf("Total notes: %d\n", totalNotes)
	fmt.Printf("Daily notes: %d\n", dailyCount)
	fmt.Printf("Floating notes: %d\n", floatingCount)
	fmt.Printf("Unique tags: %d\n", len(tagCounts))

	if len(tagCounts) > 0 {
		fmt.Println("\nTop tags:")
		// Get top 5 tags
		type tagCount struct {
			tag   string
			count int
		}
		var sorted []tagCount
		for tag, count := range tagCounts {
			sorted = append(sorted, tagCount{tag, count})
		}
		// Sort by count
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].count > sorted[i].count {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		limit := 5
		if len(sorted) < limit {
			limit = len(sorted)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("  %s (%d)\n", sorted[i].tag, sorted[i].count)
		}
	}

	return nil
}
