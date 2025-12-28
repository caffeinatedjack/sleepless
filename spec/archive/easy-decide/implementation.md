# Implementation: decide (random choice utilities)

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Add a new `decide` Cobra command group plus an internal package that implements random selection operations with a default cryptographic RNG and deterministic seeded mode.

## Prerequisites

- Confirm whether file parsing should treat `#` lines as comments (spec says MAY).

## Phases

### Phase 1: CLI scaffold + basic modes

**Goal**: Provide `sleepless decide <options...>`, `coin`, and `number`.

- [ ] Add `cmd/decide.go` and register with root.
- [ ] Implement RNG abstraction and crypto default.
- [ ] Implement `pickOne`.
- [ ] Implement `coin`.
- [ ] Implement `number <min> <max>`.

**Milestone**: Commands work end-to-end with correct exit codes.

### Phase 2: Shuffle + pick N

**Goal**: Implement `shuffle` and `pick`.

- [ ] Implement Fisher-Yates shuffle.
- [ ] Implement `pick <n>` without replacement.
- [ ] Add validation for `n` bounds.

**Milestone**: Seeded mode produces deterministic outputs for tests.

### Phase 3: Dice

**Goal**: Implement `roll NdM`.

- [ ] Parse dice notation.
- [ ] Generate rolls and totals.
- [ ] Add error cases for invalid notation.

**Milestone**: Unit tests cover parsing and result ranges.

### Phase 4: Weighted

**Goal**: Implement weighted selection.

- [ ] Parse `option:weight` arguments.
- [ ] Validate positive integer weights.
- [ ] Implement cumulative sampling.

**Milestone**: Weighted selection chooses only provided options and rejects invalid weights.

### Phase 5: `--from` and `--count`

**Goal**: Support reading options from file and repeating operations.

- [ ] Implement file reading and parsing.
- [ ] Implement `--count` loops.
- [ ] Ensure `--from` and positional options conflict is an error.

**Milestone**: Integration tests cover repeated runs.

### Phase 6: JSON output

**Goal**: Support `--json` across all modes.

- [ ] Implement typed result structs.
- [ ] Implement JSON serializer and schema stability.
- [ ] Add golden tests.

**Milestone**: JSON output matches spec for all modes.

## Testing Plan

### Unit Tests

- Deterministic test vectors with `--seed`.
- Bounds tests for `number` and dice.
- Parsing tests for weighted inputs.

### Integration Tests

- CLI argument parsing and exit codes.
- `--from` file input.
- `--json` schema.

## Rollback Plan

- Remove `decide` command registration and pkg; no data migrations.

## Open Questions

- Decide whether comment lines (`# ...`) are supported in `--from` inputs (and document behavior).
