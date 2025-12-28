// Package parser provides markdown parsing and serialization for tasks.
package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"gitlab.com/caffeinatedjack/sleepless/pkg/task"
)

var (
	// Match task lines: - [ ] or - [x] with optional {#id}
	taskLineRe = regexp.MustCompile(`^(\s*)- \[([ xX])\] (.+?)(?:\s*\{#([a-f0-9]+)\})?\s*$`)
	// Match metadata lines: - key: value
	metaLineRe = regexp.MustCompile(`^\s*- (\w+): (.+)$`)
	// Match note lines: - Note: ...
	noteLineRe = regexp.MustCompile(`^\s*- Note: (.+)$`)
)

// ParseMarkdown loads tasks from a topic markdown file.
//
// The returned tasks are top-level tasks; subtasks are attached to their parent.
func ParseMarkdown(path string, topic string) ([]*task.Task, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tasks []*task.Task
	var currentTask *task.Task
	var taskStack []*task.Task
	currentIndent := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Try to match task line
		if match := taskLineRe.FindStringSubmatch(line); match != nil {
			indent := len(match[1])
			isComplete := strings.ToLower(match[2]) == "x"
			title := match[3]
			taskID := match[4]

			t := task.New(title)
			t.Title = title
			if taskID != "" {
				t.ID = taskID
			}
			if isComplete {
				t.Status = task.StatusComplete
			}
			t.Topic = topic

			if indent == 0 {
				// Top-level task
				if currentTask != nil {
					tasks = append(tasks, currentTask)
				}
				currentTask = t
				taskStack = []*task.Task{t}
				currentIndent = 0
			} else {
				// Subtask - find parent based on indent
				for len(taskStack) > 1 && indent <= currentIndent {
					taskStack = taskStack[:len(taskStack)-1]
					currentIndent -= 2
				}

				parent := taskStack[len(taskStack)-1]
				parent.AddSubtask(t)
				taskStack = append(taskStack, t)
				currentIndent = indent
			}
			continue
		}

		// Try to match note line
		if currentTask != nil {
			if match := noteLineRe.FindStringSubmatch(line); match != nil {
				target := taskStack[len(taskStack)-1]
				target.AddNote(match[1])
				continue
			}
		}

		// Try to match metadata line
		if currentTask != nil {
			if match := metaLineRe.FindStringSubmatch(line); match != nil {
				key := strings.ToLower(match[1])
				value := match[2]
				target := taskStack[len(taskStack)-1]
				applyMetadata(target, key, value)
			}
		}
	}

	if currentTask != nil {
		tasks = append(tasks, currentTask)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func applyMetadata(t *task.Task, key, value string) {
	switch key {
	case "priority":
		t.Priority = task.ParsePriority(value)
	case "due":
		if due, err := time.Parse("2006-01-02", value); err == nil {
			t.Due = &due
		}
	case "tags":
		t.Tags = task.ParseTags(value)
	case "created":
		if created, err := time.Parse(time.RFC3339, value); err == nil {
			t.Created = created
		}
	case "completed":
		if completed, err := time.Parse(time.RFC3339, value); err == nil {
			t.Completed = &completed
		}
	}
}

// WriteMarkdown writes tasks to path, overwriting any existing file.
func WriteMarkdown(path string, tasks []*task.Task, topic string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return WriteMarkdownTo(file, tasks, topic)
}

// WriteMarkdownTo writes tasks as markdown to w.
func WriteMarkdownTo(w io.Writer, tasks []*task.Task, topic string) error {
	title := strings.ReplaceAll(topic, "-", " ")
	title = task.TitleCase(title)

	var lines []string
	lines = append(lines, fmt.Sprintf("# %s", title), "", "")

	for _, t := range tasks {
		lines = append(lines, taskToMarkdown(t, 0)...)
		lines = append(lines, "")
	}

	_, err := w.Write([]byte(strings.Join(lines, "\n")))
	return err
}

func taskToMarkdown(t *task.Task, indent int) []string {
	prefix := strings.Repeat("  ", indent)
	checkbox := "[ ]"
	if t.IsComplete() {
		checkbox = "[x]"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("%s- %s %s {#%s}", prefix, checkbox, t.Title, t.ID))

	// Metadata
	if t.Priority != task.PriorityMedium {
		lines = append(lines, fmt.Sprintf("%s  - priority: %s", prefix, t.Priority))
	}
	if t.Due != nil {
		lines = append(lines, fmt.Sprintf("%s  - due: %s", prefix, t.Due.Format("2006-01-02")))
	}
	if len(t.Tags) > 0 {
		lines = append(lines, fmt.Sprintf("%s  - tags: %s", prefix, strings.Join(t.Tags, ", ")))
	}
	if indent == 0 { // Only show created on top-level
		lines = append(lines, fmt.Sprintf("%s  - created: %s", prefix, t.Created.Format(time.RFC3339)))
	}
	if t.Completed != nil {
		lines = append(lines, fmt.Sprintf("%s  - completed: %s", prefix, t.Completed.Format(time.RFC3339)))
	}

	// Notes
	for _, note := range t.Notes {
		lines = append(lines, fmt.Sprintf("%s  - Note: %s", prefix, note))
	}

	// Subtasks
	for _, subtask := range t.Subtasks {
		lines = append(lines, taskToMarkdown(subtask, indent+1)...)
	}

	return lines
}
