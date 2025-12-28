# Design: decide (random choice utilities)
**Specification Reference**: [specification.md](specification.md)

**Status**: Draft

## 1. Context

Sleepless is implemented as a Cobra CLI. Adding a small command group that does not interact with the task storage layer is low risk. The main technical decision is the randomness source and how to support both reproducibility (`--seed`) and interactive randomness.

Go provides `math/rand` for fast deterministic pseudo-randomness, and `crypto/rand` for cryptographically secure randomness. The feature is not intended for cryptographic use, but it benefits from good default randomness and deterministic behavior when requested.

## 2. Goals and Non-Goals

### Goals

- Implement the CLI modes defined in the specification.
- Provide deterministic behavior when `--seed` is used.
- Provide a stable JSON schema.
- Keep dependencies minimal.

### Non-Goals

- Cryptographic guarantees (password generation, key material).
- Advanced probability distributions beyond the defined commands.

## 3. Options Considered

### Option 1: Always use `math/rand` (time-seeded by default)

**Advantages**:
- Simple implementation.
- Fast.

**Disadvantages**:
- Weak randomness; in some environments time seeding can be predictable.

**Complexity**: Low

### Option 2: Use `crypto/rand` by default; use `math/rand` only when seeded

**Advantages**:
- High-quality randomness by default.
- Determinism is explicit and isolated to seeded mode.

**Disadvantages**:
- Slightly more code; `crypto/rand` needs helpers for ranges.

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 2

**Rationale**: It provides the best interactive defaults while still supporting reproducibility. The additional code is contained and does not require new third-party dependencies.

**Key Factors**:
1. Better default randomness without extra dependencies.
2. Seeded mode maps naturally to `math/rand`.

## 5. Detailed Design

### Architecture Overview

- `cmd/decide.go` (new): root command.
- `cmd/decide_*.go` (new): optional subcommand files.
- `pkg/decide/` (new): RNG abstraction + implementations + algorithms.

### Component Design

#### RNG Abstraction

Define an internal interface such as:

- `Intn(n int) (int, error)`

Implementations:

- `CryptoRNG`: uses `crypto/rand` for unseeded mode.
- `SeededRNG`: wraps `math/rand.Rand` for seeded mode.

#### Algorithms

- `pickOne`: choose index in `[0, len-1]`.
- `pickN`: Fisher-Yates shuffle indices then slice first `n`.
- `shuffle`: Fisher-Yates shuffle.
- `roll`: parse `NdM`, produce `N` uniform ints in `[1, M]` and compute total.
- `weighted`: parse `name:weight`, compute cumulative sum, sample in `[1, total]`.

### API / CLI Design

The Cobra command layer validates arguments and delegates to pkg functions that return typed results suitable for either text or JSON output.

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| crypto/rand default | Higher-quality randomness | More code for ranges | Better UX defaults, still deterministic with seed |

## 7. Cross-Cutting Concerns

### Security

- Avoid claiming cryptographic suitability.
- Ensure `--seed` is the only way to force deterministic output.

### Performance

- All operations are small; performance constraints are negligible.

### Reliability

- Ensure correct error returns for invalid inputs.

### Testing

- Seeded mode enables deterministic test vectors.

## 8. Implementation Plan

- Phase 1: Cobra scaffolding + `pickOne`, `coin`, `number`.
- Phase 2: `pickN`, `shuffle`.
- Phase 3: `roll` parsing and output.
- Phase 4: `weighted`.
- Phase 5: file input (`--from`) and `--count`.
- Phase 6: JSON output + tests.

### Migration Strategy

None.

## 9. Open Questions

- Should `--from` default to trimming whitespace and allowing comments (as specified), or should that be configurable?
- Should `weighted` support floats, or keep integers only (current spec: integers only)?
