package regimen

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/task"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var remindCmd = &cobra.Command{
	Use:   "remind",
	Short: "Show due date reminders",
	Long: `Show due date reminders.

Examples:
    regimen goals remind
    regimen goals remind --days 14
    regimen goals remind --install`,
	Run: runRemind,
}

var (
	remindDays    int
	remindInstall bool
)

func init() {
	goalsCmd.AddCommand(remindCmd)
	remindCmd.Flags().IntVarP(&remindDays, "days", "d", 7, "Days to look ahead")
	remindCmd.Flags().BoolVar(&remindInstall, "install", false, "Output cron entry for daily reminders")
}

func runRemind(cmd *cobra.Command, args []string) {
	if remindInstall {
		scriptPath, _ := os.Executable()
		fmt.Println()
		fmt.Println(ui.BoldStyle.Render("Add this to your crontab (crontab -e):"))
		fmt.Println()
		fmt.Printf("  0 9 * * * %s goals remind\n", scriptPath)
		fmt.Println()
		ui.PrintDim("This will show reminders daily at 9 AM")
		fmt.Println()
		return
	}

	allTasks, err := store.LoadTasks("")
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load goals: %v", err))
		return
	}

	today := time.Now().Truncate(24 * time.Hour)

	// Overdue
	var overdue []*task.Task
	for _, t := range allTasks {
		if t.IsOverdue() {
			overdue = append(overdue, t)
		}
	}

	if len(overdue) > 0 {
		fmt.Println()
		fmt.Println(ui.ErrorStyle.Bold(true).Render("OVERDUE TASKS"))
		fmt.Println()
		for _, t := range overdue {
			daysOverdue := int(today.Sub(*t.Due).Hours() / 24)
			fmt.Printf("  %s %s %s %s\n",
				ui.ErrorStyle.Render("*"),
				t.Title,
				ui.DimStyle.Render(t.ShortID()),
				ui.ErrorStyle.Render(fmt.Sprintf("(%d day(s) overdue)", daysOverdue)))
		}
	}

	// Upcoming
	cutoff := today.AddDate(0, 0, remindDays)
	var upcoming []*task.Task
	for _, t := range allTasks {
		if t.Due != nil && !t.Due.Before(today) && !t.Due.After(cutoff) && !t.IsComplete() {
			upcoming = append(upcoming, t)
		}
	}

	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].Due.Before(*upcoming[j].Due)
	})

	if len(upcoming) > 0 {
		fmt.Println()
		fmt.Println(ui.WarningStyle.Bold(true).Render(fmt.Sprintf("Due in next %d days", remindDays)))
		fmt.Println()
		for _, t := range upcoming {
			daysUntil := int(t.Due.Sub(today).Hours() / 24)
			dayStr := "today"
			if daysUntil > 0 {
				dayStr = fmt.Sprintf("in %d day(s)", daysUntil)
			}
			style := ui.WarningStyle
			if daysUntil > 2 {
				style = ui.PriorityStyle(t.Priority)
			}
			fmt.Printf("  %s %s %s %s\n",
				style.Render("*"),
				t.Title,
				ui.DimStyle.Render(t.ShortID()),
				style.Render(fmt.Sprintf("(%s)", dayStr)))
		}
	}

	if len(overdue) == 0 && len(upcoming) == 0 {
		fmt.Println()
		ui.Success("No upcoming deadlines")
	}

	fmt.Println()
}
