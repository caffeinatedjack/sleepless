package regimen

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var archiveCmd = &cobra.Command{
	Use:   "archive [goal-id]",
	Short: "Move completed goals to archive",
	Long: `Move completed goals to archive.

Examples:
    regimen goals archive          # Archive all completed
    regimen goals archive a1b2c3   # Archive specific goal`,
	Run: runArchive,
}

func init() {
	goalsCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		// Archive specific goal
		taskID := args[0]
		t, err := resolveTask(taskID)
		if err != nil {
			ui.Error(err.Error())
			return
		}
		if t == nil {
			return
		}

		if !t.IsComplete() {
			ui.Error("Cannot archive incomplete goal. Mark it done first.")
			return
		}

		if err := store.ArchiveTask(t); err != nil {
			ui.Error(fmt.Sprintf("Failed to archive: %v", err))
			return
		}

		ui.Success(fmt.Sprintf("Archived %s", t.Title))
	} else {
		// Archive all completed goals
		allTasks, err := store.LoadTasks("")
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to load goals: %v", err))
			return
		}

		var completed []*struct {
			id    string
			title string
			topic string
		}
		for _, t := range allTasks {
			if t.IsComplete() && t.Topic != "archived" {
				completed = append(completed, &struct {
					id    string
					title string
					topic string
				}{t.ID, t.Title, t.Topic})
			}
		}

		if len(completed) == 0 {
			ui.PrintDim("No completed goals to archive")
			return
		}

		// Archive each goal
		for _, info := range completed {
			t, _ := store.FindByID(info.id)
			if t != nil {
				store.ArchiveTask(t)
			}
		}

		ui.Success(fmt.Sprintf("Archived %d goal(s)", len(completed)))
	}
}
