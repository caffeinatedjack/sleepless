package envvar

import (
	"os"
	"strings"
	"testing"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
		want     Shell
	}{
		{"bash", "/bin/bash", ShellBash},
		{"zsh", "/usr/bin/zsh", ShellZsh},
		{"fish", "/usr/bin/fish", ShellFish},
		{"pwsh", "/usr/bin/pwsh", ShellPowerShell},
		{"powershell", "/usr/bin/powershell", ShellPowerShell},
		{"unknown", "/bin/unknown", ShellBash},
		{"empty", "", ShellBash},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldShell := os.Getenv("SHELL")
			defer os.Setenv("SHELL", oldShell)

			os.Setenv("SHELL", tt.shellEnv)
			got := DetectShell()

			if got != tt.want {
				t.Errorf("DetectShell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseShell(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Shell
		wantErr bool
	}{
		{"bash", "bash", ShellBash, false},
		{"zsh", "zsh", ShellZsh, false},
		{"fish", "fish", ShellFish, false},
		{"powershell", "powershell", ShellPowerShell, false},
		{"pwsh", "pwsh", ShellPowerShell, false},
		{"uppercase", "BASH", ShellBash, false},
		{"invalid", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseShell(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseShell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseShell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExportBash(t *testing.T) {
	vars := map[string]string{
		"EDITOR": "vim",
		"PATH":   "/usr/local/bin:$PATH",
	}

	output, err := Export(ShellBash, vars)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Check both variables are present
	if !strings.Contains(output, `export EDITOR="vim"`) {
		t.Errorf("output missing EDITOR export: %s", output)
	}
	if !strings.Contains(output, `export PATH=`) {
		t.Errorf("output missing PATH export: %s", output)
	}
}

func TestExportZsh(t *testing.T) {
	vars := map[string]string{
		"EDITOR": "nvim",
	}

	output, err := Export(ShellZsh, vars)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	if !strings.Contains(output, `export EDITOR="nvim"`) {
		t.Errorf("output missing EDITOR export: %s", output)
	}
}

func TestExportFish(t *testing.T) {
	vars := map[string]string{
		"EDITOR": "vim",
	}

	output, err := Export(ShellFish, vars)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	if !strings.Contains(output, `set -gx EDITOR "vim"`) {
		t.Errorf("output missing EDITOR export: %s", output)
	}
}

func TestExportPowerShell(t *testing.T) {
	vars := map[string]string{
		"EDITOR": "code",
	}

	output, err := Export(ShellPowerShell, vars)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	if !strings.Contains(output, `$env:EDITOR = "code"`) {
		t.Errorf("output missing EDITOR export: %s", output)
	}
}

func TestExportEmpty(t *testing.T) {
	vars := map[string]string{}

	output, err := Export(ShellBash, vars)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	if output != "" {
		t.Errorf("expected empty output, got: %s", output)
	}
}

func TestExportSorted(t *testing.T) {
	vars := map[string]string{
		"ZEBRA": "last",
		"ALPHA": "first",
		"BETA":  "second",
	}

	output, err := Export(ShellBash, vars)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	lines := strings.Split(output, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// Check alphabetical order
	if !strings.Contains(lines[0], "ALPHA") {
		t.Errorf("first line should be ALPHA, got: %s", lines[0])
	}
	if !strings.Contains(lines[1], "BETA") {
		t.Errorf("second line should be BETA, got: %s", lines[1])
	}
	if !strings.Contains(lines[2], "ZEBRA") {
		t.Errorf("third line should be ZEBRA, got: %s", lines[2])
	}
}

func TestExportEscaping(t *testing.T) {
	vars := map[string]string{
		"VAR_WITH_QUOTES": `value with "quotes"`,
	}

	output, err := Export(ShellBash, vars)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// Should escape the inner quotes
	if !strings.Contains(output, `value with \"quotes\"`) {
		t.Errorf("quotes not properly escaped: %s", output)
	}
}

func TestExportPowerShellEscaping(t *testing.T) {
	vars := map[string]string{
		"VAR": `value with "quotes" and backticks`,
	}

	output, err := Export(ShellPowerShell, vars)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// PowerShell should escape quotes with backtick
	if !strings.Contains(output, "`") {
		t.Errorf("PowerShell escaping not applied: %s", output)
	}
}
