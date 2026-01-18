package dotfiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatus(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create links first
	opts := LinkOptions{
		RepoPath: repoPath,
		Target:   targetPath,
		DryRun:   false,
		Force:    false,
	}

	_, err := Link(opts)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}

	// Check status
	results, err := Status(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Status != "linked" {
			t.Errorf("expected status 'linked', got '%s' for %s", r.Status, r.Target)
		}
	}
}

func TestStatusUnlinked(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Check status without linking
	results, err := Status(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	for _, r := range results {
		if r.Status != "unlinked" {
			t.Errorf("expected status 'unlinked', got '%s'", r.Status)
		}
	}
}

func TestStatusConflict(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a regular file at target location
	conflictFile := filepath.Join(targetPath, ".bashrc")
	if err := os.WriteFile(conflictFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create conflict file: %v", err)
	}

	results, err := Status(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	// Find result for bashrc
	var bashrcResult *StatusResult
	for _, r := range results {
		if filepath.Base(r.Target) == ".bashrc" {
			bashrcResult = &r
			break
		}
	}

	if bashrcResult.Status != "conflict" {
		t.Errorf("expected status 'conflict', got '%s'", bashrcResult.Status)
	}
}

func TestStatusWrongLink(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a symlink pointing to wrong location
	wrongTarget := filepath.Join(targetPath, ".bashrc")
	if err := os.Symlink("/tmp/elsewhere", wrongTarget); err != nil {
		t.Fatalf("failed to create wrong symlink: %v", err)
	}

	results, err := Status(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	// Find result for bashrc
	var bashrcResult *StatusResult
	for _, r := range results {
		if filepath.Base(r.Target) == ".bashrc" {
			bashrcResult = &r
			break
		}
	}

	if bashrcResult.Status != "conflict" {
		t.Errorf("expected status 'conflict' for wrong link, got '%s'", bashrcResult.Status)
	}
}

func TestDiff(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create links first
	opts := LinkOptions{
		RepoPath: repoPath,
		Target:   targetPath,
		DryRun:   false,
		Force:    false,
	}

	_, err := Link(opts)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}

	// Check diff
	results, err := Diff(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	for _, r := range results {
		if r.Status != "identical" {
			t.Errorf("expected status 'identical', got '%s' for %s", r.Status, r.Target)
		}
	}
}

func TestDiffMissing(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Check diff without linking
	results, err := Diff(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	for _, r := range results {
		if r.Status != "missing" {
			t.Errorf("expected status 'missing', got '%s'", r.Status)
		}
	}
}

func TestDiffDifferent(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a file with different content at target
	diffFile := filepath.Join(targetPath, ".bashrc")
	if err := os.WriteFile(diffFile, []byte("different content"), 0644); err != nil {
		t.Fatalf("failed to create diff file: %v", err)
	}

	results, err := Diff(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	// Find result for bashrc
	var bashrcResult *DiffResult
	for _, r := range results {
		if filepath.Base(r.Target) == ".bashrc" {
			bashrcResult = &r
			break
		}
	}

	if bashrcResult == nil {
		t.Fatal("bashrc result not found")
	}

	if bashrcResult.Status != "different" {
		t.Errorf("expected status 'different', got '%s'", bashrcResult.Status)
	}

	// Check that diff output is present (if system diff is available)
	if bashrcResult.Diff == "" {
		t.Error("expected diff output to be non-empty")
	}
}

func TestDiffIdenticalButNotLinked(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Copy file content without linking
	sourceFile := filepath.Join(repoPath, "bashrc")
	targetFile := filepath.Join(targetPath, ".bashrc")

	content, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("failed to read source: %v", err)
	}

	if err := os.WriteFile(targetFile, content, 0644); err != nil {
		t.Fatalf("failed to write target: %v", err)
	}

	results, err := Diff(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	// Find result for bashrc
	var bashrcResult *DiffResult
	for _, r := range results {
		if filepath.Base(r.Target) == ".bashrc" {
			bashrcResult = &r
			break
		}
	}

	if bashrcResult.Status != "identical" {
		t.Errorf("expected status 'identical', got '%s'", bashrcResult.Status)
	}

	if !strings.Contains(bashrcResult.Message, "not linked") {
		t.Error("message should indicate file is not linked")
	}
}

// Integration test: full workflow
func TestFullWorkflow(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Step 1: Link dotfiles
	linkOpts := LinkOptions{
		RepoPath: repoPath,
		Target:   targetPath,
		DryRun:   false,
		Force:    false,
	}

	linkResults, err := Link(linkOpts)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}

	createdCount := 0
	for _, r := range linkResults {
		if r.Status == "created" {
			createdCount++
		}
	}
	if createdCount != 3 {
		t.Errorf("expected 3 files created, got %d", createdCount)
	}

	// Step 2: Check status
	statusResults, err := Status(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}

	linkedCount := 0
	for _, r := range statusResults {
		if r.Status == "linked" {
			linkedCount++
		}
	}
	if linkedCount != 3 {
		t.Errorf("expected 3 files linked, got %d", linkedCount)
	}

	// Step 3: Check diff
	diffResults, err := Diff(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	identicalCount := 0
	for _, r := range diffResults {
		if r.Status == "identical" {
			identicalCount++
		}
	}
	if identicalCount != 3 {
		t.Errorf("expected 3 files identical, got %d", identicalCount)
	}

	// Step 4: Unlink
	unlinkResults, err := Unlink(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Unlink() error = %v", err)
	}

	removedCount := 0
	for _, r := range unlinkResults {
		if r.Status == "removed" {
			removedCount++
		}
	}
	if removedCount != 3 {
		t.Errorf("expected 3 files removed, got %d", removedCount)
	}

	// Step 5: Verify all symlinks are gone
	for _, r := range unlinkResults {
		_, err := os.Lstat(r.Target)
		if !os.IsNotExist(err) {
			t.Errorf("symlink still exists after unlink: %s", r.Target)
		}
	}
}
