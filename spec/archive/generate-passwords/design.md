# Design: Generate passwords
**Specification Reference**: [specification.md](specification.md)

## 1. Context

The password generation feature adds cryptographically secure password/secret generation to Nightwatch, joining existing security utilities (JWT tools, PII redaction, fake data generation). The implementation must balance security (crypto-quality randomness), usability (sensible defaults, clear output), and flexibility (multiple generation modes, customization options).

### Current Architecture

Nightwatch is built on:
- **CLI Framework**: Cobra for command structure and flag parsing
- **Styling**: Lipgloss for terminal output formatting
- **Organization**: Commands in `internal/nightwatch/`, reusable logic in `pkg/`, data files in `internal/nightwatch/data/`

### Key Constraints

1. **Executable Independence**: Password generation is Nightwatch-only; `regimen` must not access it
2. **Security**: Production passwords must use `crypto/rand`; seeded mode is for testing only
3. **Scriptability**: Default output must be clean (single password per line) for piping
4. **Testability**: Seeded mode enables deterministic testing

### Non-Functional Requirements

- **Performance**: Generate 1000 passwords in < 1 second
- **Security**: Meet NIST SP 800-63B recommendations for password entropy
- **Usability**: Sensible defaults (16 chars, letters+digits) requiring no flags
- **Compatibility**: Work in standard POSIX shells for scripting

## 2. Goals and Non-Goals

### Goals

1. **Primary**: Provide cryptographically secure password generation for production use
2. **Secondary**: Support passphrase generation using diceware methodology
3. **Tertiary**: Enable API key generation with common format conventions
4. **Testing**: Support deterministic generation via seeding for test scenarios
5. **Education**: Display entropy to help users understand password strength
6. **Validation**: Provide password strength checking to evaluate existing passwords

### Non-Goals

1. **Password Management**: Not building a password manager (no storage, no encryption)
2. **Network Services**: No online breach checking, no API calls
3. **Advanced Cryptography**: Not generating cryptographic keys (use `ssh-keygen`, `openssl`, etc.)
4. **GUI**: CLI-only interface
5. **Password Policy Enforcement**: Not validating against external policy engines

## 3. Options Considered

### Option 1: Single Command with Modes

Use `nightwatch password generate --mode {password|phrase|api|pattern}` with mode-specific flags.

**Advantages**:
- Single entry point
- Fewer top-level commands
- Easier to document in a single place

**Disadvantages**:
- Flag conflicts between modes (e.g., `--length` means different things)
- Harder to discover mode-specific options
- Cluttered `--help` output showing all flags

**Complexity**: Low

### Option 2: Subcommands per Mode (CHOSEN)

Use `nightwatch password {generate|phrase|api|pattern|check}` with mode-specific flags.

**Advantages**:
- Clear command structure matching mental model
- Mode-specific help is focused and clean
- Follows established Nightwatch patterns (`jwt decode`, `jwt verify`, etc.)
- Easy to discover via `nightwatch password --help`

**Disadvantages**:
- More commands to implement
- Slight verbosity (but aliases can help)

**Complexity**: Low

### Option 3: Multiple Top-Level Commands

Use `nightwatch password`, `nightwatch passphrase`, `nightwatch apikey` as separate command groups.

**Advantages**:
- Shortest command invocations
- Maximum discoverability at top level

**Disadvantages**:
- Fragments related functionality
- Clutters top-level namespace
- Loses semantic grouping

**Complexity**: Low

## 4. Decision

**Chosen Option**: Option 2 (Subcommands per Mode)

**Rationale**: Matches existing Nightwatch command structure (e.g., `jwt` subcommands), provides focused help output, and clearly groups related password functionality without fragmenting the top-level namespace.

**Key Factors**:
1. **Consistency**: Aligns with `jwt` and `fake` command patterns
2. **Usability**: `nightwatch password --help` provides a navigation hub
3. **Extensibility**: Easy to add new generation modes later (e.g., `pin`, `otp`)

## 5. Detailed Design

### Architecture Overview

```
cmd/nightwatch/main.go
    ↓
internal/nightwatch/root.go (adds password command)
    ↓
internal/nightwatch/password/
    ├── root.go (password command group)
    ├── generate.go (character-based passwords)
    ├── phrase.go (passphrase generation)
    ├── api.go (API key generation)
    ├── pattern.go (pattern-based generation)
    └── check.go (strength checking)
    ↓
pkg/password/
    ├── generator.go (core generation logic)
    ├── charset.go (character set definitions)
    ├── entropy.go (entropy calculation)
    ├── passphrase.go (word list handling, word selection)
    ├── pattern.go (pattern parsing and generation)
    ├── checker.go (strength analysis)
    └── rng.go (RNG abstraction: crypto/rand vs seeded)
    ↓
internal/nightwatch/data/
    └── wordlist.txt (EFF long wordlist, 7776 words)
```

### Component Design

#### pkg/password/rng.go

Abstracts randomness source to support both crypto/rand and seeded PRNG:

```go
type RNG interface {
    Int(max int) (int, error)  // Returns [0, max)
    Bytes(n int) ([]byte, error)
}

type CryptoRNG struct{}  // Uses crypto/rand

type SeededRNG struct {
    rng *rand.Rand  // Uses math/rand with seed
}

func NewRNG(seed *int64) RNG {
    if seed != nil {
        return &SeededRNG{rng: rand.New(rand.NewSource(*seed))}
    }
    return &CryptoRNG{}
}
```

#### pkg/password/charset.go

Defines character sets and provides composition logic:

```go
type CharSet struct {
    Name  string
    Chars string
}

var (
    Lowercase = CharSet{Name: "lowercase", Chars: "abcdefghijklmnopqrstuvwxyz"}
    Uppercase = CharSet{Name: "uppercase", Chars: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"}
    Digits    = CharSet{Name: "digits", Chars: "0123456789"}
    Symbols   = CharSet{Name: "symbols", Chars: "!@#$%^&*()-_=+[]{}|;:,.<>?/~"}
    Ambiguous = CharSet{Name: "ambiguous", Chars: "0O1lI5S"}
)

func BuildCharSet(opts CharSetOptions) (string, []CharSet, error)
```

#### pkg/password/generator.go

Core generation logic:

```go
type Generator struct {
    rng RNG
}

func (g *Generator) GeneratePassword(opts PasswordOptions) (string, error)
func (g *Generator) ensureCharSetRepresentation(pwd string, sets []CharSet) bool
```

#### pkg/password/passphrase.go

Word list and passphrase logic:

```go
type WordList struct {
    words []string
    size  int
}

func LoadWordList(path string) (*WordList, error)
func LoadBuiltinWordList() (*WordList, error)
func (wl *WordList) SelectWords(rng RNG, count int) ([]string, error)
func FormatPassphrase(words []string, opts PhraseOptions) string
```

#### pkg/password/entropy.go

Entropy calculation:

```go
func CalculatePasswordEntropy(charsetSize, length int) float64
func CalculatePhraseEntropy(wordlistSize, wordCount int) float64
func CalculatePatternEntropy(pattern string) (float64, error)
```

#### pkg/password/pattern.go

Pattern parsing and generation:

```go
type PatternToken struct {
    Type  string  // "literal", "upper", "lower", "digit", "any", "alnum", "symbol"
    Value string  // For literals
}

func ParsePattern(pattern string) ([]PatternToken, error)
func GenerateFromPattern(tokens []PatternToken, rng RNG) (string, error)
```

#### pkg/password/checker.go

Password strength analysis:

```go
type StrengthReport struct {
    Length       int
    CharSets     []string
    Entropy      float64
    Strength     string  // "weak", "fair", "good", "strong", "excellent"
    HasCommon    bool
    HasKeyboard  bool
}

func CheckPassword(password string) *StrengthReport
```

### Data Design

#### Wordlist Storage

- **File**: `internal/nightwatch/data/wordlist.txt`
- **Format**: One word per line, lowercase, no special characters
- **Source**: EFF long wordlist (7776 words, 5-dice selection)
- **Embedding**: Use `//go:embed` to embed wordlist in binary

```go
//go:embed data/wordlist.txt
var wordlistData string

func LoadBuiltinWordList() (*WordList, error) {
    lines := strings.Split(wordlistData, "\n")
    // ... filter empty lines and comments
}
```

#### Character Set Data

All character sets are defined as string constants in `pkg/password/charset.go`. No external data files needed.

### API Design

Each subcommand follows this pattern:

1. **Parse flags** into an options struct
2. **Create RNG** (crypto or seeded based on `--seed`)
3. **Create Generator** with RNG
4. **Generate password(s)** using Generator
5. **Calculate entropy** if requested
6. **Format output** (text or JSON)
7. **Write to stdout**

Example flow for `generate`:

```go
func RunGenerate(cmd *cobra.Command, args []string) error {
    // 1. Parse flags
    opts := parseGenerateFlags(cmd)
    
    // 2. Validate
    if err := opts.Validate(); err != nil {
        return err
    }
    
    // 3. Warn if seeded
    if opts.Seed != nil {
        fmt.Fprintln(os.Stderr, "WARNING: Seeded passwords are for testing only.")
    }
    
    // 4. Create RNG
    rng := password.NewRNG(opts.Seed)
    gen := password.NewGenerator(rng)
    
    // 5. Generate
    results := []string{}
    for i := 0; i < opts.Count; i++ {
        pwd, err := gen.GeneratePassword(opts)
        if err != nil {
            return err
        }
        results = append(results, pwd)
    }
    
    // 6. Calculate entropy
    entropy := password.CalculatePasswordEntropy(opts.CharsetSize, opts.Length)
    
    // 7. Output
    if opts.JSON {
        return outputJSON(results, entropy, opts)
    }
    return outputText(results, entropy, opts)
}
```

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| Embed wordlist vs external file | No external dependencies, guaranteed availability | Binary size (+50KB), harder to customize | Most users won't customize; consistency matters more |
| Guarantee char representation | Stronger passwords by default | Slight performance cost (may need retries) | Security benefit outweighs minimal perf cost |
| Separate RNG abstraction | Testability via seeding | Extra indirection layer | Testing is critical; abstraction cost is minimal |
| Entropy display opt-in | Clean default output for scripting | User must request educational info | Scriptability is primary use case |
| Subcommands vs flags | Clear structure, focused help | More typing (but can alias) | Usability and discoverability trump brevity |

## 7. Cross-Cutting Concerns

### Security

1. **Randomness Source**: Use `crypto/rand.Read()` for production passwords
   - Provides cryptographically secure random bytes
   - Fails explicitly if entropy source is unavailable
   - Never fall back to weak randomness automatically

2. **Seed Warning**: Emit visible warning when `--seed` is used
   - Printed to stderr before generation
   - Included in JSON output as a warning field
   - Documentation must emphasize testing-only use

3. **Memory Clearing**: Zero out sensitive buffers after use where practical
   - Use `defer` to ensure cleanup
   - Go GC limitations prevent complete control, but best effort matters

4. **Entropy Validation**: Warn if generated passwords have < 40 bits entropy
   - Calculated, not estimated
   - Warning to stderr, doesn't block generation
   - Helps users understand weak configurations

5. **Common Password Detection**: Check against top 1000 common passwords in `check` command
   - Simple substring matching
   - Data embedded in binary
   - Flags password as weak regardless of entropy

### Performance

1. **Target**: Generate 1000 passwords in < 1 second on modern hardware
   - Crypto/rand is fast enough (~10MB/s on typical systems)
   - Charset guarantee may require a few retries, but average case is fast
   - Passphrase generation is I/O-bound on wordlist loading (mitigated by embedding)

2. **Optimization Strategy**:
   - Lazy-load and cache wordlist on first use
   - Reuse RNG instance across multiple generations in same invocation
   - Avoid unnecessary allocations in hot path

3. **Benchmarks**: Add benchmarks for all generation modes
   ```go
   BenchmarkGeneratePassword-8      50000    23456 ns/op
   BenchmarkGeneratePhrase-8       100000    12345 ns/op
   ```

### Reliability

1. **Crypto/rand Failure**: If `crypto/rand` fails, abort with clear error
   - Do not fall back to weak randomness
   - Error message: "Failed to access cryptographic random source"
   - Exit code 2 (critical failure)

2. **Wordlist Loading Failure**: Embedded wordlist should never fail, but handle gracefully
   - If custom wordlist fails to load, show clear error with path
   - Suggest using built-in wordlist

3. **Invalid Options**: Validate all options before generation
   - Empty charset after exclusions → error before generating
   - Length < 1 → error before generating
   - Prevents wasted computation and confusing output

### Testing

1. **Unit Tests**:
   - Test each generation mode with seeded RNG (deterministic)
   - Test charset composition with known seeds
   - Test entropy calculation with known inputs
   - Test pattern parsing with edge cases
   - Test strength checker with known weak/strong passwords

2. **Integration Tests**:
   - Test CLI flag parsing and validation
   - Test JSON output schema
   - Test error exit codes
   - Test stderr warning behavior

3. **Property Tests** (future enhancement):
   - Generated passwords always match requested length
   - Generated passwords always contain required character sets
   - Entropy calculation never returns negative values

4. **Test Data**:
   - Seed 42 for all deterministic tests
   - Small test wordlist (100 words) for passphrase tests
   - Known patterns for pattern generation tests

## 8. Implementation Plan

### Phase 1: Core Infrastructure

**Goal**: Set up package structure and RNG abstraction

- [x] Create `pkg/password/` directory
- [ ] Implement `rng.go` (RNG interface, CryptoRNG, SeededRNG)
- [ ] Implement `charset.go` (character set definitions and composition)
- [ ] Implement `entropy.go` (entropy calculation functions)
- [ ] Write tests for RNG and charset modules

**Milestone**: `go test ./pkg/password` passes for RNG and charset

### Phase 2: Password Generation

**Goal**: Implement character-based password generation

- [ ] Implement `generator.go` (core password generation logic)
- [ ] Add charset guarantee logic (ensure all sets represented)
- [ ] Create `internal/nightwatch/password/` directory
- [ ] Implement `root.go` (password command group)
- [ ] Implement `generate.go` (generate subcommand)
- [ ] Wire into `internal/nightwatch/root.go`
- [ ] Add tests for password generation with various options

**Milestone**: `nightwatch password generate` works with all flags

### Phase 3: Passphrase Generation

**Goal**: Implement word-based passphrase generation

- [ ] Download EFF long wordlist to `internal/nightwatch/data/wordlist.txt`
- [ ] Implement `passphrase.go` (word list loading and selection)
- [ ] Add `//go:embed` for wordlist
- [ ] Implement `phrase.go` (phrase subcommand)
- [ ] Add tests for passphrase generation with built-in and custom wordlists

**Milestone**: `nightwatch password phrase` works with all flags

### Phase 4: API Key and Pattern Generation

**Goal**: Implement specialized generation modes

- [ ] Implement `pattern.go` (pattern parsing and generation)
- [ ] Implement `api.go` (api subcommand, format handling)
- [ ] Implement `pattern.go` subcommand (pattern subcommand)
- [ ] Add tests for API key formats (hex, base32, base64, base64url)
- [ ] Add tests for pattern parsing and generation

**Milestone**: `nightwatch password api` and `nightwatch password pattern` work

### Phase 5: Strength Checking

**Goal**: Implement password strength analysis

- [ ] Implement `checker.go` (strength analysis logic)
- [ ] Add common password list to data files
- [ ] Implement `check.go` (check subcommand)
- [ ] Add tests for strength checking with known passwords

**Milestone**: `nightwatch password check` provides accurate strength reports

### Phase 6: Polish and Documentation

**Goal**: Finalize output formatting and JSON support

- [ ] Implement JSON output for all subcommands
- [ ] Add `--show-entropy` support to all generation commands
- [ ] Implement seeded mode warning to stderr
- [ ] Add examples to command help text
- [ ] Write comprehensive tests for JSON schema
- [ ] Run `make fmt && make lint && make test`

**Milestone**: All acceptance criteria in specification are met

### Migration Strategy

No migration needed (new feature). However:

1. Document integration with existing workflows (e.g., piping to `redact check`)
2. Cross-reference in existing Nightwatch documentation
3. Consider adding password generation to `fake` command's "don't use for passwords" warnings

## 9. Open Questions

1. **Should we support XKCD-style passphrase formatting (no separators, random capitalization)?**
   - **Status**: Open
   - **Owner**: Implementation phase decision
   - **Blocking**: No (can be added later as a flag)
   - **Recommendation**: Start with standard separators; add XKCD mode if requested

2. **Should `check` command detect keyboard walks (e.g., "qwerty", "asdfgh")?**
   - **Status**: Open
   - **Owner**: Implementation phase decision
   - **Blocking**: No (nice-to-have for v1)
   - **Recommendation**: Add basic pattern detection; defer advanced checks to v2

3. **Should we provide shell completion for common flags?**
   - **Status**: Open
   - **Owner**: Polish phase
   - **Blocking**: No (Cobra generates basic completion automatically)
   - **Recommendation**: Let Cobra handle it; customize later if needed

4. **What should be the maximum allowed password length?**
   - **Status**: Open
   - **Owner**: Implementation phase decision
   - **Blocking**: No
   - **Recommendation**: 1024 characters (prevents accidental huge allocations, still very generous)

5. **Should we include an interactive mode for users unsure of options?**
   - **Status**: Deferred
   - **Owner**: Future enhancement
   - **Blocking**: No
   - **Recommendation**: Keep v1 CLI-only; consider interactive TUI in future if demand exists
