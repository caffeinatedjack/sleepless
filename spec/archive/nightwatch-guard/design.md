# Design: nightwatch-guard
**Specification Reference**: [specification.md](specification.md)
Status: Draft

## 1. Context

`nightwatch` is an offline, scriptable CLI implemented with Cobra. The codebase already contains `pkg/redact`, which defines patterns and supports both rewriting (`Redact*`) and detection-only scanning (`Check`).

A guard command should reuse existing patterns, avoid external dependencies, and remain safe for use in CI (no raw secret output).

## 2. Goals and Non-Goals

### Goals

- Provide `nightwatch guard` subcommands that scan staged/worktree/path targets.
- Produce CI-friendly exit codes and optional JSON output.
- Never output raw secrets/PII.

### Non-Goals

- Network-backed secret scanning or online lookups.
- Automatic secret revocation or remediation.
- Deep git history scanning (this proposal focuses on local changes).

## 3. Options Considered

### Option 1: Scan staged snapshot via git plumbing (preferred)

Scan staged file content using git plumbing (`git diff --cached --name-only` + `git show :path`) and scan working tree files directly for `worktree`.

**Advantages**:
- Correctly reflects what will be committed.
- Does not require parsing unified diffs.
- Produces accurate file-relative line/column locations.

**Disadvantages**:
- Requires `git` availability.

**Complexity**: Medium

### Option 2: Parse `git diff` output and map line numbers

Scan unified diffs and attempt to map matches back to file/line.

**Advantages**:
- Only reads diffs; can avoid reading full files.

**Disadvantages**:
- Diff parsing is error-prone (renames, binary diffs, context-less patches).
- Line numbers in diffs can be misleading with zero context.

**Complexity**: High

## 4. Decision

**Chosen Option**: Option 1

**Rationale**: It is simpler, more reliable, and produces better locations without fragile diff parsing.

**Key Factors**:
- Accuracy of staged snapshot scanning
- Low maintenance cost
- Consistent behavior across git edge cases

## 5. Detailed Design

### Architecture Overview

- `internal/nightwatch/guard.go` registers Cobra commands and performs target selection.
- `pkg/guard` (new) contains reusable scanning and baseline logic:
  - directory traversal and UTF-8 checks
  - finding fingerprint computation
  - baseline load/save
  - integration with `pkg/redact` patterns

### Component Design

- **Scanner**: uses the same regex patterns as `pkg/redact` and produces findings with file/line/column plus a safe excerpt (masked/redacted) when needed.
- **Baseline**: JSON file storing fingerprints; does not store raw matches.

### Data Design

Baseline JSON format (normative fields):

```json
{
  "version": 1,
  "created_at": "2026-01-15T00:00:00Z",
  "fingerprints": [
    "sha256:..."
  ]
}
```

### API Design

Not applicable (CLI-only).

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| Scan full file snapshots | Accurate locations | More IO | IO cost is acceptable for local/CI use |
| No raw match output | Safer defaults | Less immediate detail | Minimizes accidental leaks in CI logs |

## 7. Cross-Cutting Concerns

### Security

- Always redact/mask excerpts.
- Avoid logging raw file content.

### Performance

- Skip binary files.
- Stream scanning where possible.

### Reliability

- Treat missing git or non-repo as explicit errors for git-based targets.

### Testing

- Unit tests for baseline, fingerprinting, and scanner.
- Integration-style tests using a temporary git repo for staged scanning.

## 8. Implementation Plan

- Implement `pkg/guard` scanning + baseline.
- Add Cobra wiring in `internal/nightwatch/guard.go`.
- Add unit tests and minimal integration tests.

### Migration Strategy

None.

## 9. Open Questions

- Baseline fingerprint stability: should it include file path or only match type/value?
