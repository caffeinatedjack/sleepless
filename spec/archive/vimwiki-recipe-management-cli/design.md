# Design: vimwiki recipe management cli
**Specification Reference**: [specification.md](specification.md)

## 1. Context

Regimen is a Go/Cobra CLI that manages Vimwiki-backed content (currently goals) by reading and writing Markdown files in a user’s wiki directory. Recipe notes already exist as Markdown files in the wiki, but there is no dedicated CLI support for discovering, categorizing, querying, or producing derived artifacts (such as a shopping list).

This feature integrates into the `regimen` executable (not `nightwatch`) and follows existing conventions in the repository: stdout for normal output, stderr for errors, minimal dependencies, and keeping executables independent by placing reusable logic in `pkg/*`.

## 2. Goals and Non-Goals

### Goals

- Add `regimen recipes` as a command group.
- Support workflows:
  - list recipes (optionally filtered by category)
  - view a recipe (full text or ingredients only)
  - search recipes (title and ingredients)
  - create a new recipe file skeleton (headings only; user edits content manually)
  - generate a shopping list Markdown output from one or more recipes
- Support categorization via a wiki-maintained index page (`~/wiki/Recipes.md`).
- Keep recipes as plain `.md` files and keep parsing permissive for existing notes.

### Non-Goals

- Full-fidelity Markdown AST parsing.
- Editing existing recipes in-place in the initial version.
- Unit normalization, ingredient deduplication, quantity parsing, or scaling.
- A database-backed recipe store.

## 3. Options Considered

### Option 1: Regex / line-scanning parser (recommended)

Scan recipe files line-by-line to find the first H1 for the title and the `## Ingredients` section for ingredient bullets. Scan `Recipes.md` for headings (categories) and Markdown links.

**Advantages**:
- Very small and robust against “real world” notes
- No new dependencies
- Easy to unit test

**Disadvantages**:
- Limited structural understanding (no full Markdown semantics)
- Edge cases (setext headings, unusual list formatting) require explicit handling if desired

**Complexity**: Low

### Option 2: Full Markdown parser (AST)

Use a Markdown parsing library to build an AST and interpret headings/lists/links.

**Advantages**:
- More correct across Markdown variants
- Easier to extend to richer features later

**Disadvantages**:
- New dependency surface
- More implementation complexity than the current need

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 1 (regex / line-scanning parser)

**Rationale**: The requirements are intentionally minimal (title + ingredients + index links) and must be compatible with hand-written wiki notes. A line-scanner aligns with existing patterns in `pkg/parser/markdown.go` and keeps dependencies low.

**Key Factors**:
1. Compatibility with existing Markdown
2. Minimal dependencies / small binary
3. Ease of testing and iteration

## 5. Detailed Design

### 5.1. Architecture Overview

- `internal/regimen/*`: Cobra wiring for `regimen recipes` commands.
- `pkg/recipes` (new): reusable library for:
  - discovering recipe files in a root directory
  - parsing a recipe (title + ingredient entries)
  - parsing the recipe index (categories + links)
  - resolving a recipe identifier to a path
  - slug generation
  - shopping list rendering

`internal/regimen` depends on `pkg/recipes`, similar to how it currently depends on `pkg/task`, `pkg/storage`, and `pkg/ui`.

### 5.2. Component Design

`pkg/recipes` responsibilities (suggested API surface):

- `Find(root) ([]RecipeRef, error)`
- `ParseTitle(path) (string, error)`
- `ParseIngredients(path) ([]string, error)`
- `Resolve(root, input) (string, error)`
- `Slugify(title) string`
- `ParseIndex(path) (Index, error)`
- `ShoppingListMarkdown(entries) (string, error)`

`internal/regimen` responsibilities:

- flags/env handling (`--dir`, `REGIMEN_RECIPES_DIR`, `--index`, `REGIMEN_RECIPES_INDEX`)
- choosing output format and writing to stdout / files
- mapping errors to user-friendly messages

### 5.3. Data Design

Minimal internal structs:

- `Recipe`:
  - `Title string`
  - `Path string`
  - `Ingredients []string`

- `Index`:
  - `Categories map[string][]string` (category name → recipe paths)

The library intentionally avoids attempting to model every possible recipe attribute.

### 5.4. Shopping list rendering

The shopping list generator should not attempt ingredient deduplication or parsing quantities (non-goal). It should render:

- `# Shopping List`
- for each recipe (in deterministic order):
  - `## <Recipe Title>`
  - `- <ingredient line>` bullets (as extracted)

This output is easy to print, edit, or import elsewhere.

## 6. Trade-offs

| Trade-off                   | Gain                     | Sacrifice                  | Justification                                                                                   |
|-----------------------------|--------------------------|----------------------------|-------------------------------------------------------------------------------------------------|
| Minimal parsing vs full AST | Simplicity + fewer deps  | Less Markdown correctness  | Wiki recipes are human-authored and varied; robustness is more important than perfect semantics |
| Categories via index file   | User-controlled taxonomy | Requires maintaining index | Leverages existing wiki workflows and avoids adding metadata requirements                       |
| No ingredient normalization | Predictable behavior     | No deduplication           | Keeps scope small; normalization can be added later                                             |

## 7. Cross-Cutting Concerns

### Security

- All path handling must prevent traversal outside the recipe root.
- Avoid following symlinks that escape the root when feasible.

### Performance

- Expected recipe counts are small; a directory scan and file reads are acceptable.
- `list` can parse only titles for speed.

### Reliability

- `list` should degrade gracefully: skip unreadable files with warnings.

### Testing

- Table-driven tests for title parsing, ingredient extraction, index parsing, and shopping list output.

## 8. Implementation Plan

Implement in phases as described in [implementation.md](implementation.md).

### Migration Strategy

None. This feature reads existing Markdown files and does not require migration.

## 9. Open Questions

1. For `regimen recipes new`, which additional headings should be included beyond `# Title` and `## Ingredients` (e.g. `## Method`, `## Notes`), or should it remain minimal?
2. Should `shopping-list` preserve the ingredient order exactly as in the file, or sort them? (Spec currently preserves source order.)
