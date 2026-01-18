package doze

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/config"
	"gitlab.com/caffeinatedjack/sleepless/pkg/envvar"
)

var (
	shellFlag string
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables",
	Long:  `Set, get, and export environment variables from profiles.`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environment variables",
	Long:  `List all environment variables in the current profile.`,
	RunE:  runEnvList,
}

var envGetCmd = &cobra.Command{
	Use:   "get <var>",
	Short: "Get a variable value",
	Long:  `Display the value of a specific environment variable.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvGet,
}

var envSetCmd = &cobra.Command{
	Use:   "set <var>=<value>",
	Short: "Set a variable",
	Long:  `Add or update an environment variable in the current profile.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvSet,
}

var envUnsetCmd = &cobra.Command{
	Use:   "unset <var>",
	Short: "Unset a variable",
	Long:  `Remove an environment variable from the current profile.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvUnset,
}

var envExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Generate shell export commands",
	Long:  `Generate shell-specific export commands for all variables in the current profile.`,
	RunE:  runEnvExport,
}

var envSourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Generate commands for shell evaluation",
	Long:  `Generate commands suitable for shell evaluation (e.g., eval "$(doze env source)").`,
	RunE:  runEnvSource,
}

func init() {
	rootCmd.AddCommand(envCmd)

	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
	envCmd.AddCommand(envExportCmd)
	envCmd.AddCommand(envSourceCmd)

	// Add flags
	envListCmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	envExportCmd.Flags().StringVar(&shellFlag, "shell", "", "target shell (bash, zsh, fish, powershell)")
	envSourceCmd.Flags().StringVar(&shellFlag, "shell", "", "target shell (bash, zsh, fish, powershell)")
}

func runEnvList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		return err
	}

	if len(profile.Variables) == 0 {
		fmt.Println("No variables defined in current profile")
		return nil
	}

	if jsonOutput {
		data, err := json.MarshalIndent(profile.Variables, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Text output - sorted by key
	keys := make([]string, 0, len(profile.Variables))
	for k := range profile.Variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("Environment variables in profile '%s':\n", cfg.CurrentProfile)
	for _, key := range keys {
		fmt.Printf("  %s=%s\n", key, profile.Variables[key])
	}

	return nil
}

func runEnvGet(cmd *cobra.Command, args []string) error {
	varName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		return err
	}

	value, exists := profile.Variables[varName]
	if !exists {
		return fmt.Errorf("variable '%s' not found in current profile", varName)
	}

	fmt.Println(value)
	return nil
}

func runEnvSet(cmd *cobra.Command, args []string) error {
	arg := args[0]

	// Parse VAR=VALUE
	parts := strings.SplitN(arg, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: expected VAR=VALUE, got '%s'", arg)
	}

	varName := parts[0]
	value := parts[1]

	if varName == "" {
		return fmt.Errorf("variable name cannot be empty")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		return err
	}

	profile.Variables[varName] = value

	// Update the profile in config
	cfg.Profiles[cfg.CurrentProfile] = *profile

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s=%s in profile '%s'\n", varName, value, cfg.CurrentProfile)
	return nil
}

func runEnvUnset(cmd *cobra.Command, args []string) error {
	varName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		return err
	}

	// Delete the variable (idempotent - no error if doesn't exist)
	delete(profile.Variables, varName)

	// Update the profile in config
	cfg.Profiles[cfg.CurrentProfile] = *profile

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Unset %s in profile '%s'\n", varName, cfg.CurrentProfile)
	return nil
}

func runEnvExport(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		return err
	}

	// Determine shell
	var shell envvar.Shell
	if shellFlag != "" {
		shell, err = envvar.ParseShell(shellFlag)
		if err != nil {
			return err
		}
	} else {
		shell = envvar.DetectShell()
	}

	exports, err := envvar.Export(shell, profile.Variables)
	if err != nil {
		return err
	}

	if exports == "" {
		fmt.Printf("# No variables to export from profile '%s'\n", cfg.CurrentProfile)
		return nil
	}

	fmt.Printf("# Exports for profile '%s' (shell: %s)\n", cfg.CurrentProfile, shell)
	fmt.Println(exports)

	return nil
}

func runEnvSource(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		return err
	}

	// Determine shell
	var shell envvar.Shell
	if shellFlag != "" {
		shell, err = envvar.ParseShell(shellFlag)
		if err != nil {
			return err
		}
	} else {
		shell = envvar.DetectShell()
	}

	exports, err := envvar.Export(shell, profile.Variables)
	if err != nil {
		return err
	}

	// Output without comments for clean eval
	if exports != "" {
		fmt.Println(exports)
	}

	return nil
}
