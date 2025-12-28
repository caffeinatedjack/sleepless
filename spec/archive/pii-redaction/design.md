# Design: redact (PII and secret redaction)
**Specification Reference**: [specification.md](specification.md)

**Status**: Draft

## 1. Context

Sleepless is a local-first CLI. A redaction command group integrates well because it can be pipeline-friendly (`stdin` → `stdout`) and does not need to modify task storage.

The main design constraints are safety (never leak unredacted values), performance (handle large logs), and correctness (avoid overly-broad matches while still catching common PII/secrets). Go’s `regexp` package (RE2) is safe from catastrophic backtracking but requires careful pattern design.

File operations must be safe, especially for `--in-place` updates. Directory processing must preserve structure and support file filtering.

## 2. Goals and Non-Goals

### Goals

- Implement `redact` commands per specification.
- Expose `redact` only under the additional security-focused executable `nightwatch`.
- Provide a clear pattern registry with default and extended sets.
- Support streaming redaction for stdin and large files.
- Provide safe in-place editing with backups.
- Provide audit-only `check` mode.

### Non-Goals

- Perfect detection for all jurisdictions/formats (e.g., every national ID format).
- Binary-file redaction.
- Semantic parsing of logs (this is regex-based).

## 3. Options Considered

### Option 0: Command placement (single vs. multiple executables)

#### Option 0A: Only `sleepless redact`

**Advantages**:
- No additional binary/build surface.
- Simplest UX (one tool).

**Disadvantages**:
- Security utilities are mixed into the task-management binary.
- Harder to install only security tooling in constrained environments.

**Complexity**: Low

#### Option 0B: Add `nightwatch` security executable

**Advantages**:
- Clear separation between productivity tooling and security tooling.
- Enables future security commands without expanding the primary UX surface.

**Disadvantages**:
- Requires a second Cobra root + build target.

**Complexity**: Medium

---

### Option 1: Load entire input into memory and apply regex replacements

**Advantages**:
- Simplest implementation.
- Easy to get correct global replacements.

**Disadvantages**:
- Memory usage scales with input size; bad for large logs.

**Complexity**: Low

### Option 2: Streaming processing with bounded buffers

**Advantages**:
- Works for very large inputs and pipelines.
- Predictable memory usage.

**Disadvantages**:
- Harder to handle patterns spanning buffer boundaries.

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 0B (add `nightwatch`) + Option 2 (streaming processing)

**Rationale**: This feature set is security-oriented and benefits from a clearly named, security-focused executable (`nightwatch`). Keeping `redact` out of the primary `sleepless` CLI reduces surface area and keeps the task-management UX focused. Additionally, the tool is intended for logs and diffs which can be large, so bounded memory behavior is more important than perfectly matching rare multi-line patterns.

**Key Factors**:
1. Pipeline friendliness.
2. Predictable memory usage.
3. Safety-first behavior for file writes.

## 5. Detailed Design

### Architecture Overview

- `cmd/redact.go` (new): `redact` command group (wired into `nightwatch`).
- `cmd/redact_*.go` (new): `stdin`, `file`, `dir`, `check`.
- `cmd/root.go` (existing): `sleepless` root command.
- `cmd/nightwatch_root.go` (new): `nightwatch` root command.
- `main.go` (existing): `sleepless` entrypoint.
- `nightwatch/main.go` or equivalent second entrypoint (new): `nightwatch` entrypoint.
- `pkg/redact/` (new): pattern registry, selection logic, redaction engine, audit reporting.

### Component Design

#### Pattern Registry

- Maintain a map of `PatternType -> regexp.Regexp + replacement behavior`.
- Provide `DefaultSet` and `ExtendedSet`.
- Add support for one custom regex via flags (`--custom`, `--custom-name`).

#### Selection Logic

Compute active patterns per spec order: defaults → extended → only → except.

#### Redaction Engine

Provide two entry points:

- `Redact(reader, writer, options) error`
- `Check(reader, options) (Report, error)`

Streaming approach:

- Read chunks using `bufio.Reader`.
- Keep an overlap tail buffer (configurable size) to reduce boundary issues.
- Apply regex replacements to each processed chunk.

Boundary limitation is acceptable for patterns that require long spans; document as a known limitation if needed.

#### In-place File Editing

Algorithm:

1. Create `path.bak` by renaming or copying original.
2. Create a temp file in the same directory.
3. Redact from backup → temp.
4. Atomically replace original with temp.
5. On failure, restore from backup.

### Data Design

- No persistent data beyond `.bak` files for `--in-place`.

### API / CLI Design

- `redact <string>` can be implemented as: join remaining args with spaces (preserving shell quoting behavior).
- `stdin` reads from `os.Stdin`.
- `check` outputs a text report summarizing counts by type and listing file paths for directory scans.

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| Streaming with overlap buffer | Handles big inputs | May miss rare cross-boundary patterns | Prioritize bounded memory and pipeline use |
| Regex-based detection | Simple + safe RE2 | False positives/negatives | Practical trade-off for CLI utility |

## 7. Cross-Cutting Concerns

### Security

- Never echo raw matches in logs/errors.
- Be conservative in audit output if it might leak sensitive values; prefer counts and locations.

### Performance

- Compile regexes once.
- Stream processing.

### Reliability

- In-place writes must be crash-safe and restore originals on errors.

### Testing

- Table tests for patterns.
- Filesystem tests for in-place editing and directory output.

## 8. Implementation Plan

- Phase 1: Scaffold commands and implement string redaction.
- Phase 2: Implement stdin streaming.
- Phase 3: Implement file mode and `--in-place`.
- Phase 4: Implement directory mode with `--output` and `--pattern`.
- Phase 5: Implement `check` report.
- Phase 6: Add `--only/--except/--extended` and mask/hash/custom.

### Migration Strategy

None.

## 9. Open Questions

- Is `nightwatch` the desired name for the security-focused executable? If not, what sleep-themed name do you prefer?
- Should `check` output include match previews (riskier) or only counts/line numbers?
- What overlap buffer size is sufficient for extended patterns like JWTs?
- Should directory mode skip hidden directories by default?
