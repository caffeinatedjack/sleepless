package regimen

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var removeCmd = &cobra.Command{
	Use:   "remove <goal-id>",
	Short: "Remove a goal",
	Long: `Remove a goal.

Examples:
    regimen goals remove a1b2c3
    regimen goals remove a1b --force`,
	Args: cobra.ExactArgs(1),
	Run:  runRemove,
}

var removeForce bool

func init() {
	goalsCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Skip confirmation")
}

func runRemove(cmd *cobra.Command, args []string) {
	taskID := args[0]

	t, err := resolveTask(taskID)
	if err != nil {
		ui.Error(err.Error())
		return
	}
	if t == nil {
		return
	}

	if !removeForce {
		_, total := t.CountSubtasks()
		if total > 0 {
			ui.Warning(fmt.Sprintf("Goal has %d subtask(s)", total))
		}
		if !confirm(fmt.Sprintf("Delete '%s'?", t.Title)) {
			return
		}
	}

	if err := store.RemoveTask(t); err != nil {
		ui.Error(fmt.Sprintf("Failed to remove: %v", err))
		return
	}

	ui.Success(fmt.Sprintf("Removed %s", t.Title))
}
