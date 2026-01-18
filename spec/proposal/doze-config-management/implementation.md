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

- [ ] 1.1 Create `cmd/doze/main.go` entry point
- [ ] 1.2 Create `internal/doze/root.go` with Cobra root command
- [ ] 1.3 Implement `pkg/config/config.go` for YAML configuration loading/saving
- [ ] 1.4 Add configuration validation and error handling
- [ ] 1.5 Add unit tests for configuration parsing

**Milestone**: `doze --help` displays command structure; configuration loads from `~/.config/doze/config.yml`.

### Phase 2: Profile Management

**Goal**: Implement profile commands for switching contexts.

- [ ] 2.1 Implement `doze profile list` command
- [ ] 2.2 Implement `doze profile current` command
- [ ] 2.3 Implement `doze profile switch <name>` command
- [ ] 2.4 Implement `doze profile create <name>` command
- [ ] 2.5 Implement `doze profile delete <name>` command
- [ ] 2.6 Implement `doze profile copy <source> <dest>` command
- [ ] 2.7 Add `--json` output support for list commands
- [ ] 2.8 Add unit tests for profile operations

**Milestone**: Can create, switch between, and manage multiple profiles.

### Phase 3: Environment Variables

**Goal**: Implement environment variable management and shell export generation.

- [ ] 3.1 Implement `pkg/envvar/export.go` with shell-specific generators
- [ ] 3.2 Implement `pkg/envvar/parser.go` for variable expansion
- [ ] 3.3 Implement `doze env list` command
- [ ] 3.4 Implement `doze env get <var>` command
- [ ] 3.5 Implement `doze env set <var>=<value>` command
- [ ] 3.6 Implement `doze env unset <var>` command
- [ ] 3.7 Implement `doze env export` command with `--shell` flag
- [ ] 3.8 Implement `doze env source` command for shell evaluation
- [ ] 3.9 Implement shell auto-detection from `$SHELL`
- [ ] 3.10 Add unit tests for each shell's export syntax
- [ ] 3.11 Add unit tests for variable expansion

**Milestone**: `eval "$(doze env source)"` works in bash, zsh, fish, and powershell.

### Phase 4: Dotfile Management

**Goal**: Implement dotfile symlinking and status tracking.

- [ ] 4.1 Implement `pkg/dotfiles/linker.go` for symlink operations
- [ ] 4.2 Implement `pkg/dotfiles/differ.go` for diff generation
- [ ] 4.3 Implement `doze dotfiles link <repo>` command
- [ ] 4.4 Implement `--target`, `--dry-run`, `--force` flags
- [ ] 4.5 Implement backup creation for `--force` overwrites
- [ ] 4.6 Implement `doze dotfiles unlink <repo>` command
- [ ] 4.7 Implement `doze dotfiles status` command
- [ ] 4.8 Implement `doze dotfiles diff <repo>` command
- [ ] 4.9 Add symlink tracking in configuration
- [ ] 4.10 Add unit tests for linking logic
- [ ] 4.11 Add integration tests for symlink creation/removal

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
