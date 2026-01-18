# doze-config-management

**Depends on**: none
**Affected files**: cmd/doze/main.go, internal/doze/, pkg/dotfiles/, pkg/envvar/

## Abstract

This specification defines a new standalone executable `doze` for managing configuration files, dotfiles, environment variables, and configuration profiles. It provides commands for linking dotfiles from repositories, generating shell-specific environment variable exports, switching between configuration profiles, and diffing configuration state.

## 1. Introduction

Developers frequently work across multiple machines and environments, each requiring different configuration: development, staging, production credentials; company vs personal machine dotfiles; project-specific environment variables. Managing these configurations manually leads to inconsistent setups, broken symlinks, and forgotten environment variables.

Existing solutions like GNU Stow, rcm, or chezmoi are either too complex for simple use cases or lack the environment variable management needed for modern development workflows. Many developers resort to maintaining shell-specific export scripts that become brittle over time.

This specification defines `doze`, a lightweight configuration management tool that handles dotfile symlinking, environment variable generation, and profile switching. It stores configuration as plain text files (YAML/TOML) and generates shell-specific output, making it scriptable and version-control friendly.

The tool is designed to be offline-first, with no network dependencies, and integrates with the sleepless CLI ecosystem's philosophy of terminal-first, local-first workflows.

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`Dotfile`: A configuration file, typically hidden (starting with `.`), stored in a user's home directory (e.g., `.bashrc`, `.gitconfig`).

`Dotfile Repository`: A directory (often version-controlled) containing dotfiles to be linked into the home directory.

`Profile`: A named set of environment variables and configuration options that can be activated.

`Environment Variable Declaration`: A key-value pair defining an environment variable, stored in a configuration file.

`Shell Export`: A shell-specific command that sets an environment variable (e.g., `export VAR=value` for bash/zsh).

`Symlink`: A symbolic link pointing from a home directory location to a file in a dotfile repository.

`Configuration Root`: The directory where doze stores its configuration files (default: `~/.config/doze`).

## 4. Concepts

### 4.1. Dotfile Management

Doze manages dotfiles by creating symlinks from the home directory (or other target directories) to files in a dotfile repository. This allows users to version-control their configuration files while keeping them in the standard locations expected by applications.

### 4.2. Environment Variable Management

Environment variables are stored in YAML/TOML configuration files and can be organized into profiles. Doze generates shell-specific export commands that can be sourced directly or added to shell initialization files.

### 4.3. Profile System

Profiles allow users to maintain multiple sets of environment variables and switch between them. Common use cases include:
- Development vs production credentials
- Personal vs work configurations
- Project-specific tool versions

### 4.4. Non-Destructive Operations

All file operations are non-destructive by default. Doze will not overwrite existing files unless explicitly instructed with a `--force` flag.

## 5. Requirements

### 5.1. Core Executable

1. The project MUST provide a standalone executable named `doze`.
2. The executable MUST be independent of other sleepless executables.
3. All commands MUST return exit code 0 on success and non-zero on failure.

### 5.2. Dotfile Commands

1. `doze dotfiles link <repo-path>` MUST create symlinks from the home directory to files in the repository.
2. `doze dotfiles link` MUST support `--target <dir>` to specify an alternative link location (default: `$HOME`).
3. `doze dotfiles link` MUST support `--dry-run` to preview changes without creating links.
4. `doze dotfiles link` MUST skip files that already exist unless `--force` is provided.
5. `doze dotfiles unlink <repo-path>` MUST remove symlinks created by previous link operations.
6. `doze dotfiles diff <repo-path>` MUST show differences between linked files and repository versions.
7. `doze dotfiles status` MUST show the state of all managed dotfiles (linked, modified, missing).

### 5.3. Environment Variable Commands

1. `doze env list` MUST display all environment variables in the current profile.
2. `doze env get <var>` MUST display the value of a specific variable.
3. `doze env set <var>=<value>` MUST add or update a variable in the current profile.
4. `doze env unset <var>` MUST remove a variable from the current profile.
5. `doze env export` MUST generate shell-specific export commands for all variables in the current profile.
6. `doze env export --shell <shell>` MUST generate exports for the specified shell (bash, zsh, fish, powershell).
7. If `--shell` is not provided, doze MUST detect the current shell from the `SHELL` environment variable.
8. `doze env source` MUST output commands suitable for shell evaluation (e.g., `eval "$(doze env source)"`).

### 5.4. Profile Commands

1. `doze profile list` MUST list all defined profiles.
2. `doze profile current` MUST display the currently active profile.
3. `doze profile switch <name>` MUST activate the specified profile.
4. `doze profile create <name>` MUST create a new empty profile.
5. `doze profile delete <name>` MUST remove a profile and its associated variables.
6. `doze profile copy <source> <dest>` MUST duplicate a profile's variables to a new profile.
7. The system MUST maintain a `default` profile that is active if no other profile is selected.

### 5.5. Configuration Storage

1. Configuration files MUST be stored in `~/.config/doze/` by default.
2. The system MUST support `DOZE_CONFIG_DIR` environment variable to override the configuration directory.
3. Profile data MUST be stored in `config.yml` or `config.toml` (user's choice).
4. The configuration format MUST support:
   - Profile definitions
   - Environment variable key-value pairs per profile
   - Dotfile repository paths
   - Linked file tracking

### 5.6. Shell Support

1. The system MUST support generating exports for: `bash`, `zsh`, `fish`, `powershell`.
2. Shell detection MUST use the `SHELL` environment variable when `--shell` is not provided.
3. Generated exports MUST be syntactically correct for the target shell.

### 5.7. Output Formats

1. All list commands MUST support `--json` for machine-readable output.
2. Text output MUST be human-readable with clear formatting.
3. Diff output MUST use unified diff format by default.

## 6. Interface

### 6.1. Commands

```bash
# Dotfile management
doze dotfiles link <repo-path> [--target <dir>] [--dry-run] [--force]
doze dotfiles unlink <repo-path>
doze dotfiles diff <repo-path>
doze dotfiles status

# Environment variables
doze env list [--json]
doze env get <var>
doze env set <var>=<value>
doze env unset <var>
doze env export [--shell <shell>]
doze env source [--shell <shell>]

# Profiles
doze profile list [--json]
doze profile current
doze profile switch <name>
doze profile create <name>
doze profile delete <name>
doze profile copy <source> <dest>
```

### 6.2. Flags

**Global Flags:**
- `--config <path>`: Override configuration directory
- `--json`: Output in JSON format

**Dotfiles Flags:**
- `--target <dir>`: Target directory for links (default: `$HOME`)
- `--dry-run`: Preview changes without applying
- `--force`: Overwrite existing files

**Env Export Flags:**
- `--shell <shell>`: Target shell (bash, zsh, fish, powershell)

### 6.3. Configuration File Format

```yaml
current_profile: default

profiles:
  default:
    variables:
      EDITOR: vim
      PATH: /usr/local/bin:$PATH
  
  work:
    variables:
      AWS_PROFILE: company-prod
      KUBECONFIG: ~/.kube/work-config
      EDITOR: code

dotfiles:
  repos:
    - path: ~/dotfiles
      links:
        - source: .bashrc
          target: ~/.bashrc
        - source: .gitconfig
          target: ~/.gitconfig
```

## 7. Behavior

### 7.1. Dotfile Linking

1. When linking dotfiles, doze MUST:
   - Verify the source file exists in the repository
   - Check if the target location already exists
   - Create parent directories if needed
   - Create a symbolic link from target to source
   - Record the link in its configuration for tracking

2. If a target file already exists and is not a symlink, doze MUST:
   - Report the conflict to stderr
   - Skip the file unless `--force` is provided
   - If `--force` is provided, backup the existing file with a `.bak` suffix before creating the link

### 7.2. Environment Variable Resolution

1. Variable values MAY reference other variables using `$VAR` or `${VAR}` syntax.
2. Variable expansion MUST be performed at export time, not storage time.
3. Undefined variable references SHOULD be preserved as-is in the output.

### 7.3. Profile Switching

1. Switching profiles MUST update the `current_profile` field in the configuration file.
2. Switching profiles MUST NOT automatically export variables to the current shell.
3. After switching, users MUST run `doze env source` to apply the new environment.

### 7.4. Shell Detection

1. Shell detection MUST extract the shell name from the `SHELL` environment variable (e.g., `/bin/bash` → `bash`).
2. If `SHELL` is not set or unrecognized, doze MUST default to `bash`.
3. Supported shell names: `bash`, `zsh`, `fish`, `powershell`, `pwsh`.

## 8. Error Handling

1. If a dotfile repository path does not exist, the command MUST exit with code 1 and report the missing path.
2. If attempting to link a file that doesn't exist in the repository, the command MUST exit with code 1.
3. If attempting to switch to a non-existent profile, the command MUST exit with code 1 and list available profiles.
4. If the configuration file is malformed, doze MUST exit with code 1 and report the parsing error.
5. If symlink creation fails (permission denied, filesystem limitations), doze MUST report the error and continue processing remaining files.
6. If unsetting a variable that doesn't exist, doze MUST succeed silently (idempotent operation).

## 9. Examples

```bash
# Link dotfiles from a repository
doze dotfiles link ~/dotfiles
doze dotfiles link ~/dotfiles --target ~/ --dry-run

# Check dotfile status
doze dotfiles status
doze dotfiles diff ~/dotfiles

# Manage environment variables
doze env list
doze env set EDITOR=nvim
doze env get EDITOR
doze env unset PAGER

# Generate shell exports
doze env export --shell bash
doze env export --shell fish

# Source into current shell
eval "$(doze env source)"

# Profile management
doze profile list
doze profile create staging
doze profile switch staging
doze env set AWS_PROFILE=staging-account
doze profile switch default

# Copy profile for modification
doze profile copy production staging
doze profile switch staging
doze env set DEBUG=true
```

## 10. Security Considerations

1. Doze MUST NOT transmit configuration data over the network.
2. Configuration files containing credentials SHOULD be stored with user-only permissions (0600).
3. When creating backups of existing files, doze MUST preserve the original file permissions.
4. Environment variable values MAY contain sensitive data (API keys, tokens). Documentation MUST warn users about:
   - Shell history exposure when using `doze env set` with sensitive values
   - Version control considerations when storing configuration files
5. Symlink creation MUST validate that the target path is not outside the intended directory tree (prevent symlink attacks).
6. The `--force` flag MUST be used cautiously as it can overwrite existing files.

## 11. Testing Considerations

Test scenarios SHOULD include:

- Linking dotfiles from various repository structures
- Handling existing files at target locations
- Creating and switching between profiles
- Variable expansion with circular references
- Shell export generation for each supported shell
- Dry-run mode accuracy
- Configuration file parsing (valid and malformed)
- Permission denied scenarios for symlink creation
- Symlink tracking and status reporting
- Profile deletion with active profile
- UTF-8 handling in variable values

## 12. References

### Normative References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997, https://www.rfc-editor.org/rfc/rfc2119

### Informative References

[STOW] GNU Stow, https://www.gnu.org/software/stow/

[RCM] thoughtbot/rcm, https://github.com/thoughtbot/rcm

[CHEZMOI] chezmoi, https://www.chezmoi.io/

[XDG] XDG Base Directory Specification, https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
