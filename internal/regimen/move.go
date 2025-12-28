package regimen

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var moveCmd = &cobra.Command{
	Use:   "move <goal-id> <new-topic>",
	Short: "Move a goal to a different topic",
	Long: `Move a goal to a different topic.

Examples:
    regimen goals move a1b2c3 work
    regimen goals move a1b home`,
	Args: cobra.ExactArgs(2),
	Run:  runMove,
}

func init() {
	goalsCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) {
	taskID := args[0]
	newTopic := args[1]

	t, err := resolveTask(taskID)
	if err != nil {
		ui.Error(err.Error())
		return
	}
	if t == nil {
		return
	}

	oldTopic := t.Topic
	if err := store.MoveTask(t, newTopic); err != nil {
		ui.Error(fmt.Sprintf("Failed to move: %v", err))
		return
	}

	ui.Success(fmt.Sprintf("Moved %s from %s to %s", t.Title, oldTopic, newTopic))
}
