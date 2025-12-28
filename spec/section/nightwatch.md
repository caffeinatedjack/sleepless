# nightwatch specification

This document consolidates the completed specifications for the **`nightwatch`** executable.

It merges the prior per-feature standalone documents under `spec/section/` that belong to Nightwatch.

---

# jwt (JSON Web Token utilities)

**Depends on**: none

## Abstract

This specification adds a new `nightwatch jwt` command group that provides utilities for decoding, verifying, creating, and inspecting JSON Web Tokens (JWTs). It defines CLI inputs for common JWT operations, supported signing algorithms, output formats (including JSON), and error handling to ensure the feature is testable, secure, and script-friendly.

## 1. Introduction

Nightwatch is a CLI-first utility tool. During development and operations work, users frequently need to inspect, verify, or generate JWTs for authentication testing, debugging, and integration work.

Existing solutions often require online services (which pose security risks for production tokens), browser-based tools (which break scriptability), or ad-hoc code snippets in various languages. A unified CLI approach provides a secure, offline, scriptable solution that integrates with the Nightwatch toolset.

The solution is a `jwt` command group that supports decoding tokens without verification, verifying signatures with secrets or public keys, creating new tokens with custom claims, and checking expiration status.

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`JWT`: A JSON Web Token as defined in RFC 7519, consisting of three base64url-encoded parts separated by dots: header, payload, and signature.

`Header`: The first part of a JWT containing metadata about the token type and signing algorithm.

`Payload`: The second part of a JWT containing claims (statements about an entity and additional data).

`Signature`: The third part of a JWT used to verify the token's authenticity.

`Claim`: A piece of information asserted about a subject, represented as a key-value pair in the payload.

`Standard Claim`: A registered claim defined in RFC 7519 (e.g., `iss`, `sub`, `aud`, `exp`, `nbf`, `iat`, `jti`).

`HMAC Algorithm`: Symmetric signing algorithms (HS256, HS384, HS512) using a shared secret.

`RSA Algorithm`: Asymmetric signing algorithms (RS256, RS384, RS512) using RSA key pairs.

`ECDSA Algorithm`: Asymmetric signing algorithms (ES256, ES384, ES512) using elliptic curve key pairs.

## 4. Concepts

### 4.1. Token Structure

A valid JWT MUST consist of three base64url-encoded segments separated by dots (`.`):
1. Header: Contains `typ` (typically "JWT") and `alg` (signing algorithm)
2. Payload: Contains claims (standard and custom)
3. Signature: Cryptographic signature computed over header and payload

### 4.2. Decoding vs. Verification

`Decode`: Parse and display JWT contents without validating the signature. Useful for inspection but MUST NOT be used for authentication decisions.

`Verify`: Decode the token AND validate the signature using the provided secret or public key. MUST fail if the signature is invalid or if required claims (exp, nbf) indicate the token should not be accepted.

## 5. Requirements

### 5.1. Core Command Group

1. The CLI MUST provide a top-level command `nightwatch jwt`.
2. All `jwt` commands MUST accept the token as either a positional argument OR from stdin.
3. All `jwt` commands MUST return exit code 0 on success and non-zero on failure.

### 5.2. Decoding Commands

1. `nightwatch jwt decode <token>` MUST decode and display both header and payload in human-readable format without verifying the signature.
2. `nightwatch jwt header <token>` MUST decode and display only the header.
3. `nightwatch jwt payload <token>` MUST decode and display only the payload.
4. Decoding commands MUST clearly indicate that verification was NOT performed.

### 5.3. Verification Commands

1. `nightwatch jwt verify <token> --secret <key>` MUST verify the token signature using the provided HMAC secret.
2. `nightwatch jwt verify <token> --key <path>` MUST verify the token signature using the RSA/ECDSA public key from the specified PEM file.
3. Verification MUST fail if the signature does not match.
4. Verification MUST check the `exp` claim if present and MUST fail if the token is expired.
5. Verification MUST check the `nbf` claim if present and MUST fail if the current time is before the "not before" time.
6. Verification MUST succeed if the signature is valid and time-based claims are valid.

### 5.4. Creation Commands

1. `nightwatch jwt create --secret <key>` MUST create a new JWT signed with the provided HMAC secret using HS256 by default.
2. `nightwatch jwt create --key <path>` MUST create a new JWT signed with the RSA/ECDSA private key from the specified PEM file.
3. The create command MUST support `--alg <algorithm>` to specify the signing algorithm.
4. The create command MUST support `--claim <key>=<value>` (repeatable) to add custom claims.
5. The create command MUST support `--payload <json>` to provide claims as a JSON object.
6. The create command MUST support standard claim shortcuts: `--iss`, `--sub`, `--aud`, `--exp`, `--nbf`, `--iat`, `--jti`.
7. The `--exp` flag MUST accept duration strings (e.g., "1h", "30m", "7d") and MUST compute the absolute expiration timestamp.
8. The `--iat` flag SHOULD default to the current timestamp if not provided.

### 5.5. Expiration Utility

1. `nightwatch jwt exp <token>` MUST display the token's issued time (`iat`), expiration time (`exp`), and current status (valid/expired/no expiry).
2. The command MUST calculate and display time remaining until expiration if the token is valid.
3. The command MUST NOT verify the signature (this is an inspection command).

### 5.6. Supported Algorithms

The implementation MUST support the following algorithms:
- HMAC: HS256, HS384, HS512
- RSA: RS256, RS384, RS512
- ECDSA: ES256, ES384, ES512

### 5.7. Output Formats

1. All commands MUST support a `--json` flag for machine-readable JSON output.
2. Human-readable output MUST use clear formatting with labeled sections.
3. The `payload` and `header` commands MUST output raw JSON suitable for piping to tools like `jq`.

## 6. Interface

### 6.1. Commands

```bash
# Decode (no verification)
nightwatch jwt decode <token>
nightwatch jwt payload <token>
nightwatch jwt header <token>

# Verify
nightwatch jwt verify <token> --secret <secret>
nightwatch jwt verify <token> --key <path>

# Create
nightwatch jwt create --secret <secret> [options]
nightwatch jwt create --key <path> [options]

# Expiry check
nightwatch jwt exp <token>
```

### 6.2. Create Command Flags

- `--secret <string>`: HMAC secret key
- `--key <path>`: Path to PEM-encoded private key file
- `--alg <algorithm>`: Signing algorithm (HS256, RS256, ES256, etc.)
- `--claim <key>=<value>`: Add custom claim (repeatable)
- `--payload <json>`: Provide complete payload as JSON string
- `--iss <issuer>`: Set issuer claim
- `--sub <subject>`: Set subject claim
- `--aud <audience>`: Set audience claim
- `--exp <duration>`: Set expiration (e.g., "1h", "30m", "7d")
- `--nbf <duration>`: Set not-before (e.g., "5m")
- `--iat <timestamp>`: Set issued-at time (defaults to now)
- `--jti <id>`: Set JWT ID claim

### 6.3. Verify Command Flags

- `--secret <string>`: HMAC secret key
- `--key <path>`: Path to PEM-encoded public key file

### 6.4. Global Flags

- `--json`: Output in JSON format

### 6.5. JSON Output (normative)

#### Decode/Header/Payload Output

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "user123",
    "exp": 1735689600
  },
  "signature": "base64url-encoded-signature",
  "verified": false,
  "raw": "full.jwt.token"
}
```

For `header` and `payload` commands without `--json`, output MUST be raw JSON of that section only.

#### Verify Output

```json
{
  "valid": true,
  "header": { ... },
  "payload": { ... },
  "verified": true,
  "algorithm": "HS256"
}
```

#### Create Output

```json
{
  "token": "eyJhbGc...",
  "header": { ... },
  "payload": { ... }
}
```

Without `--json`, create MUST output only the token string.

#### Exp Output

```json
{
  "iat": "2025-01-01T10:00:00Z",
  "exp": "2025-01-01T11:00:00Z",
  "now": "2025-01-01T10:30:00Z",
  "status": "valid",
  "remaining": "30m0s"
}
```

## 7. Behavior

1. Token input MUST accept tokens from positional arguments OR from stdin (one token per line).
2. When reading from stdin, the command MUST process each token and output results sequentially.
3. For HMAC algorithms, the secret MAY be provided as a plain string or as a base64-encoded value (auto-detected).
4. For RSA/ECDSA algorithms, keys MUST be in PEM format.
5. The `exp` command MUST handle tokens without `exp` or `iat` claims gracefully and indicate their absence.
6. Time-based claims (`exp`, `nbf`, `iat`) MUST be interpreted as Unix timestamps (seconds since epoch).
7. Duration strings MUST support: `s` (seconds), `m` (minutes), `h` (hours), `d` (days).

## 8. Error Handling

1. Invalid token format (not three base64url parts) MUST produce exit code 1 and an error message indicating format issue.
2. Invalid base64url encoding MUST produce exit code 1 and an error message.
3. Invalid JSON in header or payload MUST produce exit code 1 and an error message.
4. Signature verification failure MUST produce exit code 2 and an error message indicating verification failed.
5. Expired token during verification MUST produce exit code 3 and an error message with expiration details.
6. Token not yet valid (nbf) MUST produce exit code 4 and an error message.
7. Unsupported algorithm MUST produce exit code 1 and list supported algorithms.
8. Invalid or missing key file MUST produce exit code 1 and an error message.
9. Key type mismatch (e.g., using HMAC secret for RSA) MUST produce exit code 1 and an error message.

## 9. Examples

### Decode

```bash
# Decode and display full token
nightwatch jwt decode eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Extract just payload (pipe to jq)
nightwatch jwt payload eyJhbGc... | jq .sub

# Decode from stdin
echo "eyJhbGc..." | nightwatch jwt decode
```

### Verify

```bash
# Verify with HMAC secret
nightwatch jwt verify eyJhbGc... --secret "my-secret-key"

# Verify with RSA public key
nightwatch jwt verify eyJhbGc... --key public.pem

# JSON output
nightwatch jwt verify eyJhbGc... --secret "key" --json
```

### Create

```bash
# Create with single claim
nightwatch jwt create --secret "key" --claim sub=user123 --exp 1h

# Create with multiple claims
nightwatch jwt create --secret "key" --claim role=admin --claim tenant=acme --exp 24h

# Create with full payload
nightwatch jwt create --secret "key" --payload '{"sub":"123","role":"admin"}'

# Create with RSA private key
nightwatch jwt create --key private.pem --alg RS256 --sub user123 --exp 1h

# Create with standard claims
nightwatch jwt create --secret "key" --iss "myapp" --sub "user123" --aud "api" --exp 1h
```

### Expiration Check

```bash
# Check expiration status
nightwatch jwt exp eyJhbGc...

# JSON output
nightwatch jwt exp eyJhbGc... --json
```

## 10. Security Considerations

1. The implementation MUST NOT transmit tokens over the network.
2. Secrets and private keys MUST NOT be logged or included in error messages.
3. The implementation SHOULD warn if tokens or secrets are provided via command-line arguments (which may be visible in process lists) and SHOULD recommend stdin or environment variables.
4. The `decode` command MUST clearly indicate that the token was NOT verified, to prevent misuse for authentication decisions.
5. Generated tokens SHOULD use cryptographically secure random sources for any random claims (e.g., `jti`).
6. The implementation MUST validate key file permissions and SHOULD warn if private key files are world-readable.
7. Timing attacks on signature verification SHOULD be mitigated by using constant-time comparison where the underlying crypto library supports it.

## 11. Privacy Considerations

1. JWT payloads are NOT encrypted and contain readable claims. The implementation SHOULD NOT encourage storing sensitive information in JWT payloads.
2. When outputting tokens or claims, the implementation SHOULD provide a `--redact` option for demonstration/teaching purposes that masks claim values.

## 12. Testing Considerations

Test scenarios SHOULD include:

- Decoding valid tokens for all supported algorithms
- Verifying tokens with correct and incorrect secrets/keys
- Verifying expired and not-yet-valid tokens
- Creating tokens and verifying them round-trip
- Handling malformed tokens (invalid base64, wrong number of segments, invalid JSON)
- Testing all duration formats for `--exp`
- Parsing tokens from stdin (single and multiple tokens)
- JSON output schema validation for all commands
- Exit code verification for all error conditions
- Key file permission warnings

## 13. References

[RFC7519] Jones, M., Bradley, J., Sakimura, N., "JSON Web Token (JWT)", May 2015, https://www.rfc-editor.org/rfc/rfc7519

[RFC7515] Jones, M., Bradley, J., Sakimura, N., "JSON Web Signature (JWS)", May 2015, https://www.rfc-editor.org/rfc/rfc7515

[JWT.io] Auth0, "JWT.io - JSON Web Token Debugger", https://jwt.io

---

# redact (PII and secret redaction)

**Depends on**: none

## Abstract

This specification adds a `redact` command group that removes personally identifiable information (PII) and common secrets (tokens/keys) from text, stdin streams, files, and directories. The command group MUST be exposed via a sleep-themed security executable named `nightwatch` (`nightwatch redact ...`).

It defines default and extended pattern sets, redaction modes (replace, mask, hash), audit-only checking, and safe file modification behavior.

## 1. Introduction

Users frequently copy logs, diffs, and text snippets into tickets, chat, or documentation. These artifacts often include PII (emails, phone numbers, IP addresses) or secrets (JWTs, bearer tokens, private keys). Accidentally sharing such data creates security and privacy risks.

Manual redaction is error-prone and inconsistent. A CLI tool that can be used in pipelines, on files, or across directories reduces risk and standardizes output.

The proposed `redact` command group provides streaming-friendly redaction for stdin and files, plus an audit mode that reports findings without modifying content.

To support security-oriented usage (e.g., installing only security tools into restricted environments, or providing a clearly named tool for redaction and future security utilities), `redact` is exposed through a separate executable, `nightwatch`.

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`Finding`: A detected match of a redaction pattern within input.

`Pattern Type`: A named category such as `EMAIL` or `IP`.

`Default Patterns`: The set of Pattern Types enabled when `--extended` is not provided.

`Extended Patterns`: Additional Pattern Types enabled by `--extended`.

`Replacement Token`: The string inserted into output for a match (e.g., `[EMAIL]`).

`Mask Mode`: Redaction that preserves partial structure (e.g., `j***@***.com`).

`Hash Mode`: Redaction that replaces with a stable hash-derived token (e.g., `[EMAIL:a1b2c3]`).

`Audit Mode`: Detection-only mode that outputs findings without modifying content.

## 4. Concepts

### 4.1. Pattern Sets

The system maintains two pattern sets:

- Default pattern types: `EMAIL`, `PHONE`, `IP` (v4/v6), `CREDIT_CARD`, `UUID`
- Extended pattern types (enabled with `--extended`): `PRIVATE_KEY`, `JWT`, `BEARER_TOKEN`

### 4.2. Pattern Selection

The active pattern types are:

1. Start with Default pattern types.
2. If `--extended` is set, add Extended pattern types.
3. If `--only` is set, restrict active types to that list.
4. If `--except` is set, remove those types.

If `--only` results in an empty set, the command MUST error.

### 4.3. Processing Modes

The system supports:

- Redact mode (default): replace/mask/hash matches and emit modified output.
- Audit mode (`check`): detect matches and report them.

## 5. Requirements

### 5.1. Executables

1. The project MUST provide an executable named `regimen`.
2. The project MUST provide an additional executable named `nightwatch`.
3. `nightwatch` MUST be positioned as the security-focused companion executable and MAY grow additional security-related command groups in future specifications.

### 5.2. Command Group

1. The CLI MUST provide a top-level command `redact` under `nightwatch` (i.e., `nightwatch redact`).
2. The `regimen` executable MUST NOT expose a `redact` command.

### 5.3. Inputs

The following forms MUST be supported under `nightwatch redact`:

1. `redact <string>` MUST redact matches in the provided string and output the redacted string.
2. `redact stdin` MUST read from stdin, redact, and write to stdout.
3. `redact file <path>` MUST read the file at `<path>`, redact, and write to stdout.
4. `redact dir <path>` MUST process files under `<path>`.

### 5.4. Directory Mode

The following forms MUST be supported under `nightwatch redact`:

1. `redact dir <path> --output <dir>` MUST write redacted copies to `<dir>` while preserving relative paths.
2. `redact dir <path> --pattern <glob>` MUST restrict processed files to those matching `<glob>`. 

### 5.5. In-place File Modification

The following behavior MUST apply under `nightwatch redact`:

1. When `--in-place` is provided with `redact file <path>`, the system MUST modify the target file.
2. Before modification, the system MUST create a backup with a `.bak` suffix.
3. If writing the modified file fails, the system MUST restore the original from the backup.

### 5.6. Audit Mode

The following forms MUST be supported under `nightwatch redact`:

1. `redact check file <path>` MUST scan and report findings without modifying content.
2. `redact check dir <path>` MUST scan and report findings for all processed files.

### 5.7. Flags

- `--only <types>`: comma-separated pattern types
- `--except <types>`: comma-separated pattern types
- `--extended` (boolean)
- `--mask` (boolean)
- `--hash` (boolean)
- `--in-place` (boolean)
- `--custom <regex>`: add a custom regex pattern
- `--custom-name <name>`: replacement name for custom pattern

1. `--mask` and `--hash` MUST be mutually exclusive.
2. `--custom-name` MUST only be valid if `--custom` is provided.

## 6. Interface

### 6.1. Commands

All forms are available under `nightwatch`:

- `nightwatch redact ...`

```bash
# Redact a string
nightwatch redact "Contact john@example.com or 555-123-4567"

# Redact stdin
kubectl logs pod | nightwatch redact stdin

git diff | nightwatch redact stdin --extended

# Redact file
nightwatch redact file server.log > clean.log
nightwatch redact file server.log --in-place

# Redact directory
nightwatch redact dir ./logs --output ./logs-clean
nightwatch redact dir ./logs --pattern "*.log"

# Audit mode
nightwatch redact check file.log
nightwatch redact check dir ./src --extended
```

### 6.2. Output

- Redaction commands output the redacted content to stdout.
- `check` outputs a report to stdout.

If a JSON output mode is added in the future, it MUST be introduced with a new explicit flag to avoid breaking pipeline usage.

## 7. Behavior

1. Replacement mode MUST replace each match with `[<TYPE>]` (e.g., `[EMAIL]`) unless `--mask` or `--hash` is set.
2. In `--hash` mode, replacement MUST be `[<TYPE>:<digest>]` where `<digest>` is a short stable identifier derived from the matched value.
3. In `--mask` mode, the system MUST partially obscure values while retaining enough structure to recognize the type.
4. In directory mode, the system MUST process files as UTF-8 text. Files that are not decodable as UTF-8 SHOULD be skipped with a warning.
5. The system MUST support piping large inputs via `stdin` without unbounded memory growth.

## 8. Error Handling

1. Invalid regex values for `--custom` MUST error with a non-zero exit code.
2. If a file or directory cannot be read, the command MUST return a non-zero exit code.
3. `--in-place` used outside `file` mode MUST error.
4. Unknown pattern type names in `--only`/`--except` MUST error.

## 9. Examples

```bash
nightwatch redact "Contact john@example.com or 555-123-4567"
# => Contact [EMAIL] or [PHONE]

kubectl logs pod | nightwatch redact stdin

nightwatch redact file server.log --in-place

nightwatch redact check dir ./src --extended
```

## 10. Security Considerations

1. The redaction tool MUST avoid emitting unredacted matches in error messages.
2. In `--in-place` mode, backup files MUST be created with permissions that do not broaden access relative to the original.
3. The tool MUST NOT transmit data over the network.

## 11. Testing Considerations

- Pattern detection tests per Pattern Type.
- Mask/hash mode tests.
- In-place backup/restore tests including simulated write failures.
- Directory traversal tests with `--pattern` filtering.
- Large input streaming tests.

## 12. References

[RE2] Google, "RE2 regular expression syntax", https://github.com/google/re2/wiki/Syntax

---

# Generate passwords

**Depends on**: none

## Abstract

This specification adds a `nightwatch password` command group for generating cryptographically secure passwords, passphrases, and secrets suitable for authentication, API keys, and development. It defines CLI modes for password generation with configurable character sets, passphrase generation with word lists, API key generation, and custom pattern-based generation with reproducible seeding and machine-readable JSON output.

## 1. Introduction

Developers, security engineers, and operations teams frequently need to generate strong passwords for testing, API keys for services, passphrases for backup encryption, or temporary credentials for development environments.

Common approaches include reusing weak passwords, using online password generators (which pose security risks), manually rolling dice for diceware passphrases (slow and error-prone), or installing dedicated password manager CLIs (heavy dependencies). A lightweight, offline, CLI-first password generator integrated into Nightwatch provides a secure, scriptable solution for generating production-grade secrets without external dependencies or network access.

The proposed `password` command group provides character-based password generation with configurable complexity, diceware-style passphrase generation, API key format generation, and pattern-based custom secret generation. All generation uses cryptographically secure randomness suitable for production use, with optional seeding for reproducible test data generation.

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`Character-Based Password`: A password composed of individual characters drawn from one or more character sets (uppercase, lowercase, digits, symbols).

`Passphrase`: A password composed of words separated by a delimiter, typically drawn from a word list (e.g., diceware).

`Character Set`: A collection of characters available for password generation (e.g., lowercase letters `a-z`).

`Entropy`: A measure of password randomness and unpredictability, measured in bits.

`Diceware`: A method for creating passphrases by randomly selecting words from a curated word list.

`API Key`: A random string typically used for authentication, often with a specific format (e.g., prefix + random characters).

`Seeded Mode`: Password generation performed using a user-provided seed such that output is deterministic for the same inputs (for testing only).

`Unseeded Mode`: Password generation performed using cryptographically secure randomness (for production use).

`Symbol`: A special character such as `!@#$%^&*()-_=+[]{}|;:,.<>?`.

## 4. Concepts

### 4.1. Generation Modes

The system provides four generation modes:

1. **Password Mode** (`password generate`): Character-based passwords with configurable character sets
2. **Passphrase Mode** (`password phrase`): Word-based passphrases using word lists
3. **API Key Mode** (`password api`): API key format with optional prefix/suffix
4. **Pattern Mode** (`password pattern`): Custom pattern-based generation (e.g., `XXX-999-xxx` for license keys)

### 4.2. Character Sets

Password mode supports the following character sets:

- **Lowercase** (`a-z`): 26 characters
- **Uppercase** (`A-Z`): 26 characters
- **Digits** (`0-9`): 10 characters
- **Symbols**: `!@#$%^&*()-_=+[]{}|;:,.<>?/~` (30 characters)
- **Ambiguous characters**: Characters easily confused in certain fonts (`0O`, `1lI`, `5S`, etc.)

By default, passwords include lowercase, uppercase, and digits. Symbols are opt-in. Ambiguous characters are excluded by default.

### 4.3. Passphrase Word Lists

The system provides a built-in word list based on the EFF long word list (7776 words, each word selected by rolling 5 dice = ~12.9 bits of entropy per word).

Custom word lists MAY be provided via `--wordlist <path>`.

### 4.4. Entropy Calculation

The system MUST calculate and optionally display the entropy of generated passwords:

- Character-based: `log2(charset_size^length)` bits
- Passphrase: `log2(wordlist_size^word_count)` bits

### 4.5. Randomness Quality

All generation MUST use Go's `crypto/rand` package for cryptographically secure randomness, EXCEPT when `--seed` is explicitly provided for reproducible test data generation.

## 5. Requirements

### 5.1. Command Group

1. The CLI MUST provide a top-level command `nightwatch password`.
2. All `password` commands MUST return exit code 0 on success and non-zero on failure.
3. The `regimen` executable MUST NOT expose a `password` command (keeps password generation in the security tooling namespace).

### 5.2. Password Generation

1. `nightwatch password generate` MUST generate a character-based password.
2. `nightwatch password generate --length N` MUST generate a password of exactly N characters.
3. The default length MUST be 16 characters if `--length` is not specified.
4. `nightwatch password generate --symbols` MUST include symbol characters in the character set.
5. `nightwatch password generate --no-uppercase` MUST exclude uppercase letters.
6. `nightwatch password generate --no-lowercase` MUST exclude lowercase letters.
7. `nightwatch password generate --no-digits` MUST exclude digits.
8. `nightwatch password generate --allow-ambiguous` MUST include ambiguous characters (0O, 1lI, etc.).
9. If all character sets are excluded, the command MUST error.
10. Generated passwords MUST include at least one character from each enabled character set (unless `--no-require-all` is specified).

### 5.3. Passphrase Generation

1. `nightwatch password phrase` MUST generate a passphrase using the built-in word list.
2. `nightwatch password phrase --words N` MUST generate a passphrase with exactly N words.
3. The default word count MUST be 6 if `--words` is not specified.
4. `nightwatch password phrase --separator <string>` MUST join words with the specified separator.
5. The default separator MUST be `-` (hyphen).
6. `nightwatch password phrase --capitalize` MUST capitalize the first letter of each word.
7. `nightwatch password phrase --wordlist <path>` MUST use a custom word list from the specified file (one word per line).
8. `nightwatch password phrase --numbers` MUST append a random digit to each word.

### 5.4. API Key Generation

1. `nightwatch password api` MUST generate an API key in the format `prefix_randomchars`.
2. The default format MUST be 32 lowercase hexadecimal characters.
3. `nightwatch password api --prefix <string>` MUST prepend the specified prefix followed by an underscore.
4. `nightwatch password api --length N` MUST generate N random characters (after prefix).
5. `nightwatch password api --format <format>` MUST generate keys in the specified format.
6. Supported formats MUST include: `hex` (0-9a-f), `base32` (A-Z2-7), `base64` (A-Za-z0-9+/), `base64url` (A-Za-z0-9-_).

### 5.5. Pattern-Based Generation

1. `nightwatch password pattern <pattern>` MUST generate a password matching the specified pattern.
2. Pattern syntax MUST support:
   - `X` = random uppercase letter
   - `x` = random lowercase letter
   - `9` = random digit
   - `*` = random character (any printable ASCII)
   - `#` = random alphanumeric (a-zA-Z0-9)
   - `?` = random symbol
   - `\X`, `\x`, `\9`, `\*`, `\#`, `\?` = literal character (escape)
   - Any other character = literal character
3. Example: `XXX-999-xxx` generates patterns like `ABC-123-def`.

### 5.6. Batch Generation

1. All generation commands MUST support `--count N` to generate N passwords/passphrases.
2. When `--count` is used, each password/passphrase MUST be independently generated.
3. Multi-password output MUST print one password per line.

### 5.7. Entropy Display

1. All generation commands MUST support `--show-entropy` to display the entropy of generated passwords.
2. Entropy MUST be calculated and displayed in bits.
3. When `--count` is used with `--show-entropy`, entropy MUST be calculated once (as all passwords have the same entropy).

### 5.8. Quality Checks

1. `nightwatch password check <password>` MUST analyze a provided password and report its strength.
2. The check command MUST report:
   - Length
   - Character set composition (lowercase, uppercase, digits, symbols)
   - Estimated entropy
   - Strength rating (weak/fair/good/strong/excellent)
3. The strength rating MUST be based on entropy:
   - Weak: < 40 bits
   - Fair: 40-59 bits
   - Good: 60-79 bits
   - Strong: 80-99 bits
   - Excellent: ≥ 100 bits

### 5.9. Flags

#### Global Flags
- `--count <n>`: Generate n passwords (default: 1)
- `--seed <int>`: Enable Seeded Mode for reproducible testing (MUST NOT be used for production passwords)
- `--json`: Output in JSON format
- `--show-entropy`: Display entropy information

#### Password Mode Flags
- `--length <n>`: Password length (default: 16)
- `--symbols`: Include symbol characters
- `--no-uppercase`: Exclude uppercase letters
- `--no-lowercase`: Exclude lowercase letters
- `--no-digits`: Exclude digits
- `--allow-ambiguous`: Include ambiguous characters (0O, 1lI, etc.)
- `--no-require-all`: Don't require at least one character from each enabled set

#### Passphrase Mode Flags
- `--words <n>`: Number of words (default: 6)
- `--separator <string>`: Word separator (default: `-`)
- `--capitalize`: Capitalize first letter of each word
- `--numbers`: Append digit to each word
- `--wordlist <path>`: Use custom word list

#### API Key Mode Flags
- `--prefix <string>`: Key prefix (followed by `_`)
- `--length <n>`: Random character count (default: 32)
- `--format <format>`: Output format (hex/base32/base64/base64url) (default: hex)

### 5.10. JSON Output

When `--json` is provided, stdout MUST contain a single JSON object with:

- `mode`: generation mode (password/phrase/api/pattern)
- `seed`: integer or null
- `count`: number of passwords generated
- `entropy`: entropy in bits (if applicable)
- `results`: array of generated passwords

For `check` mode:
- `password_length`: integer
- `character_sets`: array of strings (lowercase/uppercase/digits/symbols)
- `entropy`: float
- `strength`: string (weak/fair/good/strong/excellent)

## 6. Interface

### 6.1. Commands

```bash
# Password generation
nightwatch password generate
nightwatch password generate --length 20
nightwatch password generate --length 20 --symbols
nightwatch password generate --count 5
nightwatch password generate --show-entropy

# Passphrase generation
nightwatch password phrase
nightwatch password phrase --words 8
nightwatch password phrase --separator " "
nightwatch password phrase --capitalize
nightwatch password phrase --wordlist custom.txt

# API key generation
nightwatch password api
nightwatch password api --prefix "sk_prod"
nightwatch password api --length 40 --format base64url

# Pattern-based generation
nightwatch password pattern "XXX-999-xxx"
nightwatch password pattern "####-####-####"

# Strength checking
nightwatch password check "MyP@ssw0rd123"

# JSON output
nightwatch password generate --json
nightwatch password phrase --count 3 --json
```

### 6.2. JSON Output Schema

#### Generate Output

```json
{
  "mode": "password",
  "seed": null,
  "count": 1,
  "entropy": 77.54,
  "results": [
    "aB3$kL9mP2qR5tY7"
  ]
}
```

#### Passphrase Output

```json
{
  "mode": "phrase",
  "seed": null,
  "count": 2,
  "entropy": 77.5,
  "results": [
    "correct-horse-battery-staple-robot-wizard",
    "mountain-river-sunset-keyboard-dolphin-puzzle"
  ]
}
```

#### Check Output

```json
{
  "password_length": 13,
  "character_sets": ["lowercase", "uppercase", "digits", "symbols"],
  "entropy": 63.2,
  "strength": "good"
}
```

## 7. Behavior

### 7.1. Randomness

1. In Unseeded Mode, the system MUST use `crypto/rand` for all random generation.
2. In Seeded Mode (`--seed` provided), the system MUST use a deterministic PRNG seeded with the provided value.
3. Seeded Mode MUST produce identical output for identical inputs.
4. The CLI MUST emit a warning when `--seed` is used, stating "WARNING: Seeded passwords are for testing only. DO NOT use for production."

### 7.2. Character Set Guarantee

1. By default, passwords MUST include at least one character from each enabled character set.
2. This prevents passwords like "aaaaaaaaaaaaaaaa" when all sets are enabled.
3. If `--no-require-all` is specified, this guarantee is removed.
4. The guarantee MUST be implemented by:
   - Generating random passwords until one satisfies the requirement, OR
   - Explicitly placing one character from each set, then filling the rest randomly

### 7.3. Passphrase Word Selection

1. Words MUST be selected uniformly from the word list.
2. The same word MAY appear multiple times in a single passphrase.
3. Custom word lists MUST be loaded from files with one word per line.
4. Empty lines and lines starting with `#` in word lists MUST be ignored.
5. Word lists with fewer than 100 words SHOULD produce a warning about low entropy.

### 7.4. Pattern Processing

1. Pattern strings MUST be processed left-to-right.
2. Each pattern character MUST be independently generated.
3. Invalid escape sequences (e.g., `\z`) MUST produce an error.

### 7.5. Entropy Calculation

1. For character-based passwords: `entropy = log2(charset_size) × length`
2. For passphrases: `entropy = log2(wordlist_size) × word_count`
3. For API keys: `entropy = log2(format_charset_size) × length`
4. For patterns: `entropy = sum(log2(charset_size_per_position))`

## 8. Error Handling

1. Invalid length values (< 1) MUST produce exit code 1 and an error message.
2. Invalid count values (< 1) MUST produce exit code 1 and an error message.
3. Invalid word count values (< 1) MUST produce exit code 1 and an error message.
4. Empty character set (all sets excluded) MUST produce exit code 1 and an error message listing available sets.
5. Missing or unreadable word list file MUST produce exit code 1 and an error message.
6. Invalid format in `--format` MUST produce exit code 1 and list valid formats.
7. Invalid pattern syntax MUST produce exit code 1 and indicate the error position.
8. Cryptographic randomness failure MUST produce exit code 2 and an error message (critical failure).

## 9. Examples

(Examples retained from the original standalone specification.)

## 10. Security Considerations

1. The implementation MUST use `crypto/rand` for all production password generation.
2. Generated passwords MUST NOT be logged, cached, or written to disk unless explicitly requested by the user.
3. The `--seed` flag MUST emit a prominent warning that seeded passwords are for testing only.
4. The implementation SHOULD warn if passwords are generated with low entropy (< 40 bits).
5. The implementation MUST clear sensitive data from memory after generation where practical.
6. Generated passwords SHOULD NOT be passed as command-line arguments in documentation examples (to avoid shell history exposure).
7. When checking password strength, the implementation MUST NOT transmit the password over the network.
8. The entropy calculation SHOULD assume an attacker knows the generation method (charset, length, wordlist) per Kerckhoffs's principle.
9. The implementation SHOULD detect and warn about common password patterns (e.g., "password", "123456", keyboard walks).
10. For API keys, the documentation MUST emphasize that prefix/format choices should match the target service's requirements.

## 11. Privacy Considerations

1. Generated passwords are not inherently associated with any user or service.
2. The `check` command operates entirely offline and reveals no information about where passwords are used.
3. The implementation MUST NOT phone home or send telemetry about password generation.

## 12. Testing Considerations

(Testing considerations retained from the original standalone specification.)

## 13. References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997, https://www.rfc-editor.org/rfc/rfc2119

---

# fake (fake data generation)

**Depends on**: none

## Abstract

This specification adds a `nightwatch fake` command group for generating fake/placeholder data for development, testing, and demonstration purposes. It defines CLI modes for generating names, emails, addresses, phone numbers, dates, numbers, UUIDs, lorem ipsum text, and custom templated data. It also defines reproducibility controls (`--seed`), quantity controls (`--count`), and machine-readable output (`--json`).

## 1. Introduction

Developers, testers, and content creators frequently need realistic-looking placeholder data: mock user profiles, sample database records, realistic text content for UI prototyping, or test datasets for API development.

While many online services and libraries exist, they often require network access, lack CLI integration, introduce heavyweight dependencies, or don't support reproducible generation for testing scenarios. Adding a lightweight, CLI-first fake data generator to Nightwatch provides a scriptable, offline, reproducible solution that integrates with the existing security-focused toolset.

The proposed interface centers around `nightwatch fake` with type-specific subcommands, template support for complex data structures, and seeded randomness for deterministic test data generation.

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`Data Type`: A category of generated data (e.g., `name`, `email`, `lorem`).

`Generator`: A function that produces fake data for a specific Data Type.

`Template`: A string containing placeholder tokens that are replaced with generated data.

`Seeded Mode`: Data generation performed using a user-provided seed such that output is deterministic for the same inputs.

`Unseeded Mode`: Data generation performed without a user-provided seed, using high-quality randomness.

`Lorem Ipsum`: Placeholder text derived from scrambled Latin, used as filler text in publishing and design.

## 4. Concepts

### 4.1. Data Types

The system provides built-in generators for common data types grouped by category:

**Personal Data**:
- `name`: Full name (first + last)
- `firstname`: First name only
- `lastname`: Last name only
- `email`: Email address
- `username`: Username/handle
- `phone`: Phone number

**Location Data**:
- `address`: Street address
- `city`: City name
- `state`: State/province
- `country`: Country name
- `zipcode`: Postal code

**Temporal Data**:
- `date`: Date in YYYY-MM-DD format
- `datetime`: RFC3339 timestamp
- `time`: Time in HH:MM:SS format

**Identifiers**:
- `uuid`: UUID v4
- `hex`: Hexadecimal string
- `number`: Random integer

**Text Content**:
- `lorem`: Lorem ipsum text (words, sentences, or paragraphs)
- `word`: Single word
- `sentence`: Single sentence
- `paragraph`: Single paragraph

**Internet**:
- `url`: Web URL
- `ipv4`: IPv4 address
- `ipv6`: IPv6 address
- `mac`: MAC address

### 4.2. Template Mode

Templates allow composition of multiple data types into structured output. Template strings use `{{type}}` or `{{type:arg}}` placeholders that are replaced with generated data.

### 4.3. Quantity and Repetition

Most commands support `--count N` to generate multiple items. In template mode, `--count` produces multiple records.

## 5. Requirements

### 5.1. Command Group

1. The CLI MUST provide a top-level command `nightwatch fake`.
2. The `regimen` executable MUST NOT expose a `fake` command (keeps fake data generation in the security tooling namespace).

### 5.2. Data Type Commands

1. Each Data Type listed in section 4.1 MUST have a corresponding subcommand: `nightwatch fake <type>`.
2. All data type commands MUST accept `--count N` to generate N items.
3. All data type commands MUST accept `--seed <int>` for deterministic output.
4. All data type commands MUST accept `--json` for machine-readable output.

### 5.3. Lorem Ipsum

1. `nightwatch fake lorem` MUST generate lorem ipsum text.
2. `nightwatch fake lorem --words N` MUST generate exactly N words.
3. `nightwatch fake lorem --sentences N` MUST generate exactly N sentences.
4. `nightwatch fake lorem --paragraphs N` MUST generate exactly N paragraphs.
5. If no quantity flag is provided, the default MUST be 1 paragraph.
6. Generated text MUST begin with "Lorem ipsum" when generating full paragraphs.
7. Generated text MUST use the standard lorem ipsum vocabulary.

### 5.4. Number Generation

1. `nightwatch fake number` MUST generate a random integer.
2. `nightwatch fake number --min N --max M` MUST generate an integer in the inclusive range [N, M].
3. If only `--max` is provided, `--min` MUST default to 0.
4. If neither flag is provided, the range MUST be [0, 100].

### 5.5. Date Generation

1. `nightwatch fake date` MUST generate a date in YYYY-MM-DD format.
2. `nightwatch fake date --past N` MUST generate a date up to N days in the past.
3. `nightwatch fake date --future N` MUST generate a date up to N days in the future.
4. If neither `--past` nor `--future` is provided, the date MUST be randomly chosen from the past 365 days.

### 5.6. Template Mode

1. `nightwatch fake template <string>` MUST replace placeholder tokens with generated data.
2. Placeholders MUST use the format `{{type}}` or `{{type:arg}}`.
3. Supported types MUST match the Data Type subcommands.
4. Unknown placeholders MUST produce an error.
5. Template mode MUST support `--count N` to generate N records.

### 5.7. Name Data Source

1. The system MUST use the existing name data files in `internal/nightwatch/data/` for name generation.
2. First names MUST be read from `firstnames.txt`.
3. Last names MUST be read from `surnames.txt`.

### 5.8. Flags

- `--count <n>`: Generate n items (default: 1)
- `--seed <int>`: Enable Seeded Mode for reproducibility
- `--json`: Output in JSON format
- `--separator <string>`: Line separator for multi-item output (default: newline)

**Lorem-specific**:
- `--words <n>`: Generate n words
- `--sentences <n>`: Generate n sentences
- `--paragraphs <n>`: Generate n paragraphs

**Number-specific**:
- `--min <n>`: Minimum value
- `--max <n>`: Maximum value

**Date-specific**:
- `--past <n>`: Generate date up to n days in the past
- `--future <n>`: Generate date up to n days in the future

### 5.9. JSON Output

When `--json` is provided, stdout MUST contain a single JSON object with:

- `type`: the data type generated
- `seed`: integer or null
- `count`: number of items
- `results`: array of generated values

For template mode:
- `type`: "template"
- `template`: the template string
- `results`: array of completed template strings

## 6. Interface

### 6.1. Commands

```bash
# Basic data types
nightwatch fake name
nightwatch fake email
nightwatch fake phone
nightwatch fake address
nightwatch fake city
nightwatch fake state
nightwatch fake country
nightwatch fake zipcode
nightwatch fake date
nightwatch fake datetime
nightwatch fake time
nightwatch fake uuid
nightwatch fake hex
nightwatch fake number
nightwatch fake url
nightwatch fake ipv4
nightwatch fake ipv6
nightwatch fake username

# Text generation
nightwatch fake lorem
nightwatch fake word
nightwatch fake sentence
nightwatch fake paragraph

# With flags
nightwatch fake name --count 10
nightwatch fake email --count 5 --seed 42
nightwatch fake number --min 1 --max 100 --count 20
nightwatch fake date --past 30 --count 5
nightwatch fake lorem --sentences 5
nightwatch fake lorem --paragraphs 3

# Template mode
nightwatch fake template "{{name}} <{{email}}>"
nightwatch fake template "{{firstname}} {{lastname}} lives in {{city}}, {{state}}"
nightwatch fake template '{"name":"{{name}}","age":{{number:18:65}}}' --count 5

# JSON output
nightwatch fake name --count 10 --json
nightwatch fake template "{{name}},{{email}}" --count 5 --json
```

### 6.2. Template Syntax

Basic placeholder: `{{type}}`
- Example: `{{name}}`, `{{email}}`, `{{uuid}}`

Parameterized placeholder: `{{type:arg1:arg2}}`
- `{{number:min:max}}` - number between min and max
- `{{lorem:words:N}}` - N words of lorem
- `{{lorem:sentences:N}}` - N sentences
- `{{date:past:N}}` - date up to N days ago
- `{{date:future:N}}` - date up to N days ahead

### 6.3. JSON Output Schema

```json
{
  "type": "name",
  "seed": null,
  "count": 3,
  "results": [
    "John Smith",
    "Jane Doe",
    "Bob Johnson"
  ]
}
```

Template output:
```json
{
  "type": "template",
  "template": "{{name}} <{{email}}>",
  "seed": null,
  "count": 2,
  "results": [
    "John Smith <john.smith@example.com>",
    "Jane Doe <jane.doe@example.com>"
  ]
}
```

## 7. Behavior

### 7.1. Randomness

1. In Seeded Mode, the system MUST produce the same results for the same inputs, type, and seed.
2. In Unseeded Mode, the system SHOULD use a high-quality randomness source.
3. When generating multiple items, each item MUST be independently random (or deterministically sequenced in Seeded Mode).

### 7.2. Data Realism

1. Generated names MUST come from the existing firstnames.txt and surnames.txt files.
2. Email addresses MUST be valid format (local@domain) but need not represent real addresses.
3. Phone numbers SHOULD follow North American format by default.
4. Addresses SHOULD contain realistic street types and number ranges.
5. Dates MUST be valid calendar dates.
6. UUIDs MUST follow RFC 4122 v4 format.

### 7.3. Text Generation

1. Lorem ipsum MUST use the standard vocabulary starting with "Lorem ipsum dolor sit amet".
2. Word generation MUST select from the lorem ipsum vocabulary.
3. Sentence generation MUST capitalize the first word and end with a period.
4. Paragraph generation MUST contain 4-8 sentences.

### 7.4. Template Processing

1. Templates MUST be processed left-to-right.
2. Each placeholder MUST be independently generated.
3. Malformed placeholders MUST produce an error before any data is generated.
4. When using `--count` with templates, each record MUST have newly generated values for all placeholders.

## 8. Error Handling

1. Invalid count values (less than 1) MUST produce a non-zero exit code.
2. Invalid seed values MUST produce a non-zero exit code.
3. For `number`, if `min > max` MUST produce a non-zero exit code.
4. Unknown data types in template mode MUST produce a non-zero exit code.
5. Malformed template syntax MUST produce a non-zero exit code.
6. Missing or unreadable name data files MUST produce a non-zero exit code.
7. Invalid parameter values in template placeholders MUST produce a non-zero exit code.

## 9. Examples

(Examples retained from the original standalone specification.)

## 10. Security Considerations

1. The `fake` command group MUST NOT be used for generating passwords or cryptographic material.
2. Generated data MUST NOT be represented as secure or production-ready authentication credentials.
3. The system MUST clearly document that generated emails, phone numbers, and addresses are for testing/development only.
4. Generated data SHOULD NOT attempt to validate against real databases or services.
5. When using unseeded mode, the randomness source SHOULD be suitable for generating diverse test data but need not be cryptographically secure.

## 11. Privacy Considerations

1. Generated personal data (names, emails, addresses) is fictional and MUST NOT be derived from real individuals.
2. The name data files MUST contain common names without association to real people.
3. Documentation SHOULD emphasize that generated data is for testing/development and should not be used to impersonate real individuals.

## 12. Testing Considerations

(Testing considerations retained from the original standalone specification.)

## 13. References

[FAKER] https://github.com/faker-js/faker
