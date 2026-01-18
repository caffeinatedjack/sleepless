package envvar

import (
	"os"
	"regexp"
	"strings"
)

var (
	// Match $VAR or ${VAR} patterns
	varPattern = regexp.MustCompile(`\$\{?([A-Za-z_][A-Za-z0-9_]*)\}?`)
)

// ExpandVariables expands $VAR and ${VAR} references in the given value.
// It first checks the provided variables map, then falls back to environment variables.
func ExpandVariables(value string, variables map[string]string) string {
	return varPattern.ReplaceAllStringFunc(value, func(match string) string {
		// Extract variable name
		varName := extractVarName(match)

		// Check provided variables first
		if val, exists := variables[varName]; exists {
			return val
		}

		// Fall back to environment variable
		if val := os.Getenv(varName); val != "" {
			return val
		}

		// If not found, preserve the reference as-is
		return match
	})
}

// extractVarName extracts the variable name from $VAR or ${VAR} syntax.
func extractVarName(match string) string {
	match = strings.TrimPrefix(match, "$")
	match = strings.TrimPrefix(match, "{")
	match = strings.TrimSuffix(match, "}")
	return match
}
