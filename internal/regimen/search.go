package regimen

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/task"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search goals by title, notes, or tags",
	Long: `Search goals by title, notes, or tags.

Examples:
    regimen goals search "report"
    regimen goals search "urgent" --topic work`,
	Args: cobra.ExactArgs(1),
	Run:  runSearch,
}

var (
	searchTopic  string
	searchStatus string
)

func init() {
	goalsCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVarP(&searchTopic, "topic", "t", "", "Limit search to topic")
	searchCmd.Flags().StringVarP(&searchStatus, "status", "s", "", "Filter by status (open, complete)")
}

func runSearch(cmd *cobra.Command, args []string) {
	query := strings.ToLower(args[0])

	allTasks, err := store.LoadTasks(searchTopic)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load goals: %v", err))
		return
	}

	// Apply status filter
	if searchStatus != "" {
		switch strings.ToLower(searchStatus) {
		case "open":
			allTasks = task.FilterOpen(allTasks)
		case "complete":
			allTasks = task.FilterComplete(allTasks)
		}
	}

	type match struct {
		task      *task.Task
		matchType string
	}
	var matches []match

	var searchTask func(t *task.Task)
	searchTask = func(t *task.Task) {
		if strings.Contains(strings.ToLower(t.Title), query) {
			matches = append(matches, match{t, "title"})
		} else {
			// Check tags
			foundTag := false
			for _, tag := range t.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					matches = append(matches, match{t, "tag"})
					foundTag = true
					break
				}
			}
			if !foundTag {
				// Check notes
				for _, note := range t.Notes {
					if strings.Contains(strings.ToLower(note), query) {
						matches = append(matches, match{t, "note"})
						break
					}
				}
			}
		}

		for _, subtask := range t.Subtasks {
			searchTask(subtask)
		}
	}

	for _, t := range allTasks {
		searchTask(t)
	}

	if len(matches) == 0 {
		ui.PrintDim(fmt.Sprintf("No tasks found matching '%s'", args[0]))
		fmt.Println()
		ui.PrintDim("Tips: Try partial words, check spelling, or search without filters")
		return
	}

	fmt.Println()
	fmt.Println(ui.BoldStyle.Render(fmt.Sprintf("Found %d result(s)", len(matches))))
	fmt.Println()

	for _, m := range matches {
		checkbox := ui.Checkbox(m.task.IsComplete())
		var style = ui.DimStyle
		if !m.task.IsComplete() {
			style = ui.PriorityStyle(m.task.Priority)
		}

		fmt.Printf("  %s %s %s %s %s\n",
			checkbox,
			style.Render(m.task.Title),
			ui.DimStyle.Render(m.task.ShortID()),
			ui.InfoStyle.Render("@"+m.task.Topic),
			ui.DimStyle.Render("("+m.matchType+")"))
	}
}
