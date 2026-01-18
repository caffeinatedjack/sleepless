// Package envvar provides environment variable export generation for different shells.
package envvar

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Shell represents a supported shell type.
type Shell string

const (
	ShellBash       Shell = "bash"
	ShellZsh        Shell = "zsh"
	ShellFish       Shell = "fish"
	ShellPowerShell Shell = "powershell"
	ShellPwsh       Shell = "pwsh"
)

// DetectShell detects the current shell from the SHELL environment variable.
// Returns "bash" as default if detection fails.
func DetectShell() Shell {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return ShellBash
	}

	shellName := filepath.Base(shellPath)

	switch shellName {
	case "bash":
		return ShellBash
	case "zsh":
		return ShellZsh
	case "fish":
		return ShellFish
	case "pwsh", "powershell":
		return ShellPowerShell
	default:
		return ShellBash
	}
}

// ParseShell converts a string to a Shell type.
func ParseShell(s string) (Shell, error) {
	switch strings.ToLower(s) {
	case "bash":
		return ShellBash, nil
	case "zsh":
		return ShellZsh, nil
	case "fish":
		return ShellFish, nil
	case "powershell", "pwsh":
		return ShellPowerShell, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", s)
	}
}

// Export generates shell-specific export commands for the given variables.
func Export(shell Shell, variables map[string]string) (string, error) {
	if len(variables) == 0 {
		return "", nil
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(variables))
	for k := range variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []string

	switch shell {
	case ShellBash, ShellZsh:
		for _, key := range keys {
			value := variables[key]
			// Expand variables in value
			expandedValue := ExpandVariables(value, variables)
			// Quote value to handle spaces and special characters
			lines = append(lines, fmt.Sprintf(`export %s="%s"`, key, escapeDoubleQuotes(expandedValue)))
		}

	case ShellFish:
		for _, key := range keys {
			value := variables[key]
			expandedValue := ExpandVariables(value, variables)
			// Fish uses different syntax
			lines = append(lines, fmt.Sprintf(`set -gx %s "%s"`, key, escapeDoubleQuotes(expandedValue)))
		}

	case ShellPowerShell, ShellPwsh:
		for _, key := range keys {
			value := variables[key]
			expandedValue := ExpandVariables(value, variables)
			// PowerShell syntax
			lines = append(lines, fmt.Sprintf(`$env:%s = "%s"`, key, escapePowerShell(expandedValue)))
		}

	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}

	return strings.Join(lines, "\n"), nil
}

// escapeDoubleQuotes escapes double quotes in a string for bash/zsh/fish.
func escapeDoubleQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// escapePowerShell escapes special characters for PowerShell strings.
func escapePowerShell(s string) string {
	// PowerShell uses backtick for escaping
	s = strings.ReplaceAll(s, "`", "``")
	s = strings.ReplaceAll(s, `"`, "`\"")
	return s
}
