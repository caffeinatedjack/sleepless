package regimen

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/storage"
	"gitlab.com/caffeinatedjack/sleepless/pkg/task"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var store = storage.New()

var goalsCmd = &cobra.Command{
	Use:   "goals",
	Short: "Manage goals in your wiki",
	Long: `Manage goals in your wiki.

Commands:
    add       Create a new goal
    list      Show goals with optional filters
    done      Mark a goal as complete
    edit      Modify goal properties
    remove    Delete a goal
    move      Move goal to different topic
    view      Visual goal displays
    search    Search across all goals
    remind    Show due date reminders
    archive   Move completed goals to archive
    history   Show goal change history

Examples:
    regimen goals add "My goal" --topic work
    regimen goals list --priority high
    regimen goals view progress`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		store.EnsureStructure()
	},
}

func init() {
	rootCmd.AddCommand(goalsCmd)
}

// showInboxAgingWarning shows warning if inbox has aging tasks.
func showInboxAgingWarning() {
	aging, err := store.GetAgingInboxTasks()
	if err != nil || len(aging) == 0 {
		return
	}

	fmt.Println()
	ui.Warning(fmt.Sprintf("You have %d task(s) in inbox older than 7 days. Consider organizing them:", len(aging)))
	for i, t := range aging {
		if i >= 5 {
			ui.PrintDim(fmt.Sprintf("  ... and %d more", len(aging)-5))
			break
		}
		fmt.Printf("  %s %s\n", ui.DimStyle.Render(t.ShortID()), t.Title)
	}
	fmt.Println()
}

// resolveTask resolves a task ID with prefix matching.
func resolveTask(taskID string) (*task.Task, error) {
	matches, err := store.FindAllByPrefix(taskID)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		ui.Error(fmt.Sprintf("No task found matching '%s'", taskID))
		return nil, nil
	}

	if len(matches) > 1 {
		ui.Error(fmt.Sprintf("Multiple tasks match '%s':", taskID))
		for _, t := range matches {
			fmt.Printf("  %s %s\n", ui.DimStyle.Render(t.ShortID()), t.Title)
		}
		return nil, errors.New("ambiguous task ID")
	}

	return matches[0], nil
}
