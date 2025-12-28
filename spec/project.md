# Project Specification

## Overview

Sleepless is a CLI ecosystem providing two complementary executables:

- **regimen**: Personal productivity toolset (goal/task management + small utilities) integrated with a Vimwiki-compatible Markdown directory layout.
- **nightwatch**: Security/development utilities for offline, scriptable workflows (JWT tooling, redaction, password generation, fake data generation).

Both tools are terminal-first, local-first, and designed to be scriptable.

### regimen

Goal/task management with full CRUD operations, visualization features, and hierarchical organization via topics and subtasks. Tasks are stored as human-readable Markdown files compatible with personal wiki systems.

Additional regimen utilities include `decide` (random choice helpers), `when` (timezone tools), and recipe management.

### nightwatch

Security and development utilities including JWT operations, PII/secret redaction, password/passphrase generation, and fake data generation. Designed for offline use with reproducible outputs where appropriate for testing scenarios.

## Goals

- **Terminal-first workflow**: Full task management without leaving the command line
- **Wiki integration**: Tasks stored as markdown files in `~/wiki/tasks/` for compatibility with Vim wiki and other markdown-based knowledge systems
- **Hierarchical organization**: Support for topics (categories) and subtasks within goals
- **Visual feedback**: Rich terminal output with colors, progress bars, and tree views using lipgloss
- **Simple data format**: Human-readable and editable markdown files with embedded metadata
- **Fast prefix-based ID lookup**: Short hex IDs (6 chars displayed) for quick task reference

## Non-Goals

- **GUI or web interface**: This is a terminal-only tool
- **Real-time collaboration**: Single-user, local-first design
- **Complex dependency graphs**: Only parent-child relationships, no arbitrary task dependencies

## Architecture

### Directory Structure

```
sleepless/
├── cmd/
│   ├── regimen/         # regimen executable entry point
│   │   └── main.go
│   └── nightwatch/      # nightwatch executable entry point
│       └── main.go
├── internal/
│   ├── nightwatch/      # Nightwatch-specific implementations
│   │   ├── data/        # Data files (names, lorem ipsum)
│   │   ├── fake/        # Fake data generators (implementation)
│   │   ├── jwt/         # JWT utilities (implementation)
│   │   ├── fake.go      # Fake command wiring
│   │   ├── jwt.go       # JWT command wiring
│   │   ├── password.go  # Password command wiring
│   │   ├── redact.go    # Redaction command wiring
│   │   ├── names.go     # Name utilities
│   │   └── root.go      # Root command
│   └── regimen/         # regimen-specific implementations
│       ├── add.go       # Add goal/subtask
│       ├── list.go      # List with filters
│       ├── done.go      # Mark complete
│       ├── edit.go      # Modify properties
│       ├── remove.go    # Delete goals
│       ├── move.go      # Move between topics
│       ├── view.go      # Visual displays
│       ├── search.go    # Full-text search
│       ├── remind.go    # Due date reminders
│       ├── archive.go   # Archive completed
│       ├── history.go   # Action history
│       ├── decide.go    # Random choice utilities
│       ├── recipes.go   # Recipe management
│       ├── when.go      # Timezone tools
│       └── root.go      # Root command
├── pkg/
│   ├── task/            # Core task model and utilities
│   │   └── task.go      # Task struct, status, priority, filtering
│   ├── storage/         # Persistence layer
│   │   └── storage.go   # Load/save tasks, topics, history
│   ├── parser/          # Markdown parsing
│   │   └── markdown.go  # Parse/write task markdown format
│   ├── redact/          # Redaction patterns and logic
│   │   ├── patterns.go  # PII/secret patterns
│   │   └── redact.go    # Redaction engine
│   ├── decide/          # Random choice utilities
│   │   └── rng.go       # RNG abstraction
│   └── ui/              # Terminal formatting
│       └── format.go    # Colors, styles, progress bars
└── spec/                # Project documentation
    ├── section/         # Completed specifications
    ├── proposal/        # Active proposals
    ├── archive/         # Archived specs
    └── rule/            # Project rules
```

### Data Model

**Task**: The core entity with the following properties:
- `ID`: 32-char hex UUID (6 chars displayed for brevity)
- `Title`: Task description
- `Status`: open | complete
- `Priority`: low | medium | high
- `Due`: Optional date (YYYY-MM-DD)
- `Tags`: Optional string array
- `Notes`: Optional string array
- `Created`: RFC3339 timestamp
- `Completed`: RFC3339 timestamp (when done)
- `ParentID`: Optional parent task ID (for subtasks)
- `Subtasks`: Child tasks array
- `Topic`: Category/topic name (default: "inbox")

### Storage Format

Tasks are stored as markdown files in `~/wiki/tasks/`:
- One `.md` file per topic (e.g., `work.md`, `inbox.md`)
- `index.md`: Links to all topics
- `archived.md`: Completed and archived tasks
- `.task-meta.json`: History log and metadata

Markdown format:
```markdown
# Topic Name

- [ ] Task title {#hexid}
  - priority: high
  - due: 2025-01-15
  - tags: urgent, project
  - created: 2025-01-01T10:00:00Z
  - Note: Additional details
  - [ ] Subtask {#hexid2}
```

### Key Dependencies

- **cobra** (github.com/spf13/cobra): CLI framework for all commands
- **lipgloss** (github.com/charmbracelet/lipgloss): Terminal styling and UI components
- **jwt** (github.com/golang-jwt/jwt/v5): JWT token operations

## Feature Documentation

Detailed specifications for each feature are maintained in the `spec/` directory:

### Specifications (by executable)

- **[regimen](section/regimen.md)**: Goals/tasks, `decide`, `when`, and recipes
- **[nightwatch](section/nightwatch.md)**: `jwt`, `redact`, `password`, and `fake`

The two consolidated documents above are the entry points for completed specifications, organized by executable.

### Proposals and archives

New work should be captured as proposals under `spec/proposal/<name>/...`.

Completed proposals and older spec history are kept under `spec/archive/`.

### Project Rules

- **[keep-executables-independent](rule/keep-executables-independent.md)**: Executables must remain independent

## References

### External Dependencies

- [Cobra CLI Framework](https://cobra.dev/)
- [Lipgloss Terminal Styling](https://github.com/charmbracelet/lipgloss)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt)

### Related Projects

- [Vimwiki](https://github.com/vimwiki/vimwiki) - Personal wiki system for task storage integration
