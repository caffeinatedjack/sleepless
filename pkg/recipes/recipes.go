// Package recipes provides recipe discovery, parsing, and management for Markdown-based
// recipe files stored in a Vimwiki-compatible directory structure.
package recipes

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Recipe represents a parsed recipe file.
type Recipe struct {
	Title       string
	Path        string
	Tags        []string
	Ingredients []string
}

// RecipeRef is a lightweight reference to a recipe file (path + title + tags).
type RecipeRef struct {
	Path  string
	Title string
	Tags  []string
}

// Index represents a parsed recipe index file with category mappings.
type Index struct {
	// Categories maps category names to recipe paths (resolved relative to index dir).
	Categories map[string][]string
	// CategoryOrder preserves the order categories appear in the index.
	CategoryOrder []string
}

var (
	// h1Re matches a Markdown H1 heading: # Title
	h1Re = regexp.MustCompile(`^#\s+(.+)$`)
	// h1OrH2Re matches H1 or H2 headings (used to terminate ingredient sections)
	h1OrH2Re = regexp.MustCompile(`^#{1,2}\s+`)
	// ingredientsHeaderRe matches the ingredients section header (case-insensitive)
	ingredientsHeaderRe = regexp.MustCompile(`(?i)^##\s+ingredients\s*$`)
	// listItemRe matches unordered list items: - item or * item
	listItemRe = regexp.MustCompile(`^\s*[-*]\s+(.+)$`)
	// headingRe matches any heading (for index parsing)
	headingRe = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	// linkRe matches Markdown links: [text](path)
	linkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	// tagsRe matches a Tags: line (case-insensitive)
	tagsRe = regexp.MustCompile(`(?i)^tags:\s*(.+)$`)
)

// Find discovers recipe files in the given root directory.
// It performs a non-recursive scan for *.md files, ignoring subdirectories and non-regular files.
func Find(root string) ([]RecipeRef, error) {
	root = expandPath(root)

	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("recipe directory does not exist: %s", root)
		}
		return nil, fmt.Errorf("cannot access recipe directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("recipe path is not a directory: %s", root)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("cannot read recipe directory: %w", err)
	}

	var refs []RecipeRef
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Only process .md files
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}

		// Skip non-regular files (symlinks, etc.)
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}

		path := filepath.Join(root, name)
		title, err := ParseTitle(path)
		if err != nil {
			// Skip files without valid titles (warning would be issued by caller)
			continue
		}

		tags, _ := ParseTags(path) // Ignore error, treat as empty tags

		refs = append(refs, RecipeRef{
			Path:  path,
			Title: title,
			Tags:  tags,
		})
	}

	// Sort by title for stable output
	sort.Slice(refs, func(i, j int) bool {
		return strings.ToLower(refs[i].Title) < strings.ToLower(refs[j].Title)
	})

	return refs, nil
}

// FindByTag returns recipes that have the specified tag.
func FindByTag(root, tag string) ([]RecipeRef, error) {
	refs, err := Find(root)
	if err != nil {
		return nil, err
	}

	var filtered []RecipeRef
	for _, ref := range refs {
		if HasTag(ref.Tags, tag) {
			filtered = append(filtered, ref)
		}
	}

	return filtered, nil
}

// ParseTitle extracts the recipe title from a file.
// The title is the first H1 heading (# Title) in the file.
func ParseTitle(path string) (string, error) {
	path = expandPath(path)

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open recipe file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if match := h1Re.FindStringSubmatch(line); match != nil {
			return strings.TrimSpace(match[1]), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading recipe file: %w", err)
	}

	return "", fmt.Errorf("no title found in recipe file: %s", path)
}

// ParseTags extracts tags from a recipe file.
// Tags are parsed from a line matching "Tags: tag1, tag2, tag3" (case-insensitive).
// The tags line must appear before the first H2 heading.
func ParseTags(path string) ([]string, error) {
	path = expandPath(path)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open recipe file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Stop at first H2 heading
		if strings.HasPrefix(line, "## ") {
			break
		}

		// Check for Tags: line
		if match := tagsRe.FindStringSubmatch(line); match != nil {
			return parseTags(match[1]), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading recipe file: %w", err)
	}

	// No tags found - return empty slice
	return []string{}, nil
}

// parseTags splits a comma-separated tag string into individual tags.
func parseTags(tagString string) []string {
	parts := strings.Split(tagString, ",")
	var tags []string
	for _, p := range parts {
		tag := strings.TrimSpace(p)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// HasTag checks if a recipe has a specific tag (case-insensitive).
func HasTag(tags []string, tag string) bool {
	tag = strings.ToLower(tag)
	for _, t := range tags {
		if strings.ToLower(t) == tag {
			return true
		}
	}
	return false
}

// ParseIngredients extracts ingredient entries from a recipe file.
// Ingredients are unordered list items under the ## Ingredients section.
// Parsing stops at the next H2 or higher heading.
func ParseIngredients(path string) ([]string, error) {
	path = expandPath(path)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open recipe file: %w", err)
	}
	defer file.Close()

	var ingredients []string
	inIngredients := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for ## Ingredients header
		if ingredientsHeaderRe.MatchString(line) {
			inIngredients = true
			continue
		}

		if inIngredients {
			// Stop at next H1 or H2 heading (but not ### subheadings which are allowed)
			if h1OrH2Re.MatchString(line) && !ingredientsHeaderRe.MatchString(line) {
				break
			}

			// Extract list items (including under ### subheadings)
			if match := listItemRe.FindStringSubmatch(line); match != nil {
				ingredients = append(ingredients, strings.TrimSpace(match[1]))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading recipe file: %w", err)
	}

	return ingredients, nil
}

// ParseRecipe parses a complete recipe from a file.
func ParseRecipe(path string) (*Recipe, error) {
	title, err := ParseTitle(path)
	if err != nil {
		return nil, err
	}

	tags, _ := ParseTags(path) // Ignore error, treat as empty tags

	ingredients, err := ParseIngredients(path)
	if err != nil {
		return nil, err
	}

	return &Recipe{
		Title:       title,
		Path:        path,
		Tags:        tags,
		Ingredients: ingredients,
	}, nil
}

// Resolve resolves a recipe identifier to a file path.
// It accepts either:
// - An explicit path to a recipe file
// - A base name/slug that resolves to ${root}/<name>.md
func Resolve(root, input string) (string, error) {
	root = expandPath(root)
	input = strings.TrimSpace(input)

	if input == "" {
		return "", fmt.Errorf("empty recipe identifier")
	}

	// Validate path doesn't escape root (security check)
	if err := validatePath(root, input); err != nil {
		return "", err
	}

	// If input looks like an absolute path or contains directory separators
	if filepath.IsAbs(input) || strings.Contains(input, string(filepath.Separator)) {
		expanded := expandPath(input)
		if _, err := os.Stat(expanded); err != nil {
			return "", fmt.Errorf("recipe file not found: %s", input)
		}
		return expanded, nil
	}

	// Try with .md extension
	candidate := input
	if !strings.HasSuffix(strings.ToLower(candidate), ".md") {
		candidate = candidate + ".md"
	}

	path := filepath.Join(root, candidate)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("recipe not found: %s", input)
		}
		return "", fmt.Errorf("cannot access recipe: %w", err)
	}

	return path, nil
}

// Slugify converts a title to a filename-safe slug.
// It lowercases, replaces whitespace with underscores, and removes disallowed characters.
func Slugify(title string) string {
	// Lowercase
	slug := strings.ToLower(title)

	// Replace runs of whitespace with underscore
	spaceRe := regexp.MustCompile(`\s+`)
	slug = spaceRe.ReplaceAllString(slug, "_")

	// Remove characters not in [a-z0-9_-]
	allowedRe := regexp.MustCompile(`[^a-z0-9_-]`)
	slug = allowedRe.ReplaceAllString(slug, "")

	// Collapse multiple underscores
	multiUnderRe := regexp.MustCompile(`_+`)
	slug = multiUnderRe.ReplaceAllString(slug, "_")

	// Trim leading/trailing underscores
	slug = strings.Trim(slug, "_-")

	return slug
}

// ParseIndex parses a recipe index file to extract category→recipe mappings.
// Categories are derived from headings, and recipes are derived from Markdown links.
func ParseIndex(path string) (*Index, error) {
	path = expandPath(path)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open recipe index: %w", err)
	}
	defer file.Close()

	indexDir := filepath.Dir(path)

	index := &Index{
		Categories:    make(map[string][]string),
		CategoryOrder: []string{},
	}

	var currentCategory string
	isFirstHeading := true

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for heading (category)
		if match := headingRe.FindStringSubmatch(line); match != nil {
			heading := strings.TrimSpace(match[2])

			// Skip the first heading (document title)
			if isFirstHeading {
				isFirstHeading = false
				continue
			}

			currentCategory = heading
			if _, exists := index.Categories[currentCategory]; !exists {
				index.Categories[currentCategory] = []string{}
				index.CategoryOrder = append(index.CategoryOrder, currentCategory)
			}
			continue
		}

		// Look for links in the line
		if currentCategory != "" {
			matches := linkRe.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				linkPath := match[2]

				// Only process .md links
				if !strings.HasSuffix(strings.ToLower(linkPath), ".md") {
					continue
				}

				// Resolve relative paths relative to index directory
				var resolvedPath string
				if filepath.IsAbs(linkPath) {
					resolvedPath = linkPath
				} else {
					resolvedPath = filepath.Join(indexDir, linkPath)
				}

				// Clean the path
				resolvedPath = filepath.Clean(resolvedPath)

				index.Categories[currentCategory] = append(
					index.Categories[currentCategory],
					resolvedPath,
				)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading recipe index: %w", err)
	}

	return index, nil
}

// GetCategoryRecipes returns recipe paths for a given category.
// Returns an error if the category doesn't exist.
func (idx *Index) GetCategoryRecipes(category string) ([]string, error) {
	// Case-insensitive category lookup
	for cat, paths := range idx.Categories {
		if strings.EqualFold(cat, category) {
			return paths, nil
		}
	}

	// Category not found - suggest available categories
	if len(idx.CategoryOrder) > 0 {
		return nil, fmt.Errorf("category %q not found; available categories: %s",
			category, strings.Join(idx.CategoryOrder, ", "))
	}
	return nil, fmt.Errorf("category %q not found; no categories defined in index", category)
}

// ShoppingListEntry represents ingredients from a single recipe.
type ShoppingListEntry struct {
	Title       string
	Ingredients []string
}

// BuildShoppingList generates a Markdown shopping list from recipe paths.
func BuildShoppingList(root string, recipePaths []string) (string, error) {
	var entries []ShoppingListEntry

	for _, path := range recipePaths {
		// Resolve the path if it's not absolute
		resolvedPath := path
		if !filepath.IsAbs(path) {
			var err error
			resolvedPath, err = Resolve(root, path)
			if err != nil {
				return "", fmt.Errorf("cannot resolve recipe %q: %w", path, err)
			}
		}

		recipe, err := ParseRecipe(resolvedPath)
		if err != nil {
			return "", fmt.Errorf("cannot parse recipe %q: %w", path, err)
		}

		entries = append(entries, ShoppingListEntry{
			Title:       recipe.Title,
			Ingredients: recipe.Ingredients,
		})
	}

	return renderShoppingList(entries), nil
}

// renderShoppingList renders shopping list entries as Markdown.
func renderShoppingList(entries []ShoppingListEntry) string {
	var sb strings.Builder

	sb.WriteString("# Shopping List\n")

	for _, entry := range entries {
		sb.WriteString("\n## ")
		sb.WriteString(entry.Title)
		sb.WriteString("\n\n")

		for _, ingredient := range entry.Ingredients {
			sb.WriteString("- ")
			sb.WriteString(ingredient)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// validatePath checks that the input path doesn't escape the root directory.
func validatePath(root, input string) error {
	// Don't validate absolute paths (they're used as-is)
	if filepath.IsAbs(input) {
		return nil
	}

	// Check for path traversal attempts
	cleaned := filepath.Clean(input)
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "/../") {
		return fmt.Errorf("invalid recipe path: path traversal not allowed")
	}

	return nil
}
