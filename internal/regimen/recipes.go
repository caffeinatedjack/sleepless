package regimen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/recipes"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

const (
	defaultRecipesDir   = "~/wiki/recipes"
	defaultRecipesIndex = "~/wiki/Recipes.md"
	envRecipesDir       = "REGIMEN_RECIPES_DIR"
	envRecipesIndex     = "REGIMEN_RECIPES_INDEX"
)

var (
	// Global flags for recipes commands
	recipesDir   string
	recipesIndex string
)

var recipesCmd = &cobra.Command{
	Use:   "recipes",
	Short: "Manage recipes in your wiki",
	Long: `Manage recipes in your wiki.

Commands:
    list           List recipes with optional category filter
    view           View a recipe (full text or ingredients only)
    search         Search recipes by title or ingredients
    new            Create a new recipe file skeleton
    shopping-list  Generate a shopping list from recipes

Examples:
    regimen recipes list
    regimen recipes list --category British
    regimen recipes view dhal_vhal
    regimen recipes view dhal_vhal --ingredients
    regimen recipes search chicken
    regimen recipes new "Lemon Tart"
    regimen recipes shopping-list pasta salad --out /tmp/shopping.md`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if wiki is encrypted
		wikiDir := getWikiDir()
		if err := checkWikiEncrypted(wikiDir); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(recipesCmd)

	// Persistent flags available to all subcommands
	recipesCmd.PersistentFlags().StringVar(&recipesDir, "dir", "",
		fmt.Sprintf("Recipe directory (default %s, env: %s)", defaultRecipesDir, envRecipesDir))
	recipesCmd.PersistentFlags().StringVar(&recipesIndex, "index", "",
		fmt.Sprintf("Recipe index file (default %s, env: %s)", defaultRecipesIndex, envRecipesIndex))
}

// getRecipesDir returns the recipe directory to use.
// Priority: --dir flag > REGIMEN_RECIPES_DIR env > default
func getRecipesDir() string {
	if recipesDir != "" {
		return expandPath(recipesDir)
	}
	if env := os.Getenv(envRecipesDir); env != "" {
		return expandPath(env)
	}
	return expandPath(defaultRecipesDir)
}

// getRecipesIndex returns the recipe index file to use.
// Priority: --index flag > REGIMEN_RECIPES_INDEX env > default
func getRecipesIndex() string {
	if recipesIndex != "" {
		return expandPath(recipesIndex)
	}
	if env := os.Getenv(envRecipesIndex); env != "" {
		return expandPath(env)
	}
	return expandPath(defaultRecipesIndex)
}

// slugFromPath extracts the slug (filename without .md) from a path.
func slugFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".md")
}

// --- List Command ---

var (
	listCategory string
	listTag      string
)

var recipesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recipes",
	Long: `List recipes in the recipe directory.

By default, lists all recipes sorted by title.
Use --category to filter by a category defined in the recipe index.
Use --tag to filter by a tag defined in recipe files.

Examples:
    regimen recipes list
    regimen recipes list --category British
    regimen recipes list --tag vegetarian
    regimen recipes list --dir ~/my-recipes`,
	RunE: runRecipesList,
}

func init() {
	recipesCmd.AddCommand(recipesListCmd)
	recipesListCmd.Flags().StringVarP(&listCategory, "category", "c", "",
		"Filter by category (from recipe index)")
	recipesListCmd.Flags().StringVarP(&listTag, "tag", "t", "",
		"Filter by tag (from recipe files)")
}

func runRecipesList(cmd *cobra.Command, args []string) error {
	dir := getRecipesDir()

	if listCategory != "" {
		return listByCategory(dir, listCategory)
	}

	if listTag != "" {
		return listByTag(dir, listTag)
	}

	refs, err := recipes.Find(dir)
	if err != nil {
		return err
	}

	if len(refs) == 0 {
		ui.PrintDim("No recipes found")
		return nil
	}

	for _, ref := range refs {
		slug := slugFromPath(ref.Path)
		fmt.Printf("%s  %s\n", ui.DimStyle.Render(slug), ref.Title)
	}

	return nil
}

func listByTag(dir, tag string) error {
	refs, err := recipes.FindByTag(dir, tag)
	if err != nil {
		return err
	}

	if len(refs) == 0 {
		ui.PrintDim(fmt.Sprintf("No recipes with tag %q", tag))
		return nil
	}

	for _, ref := range refs {
		slug := slugFromPath(ref.Path)
		fmt.Printf("%s  %s\n", ui.DimStyle.Render(slug), ref.Title)
	}

	return nil
}

func listByCategory(dir, category string) error {
	indexPath := getRecipesIndex()

	idx, err := recipes.ParseIndex(indexPath)
	if err != nil {
		return fmt.Errorf("cannot read recipe index: %w", err)
	}

	paths, err := idx.GetCategoryRecipes(category)
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		ui.PrintDim(fmt.Sprintf("No recipes in category %q", category))
		return nil
	}

	// Parse titles and collect with paths for sorting
	type recipeEntry struct {
		slug  string
		title string
	}
	var entries []recipeEntry
	for _, path := range paths {
		title, err := recipes.ParseTitle(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot read %s: %v\n", path, err)
			continue
		}
		entries = append(entries, recipeEntry{
			slug:  slugFromPath(path),
			title: title,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].title < entries[j].title
	})

	for _, e := range entries {
		fmt.Printf("%s  %s\n", ui.DimStyle.Render(e.slug), e.title)
	}

	return nil
}

// --- View Command ---

var (
	viewIngredients bool
)

var recipesViewCmd = &cobra.Command{
	Use:   "view <recipe>",
	Short: "View a recipe",
	Long: `View a recipe file.

By default, prints the full recipe content.
Use --ingredients to show only the ingredient list.

The recipe can be specified as:
- A slug (filename without .md extension): dhal_vhal
- A filename: dhal_vhal.md
- A full path: ~/wiki/recipes/dhal_vhal.md

Examples:
    regimen recipes view dhal_vhal
    regimen recipes view dhal_vhal --ingredients
    regimen recipes view ~/wiki/recipes/special.md`,
	Args: cobra.ExactArgs(1),
	RunE: runRecipesView,
}

func init() {
	recipesCmd.AddCommand(recipesViewCmd)
	recipesViewCmd.Flags().BoolVarP(&viewIngredients, "ingredients", "i", false,
		"Show only ingredients")
}

func runRecipesView(cmd *cobra.Command, args []string) error {
	dir := getRecipesDir()
	input := args[0]

	path, err := recipes.Resolve(dir, input)
	if err != nil {
		return err
	}

	if viewIngredients {
		ingredients, err := recipes.ParseIngredients(path)
		if err != nil {
			return err
		}

		if len(ingredients) == 0 {
			ui.PrintDim("No ingredients found")
			return nil
		}

		for _, ing := range ingredients {
			fmt.Println(ing)
		}
		return nil
	}

	// Print full file contents
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read recipe: %w", err)
	}

	fmt.Print(string(content))
	return nil
}

// --- Search Command ---

var (
	searchIngredientsOnly bool
)

var recipesSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search recipes",
	Long: `Search recipes by title and ingredients.

By default, searches both titles and ingredients.
Use --ingredients-only to search only in ingredients.

Search is case-insensitive.

Examples:
    regimen recipes search chicken
    regimen recipes search "olive oil" --ingredients-only`,
	Args: cobra.ExactArgs(1),
	RunE: runRecipesSearch,
}

func init() {
	recipesCmd.AddCommand(recipesSearchCmd)
	recipesSearchCmd.Flags().BoolVar(&searchIngredientsOnly, "ingredients-only", false,
		"Search only in ingredients")
}

func runRecipesSearch(cmd *cobra.Command, args []string) error {
	dir := getRecipesDir()
	query := strings.ToLower(args[0])

	refs, err := recipes.Find(dir)
	if err != nil {
		return err
	}

	var matches []string

	for _, ref := range refs {
		// Check title match (unless ingredients-only)
		if !searchIngredientsOnly {
			if strings.Contains(strings.ToLower(ref.Title), query) {
				matches = append(matches, ref.Title)
				continue
			}
		}

		// Check ingredients
		ingredients, err := recipes.ParseIngredients(ref.Path)
		if err != nil {
			continue
		}

		for _, ing := range ingredients {
			if strings.Contains(strings.ToLower(ing), query) {
				matches = append(matches, ref.Title)
				break
			}
		}
	}

	if len(matches) == 0 {
		ui.PrintDim("No recipes found")
		return nil
	}

	// Sort and deduplicate
	sort.Strings(matches)
	for _, title := range matches {
		fmt.Println(title)
	}

	return nil
}

// --- New Command ---

var (
	newSlug  string
	newForce bool
)

var recipesNewCmd = &cobra.Command{
	Use:   "new <title>",
	Short: "Create a new recipe file",
	Long: `Create a new recipe file with a skeleton structure.

Creates a Markdown file in the recipe directory with standard headings:
- # <Title>
- ## Ingredients
- ## Method
- ## Notes

By default, the filename is derived from the title (slugified).
Use --slug to specify a custom filename.

Examples:
    regimen recipes new "Lemon Tart"
    regimen recipes new "Sunday Roast" --slug sunday_roast
    regimen recipes new "Quick Pasta" --force`,
	Args: cobra.ExactArgs(1),
	RunE: runRecipesNew,
}

func init() {
	recipesCmd.AddCommand(recipesNewCmd)
	recipesNewCmd.Flags().StringVar(&newSlug, "slug", "", "Custom filename (without .md)")
	recipesNewCmd.Flags().BoolVar(&newForce, "force", false, "Overwrite existing file")
}

func runRecipesNew(cmd *cobra.Command, args []string) error {
	dir := getRecipesDir()
	title := args[0]

	// Determine slug
	slug := newSlug
	if slug == "" {
		slug = recipes.Slugify(title)
	}

	if slug == "" {
		return fmt.Errorf("cannot derive filename from title %q", title)
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create recipe directory: %w", err)
	}

	// Build file path
	filename := slug + ".md"
	path := filepath.Join(dir, filename)

	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		if !newForce {
			return fmt.Errorf("recipe file already exists: %s (use --force to overwrite)", path)
		}
	}

	// Create skeleton content
	content := fmt.Sprintf(`# %s

## Ingredients

## Method

## Notes
`, title)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("cannot write recipe file: %w", err)
	}

	ui.Success(fmt.Sprintf("Created %s", path))
	return nil
}

// --- Shopping List Command ---

var (
	shoppingCategory string
	shoppingTag      string
	shoppingOut      string
)

var recipesShoppingCmd = &cobra.Command{
	Use:   "shopping-list [recipe...]",
	Short: "Generate a shopping list",
	Long: `Generate a shopping list from one or more recipes.

Extracts ingredients from each recipe and outputs a Markdown shopping list.

Recipes can be specified as:
- Positional arguments: regimen recipes shopping-list pasta salad
- By category: regimen recipes shopping-list --category British
- By tag: regimen recipes shopping-list --tag vegetarian

Output goes to stdout by default. Use --out to write to a file.

Examples:
    regimen recipes shopping-list dhal_vhal
    regimen recipes shopping-list pasta salad curry
    regimen recipes shopping-list --category British
    regimen recipes shopping-list --tag quick
    regimen recipes shopping-list pasta --out /tmp/shopping.md`,
	RunE: runRecipesShopping,
}

func init() {
	recipesCmd.AddCommand(recipesShoppingCmd)
	recipesShoppingCmd.Flags().StringVarP(&shoppingCategory, "category", "c", "",
		"Generate list from all recipes in category")
	recipesShoppingCmd.Flags().StringVarP(&shoppingTag, "tag", "t", "",
		"Generate list from all recipes with tag")
	recipesShoppingCmd.Flags().StringVarP(&shoppingOut, "out", "o", "",
		"Output file path (default: stdout)")
}

func runRecipesShopping(cmd *cobra.Command, args []string) error {
	dir := getRecipesDir()

	var recipePaths []string

	if shoppingCategory != "" {
		// Get recipes from category
		indexPath := getRecipesIndex()
		idx, err := recipes.ParseIndex(indexPath)
		if err != nil {
			return fmt.Errorf("cannot read recipe index: %w", err)
		}

		paths, err := idx.GetCategoryRecipes(shoppingCategory)
		if err != nil {
			return err
		}

		recipePaths = paths
	} else if shoppingTag != "" {
		// Get recipes by tag
		refs, err := recipes.FindByTag(dir, shoppingTag)
		if err != nil {
			return err
		}

		if len(refs) == 0 {
			return fmt.Errorf("no recipes found with tag %q", shoppingTag)
		}

		for _, ref := range refs {
			recipePaths = append(recipePaths, ref.Path)
		}
	} else {
		// Use positional arguments
		if len(args) == 0 {
			return fmt.Errorf("specify recipes or use --category or --tag")
		}

		// Resolve each recipe
		for _, arg := range args {
			path, err := recipes.Resolve(dir, arg)
			if err != nil {
				return err
			}
			recipePaths = append(recipePaths, path)
		}
	}

	if len(recipePaths) == 0 {
		return fmt.Errorf("no recipes specified")
	}

	// Build shopping list
	output, err := recipes.BuildShoppingList(dir, recipePaths)
	if err != nil {
		return err
	}

	// Write output
	if shoppingOut != "" {
		outPath := expandPath(shoppingOut)
		if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("cannot write shopping list: %w", err)
		}
		ui.Success(fmt.Sprintf("Shopping list written to %s", outPath))
		return nil
	}

	fmt.Print(output)
	return nil
}
