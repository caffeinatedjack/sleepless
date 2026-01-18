package doze

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/config"
	"gitlab.com/caffeinatedjack/sleepless/pkg/dotfiles"
)

var (
	targetFlag string
	dryRunFlag bool
	forceFlag  bool
)

var dotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "Manage dotfile symlinks",
	Long:  `Create, remove, and manage symlinks for dotfile repositories.`,
}

var dotfilesLinkCmd = &cobra.Command{
	Use:   "link <repo-path>",
	Short: "Link dotfiles from a repository",
	Long:  `Create symlinks from the target directory to files in the dotfile repository.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDotfilesLink,
}

var dotfilesUnlinkCmd = &cobra.Command{
	Use:   "unlink <repo-path>",
	Short: "Remove dotfile symlinks",
	Long:  `Remove symlinks created from the dotfile repository.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDotfilesUnlink,
}

var dotfilesStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show dotfile status",
	Long:  `Display the state of all managed dotfiles.`,
	RunE:  runDotfilesStatus,
}

var dotfilesDiffCmd = &cobra.Command{
	Use:   "diff <repo-path>",
	Short: "Show differences between repository and linked files",
	Long:  `Compare files in the repository with their linked targets.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDotfilesDiff,
}

func init() {
	rootCmd.AddCommand(dotfilesCmd)

	dotfilesCmd.AddCommand(dotfilesLinkCmd)
	dotfilesCmd.AddCommand(dotfilesUnlinkCmd)
	dotfilesCmd.AddCommand(dotfilesStatusCmd)
	dotfilesCmd.AddCommand(dotfilesDiffCmd)

	// Add flags
	dotfilesLinkCmd.Flags().StringVar(&targetFlag, "target", "", "target directory for links (default: $HOME)")
	dotfilesLinkCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "preview changes without applying")
	dotfilesLinkCmd.Flags().BoolVar(&forceFlag, "force", false, "overwrite existing files")

	dotfilesUnlinkCmd.Flags().StringVar(&targetFlag, "target", "", "target directory for links (default: $HOME)")
}

func runDotfilesLink(cmd *cobra.Command, args []string) error {
	repoPath := args[0]

	// Determine target directory
	target := targetFlag
	if target == "" {
		var err error
		target, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to determine home directory: %w", err)
		}
	}

	// Perform linking
	opts := dotfiles.LinkOptions{
		RepoPath: repoPath,
		Target:   target,
		DryRun:   dryRunFlag,
		Force:    forceFlag,
	}

	results, err := dotfiles.Link(opts)
	if err != nil {
		return err
	}

	// Display results
	if dryRunFlag {
		fmt.Println("DRY RUN - No changes made")
		fmt.Println()
	}

	created := 0
	skipped := 0
	errors := 0

	for _, r := range results {
		switch r.Status {
		case "created", "would create":
			fmt.Printf("✓ %s -> %s\n", r.Target, r.Source)
			created++
		case "exists":
			// Silent for already correct links
		case "skipped":
			fmt.Printf("⊗ %s: %s\n", r.Target, r.Message)
			skipped++
		case "error":
			fmt.Printf("✗ %s: %s\n", r.Target, r.Message)
			errors++
		}
	}

	fmt.Printf("\nSummary: %d created, %d skipped, %d errors\n", created, skipped, errors)

	// Track in configuration
	if !dryRunFlag && errors == 0 {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Find or create dotfile repo entry
		found := false
		for i, repo := range cfg.Dotfiles {
			if repo.Repo == repoPath {
				cfg.Dotfiles[i].LinkedAt = time.Now().Format(time.RFC3339)
				found = true
				break
			}
		}

		if !found {
			cfg.Dotfiles = append(cfg.Dotfiles, config.DotfileRepo{
				Repo:     repoPath,
				LinkedAt: time.Now().Format(time.RFC3339),
			})
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	return nil
}

func runDotfilesUnlink(cmd *cobra.Command, args []string) error {
	repoPath := args[0]

	// Determine target directory
	target := targetFlag
	if target == "" {
		var err error
		target, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to determine home directory: %w", err)
		}
	}

	// Perform unlinking
	results, err := dotfiles.Unlink(repoPath, target)
	if err != nil {
		return err
	}

	// Display results
	removed := 0
	skipped := 0
	errors := 0

	for _, r := range results {
		switch r.Status {
		case "removed":
			fmt.Printf("✓ Removed: %s\n", r.Target)
			removed++
		case "not found":
			// Silent for non-existent links
		case "skipped":
			fmt.Printf("⊗ %s: %s\n", r.Target, r.Message)
			skipped++
		case "error":
			fmt.Printf("✗ %s: %s\n", r.Target, r.Message)
			errors++
		}
	}

	fmt.Printf("\nSummary: %d removed, %d skipped, %d errors\n", removed, skipped, errors)

	// Update configuration
	if errors == 0 {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Remove repo from tracking
		for i, repo := range cfg.Dotfiles {
			if repo.Repo == repoPath {
				cfg.Dotfiles = append(cfg.Dotfiles[:i], cfg.Dotfiles[i+1:]...)
				break
			}
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	return nil
}

func runDotfilesStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Dotfiles) == 0 {
		fmt.Println("No dotfile repositories tracked")
		return nil
	}

	// Determine target directory
	target := targetFlag
	if target == "" {
		target, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to determine home directory: %w", err)
		}
	}

	// Check status for each tracked repository
	for _, repo := range cfg.Dotfiles {
		fmt.Printf("\nRepository: %s\n", repo.Repo)
		if repo.LinkedAt != "" {
			fmt.Printf("Linked at: %s\n", repo.LinkedAt)
		}

		results, err := dotfiles.Status(repo.Repo, target)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		linked := 0
		unlinked := 0
		conflicts := 0

		for _, r := range results {
			switch r.Status {
			case "linked":
				linked++
			case "unlinked":
				unlinked++
				fmt.Printf("  ⊗ %s: not linked\n", r.Target)
			case "conflict":
				conflicts++
				fmt.Printf("  ! %s: %s\n", r.Target, r.Message)
			case "error":
				fmt.Printf("  ✗ %s: %s\n", r.Target, r.Message)
			}
		}

		fmt.Printf("  Summary: %d linked, %d unlinked, %d conflicts\n", linked, unlinked, conflicts)
	}

	return nil
}

func runDotfilesDiff(cmd *cobra.Command, args []string) error {
	repoPath := args[0]

	// Determine target directory
	target := targetFlag
	if target == "" {
		var err error
		target, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to determine home directory: %w", err)
		}
	}

	// Get diffs
	results, err := dotfiles.Diff(repoPath, target)
	if err != nil {
		return err
	}

	hasDifferences := false

	for _, r := range results {
		switch r.Status {
		case "different":
			hasDifferences = true
			fmt.Printf("\nDifference in: %s\n", r.Target)
			fmt.Println(r.Diff)
		case "missing":
			hasDifferences = true
			fmt.Printf("\nMissing: %s\n", r.Target)
		case "identical":
			// Silent for identical files
		case "error":
			fmt.Printf("\nError checking %s: %s\n", r.Target, r.Message)
		}
	}

	if !hasDifferences {
		fmt.Println("No differences found")
	}

	return nil
}
