package recipes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTitle(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		expected string
	}{
		{
			name:     "simple h1",
			content:  "# My Recipe\n\nSome content",
			expected: "My Recipe",
		},
		{
			name:     "h1 after blank lines",
			content:  "\n\n# Recipe Title\n\nContent",
			expected: "Recipe Title",
		},
		{
			name:     "h1 with extra spaces",
			content:  "#   Spaced Title  \n",
			expected: "Spaced Title",
		},
		{
			name:    "no h1 heading",
			content: "## Section\n\nNo title here",
			wantErr: true,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:     "h1 after other content",
			content:  "Some preamble\n# Actual Title\n",
			expected: "Actual Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "recipe.md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			got, err := ParseTitle(path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTitle() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ParseTitle() unexpected error: %v", err)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseTitle() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "simple tags",
			content: `# Recipe

Tags: vegetarian, quick, indian

## Ingredients
`,
			expected: []string{"vegetarian", "quick", "indian"},
		},
		{
			name: "tags with extra whitespace",
			content: `# Recipe

Tags:   spicy ,  healthy  ,  dinner  

## Ingredients
`,
			expected: []string{"spicy", "healthy", "dinner"},
		},
		{
			name: "case insensitive Tags key",
			content: `# Recipe

tags: vegan, gluten-free

## Ingredients
`,
			expected: []string{"vegan", "gluten-free"},
		},
		{
			name: "TAGS uppercase",
			content: `# Recipe

TAGS: breakfast, easy

## Ingredients
`,
			expected: []string{"breakfast", "easy"},
		},
		{
			name: "no tags",
			content: `# Recipe

## Ingredients
- Salt
`,
			expected: []string{},
		},
		{
			name: "tags after h2 ignored",
			content: `# Recipe

## Ingredients

Tags: should-be-ignored

`,
			expected: []string{},
		},
		{
			name: "single tag",
			content: `# Recipe

Tags: solo

## Method
`,
			expected: []string{"solo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "recipe.md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			got, err := ParseTags(path)
			if err != nil {
				t.Errorf("ParseTags() unexpected error: %v", err)
				return
			}

			if len(got) != len(tt.expected) {
				t.Errorf("ParseTags() got %d tags, want %d", len(got), len(tt.expected))
				t.Errorf("got: %v", got)
				t.Errorf("want: %v", tt.expected)
				return
			}

			for i, tag := range got {
				if tag != tt.expected[i] {
					t.Errorf("ParseTags()[%d] = %q, want %q", i, tag, tt.expected[i])
				}
			}
		})
	}
}

func TestHasTag(t *testing.T) {
	tags := []string{"Vegetarian", "Quick", "Indian"}

	tests := []struct {
		tag      string
		expected bool
	}{
		{"vegetarian", true},
		{"VEGETARIAN", true},
		{"Vegetarian", true},
		{"quick", true},
		{"indian", true},
		{"vegan", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := HasTag(tags, tt.tag)
			if got != tt.expected {
				t.Errorf("HasTag(%v, %q) = %v, want %v", tags, tt.tag, got, tt.expected)
			}
		})
	}
}

func TestFindByTag(t *testing.T) {
	tmpDir := t.TempDir()

	// Create recipe files with different tags
	recipes := map[string]string{
		"pasta.md": `# Pasta Carbonara

Tags: italian, quick, dinner

## Ingredients
- Pasta
`,
		"salad.md": `# Caesar Salad

Tags: healthy, quick, lunch

## Ingredients
- Lettuce
`,
		"soup.md": `# Tomato Soup

Tags: healthy, dinner

## Ingredients
- Tomatoes
`,
		"no_tags.md": `# Plain Recipe

## Ingredients
- Something
`,
	}

	for name, content := range recipes {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	tests := []struct {
		tag           string
		expectedCount int
		expectedTitle string // first expected title (for verification)
	}{
		{"quick", 2, "Caesar Salad"}, // salad and pasta (sorted)
		{"healthy", 2, "Caesar Salad"},
		{"dinner", 2, "Pasta Carbonara"},
		{"italian", 1, "Pasta Carbonara"},
		{"vegan", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			refs, err := FindByTag(tmpDir, tt.tag)
			if err != nil {
				t.Fatalf("FindByTag() error: %v", err)
			}

			if len(refs) != tt.expectedCount {
				t.Errorf("FindByTag(%q) got %d recipes, want %d", tt.tag, len(refs), tt.expectedCount)
				for _, r := range refs {
					t.Logf("  found: %s", r.Title)
				}
			}

			if tt.expectedCount > 0 && len(refs) > 0 {
				if refs[0].Title != tt.expectedTitle {
					t.Errorf("first recipe = %q, want %q", refs[0].Title, tt.expectedTitle)
				}
			}
		})
	}
}

func TestParseIngredients(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "simple ingredients",
			content: `# Recipe

## Ingredients
- Salt
- Pepper
- Olive oil

## Method
Cook it.
`,
			expected: []string{"Salt", "Pepper", "Olive oil"},
		},
		{
			name: "ingredients with subheadings",
			content: `# Vaal Dhal

## Ingredients
### Main
- Fava Beans (Broad Beans) (45%)
- Water
- Tomato Puree

### Herbs and Spices
- Cumin
- Turmeric

## Method
Mix and cook.
`,
			expected: []string{
				"Fava Beans (Broad Beans) (45%)",
				"Water",
				"Tomato Puree",
				"Cumin",
				"Turmeric",
			},
		},
		{
			name: "case insensitive header",
			content: `# Test

## INGREDIENTS
- Item 1
- Item 2

## Notes
`,
			expected: []string{"Item 1", "Item 2"},
		},
		{
			name: "asterisk list markers",
			content: `# Recipe

## Ingredients
* Salt
* Pepper

## Method
`,
			expected: []string{"Salt", "Pepper"},
		},
		{
			name: "no ingredients section",
			content: `# Recipe

## Method
Just cook it.
`,
			expected: nil,
		},
		{
			name: "ingredients at end of file",
			content: `# Recipe

## Ingredients
- Final item
`,
			expected: []string{"Final item"},
		},
		{
			name: "indented list items",
			content: `# Recipe

## Ingredients
  - Indented salt
  - Indented pepper

## Method
`,
			expected: []string{"Indented salt", "Indented pepper"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "recipe.md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			got, err := ParseIngredients(path)
			if err != nil {
				t.Errorf("ParseIngredients() unexpected error: %v", err)
				return
			}

			if len(got) != len(tt.expected) {
				t.Errorf("ParseIngredients() got %d items, want %d", len(got), len(tt.expected))
				t.Errorf("got: %v", got)
				t.Errorf("want: %v", tt.expected)
				return
			}

			for i, item := range got {
				if item != tt.expected[i] {
					t.Errorf("ParseIngredients()[%d] = %q, want %q", i, item, tt.expected[i])
				}
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Simple Recipe", "simple_recipe"},
		{"Recipe with CAPS", "recipe_with_caps"},
		{"Multiple   Spaces", "multiple_spaces"},
		{"Special!@#Characters", "specialcharacters"},
		{"Numbers 123 Here", "numbers_123_here"},
		{"Already-has-dashes", "already-has-dashes"},
		{"  Leading and Trailing  ", "leading_and_trailing"},
		{"Mix_of_underscores and spaces", "mix_of_underscores_and_spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := Slugify(tt.title)
			if got != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		})
	}
}

func TestParseIndex(t *testing.T) {
	content := `# Recipes

# British
[Broad Bean Dhal](recipes/dhal_vhal.md)
[Fish and Chips](recipes/fish_chips.md)

# Indian
[Butter Chicken](recipes/butter_chicken.md)

# Uncategorized Links
[Not a recipe](notes.txt)
[Another MD](other/file.md)
`

	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "Recipes.md")
	if err := os.WriteFile(indexPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	idx, err := ParseIndex(indexPath)
	if err != nil {
		t.Fatalf("ParseIndex() error: %v", err)
	}

	// Check categories exist
	if len(idx.Categories) != 3 {
		t.Errorf("expected 3 categories, got %d", len(idx.Categories))
	}

	// Check British category
	british, err := idx.GetCategoryRecipes("British")
	if err != nil {
		t.Errorf("GetCategoryRecipes(British) error: %v", err)
	}
	if len(british) != 2 {
		t.Errorf("expected 2 recipes in British, got %d", len(british))
	}

	// Check path resolution (should be relative to index dir)
	expectedPath := filepath.Join(tmpDir, "recipes", "dhal_vhal.md")
	if len(british) > 0 && british[0] != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, british[0])
	}

	// Check case-insensitive lookup
	_, err = idx.GetCategoryRecipes("BRITISH")
	if err != nil {
		t.Errorf("case-insensitive lookup failed: %v", err)
	}

	// Check non-existent category
	_, err = idx.GetCategoryRecipes("NonExistent")
	if err == nil {
		t.Error("expected error for non-existent category")
	}

	// Check category order
	if len(idx.CategoryOrder) != 3 {
		t.Errorf("expected 3 categories in order, got %d", len(idx.CategoryOrder))
	}
	if idx.CategoryOrder[0] != "British" {
		t.Errorf("first category should be British, got %s", idx.CategoryOrder[0])
	}

	// Verify non-.md links are ignored
	uncategorized, _ := idx.GetCategoryRecipes("Uncategorized Links")
	if len(uncategorized) != 1 {
		t.Errorf("expected 1 .md link in Uncategorized Links, got %d", len(uncategorized))
	}
}

func TestFind(t *testing.T) {
	tmpDir := t.TempDir()

	// Create recipe files
	recipes := map[string]string{
		"apple_pie.md":    "# Apple Pie\n\n## Ingredients\n- Apples\n",
		"banana_bread.md": "# Banana Bread\n\n## Ingredients\n- Bananas\n",
		"no_title.md":     "## Just a section\nNo H1 here\n",
	}

	for name, content := range recipes {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "hidden.md"), []byte("# Hidden\n"), 0644); err != nil {
		t.Fatalf("failed to write hidden.md: %v", err)
	}

	// Create a non-.md file (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte("notes"), 0644); err != nil {
		t.Fatalf("failed to write notes.txt: %v", err)
	}

	refs, err := Find(tmpDir)
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}

	// Should find 2 valid recipes (no_title.md has no H1)
	if len(refs) != 2 {
		t.Errorf("expected 2 recipes, got %d", len(refs))
		for _, r := range refs {
			t.Logf("  found: %s", r.Title)
		}
	}

	// Check sorted order (Apple Pie before Banana Bread)
	if len(refs) >= 2 {
		if refs[0].Title != "Apple Pie" {
			t.Errorf("first recipe should be Apple Pie, got %s", refs[0].Title)
		}
		if refs[1].Title != "Banana Bread" {
			t.Errorf("second recipe should be Banana Bread, got %s", refs[1].Title)
		}
	}
}

func TestResolve(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a recipe file
	recipePath := filepath.Join(tmpDir, "test_recipe.md")
	if err := os.WriteFile(recipePath, []byte("# Test Recipe\n"), 0644); err != nil {
		t.Fatalf("failed to write recipe: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"slug without extension", "test_recipe", false},
		{"slug with extension", "test_recipe.md", false},
		{"non-existent", "missing", true},
		{"path traversal attempt", "../etc/passwd", true},
		{"empty input", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := Resolve(tmpDir, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Resolve() expected error, got path: %s", path)
				}
				return
			}
			if err != nil {
				t.Errorf("Resolve() unexpected error: %v", err)
				return
			}
			if path != recipePath {
				t.Errorf("Resolve() = %q, want %q", path, recipePath)
			}
		})
	}
}

func TestBuildShoppingList(t *testing.T) {
	tmpDir := t.TempDir()

	// Create recipe files
	recipe1 := `# Pasta Carbonara

## Ingredients
- 400g spaghetti
- 200g pancetta
- 4 eggs
- 100g parmesan

## Method
Cook pasta.
`
	recipe2 := `# Caesar Salad

## Ingredients
- 1 romaine lettuce
- 50g parmesan
- Croutons

## Method
Toss salad.
`

	if err := os.WriteFile(filepath.Join(tmpDir, "pasta.md"), []byte(recipe1), 0644); err != nil {
		t.Fatalf("failed to write pasta.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "salad.md"), []byte(recipe2), 0644); err != nil {
		t.Fatalf("failed to write salad.md: %v", err)
	}

	result, err := BuildShoppingList(tmpDir, []string{"pasta", "salad"})
	if err != nil {
		t.Fatalf("BuildShoppingList() error: %v", err)
	}

	// Check structure
	if !strings.HasPrefix(result, "# Shopping List\n") {
		t.Error("shopping list should start with # Shopping List")
	}

	if !strings.Contains(result, "## Pasta Carbonara") {
		t.Error("shopping list should contain ## Pasta Carbonara")
	}

	if !strings.Contains(result, "## Caesar Salad") {
		t.Error("shopping list should contain ## Caesar Salad")
	}

	// Check ingredients are present
	if !strings.Contains(result, "- 400g spaghetti") {
		t.Error("shopping list should contain spaghetti")
	}

	if !strings.Contains(result, "- 1 romaine lettuce") {
		t.Error("shopping list should contain romaine lettuce")
	}

	// Check order (Pasta should come before Caesar since that's the input order)
	pastaIdx := strings.Index(result, "## Pasta Carbonara")
	saladIdx := strings.Index(result, "## Caesar Salad")
	if pastaIdx > saladIdx {
		t.Error("Pasta Carbonara should appear before Caesar Salad (input order)")
	}
}

func TestFindNonExistentDirectory(t *testing.T) {
	_, err := Find("/nonexistent/path/to/recipes")
	if err == nil {
		t.Error("Find() should return error for non-existent directory")
	}
}

func TestParseRecipe(t *testing.T) {
	content := `# Test Recipe

Some description.

## Ingredients
- Salt
- Pepper

## Method
Cook it.
`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	recipe, err := ParseRecipe(path)
	if err != nil {
		t.Fatalf("ParseRecipe() error: %v", err)
	}

	if recipe.Title != "Test Recipe" {
		t.Errorf("Title = %q, want %q", recipe.Title, "Test Recipe")
	}

	if len(recipe.Ingredients) != 2 {
		t.Errorf("expected 2 ingredients, got %d", len(recipe.Ingredients))
	}

	if recipe.Path != path {
		t.Errorf("Path = %q, want %q", recipe.Path, path)
	}
}
