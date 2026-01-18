package dotfiles

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DiffResult represents a comparison between a repository file and its linked target.
type DiffResult struct {
	Source  string
	Target  string
	Status  string // "identical", "different", "missing", "not_linked", "error"
	Diff    string // Unified diff output if different
	Message string
}

// Diff compares files in the repository with their linked targets.
func Diff(repoPath, target string) ([]DiffResult, error) {
	var results []DiffResult

	// Get absolute paths
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repository path: %w", err)
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Walk repository to find files
	err = filepath.Walk(absRepo, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == absRepo || info.IsDir() {
			return nil
		}

		// Skip hidden files
		if filepath.Base(path)[0] == '.' && path != absRepo {
			return nil
		}

		// Calculate target path
		relPath, err := filepath.Rel(absRepo, path)
		if err != nil {
			return err
		}

		targetName := relPath
		if targetName[0] != '.' {
			targetName = "." + targetName
		}
		targetPath := filepath.Join(absTarget, targetName)

		result := diffFile(path, targetPath)
		results = append(results, result)

		return nil
	})

	if err != nil {
		return results, fmt.Errorf("failed to walk repository: %w", err)
	}

	return results, nil
}

// diffFile compares a single repository file with its target.
func diffFile(source, target string) DiffResult {
	result := DiffResult{
		Source: source,
		Target: target,
	}

	// Check if target exists
	targetInfo, err := os.Lstat(target)
	if os.IsNotExist(err) {
		result.Status = "missing"
		result.Message = "target file does not exist"
		return result
	}
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed to stat target: %v", err)
		return result
	}

	// Check if target is a symlink pointing to source
	if targetInfo.Mode()&os.ModeSymlink != 0 {
		linkDest, err := os.Readlink(target)
		if err == nil && linkDest == source {
			// It's our symlink - they're identical by definition
			result.Status = "identical"
			result.Message = "linked and identical"
			return result
		}
	}

	// Not linked or points elsewhere - compare contents
	sourceContent, err := os.ReadFile(source)
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed to read source: %v", err)
		return result
	}

	targetContent, err := os.ReadFile(target)
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed to read target: %v", err)
		return result
	}

	// Compare contents
	if bytes.Equal(sourceContent, targetContent) {
		result.Status = "identical"
		result.Message = "content is identical (not linked)"
		return result
	}

	// Files are different - generate diff
	result.Status = "different"
	result.Message = "files differ"
	result.Diff = generateDiff(source, target, sourceContent, targetContent)

	return result
}

// generateDiff creates a unified diff between two files.
func generateDiff(sourcePath, targetPath string, sourceContent, targetContent []byte) string {
	// Try using system diff command if available
	cmd := exec.Command("diff", "-u", targetPath, sourcePath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// diff returns non-zero when files differ, which is expected
		// Check if it's because files differ (exit code 1) or an error (other codes)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return string(output)
		}
		// Fall back to simple comparison
		return fmt.Sprintf("Files differ (system diff unavailable: %v)", err)
	}

	return string(output)
}

// Status checks the status of all dotfiles from a repository.
type StatusResult struct {
	Source  string
	Target  string
	Status  string // "linked", "unlinked", "missing", "conflict", "error"
	Message string
}

// Status checks the state of dotfiles.
func Status(repoPath, target string) ([]StatusResult, error) {
	var results []StatusResult

	// Get absolute paths
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repository path: %w", err)
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Walk repository
	err = filepath.Walk(absRepo, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == absRepo || info.IsDir() {
			return nil
		}

		// Skip hidden files
		if filepath.Base(path)[0] == '.' && path != absRepo {
			return nil
		}

		// Calculate target path
		relPath, err := filepath.Rel(absRepo, path)
		if err != nil {
			return err
		}

		targetName := relPath
		if targetName[0] != '.' {
			targetName = "." + targetName
		}
		targetPath := filepath.Join(absTarget, targetName)

		result := checkStatus(path, targetPath)
		results = append(results, result)

		return nil
	})

	if err != nil {
		return results, fmt.Errorf("failed to walk repository: %w", err)
	}

	return results, nil
}

// checkStatus checks the status of a single file.
func checkStatus(source, target string) StatusResult {
	result := StatusResult{
		Source: source,
		Target: target,
	}

	// Check if target exists
	targetInfo, err := os.Lstat(target)
	if os.IsNotExist(err) {
		result.Status = "unlinked"
		result.Message = "target does not exist"
		return result
	}
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed to stat: %v", err)
		return result
	}

	// Check if it's a symlink
	if targetInfo.Mode()&os.ModeSymlink == 0 {
		result.Status = "conflict"
		result.Message = "target exists but is not a symlink"
		return result
	}

	// Check where the symlink points
	linkDest, err := os.Readlink(target)
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed to read symlink: %v", err)
		return result
	}

	if linkDest != source {
		result.Status = "conflict"
		result.Message = fmt.Sprintf("symlink points to: %s", linkDest)
		return result
	}

	result.Status = "linked"
	result.Message = "correctly linked"
	return result
}
