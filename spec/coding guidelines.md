# Coding Guidelines

## Code Style

### Go Formatting

- **Required**: All Go code MUST be formatted with `go fmt` before committing
- **Linting**: Code SHOULD pass `go vet` and `golangci-lint` (or equivalent) without warnings
- **Line Length**: Aim for 100 characters; hard limit at 120 characters
- **Indentation**: Use tabs (Go standard)

### Naming Conventions

- **Packages**: Short, lowercase, single-word names (e.g., `task`, `redact`, `jwt`)
- **Exported Identifiers**: PascalCase (e.g., `Task`, `GenerateToken`, `RedactPII`)
- **Unexported Identifiers**: camelCase (e.g., `parseTemplate`, `loadNames`, `rng`)
- **Constants**: PascalCase for exported, camelCase for unexported
- **Acronyms**: Maintain case consistency (e.g., `HTTPServer`, `URLPath`, `ID`)

### Go Best Practices

- Prefer short variable names in small scopes (e.g., `i`, `err`, `ok`)
- Use descriptive names for exported APIs and larger scopes
- Return errors as the last return value
- Accept interfaces, return concrete types where practical
- Avoid naked returns except in very short functions

## File Organization

### Package Structure

```
pkg/            # Shared, reusable packages (public API)
internal/       # Private implementation details (cannot be imported externally)
cmd/            # Executable entry points (one per binary)
```

### File Naming

- Use lowercase with underscores for multi-word files: `template_parser.go`
- Test files: `*_test.go`
- One primary type per file; file name should match primary type: `task.go` for `type Task`

### Internal Organization

- Group related functionality in packages
- Keep `main()` functions minimal (delegate to internal packages)
- Separate command handlers (`internal/`) from reusable logic (`pkg/`)

## Documentation

### Comment Requirements

**Public Functions and Types**:
- **MUST** have a doc comment starting with the name of the element
- **MUST** explain what the function does, not how it works
- **MUST** document all parameters and return values if not obvious
- **SHOULD** include usage examples for complex APIs

**Private Functions**:
- **MUST NOT** have doc comments unless the implementation is non-obvious
- **MAY** have inline comments explaining complex logic or non-obvious decisions

### Doc Comment Style

```go
// Task represents a goal or subtask with metadata and hierarchical relationships.
// Tasks are uniquely identified by a 32-character hex ID and support nesting via ParentID.
type Task struct {
    ID       string
    Title    string
    ParentID string
}

// FindByPrefix searches for a task matching the provided ID prefix.
// Returns an error if the prefix matches zero or multiple tasks.
func FindByPrefix(prefix string, tasks []Task) (*Task, error) {
    // implementation
}

// No comment needed for simple private helper
func lowercaseEmail(s string) string {
    return strings.ToLower(s)
}

// Private but complex - explain the why
func computeWeightedIndex(weights []int, random int) int {
    // Uses cumulative sum approach to avoid floating-point arithmetic
    // and ensure deterministic results across platforms
    sum := 0
    for i, w := range weights {
        sum += w
        if random < sum {
            return i
        }
    }
    return len(weights) - 1
}
```

### Package Documentation

- Every package MUST have a `doc.go` file OR a package comment in the main file
- Package comments should describe the package's purpose and primary use cases

```go
// Package task provides the core data model and operations for goal management.
// Tasks support hierarchical organization via topics and subtasks, with metadata
// including priority, due dates, tags, and notes.
package task
```

### Meaningful Comments

**Good Comments** (explain why, context, or non-obvious decisions):
```go
// Use crypto/rand for unseeded mode to ensure high-quality randomness
// even in containerized or VM environments with limited entropy
rng := NewCryptoRNG()

// Limit history to 1000 entries to prevent unbounded growth
// Oldest entries are dropped on a FIFO basis
if len(history) > 1000 {
    history = history[len(history)-1000:]
}
```

**Bad Comments** (redundant or obvious):
```go
// Set the name variable to the name parameter
name := req.Name

// Loop through tasks
for _, task := range tasks {
    // ...
}
```

## Error Handling

### Error Creation

- Use `fmt.Errorf` with `%w` to wrap errors and preserve context
- Use `errors.New` for simple error strings
- Define sentinel errors as package-level variables: `var ErrNotFound = errors.New("task not found")`

```go
if err != nil {
    return fmt.Errorf("failed to load tasks from %s: %w", path, err)
}
```

### Error Checking

- Check all errors; never ignore with `_` unless explicitly justified
- Handle errors at the appropriate level (don't pass everything to main)
- Log errors before returning them only if needed for debugging

### CLI Error Handling

- Print user-friendly error messages to stderr
- Use consistent exit codes:
  - `0`: Success
  - `1`: General errors (invalid input, file not found, etc.)
  - `2`: Verification/validation failure (e.g., JWT signature invalid)
  - `3`: Expired/time-based failure
  - `4`: Not yet valid (nbf check)

```go
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

## Testing

### Test Coverage

- All exported functions MUST have unit tests
- Focus on edge cases, error paths, and boundary conditions
- Integration tests for command-line interfaces

### Test Naming

- Test functions: `TestFunctionName_Scenario`
- Table-driven tests preferred for multiple cases

```go
func TestParseTemplate_ValidSyntax(t *testing.T) { }
func TestParseTemplate_MalformedPlaceholder(t *testing.T) { }
func TestGenerateName_Deterministic(t *testing.T) { }
```

### Table-Driven Tests

```go
func TestNumberGenerator(t *testing.T) {
    tests := []struct {
        name string
        min  int
        max  int
        want func(int) bool // validation function
    }{
        {"in range", 1, 10, func(n int) bool { return n >= 1 && n <= 10 }},
        {"single value", 5, 5, func(n int) bool { return n == 5 }},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := GenerateNumber(tt.min, tt.max)
            if !tt.want(got) {
                t.Errorf("GenerateNumber(%d, %d) = %d, want value passing validation", tt.min, tt.max, got)
            }
        })
    }
}
```

### Determinism Tests

For features with seeded randomness, include determinism tests:

```go
func TestSeededDeterminism(t *testing.T) {
    seed := int64(42)
    result1 := GenerateWithSeed(seed)
    result2 := GenerateWithSeed(seed)
    if result1 != result2 {
        t.Errorf("seeded generation not deterministic: %v != %v", result1, result2)
    }
}
```

## Dependencies

### Adding Dependencies

- **Minimize external dependencies**: Prefer Go standard library when practical
- **Justify additions**: Document why the dependency is necessary
- **Check licenses**: Ensure compatibility with project license
- **Avoid indirect bloat**: Check transitive dependencies before adding

### Dependency Guidelines

- Use `go mod tidy` to clean up unused dependencies
- Pin major versions for stability
- Update dependencies deliberately, not automatically

### Approved Dependencies

- **cobra** (github.com/spf13/cobra): CLI framework - use for all command-line interfaces
- **lipgloss** (github.com/charmbracelet/lipgloss): Terminal styling - use for all terminal UI
- **jwt** (github.com/golang-jwt/jwt/v5): JWT operations - use for JWT tools only

### Library-Specific Guidelines

#### Cobra

- One root command per executable
- Use `cobra.Command` with clear `Use`, `Short`, and `Long` descriptions
- Group related subcommands under parent commands
- Use persistent flags for options that apply to all subcommands
- Validate flags in `PreRunE` when possible

```go
var nameCmd = &cobra.Command{
    Use:   "name",
    Short: "Generate a random full name",
    Long:  "Generate a random full name using first and last name lists.",
    RunE: func(cmd *cobra.Command, args []string) error {
        // implementation
    },
}
```

#### Lipgloss

- Define styles as package-level variables for consistency
- Use `lipgloss.NewStyle()` for terminal styling
- Keep terminal output clean and readable

```go
var titleStyle = lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#FAFAFA"))
```

#### JWT

- Always validate tokens before trusting claims
- Use appropriate key types (HMAC for symmetric, RSA/ECDSA for asymmetric)
- Handle expiration (`exp`) and not-before (`nbf`) claims explicitly
