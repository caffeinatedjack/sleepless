# Implementation: Generate passwords

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

**Status**: Complete

## Overview

This implementation adds a `nightwatch password` command group with five subcommands: `generate` (character-based passwords), `phrase` (word-based passphrases), `api` (API key formats), `pattern` (custom patterns), and `check` (strength analysis). The core logic lives in `pkg/password/` for reusability, with CLI wiring in `internal/nightwatch/password.go`. All generation uses `crypto/rand` by default (via `pkg/decide.RNG`), with optional seeding for testing.

## Prerequisites

1. Go 1.21+ installed (project requirement)
2. EFF long wordlist downloaded or prepared for embedding
3. Familiarity with `crypto/rand` and `math/rand` packages
4. Understanding of entropy calculation (log2-based formulas)

## Phases

### Phase 1: Core Infrastructure

**Goal**: Set up package structure, RNG abstraction, character sets, and entropy calculation

- [x] Create `pkg/password/` directory
- [x] Implement `pkg/password/rng.go`:
  - [x] Reused existing `pkg/decide.RNG` interface (already has `Intn(n)` method)
  - [x] Reused `CryptoRNG` using `crypto/rand`
  - [x] Reused `SeededRNG` using `math/rand` with seed
  - [x] Reused `NewRNG(seed *int64)` constructor
- [x] Implement `pkg/password/charset.go`:
  - [x] Define `CharSet` struct (Name, Chars)
  - [x] Define predefined charsets (Lowercase, Uppercase, Digits, Symbols, Ambiguous)
  - [x] Implement `BuildCharSet(options)` to compose final charset
  - [x] Implement charset validation (non-empty)
- [x] Implement `pkg/password/entropy.go`:
  - [x] Implement `CalculatePasswordEntropy(charsetSize, length)`
  - [x] Implement `CalculatePhraseEntropy(wordlistSize, wordCount)`
  - [x] Implement `CalculatePatternEntropy(pattern)` in pattern.go
- [x] Write unit tests for all Phase 1 components:
  - [x] Test RNG determinism with seeded mode
  - [x] Test charset composition with various flag combinations
  - [x] Test entropy calculations with known inputs
  - [x] Test edge cases (empty charsets, zero length, etc.)

**Milestone**: `go test ./pkg/password` passes for RNG, charset, and entropy modules

### Phase 2: Password Generation

**Goal**: Implement character-based password generation and wire to CLI

- [x] Implement `pkg/password/generator.go`:
  - [x] Define `Generator` struct with RNG field
  - [x] Define `PasswordOptions` struct (length, charsets, requireAll, etc.)
  - [x] Implement `NewGenerator(rng RNG)` constructor
  - [x] Implement `GeneratePassword(opts PasswordOptions)` method
  - [x] Implement charset guarantee logic (ensure one char from each enabled set)
  - [x] Handle `--no-require-all` flag logic
- [x] Implement CLI in `internal/nightwatch/password.go` (single file approach):
  - [x] Define `passwordCmd` using Cobra
  - [x] Add common flags: `--count`, `--seed`, `--json`, `--show-entropy`
  - [x] Register all subcommands
- [x] Implement `passwordGenerateCmd`:
  - [x] Define `generateCmd` with all password-specific flags
  - [x] Implement flag parsing into `PasswordOptions`
  - [x] Implement `runPasswordGenerate` to call `generator.GeneratePassword`
  - [x] Implement text output (one password per line)
  - [x] Implement JSON output (mode, seed, count, entropy, results)
  - [x] Emit warning to stderr when `--seed` is used
- [x] Wire password command into `internal/nightwatch/root.go`:
  - [x] Added via init() in password.go
  - [x] Updated root command long description
- [x] Write tests for Phase 2:
  - [x] Test password generation with default options
  - [x] Test all charset combinations (no-uppercase, no-digits, etc.)
  - [x] Test charset guarantee with deterministic seed
  - [x] Test `--count` generates correct number of passwords
  - [x] Test error handling (empty charset, invalid length)

**Milestone**: `nightwatch password generate` works with all flags; `make test` passes

### Phase 3: Passphrase Generation

**Goal**: Implement word-based passphrase generation using EFF wordlist

- [x] Create wordlist:
  - [x] Created EFF-inspired wordlist with ~1000 common English words
  - [x] Saved as `internal/nightwatch/data/wordlist.txt`
- [x] Implement `pkg/password/passphrase.go`:
  - [x] Embed wordlist in `internal/nightwatch/password.go` using `//go:embed`
  - [x] Define `WordList` struct (words slice, size)
  - [x] Implement `LoadWordListFromString()` for embedded data
  - [x] Implement `LoadWordList(path)` for custom wordlists
  - [x] Implement word list parsing (skip empty lines, skip comments starting with `#`)
  - [x] Implement `SelectWords(rng, count)` to randomly select words
  - [x] Define `PhraseOptions` struct (separator, capitalize, numbers)
  - [x] Implement `FormatPassphrase(words, opts)` to join and format words
- [x] Implement `passwordPhraseCmd`:
  - [x] Define `phraseCmd` with passphrase-specific flags
  - [x] Implement flag parsing into `PhraseOptions`
  - [x] Implement `runPasswordPhrase` to call passphrase logic
  - [x] Implement text output (one passphrase per line)
  - [x] Implement JSON output with phrase-specific fields
  - [x] Handle custom wordlist loading errors gracefully
  - [x] Warn if custom wordlist has < 100 words (low entropy)
- [x] Write tests for Phase 3:
  - [x] Test custom wordlist loading from string
  - [x] Test word selection with deterministic seed
  - [x] Test formatting options (separator, capitalize, numbers)
  - [x] Test entropy calculation for 6-word passphrase
  - [x] Test error handling (empty wordlist)

**Milestone**: `nightwatch password phrase` works with all flags; wordlist is embedded correctly

### Phase 4: API Key and Pattern Generation

**Goal**: Implement API key formats and custom pattern generation

- [x] Implement API key generation in `pkg/password/apikey.go`:
  - [x] Define `APIKeyOptions` struct (prefix, length, format)
  - [x] Define format encoders: hex, base32, base64, base64url
  - [x] Implement `GenerateAPIKey(opts, rng)` function
  - [x] Handle prefix formatting (prefix + "_" + random)
- [x] Implement `passwordAPICmd`:
  - [x] Define `apiCmd` with API-specific flags
  - [x] Implement flag parsing and validation (valid format)
  - [x] Implement `runPasswordAPI` to call `GenerateAPIKey`
  - [x] Implement text and JSON output
- [x] Implement pattern parsing and generation in `pkg/password/pattern.go`:
  - [x] Define `PatternToken` struct (Type, Value)
  - [x] Implement `ParsePattern(pattern string)` to tokenize pattern
  - [x] Handle escape sequences (`\X`, `\9`, etc.)
  - [x] Implement `GenerateFromPattern(tokens, rng)` to generate string
  - [x] Map pattern chars to charsets (X=upper, x=lower, 9=digit, etc.)
  - [x] Calculate pattern entropy based on token types
- [x] Implement `passwordPatternCmd`:
  - [x] Define `patternCmd` with pattern argument
  - [x] Implement `runPasswordPattern` to parse and generate
  - [x] Implement text and JSON output
  - [x] Handle invalid patterns with clear error messages
- [x] Write tests for Phase 4:
  - [x] Test all API key formats (hex, base32, base64, base64url)
  - [x] Test API key prefix formatting
  - [x] Test pattern parsing (valid and invalid patterns)
  - [x] Test pattern generation with deterministic seed
  - [x] Test escape sequences in patterns

**Milestone**: `nightwatch password api` and `nightwatch password pattern` work correctly

### Phase 5: Strength Checking

**Goal**: Implement password strength analysis

- [x] Common password list (embedded in code):
  - [x] Embedded top 30 common passwords directly in checker.go
  - [x] Embedded common keyboard walks in checker.go
- [x] Implement `pkg/password/checker.go`:
  - [x] Define `StrengthReport` struct (Length, CharSets, Entropy, Strength, HasCommon, HasKeyboard)
  - [x] Implement `CheckPassword(password string)` function:
    - [x] Detect character set composition (lowercase, uppercase, digits, symbols)
    - [x] Calculate entropy based on detected charset and length
    - [x] Map entropy to strength rating (weak/fair/good/strong/excellent)
    - [x] Check against common password list
    - [x] Check for simple keyboard walks (basic patterns)
  - [x] Implement strength rating thresholds (< 40 = weak, 40-59 = fair, etc.)
- [x] Implement `passwordCheckCmd`:
  - [x] Define `checkCmd` with password argument
  - [x] Implement `runPasswordCheck` to call `CheckPassword`
  - [x] Implement human-readable text output (labeled fields)
  - [x] Implement JSON output (StrengthReport as JSON)
  - [x] Add warnings for common passwords or weak patterns
- [x] Write tests for Phase 5:
  - [x] Test strength calculation for known weak passwords
  - [x] Test strength calculation for known strong passwords
  - [x] Test common password detection
  - [x] Test character set detection

**Milestone**: `nightwatch password check` provides accurate strength analysis

### Phase 6: Polish and Documentation

**Goal**: Finalize output formatting, JSON support, warnings, and help text

- [x] Implement `--show-entropy` flag support:
  - [x] Add entropy display to text output for generate, phrase, api, pattern
  - [x] Format entropy nicely (e.g., "Entropy: 77.5 bits (strong)")
  - [x] Include entropy in JSON output automatically
- [x] Implement seeded mode warning:
  - [x] Emit warning to stderr in all generation commands when `--seed` is used
  - [x] Format warning clearly: "WARNING: Seeded passwords are for testing only. DO NOT use for production."
  - [x] Include warning in JSON output as a `warning` field when seeded
- [x] Add low-entropy warnings:
  - [x] Emit warning to stderr if generated password has < 40 bits entropy
  - [x] Suggest increasing length or enabling more character sets
- [x] Enhance command help text:
  - [x] Add examples to `--help` output for each subcommand
  - [x] Document default values clearly
  - [x] Add security warnings in long description
- [x] Implement comprehensive error messages:
  - [x] Empty charset error: "No character sets enabled. Enable at least one set."
  - [x] Invalid format error: "Invalid format 'foo'. Valid formats: hex, base32, base64, base64url"
- [x] Update help text and examples:
  - [x] Ensure `nightwatch --help` shows password command
  - [x] Ensure `nightwatch password --help` shows all subcommands
  - [x] Add rich examples to each subcommand's help
- [x] Run quality checks:
  - [x] `make fmt` - format all code
  - [x] `make lint` - run linters (vet)
  - [x] `make test` - run all tests with race detector
  - [x] `make build` - ensure binaries build successfully

**Milestone**: All acceptance criteria in specification are met; tests pass; code is clean

## Testing Plan

### Unit Tests

**Core logic in `pkg/password/`**:

- **RNG Module** (reuses pkg/decide):
  - [x] Seeded RNG produces identical results for same seed
  - [x] RNG.Intn(max) returns values in [0, max)

- **Charset Module**:
  - [x] BuildCharSet with all flags enabled produces correct composition
  - [x] BuildCharSet with exclusions produces correct composition
  - [x] BuildCharSet with all sets excluded returns error
  - [x] Ambiguous character exclusion works correctly

- **Entropy Module**:
  - [x] CalculatePasswordEntropy(62, 16) ≈ 95.3 bits (alphanumeric)
  - [x] CalculatePhraseEntropy(7776, 6) ≈ 77.5 bits (standard diceware)
  - [x] EntropyToStrength maps correctly

- **Generator Module**:
  - [x] GeneratePassword returns correct length
  - [x] GeneratePassword includes one char from each enabled set (requireAll)
  - [x] GeneratePassword is deterministic with seeded RNG
  - [x] GeneratePassword errors on empty charset
  - [x] GeneratePassword errors on invalid length

- **Passphrase Module**:
  - [x] LoadWordListFromString handles data correctly
  - [x] LoadWordListFromString skips comments and blank lines
  - [x] SelectWords with seed produces deterministic results
  - [x] FormatPassphrase applies separator correctly
  - [x] FormatPassphrase capitalizes when requested

- **Pattern Module**:
  - [x] ParsePattern tokenizes simple patterns correctly (XXX-999)
  - [x] ParsePattern handles escapes (\X becomes literal X)
  - [x] GenerateFromPattern matches pattern structure
  - [x] Invalid escape sequences return error

- **Checker Module**:
  - [x] CheckPassword detects all character set combinations
  - [x] CheckPassword maps entropy to correct strength rating
  - [x] CheckPassword detects common passwords

### Integration Tests

**CLI behavior in `internal/nightwatch/password.go`** (to be tested manually by user):

- [ ] `nightwatch password generate` outputs 16-char password by default
- [ ] `nightwatch password generate --length 20` outputs 20-char password
- [ ] `nightwatch password generate --symbols` includes symbol characters
- [ ] `nightwatch password generate --count 5` outputs 5 passwords
- [ ] `nightwatch password generate --json` produces valid JSON schema
- [ ] `nightwatch password generate --seed 42` twice produces identical output
- [ ] `nightwatch password generate --seed 42` emits warning to stderr
- [ ] `nightwatch password phrase` outputs 6-word passphrase by default
- [ ] `nightwatch password phrase --words 8` outputs 8-word passphrase
- [ ] `nightwatch password phrase --separator " "` uses space separator
- [ ] `nightwatch password api` outputs 32-char hex key
- [ ] `nightwatch password api --prefix "sk_prod"` includes prefix
- [ ] `nightwatch password pattern "XXX-999"` matches pattern format
- [ ] `nightwatch password check "password123"` reports weak strength
- [ ] Invalid flags produce non-zero exit code and error message

### Acceptance Tests

**End-to-end validation of specification requirements**:

- [ ] All subcommands accessible via `nightwatch password {subcommand}`
- [ ] Default password is 16 characters with lowercase, uppercase, digits
- [ ] `--symbols` adds symbols to charset
- [ ] `--no-uppercase` excludes uppercase letters
- [ ] Charset guarantee ensures one char from each enabled set by default
- [ ] `--count` generates multiple independent passwords
- [ ] `--seed` produces deterministic output with warning
- [ ] `--json` produces valid, parseable JSON for all commands
- [ ] `--show-entropy` displays entropy calculation
- [ ] Passphrase uses 6 words by default with `-` separator
- [ ] API key format is `prefix_randomchars` when prefix provided
- [ ] Pattern `XXX-999-xxx` generates correct structure
- [ ] Strength check reports length, charsets, entropy, strength rating
- [ ] Crypto/rand failure produces exit code 2
- [ ] Invalid options produce exit code 1
- [ ] Help text is clear and includes examples

## Rollback Plan

This is a new feature with no existing state or users. Rollback is simple:

1. **If critical bug found**: Revert the commit(s) that added the feature
2. **If design flaw identified**: Disable command by removing from root command registration
3. **If security issue found**: Immediately revert and issue advisory

No data migration or backward compatibility concerns exist.

## Open Questions

1. **Should we validate that wordlists don't contain duplicate words?**
   - **Impact**: Medium (affects entropy calculation accuracy)
   - **Recommendation**: Warn if duplicates detected, but don't block
   - **Status**: Deferred to implementation phase

2. **Should `check` command support reading from stdin for batch checking?**
   - **Impact**: Low (nice-to-have feature)
   - **Recommendation**: Keep v1 simple with single password argument; add stdin in v2 if requested
   - **Status**: Deferred to future enhancement

3. **What should be the behavior if crypto/rand is unavailable?**
   - **Impact**: High (security critical)
   - **Recommendation**: Fail immediately with exit code 2; do NOT fall back to weak randomness
   - **Status**: Resolved (design specifies fail-fast)

4. **Should we add a `--clipboard` flag to copy password to clipboard?**
   - **Impact**: Medium (usability enhancement, but adds clipboard dependency)
   - **Recommendation**: Keep v1 CLI-only; users can pipe to `xclip`/`pbcopy` themselves
   - **Status**: Deferred to future enhancement

5. **How should we handle very large `--count` values (e.g., 1 million)?**
   - **Impact**: Low (performance concern)
   - **Recommendation**: Set a reasonable maximum (e.g., 10,000) to prevent accidental resource exhaustion
   - **Status**: Deferred to implementation phase
