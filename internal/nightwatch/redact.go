package nightwatch

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/redact"
)

var (
	redactOnly       string
	redactExcept     string
	redactMask       bool
	redactHash       bool
	redactCustom     string
	redactCustomName string
)

var redactCmd = &cobra.Command{
	Use:   "redact [string]",
	Short: "Redact PII and secrets from text",
	Long: `Redact personally identifiable information (PII) and secrets from text.

Reads from argument or stdin if no argument provided.

Patterns: EMAIL, PHONE, IP, CREDIT_CARD, UUID, NAME

Examples:
    nightwatch redact "Contact john@example.com or 555-123-4567"
    nightwatch redact --mask "Email: user@example.com"
    echo "john@example.com" | nightwatch redact
    cat file.log | nightwatch redact --only EMAIL,PHONE`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRedactString,
}

func init() {
	rootCmd.AddCommand(redactCmd)

	redactCmd.PersistentFlags().StringVar(&redactOnly, "only", "", "Comma-separated list of pattern types to use")
	redactCmd.PersistentFlags().StringVar(&redactExcept, "except", "", "Comma-separated list of pattern types to exclude")
	redactCmd.PersistentFlags().BoolVar(&redactMask, "mask", false, "Partial masking instead of full replacement")
	redactCmd.PersistentFlags().BoolVar(&redactHash, "hash", false, "Replace with stable hash instead of type name")
	redactCmd.PersistentFlags().StringVar(&redactCustom, "custom", "", "Custom regex pattern to match")
	redactCmd.PersistentFlags().StringVar(&redactCustomName, "custom-name", "", "Replacement name for custom pattern")
}

func buildRedactOptions() (redact.Options, error) {
	opts := redact.Options{}

	// Parse --only
	if redactOnly != "" {
		for _, t := range strings.Split(redactOnly, ",") {
			t = strings.TrimSpace(strings.ToUpper(t))
			if t != "" {
				opts.Only = append(opts.Only, redact.PatternType(t))
			}
		}
	}

	// Parse --except
	if redactExcept != "" {
		for _, t := range strings.Split(redactExcept, ",") {
			t = strings.TrimSpace(strings.ToUpper(t))
			if t != "" {
				opts.Except = append(opts.Except, redact.PatternType(t))
			}
		}
	}

	// Mode
	if redactMask && redactHash {
		return opts, fmt.Errorf("--mask and --hash are mutually exclusive")
	}
	if redactMask {
		opts.Mode = redact.ModeMask
	} else if redactHash {
		opts.Mode = redact.ModeHash
	}

	// Custom pattern
	if redactCustom != "" {
		re, err := regexp.Compile(redactCustom)
		if err != nil {
			return opts, fmt.Errorf("invalid --custom regex: %w", err)
		}
		opts.CustomRegex = re
		opts.CustomName = redactCustomName
	} else if redactCustomName != "" {
		return opts, fmt.Errorf("--custom-name requires --custom")
	}

	return opts, nil
}

func newRedactorFromFlags() (*redact.Redactor, error) {
	opts, err := buildRedactOptions()
	if err != nil {
		return nil, err
	}
	return redact.NewRedactor(opts)
}

func walkFiles(root, pattern string, fn func(path string, info os.FileInfo) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if pattern != "" {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return fmt.Errorf("invalid pattern: %w", err)
			}
			if !matched {
				return nil
			}
		}
		return fn(path, info)
	})
}

func runRedactString(cmd *cobra.Command, args []string) error {
	r, err := newRedactorFromFlags()
	if err != nil {
		return err
	}

	// If no args, read from stdin
	if len(args) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			return r.RedactStream(os.Stdin, os.Stdout)
		}
		return fmt.Errorf("no input provided (pass string argument or pipe to stdin)")
	}

	input := strings.Join(args, " ")
	result := r.Redact(input)
	fmt.Fprintln(os.Stdout, result)
	return nil
}

var stdinCmd = &cobra.Command{
	Use:   "stdin",
	Short: "Redact PII and secrets from stdin",
	Long: `Read from stdin, redact sensitive data, and write to stdout.

Examples:
    cat server.log | nightwatch redact stdin
    kubectl logs pod | nightwatch redact stdin --extended`,
	Args: cobra.NoArgs,
	RunE: runRedactStdin,
}

func init() {
	redactCmd.AddCommand(stdinCmd)
}

func runRedactStdin(cmd *cobra.Command, args []string) error {
	r, err := newRedactorFromFlags()
	if err != nil {
		return err
	}

	return r.RedactStream(os.Stdin, os.Stdout)
}

var redactInPlace bool

var fileCmd = &cobra.Command{
	Use:   "file <path>",
	Short: "Redact PII and secrets from a file",
	Long: `Read a file, redact sensitive data, and write to stdout (or in-place with --in-place).

Examples:
    nightwatch redact file server.log > clean.log
    nightwatch redact file server.log --in-place`,
	Args: cobra.ExactArgs(1),
	RunE: runRedactFile,
}

func init() {
	redactCmd.AddCommand(fileCmd)
	fileCmd.Flags().BoolVar(&redactInPlace, "in-place", false, "Modify file in-place (creates .bak backup)")
}

func runRedactFile(cmd *cobra.Command, args []string) error {
	path := args[0]

	r, err := newRedactorFromFlags()
	if err != nil {
		return err
	}

	if redactInPlace {
		return redactFileInPlace(r, path)
	}

	// Read file and output to stdout
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return r.RedactStream(f, os.Stdout)
}

func redactFileInPlace(r *redact.Redactor, path string) error {
	// Create backup
	backupPath := path + ".bak"
	if err := copyFile(path, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Open backup for reading
	backup, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup: %w", err)
	}
	defer backup.Close()

	// Create temp file in same directory for atomic replacement
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".redact-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	// Redact to temp file
	if err := r.RedactStream(backup, tmp); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("redaction failed: %w", err)
	}
	tmp.Close()

	// Copy permissions from original
	info, err := os.Stat(backupPath)
	if err == nil {
		os.Chmod(tmpPath, info.Mode())
	}

	// Atomic replace
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		// Restore from backup
		copyFile(backupPath, path)
		return fmt.Errorf("failed to replace file: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

var (
	dirOutput  string
	dirPattern string
)

var dirCmd = &cobra.Command{
	Use:   "dir <path>",
	Short: "Redact PII and secrets from files in a directory",
	Long: `Recursively process files in a directory, redacting sensitive data.

Requires --output to specify where to write redacted files.
Use --pattern to filter which files to process.

Examples:
    nightwatch redact dir ./logs --output ./logs-clean
    nightwatch redact dir ./logs --output ./clean --pattern "*.log"
    nightwatch redact dir ./src --output ./src-clean --extended`,
	Args: cobra.ExactArgs(1),
	RunE: runRedactDir,
}

func init() {
	redactCmd.AddCommand(dirCmd)
	dirCmd.Flags().StringVar(&dirOutput, "output", "", "Output directory (required)")
	dirCmd.Flags().StringVar(&dirPattern, "pattern", "", "Glob pattern to filter files (e.g., *.log)")
	dirCmd.MarkFlagRequired("output")
}

func runRedactDir(cmd *cobra.Command, args []string) error {
	srcDir := args[0]

	r, err := newRedactorFromFlags()
	if err != nil {
		return err
	}

	// Ensure output directory exists
	if err := os.MkdirAll(dirOutput, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return walkFiles(srcDir, dirPattern, func(path string, _ os.FileInfo) error {
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		outPath := filepath.Join(dirOutput, relPath)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		return redactFileTo(r, path, outPath)
	})
}

func redactFileTo(r *redact.Redactor, src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", dst, err)
	}
	defer out.Close()

	if err := r.RedactStream(in, out); err != nil {
		return fmt.Errorf("failed to redact %s: %w", src, err)
	}

	// Copy permissions
	info, err := os.Stat(src)
	if err == nil {
		os.Chmod(dst, info.Mode())
	}

	return nil
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Audit for PII and secrets without modifying content",
	Long: `Scan for sensitive data and report findings without making changes.

Examples:
    nightwatch redact check file server.log
    nightwatch redact check dir ./src --extended`,
}

func init() {
	redactCmd.AddCommand(checkCmd)
}

var checkFileCmd = &cobra.Command{
	Use:   "file <path>",
	Short: "Audit a file for PII and secrets",
	Args:  cobra.ExactArgs(1),
	RunE:  runCheckFile,
}

func init() {
	checkCmd.AddCommand(checkFileCmd)
}

func runCheckFile(cmd *cobra.Command, args []string) error {
	path := args[0]

	r, err := newRedactorFromFlags()
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	report, err := r.Check(f)
	if err != nil {
		return err
	}

	printReport(path, report)
	return nil
}

var checkDirCmd = &cobra.Command{
	Use:   "dir <path>",
	Short: "Audit a directory for PII and secrets",
	Args:  cobra.ExactArgs(1),
	RunE:  runCheckDir,
}

func init() {
	checkCmd.AddCommand(checkDirCmd)
	checkDirCmd.Flags().StringVar(&dirPattern, "pattern", "", "Glob pattern to filter files")
}

func runCheckDir(cmd *cobra.Command, args []string) error {
	srcDir := args[0]

	r, err := newRedactorFromFlags()
	if err != nil {
		return err
	}

	totalCounts := make(map[redact.PatternType]int)
	fileCount := 0
	filesWithFindings := 0

	err = walkFiles(srcDir, dirPattern, func(path string, _ os.FileInfo) error {
		fileCount++

		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not open %s: %v\n", path, err)
			return nil
		}

		report, err := r.Check(f)
		_ = f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not check %s: %v\n", path, err)
			return nil
		}

		if report.TotalFindings() > 0 {
			filesWithFindings++
			printReport(path, report)
		}

		for t, c := range report.Counts {
			totalCounts[t] += c
		}

		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Files scanned: %d\n", fileCount)
	fmt.Printf("Files with findings: %d\n", filesWithFindings)

	if len(totalCounts) > 0 {
		fmt.Printf("\nTotal findings by type:\n")
		for t, c := range totalCounts {
			fmt.Printf("  %s: %d\n", t, c)
		}
	}

	return nil
}

func printReport(path string, report *redact.Report) {
	if report.TotalFindings() == 0 {
		fmt.Printf("%s: no findings\n", path)
		return
	}

	fmt.Printf("%s: %d finding(s)\n", path, report.TotalFindings())
	for t, c := range report.Counts {
		fmt.Printf("  %s: %d\n", t, c)
	}
}
