# Design: nightwatch-cert
**Specification Reference**: [specification.md](specification.md)
Status: Draft

## 1. Context

Nightwatch is an offline, scriptable CLI. Certificate tooling in many environments is fragmented (OpenSSL, Java `keytool`, bespoke scripts). This proposal adds a small certificate inspection command group using Go’s standard library (`crypto/x509`, `encoding/pem`).

## 2. Goals and Non-Goals

### Goals

- Parse PEM files containing certificates, chains, and keys.
- Provide common developer operations: inspect, expiry checks, fingerprints, hostname verification.
- Keep output safe and avoid printing key material.

### Non-Goals

- Network-based TLS probing (e.g., connecting to hosts).
- Truststore modification for specific runtimes (Java keystore) in this proposal.
- Certificate generation or CA management.

## 3. Options Considered

### Option 1: Implement using Go `crypto/x509` (preferred)

**Advantages**:
- No new dependencies.
- Hostname verification behavior matches the Go runtime.

**Disadvantages**:
- Output may differ slightly from OpenSSL conventions.

**Complexity**: Low

### Option 2: Shell out to OpenSSL

**Advantages**:
- Familiar output.

**Disadvantages**:
- Adds external runtime dependency.
- Harder to produce stable JSON output.

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 1

**Rationale**: Keep Nightwatch self-contained and offline with predictable parsing.

**Key Factors**:
- No external dependencies
- Scriptability
- Safe output defaults

## 5. Detailed Design

### Architecture Overview

- `internal/nightwatch/cert.go` registers Cobra commands and handles output formatting.
- `pkg/cert` (new) provides parsing and inspection helpers reusable by tests and other commands.

### Component Design

- PEM reader extracts blocks and attempts to parse them as certificate or key.
- Certificate summary normalizes common fields and collects SANs.

### Data Design

If `--json` is provided, output includes arrays of parsed certificates/keys with summary fields only.

### API Design

Not applicable (CLI-only).

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| Summary-only output | Safer sharing | Less detail | Avoids leaking internal extension data |

## 7. Cross-Cutting Concerns

### Security

- Never print private key material.

### Performance

- PEM parsing is small and bounded.

### Reliability

- Fail fast on unreadable/unparseable files.

### Testing

- Use embedded test PEM fixtures for deterministic parsing tests.

## 8. Implementation Plan

- Implement `pkg/cert` parsing and summary.
- Add Cobra wiring in `internal/nightwatch/cert.go`.
- Add tests.

### Migration Strategy

None.

## 9. Open Questions

- Should `inspect` show key usage / extended key usage by default or behind a flag?
