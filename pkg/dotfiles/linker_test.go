package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) (repoPath, targetPath string, cleanup func()) {
	// Create temporary directories
	tmpDir := t.TempDir()
	repoPath = filepath.Join(tmpDir, "dotfiles")
	targetPath = filepath.Join(tmpDir, "home")

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}

	// Create sample dotfiles in repo
	files := map[string]string{
		"bashrc":    "# bashrc content",
		"vimrc":     "set number",
		"gitconfig": "[user]\nname = test",
	}

	for name, content := range files {
		path := filepath.Join(repoPath, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", name, err)
		}
	}

	cleanup = func() {
		// t.TempDir() handles cleanup automatically
	}

	return repoPath, targetPath, cleanup
}

func TestLink(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	opts := LinkOptions{
		RepoPath: repoPath,
		Target:   targetPath,
		DryRun:   false,
		Force:    false,
	}

	results, err := Link(opts)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Check that symlinks were created
	for _, r := range results {
		if r.Status != "created" {
			t.Errorf("expected status 'created', got '%s' for %s", r.Status, r.Target)
		}

		// Verify symlink exists and points to correct source
		linkDest, err := os.Readlink(r.Target)
		if err != nil {
			t.Errorf("failed to read symlink %s: %v", r.Target, err)
			continue
		}

		if linkDest != r.Source {
			t.Errorf("symlink %s points to %s, expected %s", r.Target, linkDest, r.Source)
		}
	}
}

func TestLinkDryRun(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	opts := LinkOptions{
		RepoPath: repoPath,
		Target:   targetPath,
		DryRun:   true,
		Force:    false,
	}

	results, err := Link(opts)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}

	// Check that no symlinks were actually created
	for _, r := range results {
		if r.Status != "would create" {
			t.Errorf("expected status 'would create', got '%s'", r.Status)
		}

		_, err := os.Lstat(r.Target)
		if !os.IsNotExist(err) {
			t.Errorf("file should not exist in dry-run mode: %s", r.Target)
		}
	}
}

func TestLinkExistingFile(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create an existing file at target location
	existingFile := filepath.Join(targetPath, ".bashrc")
	if err := os.WriteFile(existingFile, []byte("existing content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	opts := LinkOptions{
		RepoPath: repoPath,
		Target:   targetPath,
		DryRun:   false,
		Force:    false,
	}

	results, err := Link(opts)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}

	// Find result for bashrc
	var bashrcResult *LinkResult
	for _, r := range results {
		if filepath.Base(r.Target) == ".bashrc" {
			bashrcResult = &r
			break
		}
	}

	if bashrcResult == nil {
		t.Fatal("bashrc result not found")
	}

	if bashrcResult.Status != "skipped" {
		t.Errorf("expected status 'skipped', got '%s'", bashrcResult.Status)
	}

	// Verify file was not overwritten
	content, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("failed to read existing file: %v", err)
	}
	if string(content) != "existing content" {
		t.Error("existing file was modified")
	}
}

func TestLinkForceOverwrite(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create an existing file
	existingFile := filepath.Join(targetPath, ".bashrc")
	if err := os.WriteFile(existingFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	opts := LinkOptions{
		RepoPath: repoPath,
		Target:   targetPath,
		DryRun:   false,
		Force:    true,
	}

	results, err := Link(opts)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}

	// Find result for bashrc
	var bashrcResult *LinkResult
	for _, r := range results {
		if filepath.Base(r.Target) == ".bashrc" {
			bashrcResult = &r
			break
		}
	}

	if bashrcResult.Status != "created" {
		t.Errorf("expected status 'created', got '%s'", bashrcResult.Status)
	}

	// Verify backup was created
	backupFile := existingFile + ".bak"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("backup file was not created")
	} else {
		content, _ := os.ReadFile(backupFile)
		if string(content) != "old content" {
			t.Error("backup does not contain original content")
		}
	}

	// Verify symlink was created
	linkDest, err := os.Readlink(existingFile)
	if err != nil {
		t.Errorf("symlink not created: %v", err)
	} else if linkDest != bashrcResult.Source {
		t.Errorf("symlink points to wrong location: %s", linkDest)
	}
}

func TestLinkAlreadyLinked(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// First linking
	opts := LinkOptions{
		RepoPath: repoPath,
		Target:   targetPath,
		DryRun:   false,
		Force:    false,
	}

	_, err := Link(opts)
	if err != nil {
		t.Fatalf("first Link() error = %v", err)
	}

	// Second linking (should detect existing correct links)
	results, err := Link(opts)
	if err != nil {
		t.Fatalf("second Link() error = %v", err)
	}

	for _, r := range results {
		if r.Status != "exists" {
			t.Errorf("expected status 'exists' for already linked file, got '%s'", r.Status)
		}
	}
}

func TestUnlink(t *testing.T) {
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

	// Now unlink
	results, err := Unlink(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Unlink() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Status != "removed" {
			t.Errorf("expected status 'removed', got '%s' for %s", r.Status, r.Target)
		}

		// Verify symlink no longer exists
		_, err := os.Lstat(r.Target)
		if !os.IsNotExist(err) {
			t.Errorf("symlink still exists: %s", r.Target)
		}
	}
}

func TestUnlinkNonExistent(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Try to unlink without creating links first
	results, err := Unlink(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Unlink() error = %v", err)
	}

	for _, r := range results {
		if r.Status != "not found" {
			t.Errorf("expected status 'not found', got '%s'", r.Status)
		}
	}
}

func TestUnlinkWrongTarget(t *testing.T) {
	repoPath, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a symlink pointing to wrong location
	wrongTarget := filepath.Join(targetPath, ".bashrc")
	wrongSource := "/tmp/somewhere/else"
	if err := os.Symlink(wrongSource, wrongTarget); err != nil {
		t.Fatalf("failed to create test symlink: %v", err)
	}

	results, err := Unlink(repoPath, targetPath)
	if err != nil {
		t.Fatalf("Unlink() error = %v", err)
	}

	// Find result for bashrc
	var bashrcResult *LinkResult
	for _, r := range results {
		if filepath.Base(r.Target) == ".bashrc" {
			bashrcResult = &r
			break
		}
	}

	if bashrcResult.Status != "skipped" {
		t.Errorf("expected status 'skipped' for wrong symlink, got '%s'", bashrcResult.Status)
	}

	// Verify symlink was not removed
	_, err = os.Lstat(wrongTarget)
	if os.IsNotExist(err) {
		t.Error("wrong symlink should not be removed")
	}
}

func TestLinkInvalidRepo(t *testing.T) {
	_, targetPath, cleanup := setupTestRepo(t)
	defer cleanup()

	opts := LinkOptions{
		RepoPath: "/nonexistent/path",
		Target:   targetPath,
		DryRun:   false,
		Force:    false,
	}

	_, err := Link(opts)
	if err == nil {
		t.Error("Link() should fail for non-existent repository")
	}
}
