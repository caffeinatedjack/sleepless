# Implementation: jwt (JSON Web Token utilities)

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Implement a new Cobra command group `jwt` with supporting package code for JWT decoding, verification, creation, and expiration checking. Use the `golang-jwt/jwt/v5` library for cryptographic operations.

## Prerequisites

- Add `github.com/golang-jwt/jwt/v5` dependency to `go.mod`
- Confirm approach for open design questions (default algorithm, secret input methods)
- Review and align PEM key format requirements with target use cases

## Phases

### Phase 1: Core infrastructure and decoding

**Goal**: Basic JWT parsing and display without verification.

**Tasks**:
- [ ] Add dependency: `go get github.com/golang-jwt/jwt/v5`
- [ ] Create `internal/nightwatch/jwt/` package directory
- [ ] Implement `decode.go` with token parsing (split, base64url decode, JSON parse)
- [ ] Implement `output.go` with formatters for text and JSON output
- [ ] Create `cmd/nightwatch/jwt.go` as root command for JWT group
- [ ] Add `decode`, `header`, `payload` subcommands
- [ ] Implement stdin and argument input handling
- [ ] Add unit tests for token parsing and decoding

**Success Criteria**:
- `nightwatch jwt decode <token>` displays header and payload for valid tokens
- `nightwatch jwt header <token>` outputs raw JSON suitable for piping to `jq`
- `nightwatch jwt payload <token>` outputs raw JSON suitable for piping to `jq`
- Invalid tokens produce clear error messages and exit code 1
- Works with tokens from both arguments and stdin

### Phase 2: Verification

**Goal**: Signature verification with HMAC secrets and RSA public keys.

**Tasks**:
- [ ] Implement `keys.go` with PEM key loading functions
  - [ ] Support PKCS1 RSA keys
  - [ ] Support PKCS8 keys (RSA and ECDSA)
  - [ ] Support PKIX public keys
  - [ ] Add key file permission checking
- [ ] Implement `verify.go` with verification logic
  - [ ] HMAC verification (HS256/384/512)
  - [ ] RSA verification (RS256/384/512)
  - [ ] Standard claim validation (exp, nbf)
  - [ ] Error type mapping to exit codes
- [ ] Add `verify` subcommand with `--secret` and `--key` flags
- [ ] Add unit tests for key loading
- [ ] Add unit tests for verification (valid and invalid cases)

**Success Criteria**:
- HMAC verification works with `--secret` flag
- RSA verification works with `--key` flag and PEM public keys
- Expired tokens return exit code 3 with clear message
- Not-yet-valid tokens return exit code 4 with clear message
- Invalid signatures return exit code 2 with clear message
- Verification output includes all required JSON fields

### Phase 3: Token creation (HMAC)

**Goal**: Create and sign tokens with HMAC algorithms.

**Tasks**:
- [ ] Implement duration parsing (`ParseDuration` supporting days)
- [ ] Implement claim parsing from `--claim` flags
- [ ] Implement `--payload` JSON parsing
- [ ] Implement `create.go` with token creation logic
  - [ ] Build claims map from flags
  - [ ] Handle standard claims (iss, sub, aud, exp, nbf, iat, jti)
  - [ ] Compute absolute timestamps for exp/nbf durations
  - [ ] Default iat to current time if not specified
  - [ ] Sign with HMAC (HS256 default)
- [ ] Add `create` subcommand with all required flags
- [ ] Add unit tests for duration parsing
- [ ] Add unit tests for claim parsing
- [ ] Add integration test for create + verify round-trip (HS256)

**Success Criteria**:
- Can create tokens with `--secret` and custom claims
- Duration formats work correctly (`1h`, `30m`, `7d`)
- Standard claim shortcuts work (--iss, --sub, --aud, --exp)
- Created tokens can be verified successfully
- Output format matches specification (token string by default, JSON with --json)

### Phase 4: Token creation (RSA/ECDSA)

**Goal**: Extend creation to support asymmetric algorithms.

**Tasks**:
- [ ] Extend `keys.go` to load private keys for signing
- [ ] Extend `create.go` to support RSA signing (RS256/384/512)
- [ ] Extend `create.go` to support ECDSA signing (ES256/384/512)
- [ ] Add `--alg` flag validation and algorithm selection
- [ ] Add unit tests for all algorithms
- [ ] Add integration tests for create + verify round-trip (all algorithms)

**Success Criteria**:
- Can create RS256/384/512 tokens with `--key` flag and PEM private keys
- Can create ES256/384/512 tokens with `--key` flag and PEM private keys
- Algorithm selection works correctly (auto-detect from key, or explicit via --alg)
- All created tokens verify successfully with corresponding public keys

### Phase 5: Expiration utility

**Goal**: Implement expiration checking command.

**Tasks**:
- [ ] Implement `exp.go` with expiration checking logic
  - [ ] Extract iat and exp claims
  - [ ] Compute status (valid/expired/no-expiry)
  - [ ] Calculate remaining time for valid tokens
  - [ ] Format output (text and JSON)
- [ ] Add `exp` subcommand
- [ ] Add unit tests with fixed timestamps
- [ ] Add tests for tokens with/without iat and exp claims

**Success Criteria**:
- `nightwatch jwt exp <token>` displays issued time, expiration time, and status
- Correctly calculates remaining time for valid tokens
- Handles tokens without exp/iat claims gracefully
- JSON output matches specification schema

### Phase 6: Security hardening and polish

**Goal**: Address security considerations and improve UX.

**Tasks**:
- [ ] Add warning for secrets in command-line arguments (visible in process list)
- [ ] Add `--secret-file` flag support for secure secret loading
- [ ] Implement key file permission validation (warn if not 600)
- [ ] Add clear "NOT VERIFIED" warning to decode command output
- [ ] Improve error messages (include examples of correct usage)
- [ ] Add support for multiple tokens from stdin (line-by-line processing)
- [ ] Add `--redact` flag for privacy-safe output demonstrations

**Success Criteria**:
- Security warnings display appropriately
- Key files with overly permissive permissions trigger warnings
- Decode output clearly indicates lack of verification
- Error messages are actionable and helpful
- Multiple token processing works correctly

### Phase 7: Testing and documentation

**Goal**: Comprehensive testing and usage documentation.

**Tasks**:
- [ ] Add integration tests for all commands
- [ ] Add security-focused tests:
  - [ ] Algorithm confusion attacks
  - [ ] Token tampering detection
  - [ ] Signature stripping
- [ ] Add golden output tests for JSON output
- [ ] Add examples to command help text
- [ ] Create usage documentation with common scenarios
- [ ] Add tests for all exit codes
- [ ] Add edge case tests (empty claims, special characters, large tokens)

**Success Criteria**:
- Test coverage >80% for jwt package
- All specified exit codes are tested
- JSON output schemas validated in tests
- Help text includes practical examples
- Security test cases cover common attacks

## Testing Plan

### Unit Tests

| Component | Test Cases |
|-----------|------------|
| Token parsing | Valid tokens (all algorithms), malformed tokens, invalid base64, invalid JSON |
| Duration parsing | Valid durations (s/m/h/d), invalid formats, edge cases (0, negative) |
| Claim parsing | Single claims, multiple claims, JSON values, special characters |
| Key loading | Various PEM formats, invalid PEM, missing files, permission checks |
| Verification | Valid/invalid signatures, expired tokens, nbf validation |
| Creation | All algorithms, all claim types, duration computation |
| Expiration | With/without iat, with/without exp, expired/valid/future tokens |

### Integration Tests

| Scenario | Description |
|----------|-------------|
| Decode flow | Token → decode → verify output format |
| Verify flow | Token + secret/key → verify → check exit code |
| Create + verify | create → verify round-trip for each algorithm |
| Stdin processing | Multiple tokens via stdin → multiple results |
| Pipeline | decode → payload → jq → extract claim value |
| Error handling | Various error conditions → correct exit codes and messages |

### Security Tests

| Attack Vector | Test |
|---------------|------|
| Algorithm confusion | Verify RS256 token with HS256 + public key as secret |
| Signature stripping | Remove signature from token → verify fails |
| Token tampering | Modify payload → verify fails |
| Key type mismatch | Use wrong key type for algorithm |
| Expired token acceptance | Verify expired token fails with exit code 3 |

### Golden Output Tests

Create reference outputs for:
- Decode output (text and JSON) for sample tokens
- Verify output (success and various failures)
- Create output for different claim combinations
- Exp output for various token states

## Dependencies

```
go.mod additions:
  github.com/golang-jwt/jwt/v5 v5.2.0
```

## Rollback Plan

If critical issues are discovered:
1. Remove `cmd/nightwatch/jwt.go` command registration
2. Remove `internal/nightwatch/jwt/` package
3. Remove jwt dependency from `go.mod`
4. No data migrations or configuration changes needed

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Crypto library vulnerability | High | Use well-vetted library (golang-jwt), monitor security advisories |
| Algorithm confusion attacks | High | Validate algorithm matches expected type, clear documentation |
| Secret exposure via CLI args | Medium | Warn users, document stdin/env var usage, provide --secret-file |
| PEM parsing edge cases | Medium | Comprehensive testing with various key formats, clear error messages |
| Breaking changes in jwt library | Low | Pin to specific version, test before upgrading |

## Performance Considerations

- Token operations are CPU-bound (base64, JSON, crypto)
- Expected usage: single tokens or small batches
- No performance optimization needed for v1
- If batch processing becomes common, consider:
  - Key caching for multiple tokens
  - Parallel processing for independent verifications

## Maintenance Notes

- Monitor `golang-jwt/jwt` releases for security updates
- Periodically review supported algorithms (add new, deprecate old)
- Update tests if JWT specifications evolve
- Consider adding algorithm-specific security warnings (e.g., HS256 vs RS256 trade-offs)

## Open Questions Resolution

Based on design discussions, recommended resolutions:

1. **Default Algorithm**: Default to HS256 when `--secret` provided; require `--alg` when using `--key`
2. **Secret Input**: Implement `--secret-file` flag in Phase 6
3. **Multiple Tokens**: Support line-by-line processing in Phase 6
4. **Additional Key Formats**: PEM only for v1; JWK as future enhancement
5. **Claim Value Types**: Attempt JSON parsing; fall back to string in Phase 3
