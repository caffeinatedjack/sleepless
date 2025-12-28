# Implementation: vimwiki recipe management cli

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Add a new `regimen recipes` Cobra command group backed by a small `pkg/recipes` library that discovers and parses recipe Markdown files stored in `~/wiki/recipes` and parses a category index in `~/wiki/Recipes.md`. Initial functionality focuses on list/view/search/new plus a shopping list generator.

## Prerequisites

- ~~Decide the exact skeleton headings produced by `regimen recipes new` beyond the required `# Title` and `## Ingredients`.~~ **Resolved**: Spec section 6.5 requires `# Title`, `## Ingredients`, `## Method`, `## Notes`.

## Phases

### Phase 1: `pkg/recipes` library (discovery + parsing + index)

**Goal**: Provide a tested library for discovering recipes, extracting titles/ingredients, parsing category index, and rendering shopping list Markdown.

- [x] Create `pkg/recipes` package with:
  - [x] `Find(root string) ([]RecipeRef, error)` (non-recursive `*.md` scan)
  - [x] `ParseTitle(path string) (string, error)`
  - [x] `ParseIngredients(path string) ([]string, error)`
  - [x] `Resolve(root, input string) (path string, error)`
  - [x] `Slugify(title string) string`
  - [x] `ParseIndex(path string) (Index, error)` (headings → categories; links → recipes)
  - [x] `BuildShoppingList(recipes []string) (string, error)` (renders Markdown)
- [x] Unit tests:
  - [x] title parsing (including missing H1)
  - [x] ingredients parsing (including `###` subheadings; stop at next `##`)
  - [x] index parsing using a fixture matching `/home/jack/wiki/Recipes.md`
  - [x] shopping list output structure and determinism

**Milestone**: `go test ./...` passes and unit tests cover the example recipe structure (`dhal_vhal.md`) and example index structure (`Recipes.md`). ✅

### Phase 2: `regimen recipes` command group (list + view + new)

**Goal**: Expose recipe operations via the regimen CLI.

- [x] Add `internal/regimen/recipes.go` command group and wire into `rootCmd`.
- [x] Shared flag/env resolution:
  - [x] `--dir` and `REGIMEN_RECIPES_DIR` (default `~/wiki/recipes`)
  - [x] `--index` and `REGIMEN_RECIPES_INDEX` (default `~/wiki/Recipes.md`)
- [x] Implement `regimen recipes list`:
  - [x] default: one title per line (sorted)
  - [x] `--category <name>`: filter via index
- [x] Implement `regimen recipes view <recipe>` with `--ingredients`.
- [x] Implement `regimen recipes new <title>`:
  - [x] creates `<slug>.md` with required headings
  - [x] does not attempt to populate ingredients; user edits manually

**Milestone**: Manual acceptance tests on a real wiki directory:

- `regimen recipes list` prints `Vaal Dhal` for `dhal_vhal.md`
- `regimen recipes list --category British` includes `Vaal Dhal` when `Recipes.md` links it
- `regimen recipes view dhal_vhal --ingredients` prints the bullet ingredients
- `regimen recipes new "Test Recipe"` creates a file skeleton in `~/wiki/recipes`

### Phase 3: Search + shopping list generator

**Goal**: Add discovery workflows and derived output.

- [x] Implement `regimen recipes search <query>` (title + ingredients; case-insensitive)
- [x] Implement `regimen recipes shopping-list`:
  - [x] accept one or more recipes by slug/path
  - [x] accept `--category <name>` to generate from all recipes in that category via index
  - [x] default output to stdout; support `--out <path>`
  - [x] output Markdown with `# Shopping List` and `## <Recipe Title>` per recipe
- [x] Add tests for:
  - [x] category selection
  - [x] output file writing vs stdout rendering (library-level tests)

**Milestone**: Running `regimen recipes shopping-list --category British` produces a Markdown shopping list including the ingredients from `dhal_vhal.md`. ✅

### Phase 4: Tag support

**Goal**: Add tag-based filtering for recipes.

- [x] Add tag parsing to `pkg/recipes`:
  - [x] `ParseTags(path string) ([]string, error)` - parse `Tags: tag1, tag2` line
  - [x] `HasTag(tags []string, tag string) bool` - case-insensitive tag matching
  - [x] `FindByTag(root, tag string) ([]RecipeRef, error)` - find recipes by tag
  - [x] Include tags in `RecipeRef` and `Recipe` structs
- [x] Add `--tag` flag to `regimen recipes list`
- [x] Add `--tag` flag to `regimen recipes shopping-list`
- [x] Unit tests for tag parsing and filtering

**Milestone**: Running `regimen recipes list --tag vegetarian` lists only recipes with that tag. ✅

## Testing Plan

### Unit Tests

- Recipe parsing:
  - H1 on first line and after leading blanks
  - missing H1 (error)
- Ingredient parsing:
  - `## Ingredients` with direct bullet list
  - `## Ingredients` with `###` subheadings
  - stop at next `##` heading
- Index parsing:
  - category headings
  - link resolution relative to index directory
  - ignore non-`.md` links
- Slugify:
  - spaces -> `_`
  - disallowed chars removed
- Tag parsing:
  - present/absent tags
  - case-insensitive matching
  - whitespace trimming
  - tag filtering for list and shopping-list
- Shopping list rendering:
  - stable/deterministic ordering
  - correct headings and bullet lists

### Integration Tests

- Temp-dir based tests that create a fake recipe root and index file and verify `list/search/view/new/shopping-list` behavior.

## Rollback Plan

- If issues arise, remove the `recipes` command group wiring from `internal/regimen` and delete `pkg/recipes`.
- No existing data is modified unless `new` is used; rollback does not require data migration.

## Open Questions

- ~~For `new`, which additional headings should be included beyond `## Ingredients` (e.g. `## Method`, `## Notes`)?~~ **Resolved**: Spec section 6.5 defines the required headings.
- ~~Should shopping list preserve ingredient order exactly as in the recipe file (current plan) or sort entries?~~ **Resolved**: Spec section 6.6 states recipes are listed in command-line order; ingredients are preserved in source order.
