package regimen

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/task"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List goals with optional filters",
	Long: `List goals with optional filters.

Examples:
    regimen goals list
    regimen goals list --topic work --priority high
    regimen goals list --overdue`,
	Run: runList,
}

var (
	listTopic    string
	listPriority string
	listStatus   string
	listOverdue  bool
	listTags     string
)

func init() {
	goalsCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&listTopic, "topic", "t", "", "Filter by topic")
	listCmd.Flags().StringVarP(&listPriority, "priority", "p", "", "Filter by priority (low, medium, high)")
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status (open, complete)")
	listCmd.Flags().BoolVar(&listOverdue, "overdue", false, "Show only overdue goals")
	listCmd.Flags().StringVar(&listTags, "tags", "", "Filter by tag")
}

func runList(cmd *cobra.Command, args []string) {
	allTasks, err := store.LoadTasks(listTopic)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load goals: %v", err))
		return
	}

	// Apply filters
	filtered := allTasks
	if listPriority != "" {
		filtered = task.FilterByPriority(filtered, task.ParsePriority(listPriority))
	}
	if listStatus != "" {
		if strings.ToLower(listStatus) == "open" {
			filtered = task.FilterOpen(filtered)
		} else if strings.ToLower(listStatus) == "complete" {
			filtered = task.FilterComplete(filtered)
		}
	}
	if listOverdue {
		filtered = task.FilterOverdue(filtered)
	}
	if listTags != "" {
		filtered = filterByTags(filtered, listTags)
	}

	if len(filtered) == 0 {
		ui.PrintDim("No goals found")
		return
	}

	// Group by topic
	byTopic, topics := task.GroupByTopic(filtered)

	for _, topic := range topics {
		topicTasks := byTopic[topic]
		fmt.Println()
		fmt.Println(ui.TopicStyle.Render(ui.TitleCase(topic)))

		for _, t := range topicTasks {
			printTask(t, 0)
		}
	}

	showInboxAgingWarning()
}

// filterByTags filters tasks that have any of the given comma-separated tags.
func filterByTags(tasks []*task.Task, tagFilter string) []*task.Task {
	tagList := strings.Split(strings.ToLower(tagFilter), ",")
	for i := range tagList {
		tagList[i] = strings.TrimSpace(tagList[i])
	}

	var filtered []*task.Task
	for _, t := range tasks {
		for _, filterTag := range tagList {
			for _, taskTag := range t.Tags {
				if strings.EqualFold(taskTag, filterTag) {
					filtered = append(filtered, t)
					goto nextTask
				}
			}
		}
	nextTask:
	}
	return filtered
}

func printTask(t *task.Task, indent int) {
	fmt.Println(ui.FormatTask(t, indent))

	// Show subtask summary (collapsed)
	complete, total := t.CountSubtasks()
	if total > 0 {
		prefix := strings.Repeat("  ", indent+2)
		fmt.Printf("%s%s\n", prefix, ui.FormatSubtaskSummary(complete, total))
	}
}
