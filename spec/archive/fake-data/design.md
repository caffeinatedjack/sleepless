# Design: fake (fake data generation)

**Specification Reference**: [specification.md](specification.md)

**Status**: Approved

## 1. Context

Nightwatch is a security-focused CLI tool built on Cobra. Adding a fake data generator expands its utility for developers who need realistic test data without compromising security (no network calls, reproducible output). The implementation will reuse the existing name data files from `internal/nightwatch/data/` and follow the established pattern of using `math/rand` for seeded mode and `crypto/rand` for unseeded mode (similar to the `decide` command group).

The main technical challenges are: (1) providing a flexible template engine that doesn't introduce heavyweight dependencies, (2) maintaining the lorem ipsum word corpus, (3) ensuring all data types generate realistic output, and (4) supporting both streaming output and JSON format.

## 2. Goals and Non-Goals

### Goals

- Implement all data type generators specified in the specification
- Provide deterministic output with `--seed` for testing scenarios
- Support template mode for structured data generation
- Reuse existing name data files from nightwatch
- Keep dependencies minimal (no external faker libraries)
- Maintain consistent JSON output schema

### Non-Goals

- Advanced data generation (credit card checksums, real address validation)
- Localization (non-English names, addresses)
- Database integration or data persistence
- Complex template logic (conditionals, loops)

## 3. Options Considered

### Option 1: Use external faker library (e.g., go-faker)

**Advantages**:
- Feature-rich with many data types
- Well-tested
- Consistent data formats

**Disadvantages**:
- External dependency
- May include features we don't need
- Less control over output format
- Harder to guarantee deterministic behavior with seeds

**Complexity**: Low

### Option 2: Build custom generators using existing patterns

**Advantages**:
- No external dependencies
- Full control over output format
- Can reuse existing name data files
- Consistent with `decide` command's RNG approach
- Lightweight

**Disadvantages**:
- More code to write and maintain
- Need to define our own lorem ipsum corpus
- Need to implement each generator

**Complexity**: Medium

### Option 3: Hybrid approach (external library with custom wrappers)

**Advantages**:
- Faster initial implementation
- Can override specific generators

**Disadvantages**:
- Still introduces dependency
- Wrapper complexity
- Seed handling may conflict

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 2 (custom generators)

**Rationale**: Aligns with the project's goal of minimal dependencies and gives full control over reproducibility and output formats. The existing name data files and `decide` command's RNG patterns provide a solid foundation.

**Key Factors**:
1. No new external dependencies
2. Full control over seeded/unseeded randomness
3. Consistent with existing codebase patterns
4. Lightweight and maintainable

## 5. Detailed Design

### Architecture Overview

```
cmd/nightwatch/
├── main.go (existing)
└── fake.go (new root command)

internal/nightwatch/
├── data/ (existing)
│   ├── firstnames.txt
│   ├── surnames.txt
│   └── lorem.txt (new)
├── fake/ (new package)
│   ├── generators.go     # Data type generators
│   ├── lorem.go          # Lorem ipsum logic
│   ├── template.go       # Template parsing and rendering
│   ├── rng.go           # RNG abstraction (shared pattern with decide)
│   └── types.go         # Type definitions and output structs
└── fake.go (new command handler)

pkg/ (potential shared code if needed)
```

### Component Design

#### RNG Abstraction (`internal/nightwatch/fake/rng.go`)

Reuse the same pattern from `decide`:

```go
type RNG interface {
    Intn(n int) int
    Int63n(n int64) int64
}

// CryptoRNG for unseeded mode
type CryptoRNG struct{}

// SeededRNG wraps math/rand.Rand for seeded mode
type SeededRNG struct {
    r *rand.Rand
}
```

#### Generator Interface (`internal/nightwatch/fake/generators.go`)

```go
type Generator interface {
    Generate(rng RNG) string
}

type GeneratorFunc func(rng RNG) string

// Data type generators
func Name(firstnames, lastnames []string) GeneratorFunc
func Email(names []string) GeneratorFunc
func Phone() GeneratorFunc
func Address() GeneratorFunc
func City() GeneratorFunc
func UUID() GeneratorFunc
func Number(min, max int) GeneratorFunc
func Date(pastDays, futureDays int) GeneratorFunc
// ... etc
```

Each generator:
- Takes an RNG interface
- Returns a string (or formatted output)
- Uses only the RNG for randomness (no direct rand calls)

#### Lorem Ipsum (`internal/nightwatch/fake/lorem.go`)

Store the lorem ipsum vocabulary in `internal/nightwatch/data/lorem.txt` (one word per line, approximately 100 words from the standard lorem ipsum text).

```go
type LoremGenerator struct {
    words []string
}

func (lg *LoremGenerator) Words(n int, rng RNG) string
func (lg *LoremGenerator) Sentences(n int, rng RNG) string
func (lg *LoremGenerator) Paragraphs(n int, rng RNG) string
```

Logic:
- Words: randomly select n words from the corpus
- Sentences: generate 5-15 words, capitalize first, add period
- Paragraphs: generate 4-8 sentences per paragraph

#### Template Engine (`internal/nightwatch/fake/template.go`)

Simple regex-based template parser:

```go
type Template struct {
    raw    string
    tokens []Token
}

type Token struct {
    Type       string   // e.g., "name", "email", "number"
    Args       []string // e.g., ["min", "max"] for number
    IsLiteral  bool
}

func ParseTemplate(s string) (*Template, error)
func (t *Template) Render(rng RNG, generators map[string]Generator) (string, error)
```

Parsing:
- Regex: `{{(\w+)(?::([^}]+))?}}`
- Group 1: type name
- Group 2: optional colon-separated args

Rendering:
- Walk tokens
- For literals, append as-is
- For placeholders, look up generator, call with args, append result

#### Command Handler (`internal/nightwatch/fake.go`)

Cobra command structure:

```go
var fakeCmd = &cobra.Command{
    Use:   "fake",
    Short: "Generate fake data for testing and development",
}

// Subcommands for each data type
var nameCmd = &cobra.Command{Use: "name", ...}
var emailCmd = &cobra.Command{Use: "email", ...}
// ... etc

var templateCmd = &cobra.Command{Use: "template <string>", ...}
```

Common flags:
- `--count` (global, default 1)
- `--seed` (global, optional)
- `--json` (global, boolean)
- Type-specific flags (e.g., `--min/--max` for number)

Flow:
1. Parse flags and arguments
2. Initialize RNG (seeded or crypto)
3. Load name data files (cached)
4. Create appropriate generator(s)
5. Generate `count` items
6. Format output (text or JSON)

#### Data Loading

Names are loaded once at command initialization:

```go
func loadNames() (firstnames, lastnames []string, err error) {
    // Read from internal/nightwatch/data/firstnames.txt
    // Read from internal/nightwatch/data/surnames.txt
    // Cache in memory for the command execution
}
```

### Data Design

#### Name Data Files

Use existing files:
- `internal/nightwatch/data/firstnames.txt`: ~100-200 common first names
- `internal/nightwatch/data/surnames.txt`: ~100-200 common last names

#### Lorem Ipsum Corpus

Create `internal/nightwatch/data/lorem.txt` with the standard lorem ipsum vocabulary (~100 words):

```
lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua ut enim ad minim veniam quis
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat
duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore
eu fugiat nulla pariatur excepteur sint occaecat cupidatat non proident sunt
in culpa qui officia deserunt mollit anim id est laborum
...
```

#### JSON Output Schema

```go
type Output struct {
    Type     string      `json:"type"`
    Seed     *int64      `json:"seed"`
    Count    int         `json:"count"`
    Template string      `json:"template,omitempty"`
    Results  []string    `json:"results"`
}
```

### Generator Implementations

#### Name Generator
- Randomly select from firstnames and lastnames
- Format: `{firstname} {lastname}`

#### Email Generator
- Generate name or use provided name
- Convert to lowercase, replace spaces with dots
- Append `@example.com` or random domain
- Format: `firstname.lastname@example.com`

#### Phone Generator
- Generate 10 random digits
- Format: `(XXX) XXX-XXXX`

#### Address Generator
- Random street number (1-9999)
- Random street name from preset list (Main, Oak, Maple, etc.)
- Random street type (St, Ave, Rd, Ln, Dr)
- Format: `1234 Main St`

#### City Generator
- Random selection from preset list (~50 common US cities)

#### State Generator
- Random selection from US state abbreviations

#### Country Generator
- Random selection from preset list (~20 countries)

#### Zipcode Generator
- Random 5 digits
- Format: `XXXXX`

#### UUID Generator
- Use standard library `crypto/rand` for bytes
- Format as UUID v4 per RFC 4122

#### Hex Generator
- Generate random bytes
- Format as lowercase hex string
- Default length: 16 characters (8 bytes)

#### Number Generator
- Random integer in [min, max]
- Use RNG.Intn for range

#### Date Generator
- Calculate range based on `--past` or `--future` days
- Random day offset
- Format: YYYY-MM-DD

#### DateTime Generator
- Random date + random time
- Format: RFC3339

#### Time Generator
- Random hour (0-23), minute (0-59), second (0-59)
- Format: HH:MM:SS

#### URL Generator
- Random protocol (http/https)
- Random subdomain (www, api, app, etc.)
- Random domain name from preset list
- Random TLD (.com, .org, .net)
- Format: `https://www.example.com`

#### IPv4 Generator
- 4 random octets (0-255)
- Format: `XXX.XXX.XXX.XXX`

#### IPv6 Generator
- 8 random 16-bit groups
- Format: `XXXX:XXXX:XXXX:XXXX:XXXX:XXXX:XXXX:XXXX`

#### MAC Address Generator
- 6 random bytes
- Format: `XX:XX:XX:XX:XX:XX`

#### Username Generator
- Random first name + random number (1-9999)
- Format: `firstname123` or `firstname_lastname`

### Algorithm Design

#### Template Rendering Algorithm

```
1. Parse template string into tokens
2. Validate all placeholder types are known
3. For each record (count times):
   a. Initialize empty result string
   b. For each token:
      - If literal: append to result
      - If placeholder:
        * Look up generator for type
        * Parse arguments if present
        * Call generator with RNG
        * Append generated value to result
   c. Add completed result to results array
4. Return results array
```

#### Seeded Generation Algorithm

```
1. If --seed provided:
   a. Create math/rand.Rand with seed
   b. Use SeededRNG wrapper
2. Else:
   a. Use CryptoRNG wrapper with crypto/rand
3. Pass RNG to all generators
```

## 6. Trade-offs

| Trade-off                       | Gain                           | Sacrifice                               | Justification                                       |
|---------------------------------|--------------------------------|-----------------------------------------|-----------------------------------------------------|
| Custom generators vs library    | No external deps, full control | More implementation work                | Aligns with project goals, manageable scope         |
| Simple template syntax          | Easy to parse, no dependencies | Less powerful than full template engine | Sufficient for use case, keeps it lightweight       |
| Preset data lists vs generation | Realistic output, fast         | Limited variety                         | Can expand lists over time, good enough for testing |

## 7. Cross-Cutting Concerns

### Security

- Use `crypto/rand` for unseeded mode (high-quality randomness)
- No network calls (all data generated locally)
- Clear documentation that this is NOT for password/key generation
- Name data files contain only common names (no PII)

### Performance

- Load name/lorem data files once at startup (cached in memory)
- Streaming output for large `--count` values (print as generated)
- Template parsing is one-time per invocation
- RNG operations are fast (both crypto and math/rand)

### Reliability

- Validate all inputs before generation
- Graceful error handling for missing data files
- Clear error messages for malformed templates
- Exit codes for different error types

### Testing

- Unit tests for each generator (spot checks for format)
- Determinism tests with seeds (same seed = same output)
- Template parsing tests (valid and invalid syntax)
- Integration tests for all commands
- JSON schema validation tests

### Maintainability

- Clear separation: RNG abstraction, generators, template engine, commands
- Each generator is independent (easy to add/modify)
- Data files are plain text (easy to update)
- Consistent patterns with existing code (`decide` command)

## 8. Implementation Plan

### Phase 1: Core Infrastructure (Week 1)
- Create package structure
- Implement RNG abstraction
- Create data files (lorem.txt)
- Implement basic generators (name, email, number, uuid)
- Basic command structure and flag handling

### Phase 2: Text Generators (Week 1)
- Implement lorem ipsum generator (words, sentences, paragraphs)
- Implement word, sentence, paragraph commands
- Add lorem.txt corpus

### Phase 3: Data Type Generators (Week 2)
- Implement remaining generators (phone, address, city, state, country, zipcode)
- Implement date/time generators
- Implement internet generators (url, ipv4, ipv6, mac)
- Implement username generator

### Phase 4: Template Engine (Week 2)
- Implement template parser
- Implement template renderer
- Add template command
- Support parameterized placeholders

### Phase 5: Output Formatting (Week 3)
- Implement JSON output mode
- Handle `--count` for all commands
- Add `--separator` flag
- Format output consistently

### Phase 6: Testing and Documentation (Week 3)
- Unit tests for all generators
- Integration tests for commands
- Determinism tests
- Update documentation
- Add examples

### Migration Strategy

No migration needed - this is a new feature. However:
- Ensure name data files in `internal/nightwatch/data/` are accessible
- Document that `fake` is nightwatch-only (not in sleepless)

## 9. Open Questions

1. **Should we support custom data files?** (e.g., `--names-file custom-names.txt`)
   - Status: Non-blocking
   - Recommendation: Add in future if needed

2. **Should we add more localization options?** (e.g., `--locale en_GB`)
   - Status: Non-blocking
   - Recommendation: Out of scope for initial implementation

3. **Should template mode support escaping?** (e.g., `\{{not_a_placeholder}}`)
   - Status: Non-blocking
   - Recommendation: Yes, add support for `\{{` and `\}}`

4. **Should we cache generated data for unique constraints?** (e.g., generate 100 unique emails)
   - Status: Non-blocking
   - Recommendation: Add `--unique` flag in future if needed
