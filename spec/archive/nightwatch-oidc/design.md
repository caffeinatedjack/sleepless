# Design: nightwatch-oidc
**Specification Reference**: [specification.md](specification.md)
Status: Draft

## 1. Context

OIDC workflows are a combination of URL construction/parsing and JWT handling. The existing Nightwatch JWT commands cover generic JWT operations, but OIDC debugging needs additional helpers: PKCE generation, authorization URL creation, callback parsing, and OIDC-focused claim checks.

The design must remain offline and avoid introducing provider-specific behavior.

## 2. Goals and Non-Goals

### Goals

- Provide `nightwatch oidc` subcommands that are useful without network access.
- Reuse existing JWT parsing/decoding where practical.
- Provide deterministic, script-friendly output (`--json`).

### Non-Goals

- Fetching discovery docs or JWKS over the network.
- Implementing a full OAuth client or browser automation.
- Provider-specific CLI wrappers.

## 3. Options Considered

### Option 1: Implement a focused `oidc` command group (preferred)

**Advantages**:
- Clear, discoverable UX.
- Keeps OIDC-specific semantics separate from generic JWT tooling.

**Disadvantages**:
- Some overlap with `nightwatch jwt` functionality.

**Complexity**: Medium

### Option 2: Add OIDC flags/subcommands under `nightwatch jwt`

**Advantages**:
- Fewer top-level groups.

**Disadvantages**:
- Conflates generic JWT tooling with OIDC rules.
- Harder to discover and document.

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 1

**Rationale**: OIDC has distinct workflows and terminology; a dedicated group keeps UX simpler.

**Key Factors**:
- Discoverability
- Keeping generic JWT tools generic

## 5. Detailed Design

### Architecture Overview

- `internal/nightwatch/oidc.go` registers the Cobra commands.
- `pkg/oidc` (new) provides:
  - PKCE generation
  - state/nonce generation
  - auth URL builder
  - callback URL parser
  - OIDC ID token lint helpers

### Component Design

- `idtoken decode/lint` reads token from arg or stdin (matching existing JWT patterns).

### Data Design

`--json` output provides stable fields for scripting (verifier/challenge, parsed callback params, lint results).

### API Design

Not applicable (CLI-only).

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| Offline-only | Safer + predictable | No auto-fetch | Matches Nightwatch goals |

## 7. Cross-Cutting Concerns

### Security

- Do not print secrets beyond what the user provided.
- Make `decode` vs `verify` semantics explicit.

### Performance

- Small string operations; negligible.

### Reliability

- Clear errors for missing required flags and invalid JWT/JWKS inputs.

### Testing

- Pure unit tests for PKCE, URL building, parsing, and linting.

## 8. Implementation Plan

- Implement `pkg/oidc` and `pkg/jwks`.
- Add Cobra wiring.
- Add tests.

### Migration Strategy

None.

## 9. Open Questions

- Should `auth-url` default `response_type=code` and `code_challenge_method=S256` when PKCE is used?
