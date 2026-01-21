package regimen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/notes"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var (
	searchOR     bool
	searchTags   string
	searchAfter  string
	searchBefore string
)

// search command
var noteSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search notes",
	Long: `Search notes by text and tags.

Examples:
    regimen note search "API design"
    regimen note search "authentication" --tags work
    regimen note search "meeting" --or`,
	Args: cobra.MinimumNArgs(1),
	RunE: runNoteSearch,
}

// tags command (list all)
var noteTagsCmd = &cobra.Command{
	Use:   "tags [tag]",
	Short: "List tags or show notes with specific tag",
	Long: `List all tags with counts, or show notes with a specific tag.

Examples:
    regimen note tags
    regimen note tags work`,
	RunE: runNoteTags,
}

// tag command (add tags)
var noteTagAddCmd = &cobra.Command{
	Use:   "tag <id> <tags>",
	Short: "Add tags to a note",
	Long: `Add tags to an existing note.

Examples:
    regimen note tag abc123 work,important`,
	Args: cobra.ExactArgs(2),
	RunE: runNoteTagAdd,
}

// untag command (remove tags)
var noteUntagCmd = &cobra.Command{
	Use:   "untag <id> <tags>",
	Short: "Remove tags from a note",
	Long: `Remove tags from an existing note.

Examples:
    regimen note untag abc123 obsolete`,
	Args: cobra.ExactArgs(2),
	RunE: runNoteUntag,
}

func init() {
	noteCmd.AddCommand(noteSearchCmd)
	noteCmd.AddCommand(noteTagsCmd)
	noteCmd.AddCommand(noteTagAddCmd)
	noteCmd.AddCommand(noteUntagCmd)

	// search flags
	noteSearchCmd.Flags().BoolVar(&searchOR, "or", false, "Use OR logic for search terms")
	noteSearchCmd.Flags().StringVar(&searchTags, "tags", "", "Filter by tags")
	noteSearchCmd.Flags().StringVar(&searchAfter, "after", "", "Filter by date (after)")
	noteSearchCmd.Flags().StringVar(&searchBefore, "before", "", "Filter by date (before)")
}

func runNoteSearch(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	query := strings.Join(args, " ")

	// Parse tag filter
	var tagList []string
	if searchTags != "" {
		for _, tag := range strings.Split(searchTags, ",") {
			tag = strings.TrimSpace(strings.ToLower(tag))
			if tag != "" {
				tagList = append(tagList, tag)
			}
		}
	}

	results, err := store.Search(query, tagList, searchOR)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		ui.PrintDim("No matches found")
		return nil
	}

	for _, result := range results {
		var id string
		if result.Note.Type == notes.NoteTypeDaily {
			id = result.Note.Date
		} else {
			id = result.Note.ID[:6] // Show short ID
		}

		fmt.Printf("%s  %s\n", ui.DimStyle.Render(id), result.Context)
	}

	fmt.Printf("\n%d result(s)\n", len(results))
	return nil
}

func runNoteTags(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	if len(args) == 0 {
		// List all tags
		tagCounts, err := store.ListAllTags()
		if err != nil {
			return err
		}

		if len(tagCounts) == 0 {
			ui.PrintDim("No tags found")
			return nil
		}

		// Sort by count descending
		type tagCount struct {
			tag   string
			count int
		}
		var sorted []tagCount
		for tag, count := range tagCounts {
			sorted = append(sorted, tagCount{tag, count})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].count > sorted[j].count
		})

		for _, tc := range sorted {
			fmt.Printf("%s  %s\n", tc.tag, ui.DimStyle.Render(fmt.Sprintf("(%d)", tc.count)))
		}

		return nil
	}

	// Show notes with specific tag
	tag := strings.ToLower(args[0])
	noteList, err := store.FindByTag(tag)
	if err != nil {
		return err
	}

	if len(noteList) == 0 {
		ui.PrintDim(fmt.Sprintf("No notes with tag %q", tag))
		return nil
	}

	for _, note := range noteList {
		var id string
		if note.Type == notes.NoteTypeDaily {
			id = note.Date
		} else {
			id = note.ID[:6]
		}

		preview := getPreview(note.Body)
		fmt.Printf("%s  %s\n", ui.DimStyle.Render(id), preview)
	}

	return nil
}

func runNoteTagAdd(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	idOrPrefix := args[0]
	tagStr := args[1]

	// Parse tags
	var tagList []string
	for _, tag := range strings.Split(tagStr, ",") {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag != "" {
			tagList = append(tagList, tag)
		}
	}

	// Load note
	note, err := store.LoadFloating(idOrPrefix)
	if err != nil {
		return err
	}

	// Add tags
	note.AddTags(tagList...)

	// Save
	if err := store.SaveFloating(note); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("Added tags to note %s", note.ID[:6]))
	return nil
}

func runNoteUntag(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	idOrPrefix := args[0]
	tagStr := args[1]

	// Parse tags
	var tagList []string
	for _, tag := range strings.Split(tagStr, ",") {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag != "" {
			tagList = append(tagList, tag)
		}
	}

	// Load note
	note, err := store.LoadFloating(idOrPrefix)
	if err != nil {
		return err
	}

	// Remove tags
	note.RemoveTags(tagList...)

	// Save
	if err := store.SaveFloating(note); err != nil {
		return err
	}

	ui.Success(fmt.Sprintf("Removed tags from note %s", note.ID[:6]))
	return nil
}
