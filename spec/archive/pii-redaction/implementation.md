# Implementation: redact (PII and secret redaction)

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Add a `redact` command group and internal package providing regex-based redaction and audit capabilities across strings, stdin, files, and directories, with safe in-place file editing.

Expose the `redact` command group under a new security-focused executable named `nightwatch` (and do not expose `redact` under `sleepless`).

## Prerequisites

- Confirm desired output format for `check` reports (counts only vs include file+line details).

## Phases

### Phase 0: Add `nightwatch` executable scaffold

**Goal**: A second executable exists that can host security-related commands.

- [x] Add a second entrypoint for `nightwatch` (new `main` package).
- [x] Add a `nightwatch` Cobra root command.
- [x] Ensure `nightwatch` has version/build info similar to `sleepless`.

**Milestone**: `nightwatch --help` runs with exit code 0.

### Phase 1: Pattern registry + string redaction

**Goal**: `nightwatch redact "..."` replaces default patterns with `[TYPE]`.

- [x] Implement default pattern registry.
- [x] Implement selection logic (defaults only).
- [x] Implement string redaction.
- [x] Wire the `redact` command group into `nightwatch`.
- [x] Ensure `sleepless` does not expose a `redact` command.

**Milestone**: Example in spec produces expected output under `nightwatch`.

### Phase 2: Stdin streaming

**Goal**: `nightwatch redact stdin` works in pipelines without large memory usage.

- [x] Implement streaming redaction with bounded buffers.
- [x] Add basic error handling and exit codes.

**Milestone**: Can process large stdin input with stable memory.

### Phase 3: File mode + in-place

**Goal**: `nightwatch redact file` outputs redacted content; `--in-place` safely modifies file with `.bak`.

- [x] Implement file read → stdout redaction.
- [x] Implement `.bak` creation and atomic replacement.
- [x] Implement restore-on-failure behavior.

**Milestone**: Integration test covers happy path and simulated failure.

### Phase 4: Directory mode

**Goal**: `nightwatch redact dir` supports `--output` and `--pattern`.

- [x] Implement recursive walk.
- [x] Implement glob filtering.
- [x] Preserve relative paths under output dir.
- [x] Skip non-UTF-8 files with warning.

**Milestone**: Directory output matches expected structure.

### Phase 5: Audit mode (`check`)

**Goal**: Implement `nightwatch redact check file/dir`.

- [x] Implement match detection returning counts by type.
- [x] Implement file report.
- [x] Implement directory report.

**Milestone**: Report is stable and does not leak sensitive values.

### Phase 6: Flags and extended/custom modes

**Goal**: Add selection flags and alternative redaction modes.

- [x] Implement `--extended` pattern set.
- [x] Implement `--only`/`--except`.
- [x] Implement `--mask` and `--hash` (mutually exclusive).
- [x] Implement `--custom` and `--custom-name`.

**Milestone**: All flags behave as specified and are covered by tests.

## Testing Plan

### Unit Tests

- Pattern matching per type.
- Mask/hash formatting tests.
- Selection logic tests (`only/except/extended`).

### Integration Tests

- Pipeline tests for stdin.
- In-place editing tests with backups.
- Directory traversal and output tests.

## Rollback Plan

- Remove `redact` commands and package; no migrations.

## Open Questions

- Is `nightwatch` the correct desired name for the security-focused executable? If not, what sleep-themed name should we use?
- Confirm whether directory traversal should follow symlinks.
- Confirm `check` report detail level.
