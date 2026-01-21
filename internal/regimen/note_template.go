package regimen

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/notes"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage note templates",
	Long: `Manage note templates for structured entries.

Templates support placeholders:
  {{DATE}}           - Current date (YYYY-MM-DD)
  {{TIME}}           - Current time (HH:MM)
  {{DATETIME}}       - Current datetime
  {{PROMPT:question}} - Prompt user for input

Built-in templates: meeting, reflection, idea, report`,
	PersistentPreRunE: noteCmd.PersistentPreRunE, // Inherit encryption check
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates",
	Long:  `Lists all available note templates.`,
	RunE:  runTemplateList,
}

var templateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new template",
	Long: `Creates a new note template by opening an editor.

The template can include placeholders that will be replaced when used:
  {{DATE}}           - Current date
  {{TIME}}           - Current time
  {{PROMPT:question}} - Ask user for input`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateCreate,
}

var templateShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show a template",
	Long:  `Displays the content of a template.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateShow,
}

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a template",
	Long:  `Deletes a template.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateDelete,
}

func init() {
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateDeleteCmd)
	noteCmd.AddCommand(templateCmd)
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	// Ensure built-in templates exist
	if err := store.EnsureBuiltInTemplates(); err != nil {
		return fmt.Errorf("failed to ensure built-in templates: %w", err)
	}

	templates, err := store.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templates) == 0 {
		fmt.Println("No templates found.")
		return nil
	}

	fmt.Println("Available templates:")
	for _, tmpl := range templates {
		// Check if it's a built-in template
		marker := ""
		if _, ok := notes.BuiltInTemplates[tmpl.Name]; ok {
			marker = " (built-in)"
		}
		fmt.Printf("  %s%s\n", tmpl.Name, marker)
	}

	return nil
}

func runTemplateCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	// Create initial template content
	initialContent := fmt.Sprintf(`# Template: %s

Write your template here. You can use placeholders:
  {{DATE}}           - Current date
  {{TIME}}           - Current time
  {{DATETIME}}       - Current datetime
  {{PROMPT:question}} - Prompt user for input

---

`, name)

	// Open editor
	content, err := editInEditor(initialContent)
	if err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	// Create template
	if err := store.CreateTemplate(name, content); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Template %q created successfully.\n", name)
	return nil
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	tmpl, err := store.GetTemplate(name)
	if err != nil {
		return err
	}

	fmt.Println(tmpl.Content)
	return nil
}

func runTemplateDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	wikiDir := getWikiDir()
	store := notes.NewStore(wikiDir)

	// Check if it's a built-in template
	if _, ok := notes.BuiltInTemplates[name]; ok {
		return fmt.Errorf("cannot delete built-in template %q", name)
	}

	if err := store.DeleteTemplate(name); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Template %q deleted successfully.\n", name)
	return nil
}

// editInEditor opens the user's preferred editor and returns the content.
func editInEditor(initialContent string) (string, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "regimen-template-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write initial content
	if _, err := tmpFile.WriteString(initialContent); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Determine editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // fallback to vi
	}

	// Open editor
	editorCmd := exec.Command(editor, tmpPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	// Read edited content
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}
