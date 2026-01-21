package notes

import (
	"regexp"
)

var taskRefRegex = regexp.MustCompile(`@task:([a-f0-9]+)`)

// ParseTaskReferences extracts task references from note content.
// Returns a deduplicated list of task IDs/prefixes.
func ParseTaskReferences(content string) []string {
	matches := taskRefRegex.FindAllStringSubmatch(content, -1)

	seen := make(map[string]bool)
	var refs []string

	for _, match := range matches {
		if len(match) > 1 {
			taskID := match[1]
			if !seen[taskID] {
				refs = append(refs, taskID)
				seen[taskID] = true
			}
		}
	}

	return refs
}
