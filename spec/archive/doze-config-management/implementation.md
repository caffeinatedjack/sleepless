# Implementation: doze-config-management

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Implement `doze` as a new standalone executable with three command groups: `dotfiles` (symlink management), `env` (environment variables), and `profile` (configuration profiles). Uses YAML configuration stored in `~/.config/doze/`.

## Prerequisites

- None (new executable, no dependencies on existing sleepless code)

## Phases

### Phase 1: Core Infrastructure

**Goal**: Establish executable structure and configuration management.

- [x] 1.1 Create `cmd/doze/main.go` entry point
- [x] 1.2 Create `internal/doze/root.go` with Cobra root command
- [x] 1.3 Implement `pkg/config/config.go` for YAML configuration loading/saving
- [x] 1.4 Add configuration validation and error handling
- [x] 1.5 Add unit tests for configuration parsing

**Milestone**: `doze --help` displays command structure; configuration loads from `~/.config/doze/config.yml`.

### Phase 2: Profile Management

**Goal**: Implement profile commands for switching contexts.

- [x] 2.1 Implement `doze profile list` command
- [x] 2.2 Implement `doze profile current` command
- [x] 2.3 Implement `doze profile switch <name>` command
- [x] 2.4 Implement `doze profile create <name>` command
- [x] 2.5 Implement `doze profile delete <name>` command
- [x] 2.6 Implement `doze profile copy <source> <dest>` command
- [x] 2.7 Add `--json` output support for list commands
- [x] 2.8 Add unit tests for profile operations

**Milestone**: Can create, switch between, and manage multiple profiles.

### Phase 3: Environment Variables

**Goal**: Implement environment variable management and shell export generation.

- [x] 3.1 Implement `pkg/envvar/export.go` with shell-specific generators
- [x] 3.2 Implement `pkg/envvar/parser.go` for variable expansion
- [x] 3.3 Implement `doze env list` command
- [x] 3.4 Implement `doze env get <var>` command
- [x] 3.5 Implement `doze env set <var>=<value>` command
- [x] 3.6 Implement `doze env unset <var>` command
- [x] 3.7 Implement `doze env export` command with `--shell` flag
- [x] 3.8 Implement `doze env source` command for shell evaluation
- [x] 3.9 Implement shell auto-detection from `$SHELL`
- [x] 3.10 Add unit tests for each shell's export syntax
- [x] 3.11 Add unit tests for variable expansion

**Milestone**: `eval "$(doze env source)"` works in bash, zsh, fish, and powershell.

### Phase 4: Dotfile Management

**Goal**: Implement dotfile symlinking and status tracking.

- [x] 4.1 Implement `pkg/dotfiles/linker.go` for symlink operations
- [x] 4.2 Implement `pkg/dotfiles/differ.go` for diff generation
- [x] 4.3 Implement `doze dotfiles link <repo>` command
- [x] 4.4 Implement `--target`, `--dry-run`, `--force` flags
- [x] 4.5 Implement backup creation for `--force` overwrites
- [x] 4.6 Implement `doze dotfiles unlink <repo>` command
- [x] 4.7 Implement `doze dotfiles status` command
- [x] 4.8 Implement `doze dotfiles diff <repo>` command
- [x] 4.9 Add symlink tracking in configuration
- [x] 4.10 Add unit tests for linking logic
- [x] 4.11 Add integration tests for symlink creation/removal

**Milestone**: Can link dotfiles from a repository, view status, and see diffs.

## Testing Plan

### Unit Tests

- Configuration parsing: valid YAML, invalid YAML, missing file, malformed structure
- Profile operations: create, switch, delete, copy, list
- Variable expansion: simple `$VAR`, braced `${VAR}`, undefined variables, nested references
- Shell export generation: bash, zsh, fish, powershell syntax correctness
- Symlink logic: collision detection, backup creation, path validation

### Integration Tests

- Full workflow: create profile → set variables → export → source in shell
- Dotfile linking: link → status → modify file → diff → unlink
- Configuration persistence: changes survive process restart
- Cross-platform: symlink behavior on Linux/macOS/Windows

## Rollback Plan

- Remove `cmd/doze/`, `internal/doze/`, `pkg/dotfiles/`, `pkg/envvar/`, `pkg/config/`
- Remove `doze` from Makefile build targets

## Open Questions

- Should we add shell completion generation in Phase 1 or defer to a later phase?
- Should `doze env set` support reading value from stdin for multiline values?
