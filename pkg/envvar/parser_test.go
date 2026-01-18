package envvar

import (
	"os"
	"testing"
)

func TestExpandVariables(t *testing.T) {
	vars := map[string]string{
		"HOME":   "/home/user",
		"EDITOR": "vim",
		"PATH":   "/usr/local/bin:/usr/bin",
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple variable",
			input: "$HOME/config",
			want:  "/home/user/config",
		},
		{
			name:  "braced variable",
			input: "${HOME}/config",
			want:  "/home/user/config",
		},
		{
			name:  "multiple variables",
			input: "$HOME uses $EDITOR",
			want:  "/home/user uses vim",
		},
		{
			name:  "mixed syntax",
			input: "$HOME and ${EDITOR}",
			want:  "/home/user and vim",
		},
		{
			name:  "undefined variable",
			input: "$UNDEFINED",
			want:  "$UNDEFINED",
		},
		{
			name:  "no variables",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "path append",
			input: "$PATH:/opt/bin",
			want:  "/usr/local/bin:/usr/bin:/opt/bin",
		},
		{
			name:  "escaped dollar (plain)",
			input: "literal $$ sign",
			want:  "literal $$ sign",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandVariables(tt.input, vars)
			if got != tt.want {
				t.Errorf("ExpandVariables() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpandVariablesFromEnv(t *testing.T) {
	// Set an actual environment variable
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	vars := map[string]string{
		"OTHER": "other-value",
	}

	input := "$TEST_VAR and $OTHER"
	want := "test-value and other-value"

	got := ExpandVariables(input, vars)
	if got != want {
		t.Errorf("ExpandVariables() = %q, want %q", got, want)
	}
}

func TestExpandVariablesPriority(t *testing.T) {
	// Set environment variable
	os.Setenv("PRIORITY_VAR", "env-value")
	defer os.Unsetenv("PRIORITY_VAR")

	// Override with local variable
	vars := map[string]string{
		"PRIORITY_VAR": "local-value",
	}

	input := "$PRIORITY_VAR"
	want := "local-value" // Local should take priority

	got := ExpandVariables(input, vars)
	if got != want {
		t.Errorf("ExpandVariables() = %q, want %q (local should override env)", got, want)
	}
}

func TestExtractVarName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "$VAR", "VAR"},
		{"braced", "${VAR}", "VAR"},
		{"with underscore", "$MY_VAR", "MY_VAR"},
		{"with numbers", "$VAR123", "VAR123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVarName(tt.input)
			if got != tt.want {
				t.Errorf("extractVarName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpandVariablesCircular(t *testing.T) {
	// Test circular reference handling
	vars := map[string]string{
		"A": "$B",
		"B": "$A",
	}

	input := "$A"
	// Should not infinite loop, will expand once and stop
	got := ExpandVariables(input, vars)

	// Expected: $A -> $B (one level expansion)
	if got != "$B" {
		t.Errorf("ExpandVariables() with circular ref = %q, want %q", got, "$B")
	}
}

func TestExpandVariablesEmpty(t *testing.T) {
	vars := map[string]string{
		"EMPTY": "",
	}

	input := "$EMPTY/path"
	want := "/path"

	got := ExpandVariables(input, vars)
	if got != want {
		t.Errorf("ExpandVariables() = %q, want %q", got, want)
	}
}

func TestExpandVariablesComplex(t *testing.T) {
	vars := map[string]string{
		"USER":    "alice",
		"PROJECT": "myapp",
		"ENV":     "staging",
	}

	input := "/home/$USER/projects/${PROJECT}/config-${ENV}.yml"
	want := "/home/alice/projects/myapp/config-staging.yml"

	got := ExpandVariables(input, vars)
	if got != want {
		t.Errorf("ExpandVariables() = %q, want %q", got, want)
	}
}
