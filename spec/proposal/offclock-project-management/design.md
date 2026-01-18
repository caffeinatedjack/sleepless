# Design: offclock-project-management
**Specification Reference**: [specification.md](specification.md)
Status: Draft

## 1. Context

Offclock adds a new executable to the sleepless ecosystem focused on project bootstrapping. Unlike regimen (task management) and nightwatch (security/dev utilities), offclock is a scaffolding tool: it creates files and directory structures based on templates.

Key constraints:

- Must work offline (templates embedded + local user templates)
- Must be safe by default (no overwrite unless requested)
- Must not execute arbitrary code from templates
- Must be extensible without recompiling

## 2. Goals and Non-Goals

### Goals

- One command to scaffold a usable project: layout, dependency manifest(s), Dockerfile, CI/CD config
- Embedded templates for "works out of the box" offline scaffolding
- User templates that override/extend embedded ones
- Predictable output and safe overwrites (plan + dry-run)
- Standard script generation (starting with Python)

### Non-Goals

- Remote template registries or network fetching (local-only)
- Running dependency installers or provisioning external services
- Highly opinionated monorepo management

## 3. Options Considered

### Option 1: Minimal string replacement engine

Templates are simple directories; variables are replaced using token substitution.

**Advantages**:
- Simple to implement
- Easy to reason about

**Disadvantages**:
- Hard to express conditionals (e.g. include CI only when requested)
- Poor ergonomics for template authors

**Complexity**: Low

### Option 2: Go text/template rendering with a template manifest

Templates use `text/template` for file contents and file paths. A manifest describes required variables, defaults, and which files belong to the template.

**Advantages**:
- Strong, battle-tested templating engine
- Conditionals supported without inventing a DSL
- Manifest gives structure and validation

**Disadvantages**:
- Need to constrain template functions for safety

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 2 (text/template + manifest)

**Rationale**: It provides enough expressiveness for "project starter kits" (CI options, optional Docker, optional scripts) without requiring plugins or executing arbitrary code.

**Key Factors**:

1. Safety: keep template functions minimal and deterministic
2. Extensibility: allow user templates and future embedded templates
3. Developer experience: templates can be authored and tested locally

## 5. Detailed Design

### Architecture Overview

```
cmd/offclock/main.go
internal/offclock/
  root.go
  init.go
  template.go
  script.go
pkg/scaffold/
  plan.go             # build plan (create/write/skip)
  apply.go            # apply plan with overwrite rules
  errors.go
pkg/scaffold/template/
  loader.go           # embedded + user template discovery
  manifest.go         # manifest schema + validation
pkg/scaffold/render/
  render.go            # render file paths and file contents
  funcs.go             # restricted template funcs (safe)
```

### Template Format

Each template is a directory containing:

- `template.yaml` (manifest)
- `files/` (tree of files to render)

Example manifest shape (illustrative):

```yaml
name: python-service
description: "Python service with Docker and CI"
vars:
  project_name:
    required: true
  module_name:
    default: "{{ slug project_name }}"
components:
  docker: true
  ci:
    - github-actions
    - gitlab-ci
```

Rendering rules:

- File contents are rendered with `text/template`.
- File paths are also rendered with `text/template`.
- Allowed template funcs are strictly controlled (e.g. `lower`, `upper`, `slug`, `snake`, `kebab`, `trim`). No filesystem, no exec, no network.

### Config + Discovery

- Embedded templates are compiled with `go:embed`.
- User templates are loaded from `~/.config/offclock/templates/` (override embedded by name).
- `OFFCLOCK_CONFIG_DIR` overrides the root config path.

### Init Planning + Apply

`offclock init` builds a plan first:

- enumerate the rendered output paths
- classify each operation: create, overwrite, skip
- emit plan for `--dry-run` and for human inspection

Apply behavior:

- default: refuse to overwrite and skip collisions
- `--force`: allow overwrite

### Script Generation

`offclock script new <name> --python` writes a standardized script stub.

The generated script includes:

- argparse CLI with placeholder parameters
- structured logging configuration
- `main()` that returns an int exit code

This is implemented as a small built-in "single-file template" rather than a full project template.

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| Offline-only templates | Works everywhere | No central registry | Matches project philosophy |
| text/template | Expressive templates | Need to constrain funcs | Restrict funcs for safety |
| Plan/apply step | Safe writes | More code | Prevents clobbering user work |

## 7. Cross-Cutting Concerns

### Security

- No template command execution
- No network access
- Restricted templating functions

### Performance

- Rendering is filesystem-bound; acceptable for typical scaffolds (< 500 files)

### Reliability

- Use atomic writes where feasible (temp + rename) when overwriting
- Deterministic plan output

### Testing

- Golden tests for rendering + plan output
- Temp-dir integration tests for init/script/template commands

## 8. Implementation Plan

See [implementation.md](implementation.md).

## 9. Open Questions

- Do we want to treat CI selection as a single `--ci` flag across all templates, or allow template-defined component flags (e.g. `--ci github --docker --editorconfig`)?
- Should the template default CI always be `github` when supported, or can templates choose their own default (recommended: template-defined default)?

- Should `offclock init` optionally run `git init`?
- Should templates support "components" (docker/ci/scripts) as flags, or be separate templates?
