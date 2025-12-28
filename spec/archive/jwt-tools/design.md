# Design: jwt (JSON Web Token utilities)

**Specification Reference**: [specification.md](specification.md)

**Status**: Draft

## 1. Context

Nightwatch is structured as a Cobra-based CLI with modular command groups (currently focusing on `when` for timezone utilities). The project emphasizes local-first, scriptable tools that provide predictable output and integrate well with Unix pipelines.

JWT handling requires cryptographic operations for signing and verification. Go's standard library provides `crypto/hmac`, `crypto/rsa`, `crypto/ecdsa`, and related packages, but does not include complete JWT parsing/generation. Third-party libraries like `golang-jwt/jwt` or `lestrrat-go/jwx` provide production-ready JWT implementations with support for all standard algorithms.

Key design constraints:
- Must support stdin input for secure token/secret handling
- Must provide both human-readable and JSON output
- Must handle PEM-encoded keys for asymmetric algorithms
- Must integrate cleanly with existing Cobra command structure

## 2. Goals and Non-Goals

### Goals

- Implement `nightwatch jwt` command group as defined in the specification
- Support all specified algorithms (HS256/384/512, RS256/384/512, ES256/384/512)
- Provide secure handling of secrets and tokens (stdin support, no logging)
- Ensure round-trip compatibility (create → verify cycles work correctly)
- Generate stable JSON output suitable for automation

### Non-Goals

- Supporting legacy/deprecated algorithms (none, RSA-with-MD5, etc.)
- Implementing JWT encryption (JWE) - only signatures (JWS)
- Online token validation against JWKS endpoints
- Interactive key generation utilities (use existing tools like `openssl`)
- Token refresh/renewal flows (application-level concern)

## 3. Options Considered

### Option 1: Use golang-jwt/jwt library

`github.com/golang-jwt/jwt/v5` - Fork of the original dgrijalva/jwt-go library, actively maintained.

**Advantages**:
- Mature, widely-used library
- Simple API for common operations
- Good documentation and examples
- Built-in support for all standard algorithms
- Clear error types for verification failures

**Disadvantages**:
- Slightly verbose API for custom claims
- Some deprecated patterns in older examples

**Complexity**: Low

### Option 2: Use lestrrat-go/jwx library

`github.com/lestrrat-go/jwx/v2` - Comprehensive JWT/JWS/JWE/JWK library.

**Advantages**:
- More complete feature set (JWE, JWK, etc.)
- Modern API design
- Better support for advanced features
- Active development

**Disadvantages**:
- Larger dependency footprint
- More complex API for simple operations
- Overkill for current requirements

**Complexity**: Medium

### Option 3: Implement JWT from scratch

Use Go standard library crypto packages directly.

**Advantages**:
- No external dependencies
- Complete control over implementation
- Minimal binary size impact

**Disadvantages**:
- Significant implementation effort
- Higher risk of security bugs
- Need to handle base64url encoding edge cases
- Must maintain algorithm support manually

**Complexity**: High

## 4. Decision

**Chosen Option**: Option 1 (golang-jwt/jwt library)

**Rationale**: The golang-jwt/jwt library provides the right balance of simplicity and completeness for this use case. It handles all the required algorithms, has a proven security track record, and its API is straightforward for CLI usage. The library is widely adopted and actively maintained.

**Key Factors**:
1. Production-ready security implementation
2. Simple API suitable for CLI operations
3. Minimal dependencies
4. Active maintenance and community support
5. Sufficient for current requirements without over-engineering

## 5. Detailed Design

### Architecture Overview

```
cmd/
  nightwatch/
    main.go
internal/
  nightwatch/
    jwt/
      decode.go      # Decode/header/payload commands
      verify.go      # Verification logic
      create.go      # Token creation
      exp.go         # Expiration checking
      keys.go        # Key loading (PEM files)
      output.go      # Formatting (text/JSON)
    root.go          # Root command
```

### Component Design

#### JWT Command Group (`internal/nightwatch/jwt/`)

**Core Components**:

1. **Token Parser**: Wraps `jwt.Parse()` for decoding without verification
2. **Verifier**: Uses `jwt.ParseWithClaims()` with appropriate key function
3. **Creator**: Uses `jwt.NewWithClaims()` and signs with provided key/secret
4. **Key Loader**: Reads PEM files and extracts RSA/ECDSA keys
5. **Output Formatter**: Converts results to text or JSON format

#### Data Structures

```go
// DecodedToken represents a parsed JWT
type DecodedToken struct {
    Header    map[string]interface{} `json:"header"`
    Payload   map[string]interface{} `json:"payload"`
    Signature string                 `json:"signature"`
    Verified  bool                   `json:"verified"`
    Raw       string                 `json:"raw"`
}

// VerifyResult represents verification output
type VerifyResult struct {
    Valid     bool                   `json:"valid"`
    Header    map[string]interface{} `json:"header"`
    Payload   map[string]interface{} `json:"payload"`
    Verified  bool                   `json:"verified"`
    Algorithm string                 `json:"algorithm"`
    Error     string                 `json:"error,omitempty"`
}

// CreateResult represents token creation output
type CreateResult struct {
    Token   string                 `json:"token"`
    Header  map[string]interface{} `json:"header"`
    Payload map[string]interface{} `json:"payload"`
}

// ExpResult represents expiration check output
type ExpResult struct {
    IssuedAt  *time.Time `json:"iat,omitempty"`
    ExpiresAt *time.Time `json:"exp,omitempty"`
    Now       time.Time  `json:"now"`
    Status    string     `json:"status"` // "valid", "expired", "no-expiry"
    Remaining string     `json:"remaining,omitempty"`
}
```

### Algorithm Implementation

#### Decode Flow

1. Accept token from arg or stdin
2. Split token by dots, validate 3 parts
3. Base64url decode each part
4. Parse JSON from header and payload parts
5. Format output (raw JSON for header/payload, structured for decode)

#### Verify Flow

1. Decode token (as above)
2. Determine algorithm from header
3. Load appropriate key/secret based on flags and algorithm
4. Use `jwt.Parse()` with key function
5. Check standard claims (exp, nbf) - library handles this
6. Return verification result with appropriate exit code

#### Create Flow

1. Parse claim flags and build claims map
2. Handle duration strings for `exp` and `nbf` (compute absolute times)
3. Set `iat` if not provided
4. Create token with specified algorithm
5. Load key/secret based on flags and algorithm
6. Sign token
7. Output token string (or JSON with full details)

#### Key Loading

```go
// LoadKey determines key type and loads it
func LoadKey(path string, forSigning bool) (interface{}, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    block, _ := pem.Decode(data)
    if block == nil {
        return nil, errors.New("invalid PEM format")
    }
    
    // Try RSA
    if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
        if forSigning {
            return key, nil
        }
        return &key.PublicKey, nil
    }
    
    // Try PKCS8 (handles RSA and ECDSA)
    if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
        if forSigning {
            return key, nil
        }
        // Extract public key from private key
        switch k := key.(type) {
        case *rsa.PrivateKey:
            return &k.PublicKey, nil
        case *ecdsa.PrivateKey:
            return &k.PublicKey, nil
        }
    }
    
    // Try public keys
    if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
        return key, nil
    }
    
    return nil, errors.New("unable to parse key")
}
```

### Input Handling

```go
// GetTokenInput reads token from arg or stdin
func GetTokenInput(args []string) (string, error) {
    if len(args) > 0 {
        return args[0], nil
    }
    
    // Check if stdin has data
    stat, _ := os.Stdin.Stat()
    if (stat.Mode() & os.ModeCharDevice) == 0 {
        scanner := bufio.NewScanner(os.Stdin)
        if scanner.Scan() {
            return scanner.Text(), nil
        }
        return "", scanner.Err()
    }
    
    return "", errors.New("no token provided")
}
```

### Duration Parsing

```go
// ParseDuration handles formats like "1h", "30m", "7d"
func ParseDuration(s string) (time.Duration, error) {
    // Handle days (not in stdlib)
    if strings.HasSuffix(s, "d") {
        days, err := strconv.Atoi(s[:len(s)-1])
        if err != nil {
            return 0, err
        }
        return time.Duration(days) * 24 * time.Hour, nil
    }
    return time.ParseDuration(s)
}
```

### Error Handling Strategy

Map JWT library errors to exit codes:
- Parse errors: exit 1
- Verification failure: exit 2
- Expired token: exit 3
- Not yet valid: exit 4

Use typed errors for clear messaging:

```go
type ErrorType int

const (
    ErrInvalidFormat ErrorType = iota
    ErrVerificationFailed
    ErrExpired
    ErrNotYetValid
    ErrUnsupportedAlgorithm
    ErrKeyLoad
)

func (e *JWTError) ExitCode() int {
    switch e.Type {
    case ErrInvalidFormat, ErrUnsupportedAlgorithm, ErrKeyLoad:
        return 1
    case ErrVerificationFailed:
        return 2
    case ErrExpired:
        return 3
    case ErrNotYetValid:
        return 4
    default:
        return 1
    }
}
```

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| Use external JWT library | Security, correctness, maintenance | Small dependency | Cryptographic code should not be written from scratch |
| Support stdin for tokens | Security (no process list exposure) | Slightly more complex input handling | Critical for production use |
| Raw JSON for header/payload commands | Pipe-ability to jq | Less user-friendly standalone | Aligns with Unix philosophy |
| Distinct exit codes for error types | Script-friendly error handling | More complex error management | Enables better automation |

## 7. Cross-Cutting Concerns

### Security

- **Key Material**: Never log secrets or keys; use secure memory clearing where possible
- **File Permissions**: Check and warn if private key files are not user-only (600)
- **Process List**: Warn when secrets provided via CLI args (visible in `ps`)
- **Timing Attacks**: Rely on crypto library's constant-time comparisons
- **Algorithm Verification**: Always check algorithm matches expected type (prevent key confusion attacks)

### Performance

- Token parsing and verification are O(1) operations relative to token count
- Key loading: Cache parsed keys within command execution if processing multiple tokens
- No performance concerns for typical CLI usage

### Reliability

- Invalid tokens must fail gracefully with clear error messages
- Malformed PEM files must produce actionable errors
- Base64 decoding errors must be distinguishable from JSON errors

### Testing

**Unit Tests**:
- Token parsing (valid and malformed)
- Claim extraction and formatting
- Duration parsing
- Key loading (various PEM formats)
- Algorithm matching

**Integration Tests**:
- Create + verify round-trip for each algorithm
- Expiration validation
- Error exit codes
- JSON output schema validation
- Stdin input handling

**Security Tests**:
- Algorithm confusion (verify RS256 token with HS256)
- Invalid signatures
- Expired tokens
- Modified tokens (tampered header/payload)

## 8. Implementation Plan

### Phase 1: Core infrastructure
- Add `cmd/jwt.go` command group registration
- Implement token parsing and decoding (no verification)
- Add output formatting (text and JSON)
- Implement `decode`, `header`, `payload` commands

**Milestone**: `nightwatch jwt decode <token>` works for all algorithm types

### Phase 2: Verification
- Implement key/secret loading
- Add verification logic with `golang-jwt/jwt`
- Implement claim validation (exp, nbf)
- Add `verify` command

**Milestone**: Can verify tokens with HMAC and RSA keys

### Phase 3: Creation
- Implement claim parsing from flags
- Add duration parsing for exp/nbf
- Implement signing with HMAC
- Implement signing with RSA/ECDSA keys
- Add `create` command

**Milestone**: Can create and verify round-trip for HS256 and RS256

### Phase 4: Utilities and polish
- Implement `exp` command
- Add stdin input support for all commands
- Add ECDSA algorithm support
- Add comprehensive error handling with exit codes
- Add key file permission warnings

**Milestone**: All commands work with all supported algorithms

### Phase 5: Testing and documentation
- Add unit tests for all components
- Add integration tests with golden output
- Add security test cases
- Document usage with examples

**Milestone**: Full test coverage and user documentation

### Migration Strategy

None - new command group with no existing functionality.

## 9. Open Questions

1. **Default Algorithm**: Should `create` default to HS256 if `--alg` not specified, or require explicit algorithm selection?
   - **Recommendation**: Default to HS256 for HMAC operations, require `--alg` for asymmetric algorithms

2. **Secret Input**: Should we support reading secrets from environment variables or files in addition to CLI args and stdin?
   - **Recommendation**: Support `--secret-file` flag for reading secrets from files

3. **Multiple Token Processing**: Should stdin mode support processing multiple tokens (one per line) and output multiple results?
   - **Recommendation**: Yes, process line-by-line and output results sequentially (easier for batch operations)

4. **Key Format Auto-detection**: Should we support additional key formats beyond PEM (e.g., JWK)?
   - **Recommendation**: PEM only initially; JWK can be future work

5. **Claim Value Types**: How should we handle non-string claim values when specified via `--claim`?
   - **Recommendation**: Support JSON parsing - if value is valid JSON, parse it; otherwise treat as string
