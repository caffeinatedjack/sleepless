package regimen

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var historyCmd = &cobra.Command{
	Use:   "history [goal-id]",
	Short: "Show goal change history",
	Long: `Show goal change history.

Examples:
    regimen goals history
    regimen goals history a1b2c3
    regimen goals history --limit 50`,
	Run: runHistory,
}

var historyLimit int

func init() {
	goalsCmd.AddCommand(historyCmd)
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "l", 20, "Number of entries to show")
}

func runHistory(cmd *cobra.Command, args []string) {
	var taskID string
	if len(args) > 0 {
		taskID = args[0]
	}

	entries, err := store.GetHistory(historyLimit, taskID)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to get history: %v", err))
		return
	}

	if len(entries) == 0 {
		ui.PrintDim("No history found")
		return
	}

	fmt.Println()
	fmt.Println(ui.BoldStyle.Render("Goal History"))
	fmt.Println()

	for _, entry := range entries {
		// Format timestamp
		timestamp := entry.Timestamp
		if len(timestamp) > 16 {
			timestamp = timestamp[:16]
		}
		timestamp = strings.Replace(timestamp, "T", " ", 1)

		// Get action color
		var actionStyle = ui.DimStyle
		switch entry.Action {
		case "save":
			actionStyle = ui.SuccessStyle
		case "remove":
			actionStyle = ui.ErrorStyle
		case "move":
			actionStyle = ui.WarningStyle
		case "archive":
			actionStyle = ui.InfoStyle
		}

		taskShort := entry.TaskID
		if len(taskShort) > 6 {
			taskShort = taskShort[:6]
		}

		fmt.Printf("  %s %s %s %s\n",
			ui.DimStyle.Render(timestamp),
			actionStyle.Render(fmt.Sprintf("%-8s", entry.Action)),
			ui.DimStyle.Render(taskShort),
			entry.Details)
	}

	fmt.Println()
}
