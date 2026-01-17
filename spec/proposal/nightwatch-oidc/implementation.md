# Implementation: nightwatch-oidc

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Add an offline `nightwatch oidc` command group with PKCE generation, authorization URL building, callback parsing, and OIDC ID token decode/lint.

## Prerequisites

- None (avoid network and heavy dependencies).

## Phases

### Phase 1: Libraries

**Goal**: Implement reusable helpers.

- [ ] Add `pkg/oidc` for PKCE and URL parsing/building.
- [ ] Add unit tests for PKCE, auth URL, callback parsing.

**Milestone**: `go test ./...` passes with new packages.

### Phase 2: CLI

**Goal**: Wire commands and outputs.

- [ ] Add `internal/nightwatch/oidc.go` Cobra commands: `pkce`, `state`, `nonce`, `auth-url`, `callback`, `idtoken decode|lint`.
- [ ] Implement stdin token input for ID token commands.
- [ ] Implement `--json` output.

**Milestone**: Manual runs produce correct outputs for sample tokens and URLs.

## Testing Plan

### Unit Tests

- PKCE challenge correctness for known test vectors.
- Auth URL encoding includes all required parameters.
- Callback parsing handles `error` and `code` flows.
- Lint checks produce expected failures.

### Integration Tests


## Rollback Plan

- Remove `internal/nightwatch/oidc.go` and `pkg/oidc`.

## Open Questions

- None.
