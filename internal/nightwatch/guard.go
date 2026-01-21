package nightwatch

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/guard"
	"gitlab.com/caffeinatedjack/sleepless/pkg/redact"
)

var (
	guardJSON     bool
	guardBaseline string
)

var guardCmd = &cobra.Command{
	Use:   "guard",
	Short: "Scan for secrets and PII in code",
	Long: `Guard scans local code for likely secrets and PII, returning non-zero exit codes when findings exist.
Designed for offline use in pre-commit hooks and CI pipelines.

Output is safe-by-default: raw secret values are NEVER printed.

Examples:
    nightwatch guard staged
    nightwatch guard staged --json
    nightwatch guard staged --baseline .nightwatch-baseline.json
    nightwatch guard worktree
    nightwatch guard path src/
    nightwatch guard baseline staged --out .nightwatch-baseline.json`,
}

func init() {
	rootCmd.AddCommand(guardCmd)
}

var guardStagedCmd = &cobra.Command{
	Use:   "staged",
	Short: "Scan staged files (git)",
	Long: `Scan the staged snapshot (files that would be committed).
Requires being in a git repository.

Exit codes:
  0 - No findings
  1 - Findings detected or error occurred`,
	Args: cobra.NoArgs,
	RunE: runGuardStaged,
}

func init() {
	guardCmd.AddCommand(guardStagedCmd)
	guardStagedCmd.Flags().BoolVar(&guardJSON, "json", false, "Output JSON instead of human-readable format")
	guardStagedCmd.Flags().StringVar(&guardBaseline, "baseline", "", "Path to baseline file (suppresses known findings)")
}

var guardWorktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Scan working tree files (git)",
	Long: `Scan the working tree snapshot (current file contents, not staged).
Requires being in a git repository.

Exit codes:
  0 - No findings
  1 - Findings detected or error occurred`,
	Args: cobra.NoArgs,
	RunE: runGuardWorktree,
}

func init() {
	guardCmd.AddCommand(guardWorktreeCmd)
	guardWorktreeCmd.Flags().BoolVar(&guardJSON, "json", false, "Output JSON instead of human-readable format")
	guardWorktreeCmd.Flags().StringVar(&guardBaseline, "baseline", "", "Path to baseline file (suppresses known findings)")
}

var guardPathCmd = &cobra.Command{
	Use:   "path <path>",
	Short: "Scan a file or directory",
	Long: `Scan the provided file or directory for secrets and PII.

Exit codes:
  0 - No findings
  1 - Findings detected or error occurred`,
	Args: cobra.ExactArgs(1),
	RunE: runGuardPath,
}

func init() {
	guardCmd.AddCommand(guardPathCmd)
	guardPathCmd.Flags().BoolVar(&guardJSON, "json", false, "Output JSON instead of human-readable format")
	guardPathCmd.Flags().StringVar(&guardBaseline, "baseline", "", "Path to baseline file (suppresses known findings)")
}

var guardBaselineCmd = &cobra.Command{
	Use:   "baseline [staged|worktree|path <path>]",
	Short: "Generate a baseline file from scan results",
	Long: `Scan a target and create a baseline file containing fingerprints of all findings.
The baseline can be used with --baseline to suppress known findings.

Defaults to scanning staged files if no subcommand is provided.

Examples:
    nightwatch guard baseline staged --out .nightwatch-baseline.json
    nightwatch guard baseline worktree --out baseline.json
    nightwatch guard baseline path src/ --out baseline.json`,
	Args: cobra.MaximumNArgs(2),
	RunE: runGuardBaseline,
}

var baselineOut string

func init() {
	guardCmd.AddCommand(guardBaselineCmd)
	guardBaselineCmd.Flags().StringVar(&baselineOut, "out", "", "Output file path (defaults to stdout)")
}

// runGuardStaged scans staged files in a git repository
func runGuardStaged(cmd *cobra.Command, args []string) error {
	// Check if we're in a git repo
	if !isGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	// Get list of staged files
	stagedFiles, err := getStagedFiles()
	if err != nil {
		return fmt.Errorf("failed to get staged files: %w", err)
	}

	if len(stagedFiles) == 0 {
		if guardJSON {
			output := map[string]interface{}{
				"target":   "staged",
				"findings": []interface{}{},
				"counts":   map[string]int{},
				"total":    0,
			}
			return outputJSON(output)
		}
		fmt.Println("No staged files to scan")
		return nil
	}

	// Scan staged content
	scanner, err := guard.NewScanner()
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	results := &guard.ScanResult{
		Findings: make([]guard.Finding, 0),
		Counts:   make(map[redact.PatternType]int),
		Total:    0,
	}

	for _, file := range stagedFiles {
		content, err := getStagedFileContent(file)
		if err != nil {
			// Skip files we can't read
			continue
		}

		// Write content to temp file for scanning
		tmpFile, err := os.CreateTemp("", "nightwatch-guard-*")
		if err != nil {
			continue
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := tmpFile.Write(content); err != nil {
			tmpFile.Close()
			continue
		}
		tmpFile.Close()

		fileResult, err := scanner.ScanFile(tmpPath)
		if err != nil {
			// Skip files that fail to scan
			continue
		}

		// Update file paths in results to show original file
		for i := range fileResult.Findings {
			fileResult.Findings[i].File = file
		}

		results.Findings = append(results.Findings, fileResult.Findings...)
		results.Total += fileResult.Total
		for t, c := range fileResult.Counts {
			results.Counts[t] += c
		}
	}

	// Apply baseline if provided
	if guardBaseline != "" {
		baseline, err := guard.LoadBaseline(guardBaseline)
		if err != nil {
			return fmt.Errorf("failed to load baseline: %w", err)
		}
		results = guard.ApplyBaseline(results, baseline)
	}

	// Output results
	return outputGuardResults("staged", results)
}

// runGuardWorktree scans working tree files
func runGuardWorktree(cmd *cobra.Command, args []string) error {
	// Check if we're in a git repo
	if !isGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	// Get list of modified files in working tree
	modifiedFiles, err := getModifiedFiles()
	if err != nil {
		return fmt.Errorf("failed to get modified files: %w", err)
	}

	if len(modifiedFiles) == 0 {
		if guardJSON {
			output := map[string]interface{}{
				"target":   "worktree",
				"findings": []interface{}{},
				"counts":   map[string]int{},
				"total":    0,
			}
			return outputJSON(output)
		}
		fmt.Println("No modified files to scan")
		return nil
	}

	// Scan files
	scanner, err := guard.NewScanner()
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	results := &guard.ScanResult{
		Findings: make([]guard.Finding, 0),
		Counts:   make(map[redact.PatternType]int),
		Total:    0,
	}

	for _, file := range modifiedFiles {
		fileResult, err := scanner.ScanFile(file)
		if err != nil {
			// Skip files that fail to scan
			continue
		}

		results.Findings = append(results.Findings, fileResult.Findings...)
		results.Total += fileResult.Total
		for t, c := range fileResult.Counts {
			results.Counts[t] += c
		}
	}

	// Apply baseline if provided
	if guardBaseline != "" {
		baseline, err := guard.LoadBaseline(guardBaseline)
		if err != nil {
			return fmt.Errorf("failed to load baseline: %w", err)
		}
		results = guard.ApplyBaseline(results, baseline)
	}

	// Output results
	return outputGuardResults("worktree", results)
}

// runGuardPath scans a specified path
func runGuardPath(cmd *cobra.Command, args []string) error {
	path := args[0]

	scanner, err := guard.NewScanner()
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	var results *guard.ScanResult

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		results, err = scanner.ScanDirectory(path)
	} else {
		results, err = scanner.ScanFile(path)
	}

	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Apply baseline if provided
	if guardBaseline != "" {
		baseline, err := guard.LoadBaseline(guardBaseline)
		if err != nil {
			return fmt.Errorf("failed to load baseline: %w", err)
		}
		results = guard.ApplyBaseline(results, baseline)
	}

	// Output results
	return outputGuardResults(fmt.Sprintf("path:%s", path), results)
}

// runGuardBaseline generates a baseline file
func runGuardBaseline(cmd *cobra.Command, args []string) error {
	// Determine target
	target := "staged"
	if len(args) > 0 {
		target = args[0]
	}

	var results *guard.ScanResult
	var err error

	scanner, err := guard.NewScanner()
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}

	switch target {
	case "staged":
		if !isGitRepo() {
			return fmt.Errorf("not a git repository")
		}
		stagedFiles, err := getStagedFiles()
		if err != nil {
			return fmt.Errorf("failed to get staged files: %w", err)
		}

		results = &guard.ScanResult{
			Findings: make([]guard.Finding, 0),
			Counts:   make(map[redact.PatternType]int),
			Total:    0,
		}

		for _, file := range stagedFiles {
			content, err := getStagedFileContent(file)
			if err != nil {
				continue
			}

			tmpFile, err := os.CreateTemp("", "nightwatch-guard-*")
			if err != nil {
				continue
			}
			tmpPath := tmpFile.Name()
			defer os.Remove(tmpPath)

			if _, err := tmpFile.Write(content); err != nil {
				tmpFile.Close()
				continue
			}
			tmpFile.Close()

			fileResult, err := scanner.ScanFile(tmpPath)
			if err != nil {
				continue
			}

			for i := range fileResult.Findings {
				fileResult.Findings[i].File = file
			}

			results.Findings = append(results.Findings, fileResult.Findings...)
			results.Total += fileResult.Total
			for t, c := range fileResult.Counts {
				results.Counts[t] += c
			}
		}

	case "worktree":
		if !isGitRepo() {
			return fmt.Errorf("not a git repository")
		}
		modifiedFiles, err := getModifiedFiles()
		if err != nil {
			return fmt.Errorf("failed to get modified files: %w", err)
		}

		results = &guard.ScanResult{
			Findings: make([]guard.Finding, 0),
			Counts:   make(map[redact.PatternType]int),
			Total:    0,
		}

		for _, file := range modifiedFiles {
			fileResult, err := scanner.ScanFile(file)
			if err != nil {
				continue
			}

			results.Findings = append(results.Findings, fileResult.Findings...)
			results.Total += fileResult.Total
			for t, c := range fileResult.Counts {
				results.Counts[t] += c
			}
		}

	case "path":
		if len(args) < 2 {
			return fmt.Errorf("path target requires a path argument")
		}
		path := args[1]

		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat path: %w", err)
		}

		if info.IsDir() {
			results, err = scanner.ScanDirectory(path)
		} else {
			results, err = scanner.ScanFile(path)
		}

		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

	default:
		return fmt.Errorf("unknown target: %s (use staged, worktree, or path)", target)
	}

	// Create baseline
	baseline := guard.CreateBaseline(results)

	// Save baseline
	if err := guard.SaveBaseline(baseline, baselineOut); err != nil {
		return fmt.Errorf("failed to save baseline: %w", err)
	}

	// Print summary to stderr
	fmt.Fprintf(os.Stderr, "Baseline created with %d unique fingerprints\n", len(baseline.Fingerprints))

	return nil
}

// Helper functions

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func getStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=AM")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func getStagedFileContent(file string) ([]byte, error) {
	cmd := exec.Command("git", "show", ":"+file)
	return cmd.Output()
}

func getModifiedFiles() ([]string, error) {
	// Get files modified in working tree relative to HEAD
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Check if file exists in working tree
			if _, err := os.Stat(line); err == nil {
				files = append(files, line)
			}
		}
	}
	return files, nil
}

func outputGuardResults(target string, results *guard.ScanResult) error {
	if guardJSON {
		output := map[string]interface{}{
			"target":   target,
			"findings": results.Findings,
			"counts":   results.Counts,
			"total":    results.Total,
		}
		if err := outputJSON(output); err != nil {
			return err
		}
	} else {
		printGuardResults(target, results)
	}

	// Exit code: 0 if no findings, 1 if findings exist
	if results.Total > 0 {
		os.Exit(1)
	}

	return nil
}

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func printGuardResults(target string, results *guard.ScanResult) {
	fmt.Printf("Target: %s\n", target)
	fmt.Printf("Findings: %d\n\n", results.Total)

	if results.Total == 0 {
		fmt.Println("✓ No secrets or PII detected")
		return
	}

	// Group findings by file
	fileFindings := make(map[string][]guard.Finding)
	for _, f := range results.Findings {
		fileFindings[f.File] = append(fileFindings[f.File], f)
	}

	// Print findings by file
	for file, findings := range fileFindings {
		fmt.Printf("File: %s\n", file)
		for _, f := range findings {
			fmt.Printf("  Line %d, Column %d: %s\n", f.Line, f.Column, f.Type)
			if f.Excerpt != "" {
				fmt.Printf("    %s\n", f.Excerpt)
			}
		}
		fmt.Println()
	}

	// Print summary
	if len(results.Counts) > 0 {
		fmt.Println("Summary by type:")
		for t, c := range results.Counts {
			fmt.Printf("  %s: %d\n", t, c)
		}
	}
}
