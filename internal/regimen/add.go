package regimen

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/task"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new goal",
	Long: `Add a new goal.

Examples:
    regimen goals add "Buy groceries"
    regimen goals add "Finish report" --topic work --priority high
    regimen goals add "Review section 1" --parent a1b2c3`,
	Args: cobra.ExactArgs(1),
	Run:  runAdd,
}

var (
	addTopic    string
	addPriority string
	addDue      string
	addTags     string
	addParent   string
)

func init() {
	goalsCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&addTopic, "topic", "t", "inbox", "Topic to add goal to")
	addCmd.Flags().StringVarP(&addPriority, "priority", "p", "medium", "Goal priority (low, medium, high)")
	addCmd.Flags().StringVarP(&addDue, "due", "d", "", "Due date (YYYY-MM-DD)")
	addCmd.Flags().StringVar(&addTags, "tags", "", "Comma-separated tags")
	addCmd.Flags().StringVar(&addParent, "parent", "", "Parent goal ID for subtask")
}

func runAdd(cmd *cobra.Command, args []string) {
	title := args[0]

	t := task.New(title)
	t.Topic = addTopic
	t.Priority = task.ParsePriority(addPriority)

	// Set due date
	if addDue != "" {
		due, err := time.Parse("2006-01-02", addDue)
		if err != nil {
			ui.Error(fmt.Sprintf("Invalid date format: %s (use YYYY-MM-DD)", addDue))
			return
		}
		t.Due = &due
	}

	// Set tags
	if tags := task.ParseTags(addTags); tags != nil {
		t.Tags = tags
	}

	// Handle parent goal (subgoal)
	if addParent != "" {
		parentTask, err := resolveTask(addParent)
		if err != nil {
			ui.Error(err.Error())
			return
		}
		if parentTask == nil {
			return
		}
		parentTask.AddSubtask(t)
		if err := store.SaveTask(parentTask); err != nil {
			ui.Error(fmt.Sprintf("Failed to save: %v", err))
			return
		}
		ui.Success(fmt.Sprintf("Added subgoal %s to %s", ui.DimStyle.Render(t.ShortID()), parentTask.Title))
	} else {
		if err := store.SaveTask(t); err != nil {
			ui.Error(fmt.Sprintf("Failed to save: %v", err))
			return
		}
		ui.Success(fmt.Sprintf("Added goal %s %s", ui.DimStyle.Render(t.ShortID()), title))
	}

	showInboxAgingWarning()
}
