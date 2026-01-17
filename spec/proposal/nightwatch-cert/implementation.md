# Implementation: nightwatch-cert

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Add a `nightwatch cert` Cobra command group backed by a small `pkg/cert` library to parse PEM certificates/keys and provide inspect/expiry/fingerprint/hostname verification workflows.

## Prerequisites

- None (standard library only).

## Phases

### Phase 1: Parsing and summary library

**Goal**: Implement deterministic parsing and summaries.

- [ ] Add `pkg/cert` with PEM parsing for certificates and keys.
- [ ] Add helpers: earliest expiry, SHA-256 fingerprint, hostname verification.
- [ ] Add unit tests using small PEM fixtures.

**Milestone**: `go test ./...` passes with new `pkg/cert` tests.

### Phase 2: CLI wiring

**Goal**: Expose subcommands and output.

- [ ] Add `internal/nightwatch/cert.go` Cobra commands: `inspect`, `expires`, `fingerprint`, `verify-host`.
- [ ] Add `--json` output.
- [ ] Ensure key material is never printed.

**Milestone**: Manual `nightwatch cert inspect <pem>` prints expected summary.

## Testing Plan

### Unit Tests

- Parse cert PEM and chain PEM.
- Parse key PEM and assert output is summary-only.
- Hostname verification with SAN DNS names.
- Expiry threshold behavior.

### Integration Tests

- Not required (no external dependencies, no network).

## Rollback Plan

- Remove `internal/nightwatch/cert.go` and `pkg/cert`; no migrations.

## Open Questions

- Should `expires` return distinct exit codes for "expired" vs "expiring soon"?
