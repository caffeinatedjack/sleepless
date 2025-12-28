// Package storage provides task persistence to markdown files.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"gitlab.com/caffeinatedjack/sleepless/pkg/parser"
	"gitlab.com/caffeinatedjack/sleepless/pkg/task"
)

const (
	metaFile        = ".task-meta.json"
	inboxAgingDays  = 7
	maxHistoryItems = 1000
)

// Storage handles reading/writing tasks to markdown files.
type Storage struct {
	Path     string // ~/wiki/tasks
	MetaPath string
}

// HistoryEntry represents a single history record.
type HistoryEntry struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	TaskID    string `json:"task_id"`
	Details   string `json:"details,omitempty"`
}

// Meta represents the metadata JSON structure.
type Meta struct {
	History []HistoryEntry `json:"history"`
	Version int            `json:"version"`
}

// New points Storage at a task directory.
//
// If no path is provided, it defaults to ~/wiki/tasks.
func New(path ...string) *Storage {
	var tasksPath string
	if len(path) > 0 && path[0] != "" {
		tasksPath = path[0]
	} else {
		home, _ := os.UserHomeDir()
		tasksPath = filepath.Join(home, "wiki", "tasks")
	}
	return &Storage{
		Path:     tasksPath,
		MetaPath: filepath.Join(tasksPath, metaFile),
	}
}

// EnsureStructure makes sure the task folder and its baseline files exist.
func (s *Storage) EnsureStructure() error {
	if err := os.MkdirAll(s.Path, 0755); err != nil {
		return err
	}

	// Create inbox if missing
	inboxPath := s.topicPath("inbox")
	if _, err := os.Stat(inboxPath); os.IsNotExist(err) {
		if err := s.writeTopicFile("inbox", "Inbox"); err != nil {
			return err
		}
	}

	// Create meta if missing
	if _, err := os.Stat(s.MetaPath); os.IsNotExist(err) {
		if err := s.saveMeta(&Meta{History: []HistoryEntry{}, Version: 1}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) topicPath(topic string) string {
	return filepath.Join(s.Path, topic+".md")
}

func (s *Storage) ensureTopicFile(topic, title string) error {
	path := s.topicPath(topic)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return s.writeTopicFile(topic, title)
	}
	if topic != "inbox" && topic != "archived" {
		return s.updateIndex(topic, title)
	}
	return nil
}

func (s *Storage) readTopic(topic string) (tasks []*task.Task, path string, exists bool, err error) {
	path = s.topicPath(topic)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return []*task.Task{}, path, false, nil
		}
		return nil, path, false, err
	}

	tasks, err = parser.ParseMarkdown(path, topic)
	if err != nil {
		return nil, path, true, err
	}
	return tasks, path, true, nil
}

func filterOutTaskByID(tasks []*task.Task, id string) []*task.Task {
	filtered := tasks[:0]
	for _, t := range tasks {
		if t.ID != id {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func (s *Storage) writeTopicFile(filename, title string) error {
	content := "# " + title + "\n\n"
	if err := os.WriteFile(s.topicPath(filename), []byte(content), 0644); err != nil {
		return err
	}
	// Update index with the new topic (skip special files)
	if filename != "inbox" && filename != "archived" {
		s.updateIndex(filename, title)
	}
	return nil
}

func (s *Storage) updateIndex(filename, title string) error {
	indexPath := filepath.Join(s.Path, "index.md")

	// Read existing content
	content, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create index with default structure
			content = []byte("# Tasks\n\nWelcome to your task dashboard.\n\n## Quick Links\n\n- [Inbox](inbox.md) - Uncategorized tasks\n- [Archived](archived.md) - Completed tasks\n\n## Topics\n\n")
		} else {
			return err
		}
	}

	// Check if topic already in index
	topicLink := fmt.Sprintf("[%s](%s.md)", title, filename)
	if strings.Contains(string(content), topicLink) || strings.Contains(string(content), "("+filename+".md)") {
		return nil
	}

	// Find "## Topics" section and add link
	contentStr := string(content)
	topicsMarker := "## Topics"
	if idx := strings.Index(contentStr, topicsMarker); idx != -1 {
		// Find end of topics header line
		endOfLine := idx + len(topicsMarker)
		for endOfLine < len(contentStr) && contentStr[endOfLine] != '\n' {
			endOfLine++
		}
		if endOfLine < len(contentStr) {
			endOfLine++ // Include the newline
		}
		// Skip any blank line after header
		if endOfLine < len(contentStr) && contentStr[endOfLine] == '\n' {
			endOfLine++
		}
		// Insert the new topic link
		newLink := fmt.Sprintf("- [%s](%s.md)\n", title, filename)
		contentStr = contentStr[:endOfLine] + newLink + contentStr[endOfLine:]
	} else {
		// No Topics section, append one
		contentStr += "\n## Topics\n\n" + fmt.Sprintf("- [%s](%s.md)\n", title, filename)
	}

	return os.WriteFile(indexPath, []byte(contentStr), 0644)
}

// LoadTasks reads tasks from disk.
//
// If topic is empty, it loads every topic file.
func (s *Storage) LoadTasks(topic string) ([]*task.Task, error) {
	if err := s.EnsureStructure(); err != nil {
		return nil, err
	}

	topics := []string{topic}
	if topic == "" {
		topics = s.ListTopics()
	}

	var all []*task.Task
	for _, t := range topics {
		path := s.topicPath(t)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		tasks, err := parser.ParseMarkdown(path, t)
		if err != nil {
			return nil, err
		}
		all = append(all, tasks...)
	}
	return all, nil
}

// SaveTask writes a task back to its topic file (creating the topic if needed).
func (s *Storage) SaveTask(t *task.Task) error {
	if err := s.EnsureStructure(); err != nil {
		return err
	}

	title := strings.ReplaceAll(t.Topic, "-", " ")
	if err := s.ensureTopicFile(t.Topic, title); err != nil {
		return err
	}

	tasks, path, _, err := s.readTopic(t.Topic)
	if err != nil {
		return err
	}

	// Update or append
	found := false
	for i, existing := range tasks {
		if existing.ID == t.ID {
			tasks[i] = t
			found = true
			break
		}
	}
	if !found {
		tasks = append(tasks, t)
	}

	if err := s.writeTasksLocked(path, tasks, t.Topic); err != nil {
		return err
	}
	return s.AddHistory("save", t.ID, t.Title)
}

func (s *Storage) writeTasksLocked(path string, tasks []*task.Task, topic string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	return parser.WriteMarkdownTo(file, tasks, topic)
}

// SaveTaskOrParent persists a change to t.
//
// If t is a subtask, it saves the parent so the whole tree stays consistent.
func (s *Storage) SaveTaskOrParent(t *task.Task) error {
	if t.ParentID != nil {
		parent, err := s.FindByID(*t.ParentID)
		if err != nil {
			return err
		}
		if parent != nil {
			return s.SaveTask(parent)
		}
	}
	return s.SaveTask(t)
}

// RemoveTask deletes a task from its topic file.
func (s *Storage) RemoveTask(t *task.Task) error {
	if err := s.EnsureStructure(); err != nil {
		return err
	}

	tasks, path, exists, err := s.readTopic(t.Topic)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	filtered := filterOutTaskByID(tasks, t.ID)
	if err := s.writeTasksLocked(path, filtered, t.Topic); err != nil {
		return err
	}
	return s.AddHistory("remove", t.ID, t.Title)
}

// MoveTask moves a task (and its subtasks) to a different topic file.
func (s *Storage) MoveTask(t *task.Task, newTopic string) error {
	if err := s.EnsureStructure(); err != nil {
		return err
	}

	oldTopic := t.Topic

	// Remove from old topic (without logging)
	oldTasks, oldPath, oldExists, err := s.readTopic(oldTopic)
	if err != nil {
		return err
	}
	if oldExists {
		filtered := filterOutTaskByID(oldTasks, t.ID)
		if err := s.writeTasksLocked(oldPath, filtered, oldTopic); err != nil {
			return err
		}
	}

	// Update topic for task and subtasks
	t.Topic = newTopic
	for _, sub := range t.Subtasks {
		sub.Topic = newTopic
	}

	title := strings.ReplaceAll(newTopic, "-", " ")
	if err := s.ensureTopicFile(newTopic, title); err != nil {
		return err
	}

	newTasks, newPath, _, err := s.readTopic(newTopic)
	if err != nil {
		return err
	}

	newTasks = append(newTasks, t)
	if err := s.writeTasksLocked(newPath, newTasks, newTopic); err != nil {
		return err
	}

	return s.AddHistory("move", t.ID, oldTopic+" -> "+newTopic)
}

// ArchiveTask moves a completed task into archived.md.
func (s *Storage) ArchiveTask(t *task.Task) error {
	if !t.IsComplete() {
		return nil
	}
	if err := s.EnsureStructure(); err != nil {
		return err
	}

	oldTopic := t.Topic

	// Remove from current topic
	oldTasks, oldPath, oldExists, err := s.readTopic(oldTopic)
	if err != nil {
		return err
	}
	if oldExists {
		filtered := filterOutTaskByID(oldTasks, t.ID)
		if err := s.writeTasksLocked(oldPath, filtered, oldTopic); err != nil {
			return err
		}
	}

	// Add to archive (prepend for most recent first)
	t.Topic = "archived"
	if err := s.ensureTopicFile("archived", "Archived Tasks"); err != nil {
		return err
	}

	archived, archivedPath, _, err := s.readTopic("archived")
	if err != nil {
		return err
	}

	archived = append([]*task.Task{t}, archived...)
	if err := s.writeTasksLocked(archivedPath, archived, "archived"); err != nil {
		return err
	}

	return s.AddHistory("archive", t.ID, t.Title)
}

// ListTopics returns topic names (excluding index, archived, and metadata files).
func (s *Storage) ListTopics() []string {
	excluded := map[string]bool{"index.md": true, "archived.md": true, metaFile: true}

	files, err := os.ReadDir(s.Path)
	if err != nil {
		return nil
	}

	var topics []string
	for _, f := range files {
		name := f.Name()
		if !f.IsDir() && strings.HasSuffix(name, ".md") && !excluded[name] {
			topics = append(topics, strings.TrimSuffix(name, ".md"))
		}
	}
	sort.Strings(topics)
	return topics
}

// FindByID finds a task by full ID or a unique prefix.
func (s *Storage) FindByID(id string) (*task.Task, error) {
	matches, err := s.FindAllByPrefix(id)
	if err != nil {
		return nil, err
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return nil, nil
}

// FindAllByPrefix returns every task whose ID starts with prefix.
func (s *Storage) FindAllByPrefix(prefix string) ([]*task.Task, error) {
	tasks, err := s.LoadTasks("")
	if err != nil {
		return nil, err
	}
	return task.FindTasksByPrefix(tasks, prefix), nil
}

// GetAgingInboxTasks returns incomplete inbox tasks older than a week.
func (s *Storage) GetAgingInboxTasks() ([]*task.Task, error) {
	tasks, err := s.LoadTasks("inbox")
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().AddDate(0, 0, -inboxAgingDays)
	var aging []*task.Task
	for _, t := range tasks {
		if t.Created.Before(cutoff) && !t.IsComplete() {
			aging = append(aging, t)
		}
	}
	return aging, nil
}

// AddHistory appends an entry to the task history log.
func (s *Storage) AddHistory(action, taskID, details string) error {
	meta, err := s.loadMeta()
	if err != nil {
		return err
	}

	meta.History = append(meta.History, HistoryEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Action:    action,
		TaskID:    taskID,
		Details:   details,
	})

	// Keep last N entries
	if len(meta.History) > maxHistoryItems {
		meta.History = meta.History[len(meta.History)-maxHistoryItems:]
	}

	return s.saveMeta(meta)
}

// GetHistory returns the most recent history entries.
//
// If taskID is set, it filters by ID prefix.
func (s *Storage) GetHistory(limit int, taskID string) ([]HistoryEntry, error) {
	meta, err := s.loadMeta()
	if err != nil {
		return nil, err
	}

	history := meta.History
	if taskID != "" {
		var filtered []HistoryEntry
		for _, h := range history {
			if strings.HasPrefix(h.TaskID, taskID) {
				filtered = append(filtered, h)
			}
		}
		history = filtered
	}

	// Take last N and reverse
	if len(history) > limit {
		history = history[len(history)-limit:]
	}
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

func (s *Storage) loadMeta() (*Meta, error) {
	data, err := os.ReadFile(s.MetaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Meta{History: []HistoryEntry{}, Version: 1}, nil
		}
		return nil, err
	}

	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *Storage) saveMeta(meta *Meta) error {
	os.MkdirAll(filepath.Dir(s.MetaPath), 0755)

	file, err := os.OpenFile(s.MetaPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}
