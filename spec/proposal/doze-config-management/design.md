# Design: doze-config-management
**Specification Reference**: [specification.md](specification.md)
Status: Draft

## 1. Context

Doze is a new standalone executable for managing dotfiles, environment variables, and configuration profiles. Unlike regimen (productivity) and nightwatch (security), doze addresses developer environment management—a distinct domain that benefits from separation.

The sleepless project already establishes patterns for CLI structure (Cobra), storage (YAML/Markdown files), and cross-platform support. Doze follows these patterns while introducing configuration-specific functionality.

Key constraints:
- Must work offline without network dependencies
- Must support multiple shells (bash, zsh, fish, powershell)
- Must not interfere with existing dotfile management tools
- Storage format must be human-readable and version-control friendly

## 2. Goals and Non-Goals

### Goals

- Provide a simple CLI for symlinking dotfiles from a repository
- Enable shell-agnostic environment variable management with shell-specific export generation
- Support profile switching for different contexts (work, personal, project-specific)
- Maintain human-readable YAML configuration files
- Integrate with existing shell initialization workflows via `eval "$(doze env source)"`

### Non-Goals

- Template rendering for dotfiles (use chezmoi or similar for complex templating)
- Secret management or encryption (use dedicated tools like age, sops)
- Remote synchronization (use git or syncthing separately)
- GUI or TUI interfaces

## 3. Options Considered

### Option 1: Single YAML Configuration File

Store all configuration (profiles, variables, dotfile mappings) in a single `~/.config/doze/config.yml` file.

**Advantages**:
- Simple to implement and understand
- Single source of truth
- Easy to backup/version control

**Disadvantages**:
- Large configurations become unwieldy
- All profiles loaded even when only one is used

**Complexity**: Low

### Option 2: Directory-Based Configuration with File Per Profile

Store base configuration in `config.yml` and each profile in `profiles/<name>.yml`.

**Advantages**:
- Cleaner organization for many profiles
- Profiles can be added/removed as files
- Easier to share individual profiles

**Disadvantages**:
- More complex file discovery
- Potential for orphaned profile files

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 1 (Single YAML Configuration File)

**Rationale**: Most users will have 2-5 profiles. A single file keeps implementation simple and makes configuration easily portable. If scaling becomes an issue, Option 2 can be introduced later as a non-breaking enhancement.

**Key Factors**:
1. Simplicity for typical use cases
2. Easier implementation and testing
3. Single file to backup/version control

## 5. Detailed Design

### Architecture Overview

```
cmd/doze/main.go          Entry point
internal/doze/
  root.go                 Root command setup
  dotfiles.go             Dotfile link/unlink/diff/status commands
  env.go                  Environment variable commands
  profile.go              Profile management commands
pkg/dotfiles/
  linker.go               Symlink creation/removal logic
  differ.go               Diff generation for dotfiles
pkg/envvar/
  export.go               Shell-specific export generation
  parser.go               Variable expansion logic
pkg/config/
  config.go               Configuration file loading/saving
```

### Component Design

**Config Manager (`pkg/config`)**:
- Loads/saves YAML configuration from `~/.config/doze/config.yml`
- Handles `DOZE_CONFIG_DIR` override
- Validates configuration structure on load

**Dotfile Linker (`pkg/dotfiles`)**:
- Creates symlinks with collision detection
- Tracks linked files in configuration for status reporting
- Creates backups before `--force` overwrites
- Validates source paths exist before linking

**Environment Exporter (`pkg/envvar`)**:
- Generates shell-specific export commands
- Supports variable expansion (`$VAR`, `${VAR}`)
- Detects shell from `$SHELL` environment variable

**Shell Support Matrix**:
| Shell | Export Syntax | Detection |
|-------|--------------|-----------|
| bash | `export VAR="value"` | `/bin/bash`, `/usr/bin/bash` |
| zsh | `export VAR="value"` | `/bin/zsh`, `/usr/bin/zsh` |
| fish | `set -gx VAR "value"` | `/usr/bin/fish` |
| powershell | `$env:VAR = "value"` | `pwsh`, `powershell` |

### Data Design

**Configuration Schema** (`~/.config/doze/config.yml`):

```yaml
version: 1
current_profile: default

profiles:
  default:
    variables:
      EDITOR: vim
      PAGER: less
  work:
    variables:
      AWS_PROFILE: company-prod
      KUBECONFIG: ~/.kube/work-config

dotfiles:
  - repo: ~/dotfiles
    links:
      - source: bashrc
        target: ~/.bashrc
      - source: gitconfig
        target: ~/.gitconfig
    linked_at: 2026-01-18T10:00:00Z
```

### API Design

Not applicable (CLI-only).

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| Single config file | Simplicity | Scalability for large configs | Most users have few profiles |
| No templating | Focused scope | Advanced dotfile features | Users can use chezmoi alongside |
| No encryption | Simplicity | Secret protection | Credentials should use dedicated tools |

## 7. Cross-Cutting Concerns

### Security

- Configuration files may contain paths to sensitive files; recommend 0600 permissions
- Variable values displayed via `doze env list` may contain secrets; warn in documentation
- Symlink creation validates target paths to prevent symlink attacks outside intended directories

### Performance

- Configuration file is small; no caching needed
- Dotfile linking is infrequent; no performance concerns
- Shell export generation is instant for typical variable counts (<100)

### Reliability

- Backup existing files before `--force` overwrites
- Validate configuration on load; fail fast with clear errors
- Atomic profile switching (update config, then report success)

### Testing

- Unit tests for each shell's export syntax generation
- Unit tests for variable expansion (including edge cases like circular references)
- Integration tests for symlink creation/removal
- Tests for configuration validation and error messages

## 8. Implementation Plan

See [implementation.md](implementation.md) for phased task breakdown.

### Migration Strategy

Not applicable (new executable).

## 9. Open Questions

- Should `doze env source` output a comment header identifying the profile? (Helps debugging)
- Should circular variable references be detected and reported, or silently passed through?
- Should `doze dotfiles link` support glob patterns in the repository (e.g., `*.conf`)?
