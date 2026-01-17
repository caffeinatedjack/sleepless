# nightwatch-oidc

**Depends on**: none
**Affected files**: internal/nightwatch/oidc.go, pkg/oidc/*

## Abstract

This specification adds a `nightwatch oidc` command group containing small, offline utilities that make OpenID Connect (OIDC) integration and debugging faster. It includes PKCE generation, authorization URL construction, callback URL parsing, and ID token (JWT) inspection and linting. The command group is designed to avoid sending tokens to third parties and to be usable in scripts.

## 1. Introduction

OIDC/OAuth integrations are frequent sources of friction in large environments: subtle URL encoding mistakes, `state`/`nonce` mismatches, token expiry confusion, and signature verification failures. Debugging often requires bouncing between browser tools, provider dashboards, and ad-hoc scripts.

Nightwatch already provides JWT tooling; however, OIDC work commonly needs additional glue around authorization URLs, callback parsing, PKCE values, and OIDC-specific claim checks.

This proposal adds `nightwatch oidc` as a focused developer command group for local-first OIDC debugging.

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`OIDC`: OpenID Connect, an identity layer on top of OAuth 2.0.

`Authorization Request`: A request sent to the provider authorization endpoint (typically via browser redirect).

`Callback URL`: The redirect URI invoked by the provider after authorization.

`PKCE`: Proof Key for Code Exchange, defined in RFC 7636.

`ID Token`: A JWT issued by an OIDC provider containing identity claims.


## 4. Concepts

### 4.1. Offline-first debugging

All OIDC commands in this proposal are local computations:

- constructing URLs
- parsing URLs
- decoding/verifying JWTs using locally provided keys

The implementation MUST NOT perform network requests.

### 4.2. OIDC-specific JWT checks

OIDC ID tokens have conventions and requirements beyond generic JWTs (issuer/audience rules, nonce usage, time-based claims). `nightwatch oidc idtoken lint` performs these checks.

## 5. Requirements

### 5.1. Command Group

1. The CLI MUST provide `nightwatch oidc`.
2. Commands MUST be offline and MUST NOT make network requests.
3. Token input for ID token subcommands MUST accept a positional token argument OR stdin.
4. All commands MUST support `--json` output.

### 5.2. Subcommands

1. `nightwatch oidc pkce` MUST generate a PKCE verifier and S256 challenge.
2. `nightwatch oidc state` MUST generate a state value suitable for CSRF protection.
3. `nightwatch oidc nonce` MUST generate a nonce value suitable for OIDC ID token binding.
4. `nightwatch oidc auth-url` MUST construct an authorization URL from flags.
5. `nightwatch oidc callback <url>` MUST parse an authorization callback URL and extract relevant parameters.
6. `nightwatch oidc idtoken decode <jwt>` MUST decode and summarize OIDC-relevant claims.
7. `nightwatch oidc idtoken lint <jwt>` MUST apply opinionated OIDC checks and return non-zero when checks fail.

### 5.3. Core flags

For `auth-url`:
- `--auth-endpoint <url>` (REQUIRED)
- `--client-id <string>` (REQUIRED)
- `--redirect-uri <url>` (REQUIRED)
- `--scope <string>` (REQUIRED)
- `--state <string>` (RECOMMENDED)
- `--nonce <string>` (RECOMMENDED)
- `--pkce-verifier <string>` (OPTIONAL)

For `idtoken lint`:
- `--issuer <url>` (OPTIONAL; if provided, MUST match `iss`)
- `--audience <string>` (OPTIONAL; if provided, MUST be present in `aud`)
- `--clock-skew <duration>` (OPTIONAL; default `0s`)

## 6. Interface

```bash
# PKCE
nightwatch oidc pkce

# state / nonce
nightwatch oidc state
nightwatch oidc nonce

# Authorization URL
nightwatch oidc auth-url \
  --auth-endpoint "https://issuer.example.com/oauth2/v2.0/authorize" \
  --client-id "my-client" \
  --redirect-uri "http://localhost:3000/callback" \
  --scope "openid profile email" \
  --state "..." \
  --nonce "..."

# Callback parsing
nightwatch oidc callback "http://localhost:3000/callback?code=...&state=..."

# ID token operations
nightwatch oidc idtoken decode <jwt>
nightwatch oidc idtoken lint <jwt> --issuer https://issuer.example.com --audience my-client
```

## 7. Behavior

1. `pkce` MUST output both verifier and challenge.
2. `state` MUST output a single URL-safe string.
3. `nonce` MUST output a single URL-safe string.
4. `state` and `nonce` MUST use a cryptographically secure randomness source.
5. `auth-url` MUST URL-encode query parameters and MUST NOT reorder or drop user-provided values.
6. `callback` MUST report the presence of `error` and `error_description` parameters.
7. `idtoken decode` MUST NOT verify signatures.
8. `idtoken lint` SHOULD check:
   - presence and basic types of `iss`, `aud`, `sub`, `exp`, `iat`
   - `exp` is not in the past (with clock skew)
   - `iat` is not unreasonably in the future (with clock skew)

## 8. Error Handling

1. Invalid URLs MUST return a non-zero exit code and a clear message.
2. Invalid JWT formatting MUST return a non-zero exit code.
3. Missing required flags MUST return a non-zero exit code.

## 9. Examples

```bash
# Generate PKCE values
nightwatch oidc pkce --json

```

## 10. Security Considerations

1. The implementation MUST NOT make network requests.
2. The implementation SHOULD warn users that passing tokens via CLI args may leak through shell history/process listings.
3. `idtoken decode` MUST clearly indicate that it does not verify signatures.

## 11. Testing Considerations

- PKCE verifier and challenge match RFC 7636 S256 requirements.
- `state` and `nonce` output are URL-safe and non-empty.
- Authorization URL construction encodes and includes required parameters.
- Callback parsing extracts parameters correctly.
- ID token lint checks behave as expected for tokens with missing/invalid claims.

## 12. References

### Normative References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997, https://www.rfc-editor.org/rfc/rfc2119

[RFC7636] Sakimura, N., et al., "Proof Key for Code Exchange by OAuth Public Clients", September 2015, https://www.rfc-editor.org/rfc/rfc7636

### Informative References

[OIDC] OpenID Foundation, "OpenID Connect Core 1.0", https://openid.net/specs/openid-connect-core-1_0.html
