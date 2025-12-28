package regimen

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/task"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var editCmd = &cobra.Command{
	Use:   "edit <goal-id>",
	Short: "Edit goal properties",
	Long: `Edit goal properties.

Examples:
    regimen goals edit a1b2c3 --title "New title"
    regimen goals edit a1b --priority high --due 2025-02-01
    regimen goals edit a1b --note "Remember to check X"`,
	Args: cobra.ExactArgs(1),
	Run:  runEdit,
}

var (
	editTitle    string
	editPriority string
	editDue      string
	editTags     string
	editNote     string
)

func init() {
	goalsCmd.AddCommand(editCmd)
	editCmd.Flags().StringVar(&editTitle, "title", "", "New title")
	editCmd.Flags().StringVarP(&editPriority, "priority", "p", "", "New priority (low, medium, high)")
	editCmd.Flags().StringVarP(&editDue, "due", "d", "", "New due date (YYYY-MM-DD)")
	editCmd.Flags().StringVar(&editTags, "tags", "", "New tags (comma-separated)")
	editCmd.Flags().StringVarP(&editNote, "note", "n", "", "Add a note")
}

func runEdit(cmd *cobra.Command, args []string) {
	taskID := args[0]

	t, err := resolveTask(taskID)
	if err != nil {
		ui.Error(err.Error())
		return
	}
	if t == nil {
		return
	}

	var changes []string

	if editTitle != "" {
		t.Title = editTitle
		changes = append(changes, "title")
	}

	if editPriority != "" {
		t.Priority = task.ParsePriority(editPriority)
		changes = append(changes, "priority")
	}

	if editDue != "" {
		due, err := time.Parse("2006-01-02", editDue)
		if err != nil {
			ui.Error(fmt.Sprintf("Invalid date format: %s (use YYYY-MM-DD)", editDue))
			return
		}
		t.Due = &due
		changes = append(changes, "due")
	}

	if editTags != "" {
		t.Tags = task.ParseTags(editTags)
		changes = append(changes, "tags")
	}

	if editNote != "" {
		t.AddNote(editNote)
		changes = append(changes, "note")
	}

	if len(changes) == 0 {
		ui.Warning("No changes specified")
		return
	}

	if err := store.SaveTaskOrParent(t); err != nil {
		ui.Error(fmt.Sprintf("Failed to save: %v", err))
		return
	}

	ui.Success(fmt.Sprintf("Updated %s (%s)", t.Title, strings.Join(changes, ", ")))
}
