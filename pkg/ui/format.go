// Package ui provides terminal UI formatting helpers.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"gitlab.com/caffeinatedjack/sleepless/pkg/task"
)

// TitleCase is an alias to task.TitleCase for convenience.
var TitleCase = task.TitleCase

var (
	// Colors
	Red      = lipgloss.Color("9")
	Yellow   = lipgloss.Color("11")
	Green    = lipgloss.Color("10")
	Blue     = lipgloss.Color("12")
	DimColor = lipgloss.Color("8")
	White    = lipgloss.Color("15")

	// Styles
	ErrorStyle   = lipgloss.NewStyle().Foreground(Red)
	WarningStyle = lipgloss.NewStyle().Foreground(Yellow)
	SuccessStyle = lipgloss.NewStyle().Foreground(Green)
	InfoStyle    = lipgloss.NewStyle().Foreground(Blue)
	DimStyle     = lipgloss.NewStyle().Foreground(DimColor)
	BoldStyle    = lipgloss.NewStyle().Bold(true)
	TopicStyle   = lipgloss.NewStyle().Foreground(Blue).Bold(true)
)

// PriorityColor returns the color for a priority level.
func PriorityColor(p task.Priority) lipgloss.Color {
	switch p {
	case task.PriorityHigh:
		return Red
	case task.PriorityMedium:
		return Yellow
	case task.PriorityLow:
		return Green
	default:
		return White
	}
}

// PriorityStyle returns the style for a priority level.
func PriorityStyle(p task.Priority) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(PriorityColor(p))
}

// Checkbox returns a formatted checkbox string.
func Checkbox(complete bool) string {
	if complete {
		return "[x]"
	}
	return "[ ]"
}

// FormatDue formats a due date with color for overdue.
func FormatDue(t *task.Task) string {
	if t.Due == nil {
		return ""
	}

	if t.IsOverdue() {
		style := lipgloss.NewStyle().Foreground(Red).Bold(true)
		return style.Render(fmt.Sprintf("OVERDUE %s", t.Due.Format("2006-01-02")))
	}

	today := time.Now().Truncate(24 * time.Hour)
	dueDate := t.Due.Truncate(24 * time.Hour)
	days := int(dueDate.Sub(today).Hours() / 24)

	style := lipgloss.NewStyle().Foreground(Yellow)
	switch {
	case days == 0:
		return style.Render("today")
	case days == 1:
		return style.Render("tomorrow")
	case days <= 7:
		return style.Render(t.Due.Format("2006-01-02"))
	default:
		return t.Due.Format("2006-01-02")
	}
}

// PriorityIndicator returns a priority indicator string.
func PriorityIndicator(p task.Priority) string {
	switch p {
	case task.PriorityHigh:
		return lipgloss.NewStyle().Foreground(Red).Render("!")
	case task.PriorityLow:
		return lipgloss.NewStyle().Foreground(Green).Render("-")
	default:
		return ""
	}
}

// ProgressBar creates a simple progress bar.
func ProgressBar(percent float64, color lipgloss.Color) string {
	filled := int(percent / 5)
	if filled > 20 {
		filled = 20
	}
	empty := 20 - filled

	filledStyle := lipgloss.NewStyle().Foreground(color)
	emptyStyle := DimStyle

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("%s %5.1f%%", bar, percent)
}

// FormatTask formats a task for display.
func FormatTask(t *task.Task, indent int) string {
	prefix := strings.Repeat("  ", indent)
	checkbox := Checkbox(t.IsComplete())

	var style lipgloss.Style
	if t.IsComplete() {
		style = DimStyle
	} else {
		style = lipgloss.NewStyle().Foreground(White)
	}

	priorityInd := ""
	if !t.IsComplete() {
		priorityInd = PriorityIndicator(t.Priority)
		if priorityInd != "" {
			priorityInd += " "
		}
	}

	dueStr := FormatDue(t)
	duePart := ""
	if dueStr != "" {
		duePart = " " + dueStr
	}

	shortID := DimStyle.Render(t.ShortID())

	return fmt.Sprintf("%s%s %s%s %s%s",
		prefix, checkbox, priorityInd, style.Render(t.Title), shortID, duePart)
}

// FormatSubtaskSummary formats a subtask summary.
func FormatSubtaskSummary(complete, total int) string {
	return DimStyle.Render(fmt.Sprintf("(%d/%d subtasks)", complete, total))
}

// Success prints a success message.
func Success(msg string) {
	fmt.Println(SuccessStyle.Render(msg))
}

// Error prints an error message.
func Error(msg string) {
	fmt.Println(ErrorStyle.Render(msg))
}

// Warning prints a warning message.
func Warning(msg string) {
	fmt.Println(WarningStyle.Render(msg))
}

// Info prints an info message.
func Info(msg string) {
	fmt.Println(InfoStyle.Render(msg))
}

// PrintDim prints a dimmed message.
func PrintDim(msg string) {
	fmt.Println(DimStyle.Render(msg))
}
