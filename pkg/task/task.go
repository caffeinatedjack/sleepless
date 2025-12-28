// Package task provides task models and types for the regimen task manager.
package task

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
	"unicode"
)

// Status represents the completion state of a task.
type Status string

const (
	StatusOpen     Status = "open"
	StatusComplete Status = "complete"
)

// Priority represents the importance level of a task.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Task represents a single task with metadata and optional subtasks.
type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Status    Status     `json:"status"`
	Priority  Priority   `json:"priority"`
	Due       *time.Time `json:"due,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Notes     []string   `json:"notes,omitempty"`
	Created   time.Time  `json:"created"`
	Completed *time.Time `json:"completed,omitempty"`
	ParentID  *string    `json:"parent_id,omitempty"`
	Subtasks  []*Task    `json:"subtasks,omitempty"`
	Topic     string     `json:"topic"`
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// New creates an open task with a fresh ID and Created timestamp.
//
// The default topic is "inbox".
func New(title string) *Task {
	return &Task{
		ID:       generateID(),
		Title:    title,
		Status:   StatusOpen,
		Priority: PriorityMedium,
		Tags:     []string{},
		Notes:    []string{},
		Created:  time.Now(),
		Subtasks: []*Task{},
		Topic:    "inbox",
	}
}

// ShortID is a compact ID for display in lists.
func (t *Task) ShortID() string {
	if len(t.ID) >= 6 {
		return t.ID[:6]
	}
	return t.ID
}

// IsComplete returns true if the task is marked complete.
func (t *Task) IsComplete() bool {
	return t.Status == StatusComplete
}

// IsOverdue returns true if the task is past its due date and not complete.
func (t *Task) IsOverdue() bool {
	if t.Due == nil || t.IsComplete() {
		return false
	}
	today := time.Now().Truncate(24 * time.Hour)
	dueDate := t.Due.Truncate(24 * time.Hour)
	return dueDate.Before(today)
}

// Complete marks the task as complete with the current timestamp.
func (t *Task) Complete() {
	t.Status = StatusComplete
	now := time.Now()
	t.Completed = &now
}

// Reopen reopens a completed task.
func (t *Task) Reopen() {
	t.Status = StatusOpen
	t.Completed = nil
}

// AddSubtask adds a subtask to this task.
func (t *Task) AddSubtask(subtask *Task) {
	subtask.ParentID = &t.ID
	subtask.Topic = t.Topic
	t.Subtasks = append(t.Subtasks, subtask)
}

// AddNote adds a note to this task.
func (t *Task) AddNote(note string) {
	t.Notes = append(t.Notes, note)
}

// CountSubtasks returns (complete, total) subtask counts recursively.
func (t *Task) CountSubtasks() (complete, total int) {
	total = len(t.Subtasks)
	for _, s := range t.Subtasks {
		if s.IsComplete() {
			complete++
		}
		subComplete, subTotal := s.CountSubtasks()
		complete += subComplete
		total += subTotal
	}
	return complete, total
}

// AllSubtasksComplete returns true if all subtasks are complete.
func (t *Task) AllSubtasksComplete() bool {
	for _, s := range t.Subtasks {
		if !s.IsComplete() {
			return false
		}
	}
	return true
}

// Validate returns human-readable problems with the task.
func (t *Task) Validate() []string {
	var errors []string
	if t.Title == "" {
		errors = append(errors, "Task title cannot be empty")
	}
	if t.Due != nil && t.Due.Before(t.Created) {
		errors = append(errors, "Due date cannot be before creation date")
	}
	return errors
}

// MatchesPrefix checks if the task ID starts with the given prefix.
func (t *Task) MatchesPrefix(prefix string) bool {
	if len(prefix) > len(t.ID) {
		return false
	}
	return t.ID[:len(prefix)] == prefix
}

// FindByPrefix finds all tasks (including subtasks) matching the ID prefix.
func (t *Task) FindByPrefix(prefix string) []*Task {
	var matches []*Task
	if t.MatchesPrefix(prefix) {
		matches = append(matches, t)
	}
	for _, subtask := range t.Subtasks {
		matches = append(matches, subtask.FindByPrefix(prefix)...)
	}
	return matches
}

// MarshalJSON formats dates consistently when writing a Task to JSON.
func (t *Task) MarshalJSON() ([]byte, error) {
	type Alias Task
	aux := &struct {
		Due       string `json:"due,omitempty"`
		Created   string `json:"created"`
		Completed string `json:"completed,omitempty"`
		*Alias
	}{
		Alias:   (*Alias)(t),
		Created: t.Created.Format(time.RFC3339),
	}
	if t.Due != nil {
		aux.Due = t.Due.Format("2006-01-02")
	}
	if t.Completed != nil {
		aux.Completed = t.Completed.Format(time.RFC3339)
	}
	return json.Marshal(aux)
}

// FindTasksByPrefix searches a task list (and all subtasks) by ID prefix.
func FindTasksByPrefix(tasks []*Task, prefix string) []*Task {
	var matches []*Task
	for _, task := range tasks {
		matches = append(matches, task.FindByPrefix(prefix)...)
	}
	return matches
}

// TitleCase capitalizes the first letter of each word.
func TitleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

// ParsePriority turns user input into a Priority value.
//
// Unknown values fall back to PriorityMedium.
func ParsePriority(s string) Priority {
	switch strings.ToLower(s) {
	case "low":
		return PriorityLow
	case "high":
		return PriorityHigh
	default:
		return PriorityMedium
	}
}

// ParseTags splits a comma-separated tag string.
func ParseTags(s string) []string {
	if s == "" {
		return nil
	}
	tags := strings.Split(s, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	return tags
}

// FilterByStatus filters tasks by status.
func FilterByStatus(tasks []*Task, status Status) []*Task {
	var filtered []*Task
	for _, t := range tasks {
		if t.Status == status {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// FilterByPriority filters tasks by priority.
func FilterByPriority(tasks []*Task, p Priority) []*Task {
	var filtered []*Task
	for _, t := range tasks {
		if t.Priority == p {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// FilterOpen returns only open (incomplete) tasks.
func FilterOpen(tasks []*Task) []*Task {
	return FilterByStatus(tasks, StatusOpen)
}

// FilterComplete returns only complete tasks.
func FilterComplete(tasks []*Task) []*Task {
	return FilterByStatus(tasks, StatusComplete)
}

// FilterOverdue returns only overdue tasks.
func FilterOverdue(tasks []*Task) []*Task {
	var filtered []*Task
	for _, t := range tasks {
		if t.IsOverdue() {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// GroupByTopic groups tasks and returns both the map and a sorted topic list.
func GroupByTopic(tasks []*Task) (map[string][]*Task, []string) {
	byTopic := make(map[string][]*Task)
	for _, t := range tasks {
		byTopic[t.Topic] = append(byTopic[t.Topic], t)
	}

	var topics []string
	for topic := range byTopic {
		topics = append(topics, topic)
	}
	sort.Strings(topics)

	return byTopic, topics
}
