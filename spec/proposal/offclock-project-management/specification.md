# offclock-project-management

**Depends on**: none
**Affected files**: cmd/offclock/main.go, internal/offclock/, pkg/scaffold/, pkg/scaffold/template/, pkg/scaffold/render/, Makefile

## Abstract

This specification adds a new standalone executable, `offclock`, that bootstraps a complete project from a template: repository layout, dependency manifests, Dockerfiles, CI/CD pipeline config, and starter scripts. Templates are embedded in the executable for offline use, and users can add custom templates locally.

## 1. Introduction

Starting a new project is usually a tedious series of repetitive steps: creating directories, wiring dependency manifests, writing a Dockerfile, adding CI, setting up a few scripts, and standardizing conventions (formatters, linting, .gitignore, editor config). Teams often solve this with bespoke cookiecutters, but those frequently require network access, drift over time, or are hard to extend.

`offclock` provides a terminal-first, local-first way to create a ready-to-work project skeleton in seconds. It focuses on reproducibility (templates are versioned and embedded), safety (no clobbering by default), and extensibility (users can install local templates).

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`Template`: A named project blueprint consisting of a manifest plus a set of files to render.

`Embedded Template`: A template packaged inside the `offclock` binary.

`User Template`: A template installed under the user's config directory.

`Render`: The process of expanding a template into concrete files with variable substitution.

`Project Root`: The output directory where `offclock init` renders a project.

`Plan`: A list of file operations (create/write/skip) that would be performed.

## 4. Concepts

### 4.1. Template-Driven Project Scaffolding

`offclock` renders a project by:

1. Selecting a template.
2. Collecting template variables (flags, defaults, optional interactive prompts).
3. Rendering files into a target directory.

### 4.2. Embedded + User Templates

`offclock` ships with a small set of embedded templates that work offline. Users can add their own templates without modifying the binary.

### 4.3. Safe, Idempotent Writes

Scaffolding tools should be safe by default. `offclock` produces a plan, avoids overwriting existing files unless explicitly asked, and supports mixed scenarios (e.g. adding CI files later).

## 5. Requirements

### 5.1. Core Executable

1. The project MUST provide a standalone executable named `offclock`.
2. The executable MUST be independent of other sleepless executables.
3. All commands MUST return exit code 0 on success and non-zero on failure.

### 5.2. Templates

1. `offclock template list` MUST list all available templates (embedded and user-installed).
2. `offclock template show <name>` MUST print template metadata (description, variables, included components).
3. `offclock template add <path>` MUST install a template directory into the user templates directory.
4. `offclock template remove <name>` MUST remove a user-installed template.
5. Embedded templates MUST be read-only.
6. Templates MUST be usable offline.

### 5.3. Project Initialization

1. `offclock init` MUST scaffold a project from a template into a target directory.
2. `offclock init --template <name>` MUST select a specific template.
3. `offclock init --name <project>` MUST set the project name variable.
4. `offclock init` MUST NOT overwrite existing files by default.
5. `offclock init --force` MAY overwrite existing files.
6. `offclock init --dry-run` MUST output a plan and MUST NOT write any files.
7. If the target directory does not exist, `offclock init` MUST create it.

### 5.4. Dockerfile + Container Support

1. Templates MAY include a `Dockerfile` and `.dockerignore`.
2. The default embedded templates MUST include at least one template that generates a `Dockerfile`.

### 5.5. Dependencies and Tooling

1. Templates MAY include dependency manifests appropriate to the template (e.g., `pyproject.toml`, `requirements.txt`, `go.mod`, `package.json`).
2. `offclock init` MUST render dependency manifests from the template.
3. `offclock` MUST NOT download dependencies automatically.

### 5.6. CI/CD Pipeline

1. Templates MAY include CI/CD pipeline configuration.
2. Embedded templates MUST include at least one CI/CD pipeline option.
3. `offclock init` MUST support selecting the CI/CD target via a parameter (e.g. `--ci github` or `--ci gitlab`) when the chosen template supports multiple CI targets.
4. If `--ci` is omitted and the template supports CI, `offclock init` MUST use the template-defined default CI target.
5. If a template does not support the requested CI target, `offclock init` MUST exit non-zero and list supported CI targets.
6. CI/CD pipeline configuration MUST be generated exclusively from the chosen template (no network calls).

### 5.7. Script Generation

1. `offclock script new <name> --python` MUST generate a standardized Python script with parameters already set up.
2. The generated Python script MUST include:
   - argparse-based CLI parsing
   - structured logging
   - a `main()` entrypoint and `if __name__ == "__main__":` guard
   - exit codes that reflect success/failure
3. `offclock script new` MUST NOT overwrite an existing script by default.

### 5.8. Configuration

1. User templates MUST be stored under `~/.config/offclock/templates/` on Unix-like systems.
2. `offclock` MUST allow overriding the config directory via an environment variable: `OFFCLOCK_CONFIG_DIR`.

## 6. Interface

### 6.1. Commands

```bash
# Scaffold a project
offclock init --template <name> --name <project> [--dir <path>] [--ci <target>] [--force] [--dry-run]

# Templates
offclock template list
offclock template show <name>
offclock template add <path>
offclock template remove <name>

# Script generation
offclock script new <name> --python [--dir <path>] [--force]
```

### 6.2. Global Flags

- `--dir <path>`: Output directory (defaults to current directory for `script`, and `./<name>` for `init` when `--name` is set)
- `--json`: Optional machine-readable output for list/show/plan

## 7. Behavior

### 7.1. Template Selection

1. When a template name exists as both embedded and user-installed, the user-installed template MUST take precedence.
2. Template names MUST be treated case-insensitively for lookup.

### 7.2. Variable Resolution

1. Variables MUST be resolved in the following order:
   - explicit flags
   - environment variables (if supported by the template)
   - template defaults
2. Missing required variables MUST cause a non-zero exit with an actionable error.

### 7.3. Rendering

1. Rendering MUST support variable substitution in file contents.
2. Rendering MUST support variable substitution in file paths (e.g. `{{project_name}}/src`).
3. When a file already exists:
   - default behavior MUST be skip with a warning
   - with `--force`, behavior MUST be overwrite

## 8. Error Handling

1. If the template cannot be found, `offclock` MUST list available templates and exit non-zero.
2. If `template add` is given an invalid template directory, it MUST describe the missing/invalid components and exit non-zero.
3. If rendering fails (e.g. template syntax error), `offclock` MUST identify the file/path where the error occurred.

## 9. Examples

```bash
# Create a new project scaffold
offclock init --template python-service --name aurora

# See what would happen without writing files
offclock init --template python-service --name aurora --dry-run

# Install a custom template
offclock template add ./my-templates/python-cli

# Generate a standardized script
offclock script new ingest --python --dir scripts
```

## 10. Security Considerations

1. `offclock` MUST NOT transmit project data over the network.
2. `offclock` MUST NOT execute arbitrary commands from templates.
3. Template rendering MUST treat template variables as data (no shell evaluation).

## 11. Testing Considerations

Test scenarios SHOULD include:

- Rendering to an empty directory
- Rendering to a non-empty directory (skip vs force)
- Template precedence (user overrides embedded)
- Template add/remove validation
- Dry-run plan correctness
- Path templating and content templating
- Script generation behavior and overwrite protections

## 12. References

### Normative References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997, https://www.rfc-editor.org/rfc/rfc2119
