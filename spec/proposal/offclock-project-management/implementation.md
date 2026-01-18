# Implementation: offclock-project-management

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Implement `offclock` as a standalone scaffolding executable. It discovers embedded and user templates, validates template manifests, builds a safe plan for rendering, and applies the plan to produce a ready-to-use project skeleton. It also includes a small script generator (Python first).

## Prerequisites

- Add `golang.org/x/crypto` only if needed later (not required for scaffolding)
- Use existing project dependency on cobra

## Phases

### Phase 1: CLI Skeleton + Config

**Goal**: Create executable, root command, and config paths.

- [ ] 1.1 Create `cmd/offclock/main.go`
- [ ] 1.2 Create `internal/offclock/root.go` (cobra root)
- [ ] 1.3 Implement config dir resolution (`OFFCLOCK_CONFIG_DIR`, default `~/.config/offclock`)
- [ ] 1.4 Wire `template`, `init`, `script` subcommands

**Milestone**: `offclock --help` shows commands and global flags.

### Phase 2: Template System (Embedded + User)

**Goal**: Load templates from embed and user directory with precedence rules.

- [ ] 2.1 Add `pkg/scaffold/template/manifest.go` + validation
- [ ] 2.2 Add `pkg/scaffold/template/loader.go` (embedded + user discovery)
- [ ] 2.3 Implement `offclock template list`
- [ ] 2.4 Implement `offclock template show <name>`
- [ ] 2.5 Implement `offclock template add <path>` (copy into user templates)
- [ ] 2.6 Implement `offclock template remove <name>` (user-only)
- [ ] 2.7 Add unit tests for manifest validation and precedence

**Milestone**: Can list embedded templates and install/remove user templates.

### Phase 3: Rendering Engine

**Goal**: Render file contents + paths with a restricted function set.

- [ ] 3.1 Add `pkg/scaffold/render/funcs.go` (safe funcs: slug/snake/kebab/lower/upper/trim)
- [ ] 3.2 Add `pkg/scaffold/render/render.go` (path + content rendering)
- [ ] 3.3 Add unit tests for templating, including path templating

**Milestone**: Can render a small in-memory template tree deterministically.

### Phase 4: Plan + Apply

**Goal**: Build a plan (dry-run) and safely apply it.

- [ ] 4.1 Add `pkg/scaffold/plan.go` (create/write/skip)
- [ ] 4.2 Add `pkg/scaffold/apply.go` (force vs skip)
- [ ] 4.3 Implement `offclock init` with `--dry-run`, `--force`, `--dir`, `--name`, `--template`, `--ci`
- [ ] 4.4 Add integration tests using temp dirs

**Milestone**: `offclock init --dry-run` prints a plan; running without `--dry-run` creates files.

### Phase 5: Embedded Templates

**Goal**: Provide at least one high-quality embedded template that includes Docker + CI.

- [ ] 5.1 Create embedded template(s) under `internal/offclock/templates/` and embed with `go:embed`
- [ ] 5.2 Include Dockerfile + .dockerignore
- [ ] 5.3 Include CI config variants (GitHub Actions + GitLab CI) selectable via `--ci`
- [ ] 5.4 Add a template that generates a Python service (pyproject/requirements, lint config)
- [ ] 5.5 Add a template that generates a Go service (go.mod, basic layout, lint/test config)
- [ ] 5.6 Add golden tests that validate rendered output

**Milestone**: `offclock init --template python-service --name demo --ci github` generates a runnable skeleton.

### Phase 6: Script Generator

**Goal**: Generate standardized scripts.

- [ ] 6.1 Implement `offclock script new <name> --python`
- [ ] 6.2 Add overwrite protections and `--force`
- [ ] 6.3 Add unit/integration tests for script generation

**Milestone**: `offclock script new ingest --python` writes a usable script with argparse + logging.

## Testing Plan

### Unit Tests

- Manifest validation
- Template discovery + precedence
- Render functions (safe funcs)
- Path templating
- Plan creation (collision detection)

### Integration Tests

- Init into empty dir
- Init into non-empty dir (skip vs force)
- Dry-run plan output does not write
- Install template then init using it
- Script generation into specified directory

## Rollback Plan

- Remove `cmd/offclock/`, `internal/offclock/`, `pkg/scaffold/`
- Remove `offclock` from Makefile build/install targets

## Open Questions

- CI support scope: GitHub Actions only, or also GitLab CI?
- Do we want `offclock init` to optionally run `git init` and create an initial commit?
- Should templates support components (docker/ci) toggles or separate templates?