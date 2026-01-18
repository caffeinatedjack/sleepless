package doze

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/config"
)

var (
	jsonOutput bool
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage configuration profiles",
	Long:  `Create, switch between, and manage configuration profiles.`,
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Long:  `List all defined configuration profiles.`,
	RunE:  runProfileList,
}

var profileCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current profile",
	Long:  `Display the currently active profile.`,
	RunE:  runProfileCurrent,
}

var profileSwitchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch to a different profile",
	Long:  `Activate the specified profile.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileSwitch,
}

var profileCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new profile",
	Long:  `Create a new empty profile.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileCreate,
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a profile",
	Long:  `Remove a profile and its associated variables.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileDelete,
}

var profileCopyCmd = &cobra.Command{
	Use:   "copy <source> <dest>",
	Short: "Copy a profile",
	Long:  `Duplicate a profile's variables to a new profile.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runProfileCopy,
}

func init() {
	rootCmd.AddCommand(profileCmd)

	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileCurrentCmd)
	profileCmd.AddCommand(profileSwitchCmd)
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	profileCmd.AddCommand(profileCopyCmd)

	// Add --json flag to list command
	profileListCmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
}

func runProfileList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if jsonOutput {
		type ProfileInfo struct {
			Name      string `json:"name"`
			IsCurrent bool   `json:"is_current"`
			VarCount  int    `json:"variable_count"`
		}

		var profiles []ProfileInfo
		for name, profile := range cfg.Profiles {
			profiles = append(profiles, ProfileInfo{
				Name:      name,
				IsCurrent: name == cfg.CurrentProfile,
				VarCount:  len(profile.Variables),
			})
		}

		// Sort by name
		sort.Slice(profiles, func(i, j int) bool {
			return profiles[i].Name < profiles[j].Name
		})

		data, err := json.MarshalIndent(profiles, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Text output
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Profiles:")
	for _, name := range names {
		profile := cfg.Profiles[name]
		marker := " "
		if name == cfg.CurrentProfile {
			marker = "*"
		}
		fmt.Fprintf(out, "  %s %s (%d variables)\n", marker, name, len(profile.Variables))
	}

	return nil
}

func runProfileCurrent(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), cfg.CurrentProfile)
	return nil
}

func runProfileSwitch(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.SetCurrentProfile(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Switched to profile: %s\n", name)
	return nil
}

func runProfileCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.CreateProfile(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Created profile: %s\n", name)
	return nil
}

func runProfileDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.DeleteProfile(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Deleted profile: %s\n", name)
	return nil
}

func runProfileCopy(cmd *cobra.Command, args []string) error {
	source := args[0]
	dest := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.CopyProfile(source, dest); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Copied profile '%s' to '%s'\n", source, dest)
	return nil
}
