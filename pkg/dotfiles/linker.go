// Package dotfiles provides symlink management for dotfile repositories.
package dotfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LinkResult represents the result of a linking operation.
type LinkResult struct {
	Source  string
	Target  string
	Status  string // "created", "exists", "skipped", "error"
	Message string
}

// LinkOptions configures the linking behavior.
type LinkOptions struct {
	RepoPath string
	Target   string
	DryRun   bool
	Force    bool
}

// Link creates symlinks from the target directory to files in the repository.
// It returns a list of results for each file processed.
func Link(opts LinkOptions) ([]LinkResult, error) {
	var results []LinkResult

	// Validate repository path
	repoInfo, err := os.Stat(opts.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("repository path does not exist: %w", err)
	}
	if !repoInfo.IsDir() {
		return nil, fmt.Errorf("repository path is not a directory: %s", opts.RepoPath)
	}

	// Get absolute paths
	absRepo, err := filepath.Abs(opts.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repository path: %w", err)
	}

	absTarget, err := filepath.Abs(opts.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Walk repository and link dotfiles
	err = filepath.Walk(absRepo, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == absRepo {
			return nil
		}

		// Skip hidden files/directories in subdirectories (like .git)
		if filepath.Base(path)[0] == '.' && path != absRepo {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process files
		if info.IsDir() {
			return nil
		}

		// Calculate relative path from repo
		relPath, err := filepath.Rel(absRepo, path)
		if err != nil {
			return err
		}

		// Target path (prepend dot if not already present)
		targetName := relPath
		if targetName[0] != '.' {
			targetName = "." + targetName
		}
		targetPath := filepath.Join(absTarget, targetName)

		result := linkFile(path, targetPath, opts)
		results = append(results, result)

		return nil
	})

	if err != nil {
		return results, fmt.Errorf("failed to walk repository: %w", err)
	}

	return results, nil
}

// linkFile creates a single symlink.
func linkFile(source, target string, opts LinkOptions) LinkResult {
	result := LinkResult{
		Source: source,
		Target: target,
	}

	// Check if target already exists
	targetInfo, err := os.Lstat(target)
	if err == nil {
		// Target exists
		if targetInfo.Mode()&os.ModeSymlink != 0 {
			// It's a symlink - check if it points to our source
			linkDest, err := os.Readlink(target)
			if err == nil && linkDest == source {
				result.Status = "exists"
				result.Message = "already linked correctly"
				return result
			}
		}

		// Target exists but is not our symlink
		if !opts.Force {
			result.Status = "skipped"
			result.Message = "target exists (use --force to overwrite)"
			return result
		}

		// Force mode - backup and overwrite
		if !opts.DryRun {
			backupPath := target + ".bak"
			if err := os.Rename(target, backupPath); err != nil {
				result.Status = "error"
				result.Message = fmt.Sprintf("failed to backup: %v", err)
				return result
			}
			result.Message = fmt.Sprintf("backed up to %s", backupPath)
		}
	}

	// Create parent directories if needed
	if !opts.DryRun {
		targetDir := filepath.Dir(target)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("failed to create directory: %v", err)
			return result
		}
	}

	// Create symlink
	if opts.DryRun {
		result.Status = "would create"
		result.Message = "dry-run mode"
	} else {
		if err := os.Symlink(source, target); err != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("failed to create symlink: %v", err)
		} else {
			result.Status = "created"
			result.Message = "symlink created"
		}
	}

	return result
}

// Unlink removes symlinks created from the repository.
func Unlink(repoPath, target string) ([]LinkResult, error) {
	var results []LinkResult

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

		result := unlinkFile(path, targetPath)
		results = append(results, result)

		return nil
	})

	if err != nil {
		return results, fmt.Errorf("failed to walk repository: %w", err)
	}

	return results, nil
}

// unlinkFile removes a single symlink if it points to the source.
func unlinkFile(source, target string) LinkResult {
	result := LinkResult{
		Source: source,
		Target: target,
	}

	// Check if target exists
	targetInfo, err := os.Lstat(target)
	if os.IsNotExist(err) {
		result.Status = "not found"
		result.Message = "symlink does not exist"
		return result
	}
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed to stat: %v", err)
		return result
	}

	// Check if it's a symlink
	if targetInfo.Mode()&os.ModeSymlink == 0 {
		result.Status = "skipped"
		result.Message = "not a symlink"
		return result
	}

	// Check if it points to our source
	linkDest, err := os.Readlink(target)
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed to read symlink: %v", err)
		return result
	}

	if linkDest != source {
		result.Status = "skipped"
		result.Message = fmt.Sprintf("points to different location: %s", linkDest)
		return result
	}

	// Remove the symlink
	if err := os.Remove(target); err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed to remove: %v", err)
	} else {
		result.Status = "removed"
		result.Message = "symlink removed"
	}

	return result
}

// LinkSpec represents metadata about a created link for tracking.
type LinkSpec struct {
	Source   string
	Target   string
	LinkedAt time.Time
}
