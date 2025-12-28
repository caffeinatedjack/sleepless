# Implementation: when (world clock and time conversion)

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Implement a new Cobra command group `when` plus supporting package code for time parsing, zone resolution, configuration persistence, and output formatting (text and JSON).

## Prerequisites

- Decide config file location and schema (design proposes XDG config JSON).
- Confirm desired initial set of built-in abbreviations/city tokens.

## Phases

### Phase 1: Command skeleton + basic world clock

**Goal**: `regimen when` prints the current time in local zone and configured zones.

- [x] Add `internal/regimen/when.go` and register it with root.
- [x] Define output formatting for the default view.
- [x] Add minimal config loading (even if empty).

**Milestone**: `regimen when` runs with exit code 0 and prints stable output.

### Phase 2: Time parsing + conversions

**Goal**: Support `regimen when <time-expr> [in <zone>] [to <zone>]`.

- [x] Implement time expression parsing (`now`, `3pm`, `3:30pm`, `17:00`).
- [x] Implement zone resolution (IANA + minimal abbreviation/city mapping).
- [x] Implement direct conversion output for `to <zone>`.

**Milestone**: Conversions match expected results for a set of fixed test cases.

### Phase 3: Alias/configured zone management

**Goal**: Support `add/remove` and persist aliases + configured zones.

- [x] Implement config file read/write with user-only permissions.
- [x] `regimen when add <alias> <zone>`.
- [x] `regimen when remove <alias>`.

**Milestone**: Aliases persist across runs and affect zone resolution.

### Phase 4: Zone listing

**Goal**: Implement `regimen when zones [query]`.

- [x] Enumerate system zoneinfo directory.
- [x] Add substring filter.
- [x] Add error message for enumeration failures.

**Milestone**: `zones` prints a sorted list and `zones europe` filters correctly.

### Phase 5: Utilities

**Goal**: Implement `diff`, `until`, `overlap`.

- [x] Implement `diff <a> <b>`.
- [x] Implement `until <time-expr>`.
- [x] Implement `overlap <a> <b>` using default 09:00-17:00.

**Milestone**: Utility outputs match expected behavior in unit tests.

### Phase 6: JSON output + polish

**Goal**: Add `--json`, `--short`, `--date`, `--12h/--24h`.

- [x] Implement JSON schemas per spec.
- [x] Ensure flags apply consistently across subcommands.
- [x] Add unit tests for time parsing.

**Milestone**: JSON output validates against a documented schema and is stable.

## Testing Plan

### Unit Tests

- Time parsing table tests.
- Zone resolution tests (alias/IANA/abbrev/city).
- DST conversion tests using fixed timestamps.
- Overlap intersection logic tests.

### Integration Tests

- Golden output tests for `--json`.
- CLI error code tests for invalid zones/time expressions.

## Rollback Plan

- Remove the `when` command registration and associated package; no data migrations.

## Open Questions

- Confirm whether `--24h` default is acceptable or should follow locale.
- Confirm how to represent date differences in text output when `--date` is not set.
