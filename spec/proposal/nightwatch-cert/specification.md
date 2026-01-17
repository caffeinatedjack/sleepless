# nightwatch-cert

**Depends on**: none
**Affected files**: internal/nightwatch/cert.go, pkg/cert/*

## Abstract

This specification adds a `nightwatch cert` command group for inspecting X.509 certificates and private keys from local PEM files. It provides quick answers to common development and operations questions: what is in this certificate, when does it expire, what is its fingerprint, and does it match a hostname. The tool is offline and safe-by-default: it MUST NOT print private key material.

## 1. Introduction

Certificate-related failures are a recurring source of lost time in development organizations: TLS handshakes fail in staging but not production, a rotated certificate lacks required SANs, or an expiring certificate breaks a deployment. Debugging these issues often requires a mix of language-specific tooling (`keytool` for Java), OpenSSL invocations, and ad-hoc scripts.

Nightwatch is positioned as an offline, scriptable toolbox for developers and operators. Adding certificate inspection utilities fits that scope and complements existing security-oriented commands.

This proposal introduces `nightwatch cert` as a small, local-only command group that operates on PEM files and produces human-readable output with optional JSON.

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`PEM`: Base64-encoded DER blocks wrapped in `-----BEGIN ...-----` / `-----END ...-----` markers.

`Certificate`: An X.509 certificate, typically encoded as PEM.

`Chain`: Multiple certificates in a single PEM file (leaf + intermediates).

`SAN`: Subject Alternative Name extension, containing hostnames/IPs the certificate is valid for.

`Fingerprint`: A digest (typically SHA-256) of a certificate DER encoding.

## 4. Concepts

### 4.1. Inputs

The command group operates on local files. A single input file MAY contain:

- one certificate
- a chain (multiple certificates)
- a private key (RSA/ECDSA/Ed25519)

### 4.2. Hostname Verification

Hostname verification MUST follow standard X.509 rules (SAN is authoritative; CN is fallback when appropriate) as implemented by the platform’s X.509 hostname verification.

## 5. Requirements

### 5.1. Command Group

1. The CLI MUST provide a top-level command group: `nightwatch cert`.
2. Commands MUST be offline and MUST NOT make network requests.
3. Commands MUST return exit code 0 on success and non-zero on failure.
4. Commands MUST NOT print private key material.

### 5.2. Subcommands

1. `nightwatch cert inspect <file>` MUST parse the input PEM file and print a summary for each contained certificate and/or key.
2. `nightwatch cert expires <file>` MUST print the earliest certificate expiration in the file and support a threshold.
3. `nightwatch cert fingerprint <file>` MUST print the SHA-256 fingerprint for each certificate found.
4. `nightwatch cert verify-host <file> <hostname>` MUST check whether at least one certificate in the file is valid for the hostname.

### 5.3. Flags

- `--json` (boolean): output machine-readable JSON

For `expires`:
- `--within <duration>` (optional): exit non-zero if expiration is within the provided duration (e.g., `30d`, `72h`)

## 6. Interface

```bash
nightwatch cert inspect server.pem
nightwatch cert expires server.pem --within 30d
nightwatch cert fingerprint server.pem
nightwatch cert verify-host server.pem api.internal.example.com
```

## 7. Behavior

1. `inspect` MUST display for each certificate:
   - Subject and Issuer
   - NotBefore / NotAfter
   - SAN DNS names (and IPs if present)
   - Key algorithm and key size where applicable
2. `inspect` SHOULD warn on weak keys (e.g., RSA < 2048 bits).
3. `expires` MUST consider all certificates in the file and use the earliest `NotAfter`.
4. `verify-host` MUST use standard hostname verification rules.

## 8. Error Handling

1. If the file cannot be read, the command MUST fail with a clear error.
2. If the file contains no supported PEM blocks, the command MUST fail.
3. If a PEM block cannot be parsed as a certificate or key, the command MUST fail.

## 9. Examples

```bash
# Quick inspection
nightwatch cert inspect ./certs/tls.pem

# Fail the build if expiring in the next 14 days
nightwatch cert expires ./certs/tls.pem --within 14d

# Verify certificate contains the expected hostname
nightwatch cert verify-host ./certs/tls.pem service.example.com
```

## 10. Security Considerations

1. The tool MUST NOT print private key material.
2. The tool MUST NOT transmit certificate contents over the network.
3. The tool SHOULD avoid printing overly verbose extensions by default to reduce accidental sharing of internal details.

## 11. Testing Considerations

- Parse single-cert PEM and chain PEM.
- Parse private key PEM and confirm no key material is printed.
- Hostname verification matches expected SAN behavior.
- Expiry threshold handling and exit codes.

## 12. References

### Normative References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997, https://www.rfc-editor.org/rfc/rfc2119

[RFC5280] Cooper, D., et al., "Internet X.509 Public Key Infrastructure Certificate and CRL Profile", May 2008, https://www.rfc-editor.org/rfc/rfc5280

### Informative References

[GoX509] Go standard library, `crypto/x509`, https://pkg.go.dev/crypto/x509
