# Implementation: nightwatch-guard

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Implement a new `nightwatch guard` Cobra command group that scans staged snapshots, working tree changes, or explicit filesystem paths for `pkg/redact` patterns. Add baseline support for incremental rollout.

## Prerequisites

- None (must remain offline, no new external dependencies).

## Phases

### Phase 1: Core scanner and baseline

**Goal**: Provide reusable scanning logic and baseline suppression.

- [ ] Add `pkg/guard` scanner that returns file/line/column and counts.
- [ ] Add baseline JSON load/save and fingerprint suppression.
- [ ] Add unit tests for fingerprinting and baseline behavior.

**Milestone**: `go test ./...` includes new guard tests and passes.

### Phase 2: CLI wiring

**Goal**: Expose commands and output formats.

- [ ] Add `internal/nightwatch/guard.go` Cobra commands: `staged`, `worktree`, `path`, `baseline`.
- [ ] Implement `--json` output and safe human-readable output.
- [ ] Ensure exit codes match spec (no findings vs findings vs error).

**Milestone**: Manual run of `nightwatch guard staged` works in a git repo and returns correct exit code.

## Testing Plan

### Unit Tests

- Findings detected for representative patterns.
- Baseline suppression works.
- JSON output does not include raw matches.

### Integration Tests

- Temporary git repo test that stages content differing from worktree and verifies `staged` scans the staged snapshot.

## Rollback Plan

- Remove `internal/nightwatch/guard.go` and `pkg/guard` if needed; no data migrations.

## Open Questions

- Should `worktree` include untracked files by default or behind a flag?
