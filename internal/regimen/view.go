package regimen

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/task"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Visual goal displays",
	Long: `Visual goal displays.

Views:
    tree      Hierarchical goal tree
    progress  Progress bars by topic/priority
    deps      ASCII dependency diagram
    calendar  Upcoming due dates grid`,
}

func init() {
	goalsCmd.AddCommand(viewCmd)
	viewCmd.AddCommand(viewTreeCmd)
	viewCmd.AddCommand(viewProgressCmd)
	viewCmd.AddCommand(viewDepsCmd)
	viewCmd.AddCommand(viewCalendarCmd)
}

var viewTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Show hierarchical goal tree",
	Long: `Show hierarchical goal tree.

Examples:
    regimen goals view tree
    regimen goals view tree --topic work
    regimen goals view tree --expand-completed`,
	Run: runViewTree,
}

var (
	viewTreeTopic           string
	viewTreeExpandCompleted bool
)

func init() {
	viewTreeCmd.Flags().StringVarP(&viewTreeTopic, "topic", "t", "", "Filter by topic")
	viewTreeCmd.Flags().BoolVar(&viewTreeExpandCompleted, "expand-completed", false, "Show completed subtasks")
}

func runViewTree(cmd *cobra.Command, args []string) {
	allTasks, err := store.LoadTasks(viewTreeTopic)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load goals: %v", err))
		return
	}

	tasksToShow := allTasks
	if !viewTreeExpandCompleted {
		tasksToShow = task.FilterOpen(allTasks)
	}

	if len(tasksToShow) == 0 {
		ui.PrintDim("No open goals")
		return
	}

	// Group by topic
	byTopic, topics := task.GroupByTopic(tasksToShow)

	for _, topic := range topics {
		topicTasks := byTopic[topic]
		fmt.Println(ui.TopicStyle.Render(ui.TitleCase(topic)))
		for _, t := range topicTasks {
			printTreeTask(t, 0, "")
		}
		fmt.Println()
	}
}

func printTreeTask(t *task.Task, depth int, prefix string) {
	checkbox := ui.Checkbox(t.IsComplete())
	var style = ui.DimStyle
	if !t.IsComplete() {
		style = ui.PriorityStyle(t.Priority)
	}

	label := fmt.Sprintf("%s %s %s", checkbox, style.Render(t.Title), ui.DimStyle.Render(t.ShortID()))
	if t.Due != nil {
		label += " " + ui.FormatDue(t)
	}

	connector := ""
	if depth > 0 {
		connector = "├── "
	}
	fmt.Printf("%s%s%s\n", prefix, connector, label)

	complete, total := t.CountSubtasks()
	if total > 0 && !viewTreeExpandCompleted {
		// Show collapsed summary
		childPrefix := prefix
		if depth > 0 {
			childPrefix += "│   "
		}
		fmt.Printf("%s    %s\n", childPrefix, ui.FormatSubtaskSummary(complete, total))
	} else {
		for i, subtask := range t.Subtasks {
			if viewTreeExpandCompleted || !subtask.IsComplete() {
				childPrefix := prefix
				if depth > 0 {
					childPrefix += "│   "
				}
				isLast := i == len(t.Subtasks)-1
				_ = isLast // TODO: handle last child differently
				printTreeTask(subtask, depth+1, childPrefix)
			}
		}
	}
}

// ===== view progress =====

var viewProgressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Show progress bars by topic and priority",
	Long: `Show progress bars by topic and priority.

Example:
    regimen goals view progress`,
	Run: runViewProgress,
}

func runViewProgress(cmd *cobra.Command, args []string) {
	allTasks, err := store.LoadTasks("")
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load goals: %v", err))
		return
	}

	// By topic
	fmt.Println()
	fmt.Println(ui.BoldStyle.Render("Progress by Topic"))
	fmt.Println()

	byTopic := make(map[string][2]int) // [complete, total]
	for _, t := range allTasks {
		complete, total := 0, 1
		if t.IsComplete() {
			complete = 1
		}
		subComplete, subTotal := t.CountSubtasks()
		complete += subComplete
		total += subTotal

		prev := byTopic[t.Topic]
		byTopic[t.Topic] = [2]int{prev[0] + complete, prev[1] + total}
	}

	var topics []string
	for topic := range byTopic {
		topics = append(topics, topic)
	}
	sort.Strings(topics)

	for _, topic := range topics {
		counts := byTopic[topic]
		complete, total := counts[0], counts[1]
		pct := float64(0)
		if total > 0 {
			pct = float64(complete) / float64(total) * 100
		}
		bar := ui.ProgressBar(pct, ui.Green)
		fmt.Printf("  %-15s %s %d/%d\n", topic, bar, complete, total)
	}

	// By priority
	fmt.Println()
	fmt.Println(ui.BoldStyle.Render("Progress by Priority"))
	fmt.Println()

	byPriority := make(map[task.Priority][2]int)
	for _, t := range allTasks {
		complete := 0
		if t.IsComplete() {
			complete = 1
		}
		prev := byPriority[t.Priority]
		byPriority[t.Priority] = [2]int{prev[0] + complete, prev[1] + 1}
	}

	for _, priority := range []task.Priority{task.PriorityHigh, task.PriorityMedium, task.PriorityLow} {
		counts := byPriority[priority]
		complete, total := counts[0], counts[1]
		pct := float64(0)
		if total > 0 {
			pct = float64(complete) / float64(total) * 100
		}
		color := ui.PriorityColor(priority)
		bar := ui.ProgressBar(pct, color)
		priorityStyle := ui.PriorityStyle(priority)
		fmt.Printf("  %s %s %d/%d\n", priorityStyle.Render(fmt.Sprintf("%-8s", priority)), bar, complete, total)
	}

	fmt.Println()
}

// ===== view deps =====

var viewDepsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Show ASCII dependency diagram",
	Long: `Show ASCII dependency diagram (parent-child relationships).

Example:
    regimen goals view deps`,
	Run: runViewDeps,
}

func runViewDeps(cmd *cobra.Command, args []string) {
	allTasks, err := store.LoadTasks("")
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load goals: %v", err))
		return
	}

	openTasks := task.FilterOpen(allTasks)

	if len(openTasks) == 0 {
		ui.PrintDim("No open goals")
		return
	}

	fmt.Println()
	fmt.Println(ui.BoldStyle.Render("Goal Dependencies"))
	fmt.Println()

	for _, t := range openTasks {
		if len(t.Subtasks) > 0 {
			printDeps(t, 0)
			fmt.Println()
		}
	}
}

func printDeps(t *task.Task, depth int) {
	prefix := strings.Repeat("│   ", depth)
	connector := ""
	if depth > 0 {
		connector = "├── "
	}

	status := "○"
	if t.IsComplete() {
		status = "✓"
	}

	var style = ui.DimStyle
	if !t.IsComplete() {
		style = ui.PriorityStyle(t.Priority)
	}

	fmt.Printf("%s%s%s %s %s\n", prefix, connector, style.Render(status), style.Render(t.Title), ui.DimStyle.Render(t.ShortID()))

	for _, subtask := range t.Subtasks {
		printDeps(subtask, depth+1)
	}
}

// ===== view calendar =====

var viewCalendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Show upcoming due dates",
	Long: `Show upcoming due dates.

Example:
    regimen goals view calendar
    regimen goals view calendar --days 30`,
	Run: runViewCalendar,
}

var viewCalendarDays int

func init() {
	viewCalendarCmd.Flags().IntVarP(&viewCalendarDays, "days", "d", 14, "Number of days to show")
}

func runViewCalendar(cmd *cobra.Command, args []string) {
	allTasks, err := store.LoadTasks("")
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load goals: %v", err))
		return
	}

	today := time.Now().Truncate(24 * time.Hour)

	// Collect tasks with due dates
	type taskWithDue struct {
		task *task.Task
		due  time.Time
	}
	var withDue []taskWithDue
	for _, t := range allTasks {
		if t.Due != nil && !t.IsComplete() {
			withDue = append(withDue, taskWithDue{t, *t.Due})
		}
	}

	sort.Slice(withDue, func(i, j int) bool {
		return withDue[i].due.Before(withDue[j].due)
	})

	// Overdue
	var overdue []taskWithDue
	for _, tw := range withDue {
		if tw.due.Before(today) {
			overdue = append(overdue, tw)
		}
	}

	if len(overdue) > 0 {
		fmt.Println()
		fmt.Println(ui.ErrorStyle.Bold(true).Render("OVERDUE"))
		for _, tw := range overdue {
			daysOverdue := int(today.Sub(tw.due).Hours() / 24)
			fmt.Printf("  %s (%dd ago) - %s %s\n",
				ui.ErrorStyle.Render(tw.due.Format("2006-01-02")),
				daysOverdue,
				tw.task.Title,
				ui.DimStyle.Render(tw.task.ShortID()))
		}
	}

	// Upcoming
	cutoff := today.AddDate(0, 0, viewCalendarDays)
	var upcoming []taskWithDue
	for _, tw := range withDue {
		if !tw.due.Before(today) && !tw.due.After(cutoff) {
			upcoming = append(upcoming, tw)
		}
	}

	if len(upcoming) > 0 {
		fmt.Println()
		fmt.Println(ui.BoldStyle.Render(fmt.Sprintf("Next %d Days", viewCalendarDays)))

		var currentDate *time.Time
		for _, tw := range upcoming {
			dueDate := tw.due.Truncate(24 * time.Hour)
			if currentDate == nil || !dueDate.Equal(*currentDate) {
				currentDate = &dueDate
				dayName := "Today"
				if !dueDate.Equal(today) {
					dayName = dueDate.Format("Mon Jan 02")
				}
				fmt.Println()
				fmt.Printf("  %s\n", ui.WarningStyle.Render(dayName))
			}
			priorityStyle := ui.PriorityStyle(tw.task.Priority)
			fmt.Printf("    %s %s %s\n",
				priorityStyle.Render("○"),
				tw.task.Title,
				ui.DimStyle.Render(tw.task.ShortID()))
		}
	}

	if len(overdue) == 0 && len(upcoming) == 0 {
		fmt.Println()
		ui.PrintDim("No upcoming due dates")
	}

	fmt.Println()
}
