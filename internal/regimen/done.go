package regimen

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var doneCmd = &cobra.Command{
	Use:   "done <goal-id>",
	Short: "Mark a goal as complete",
	Long: `Mark a goal as complete.

Examples:
    regimen goals done a1b2c3
    regimen goals done a1b --auto-archive`,
	Args: cobra.ExactArgs(1),
	Run:  runDone,
}

var doneAutoArchive bool

func init() {
	goalsCmd.AddCommand(doneCmd)
	doneCmd.Flags().BoolVarP(&doneAutoArchive, "auto-archive", "a", false, "Automatically archive after completing")
}

func runDone(cmd *cobra.Command, args []string) {
	taskID := args[0]

	t, err := resolveTask(taskID)
	if err != nil {
		ui.Error(err.Error())
		return
	}
	if t == nil {
		return
	}

	// Check for incomplete subtasks
	if len(t.Subtasks) > 0 {
		var incomplete int
		for _, s := range t.Subtasks {
			if !s.IsComplete() {
				incomplete++
			}
		}
		if incomplete > 0 {
			ui.Warning(fmt.Sprintf("Goal has %d incomplete subtask(s):", incomplete))
			count := 0
			for _, s := range t.Subtasks {
				if !s.IsComplete() {
					if count >= 5 {
						break
					}
					fmt.Printf("  %s %s\n", ui.DimStyle.Render(s.ShortID()), s.Title)
					count++
				}
			}
			if !confirm("Complete all subtasks too?") {
				return
			}
			for _, s := range t.Subtasks {
				s.Complete()
			}
		}
	}

	t.Complete()

	if err := store.SaveTaskOrParent(t); err != nil {
		ui.Error(fmt.Sprintf("Failed to save: %v", err))
		return
	}

	ui.Success(fmt.Sprintf("Completed %s", t.Title))

	if doneAutoArchive {
		if err := store.ArchiveTask(t); err != nil {
			ui.Error(fmt.Sprintf("Failed to archive: %v", err))
			return
		}
		ui.PrintDim("Archived to archived.md")
	}
}

// confirm asks for user confirmation.
func confirm(question string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", question)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
