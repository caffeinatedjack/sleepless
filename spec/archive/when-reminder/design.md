# Design: when (world clock and time conversion)
**Specification Reference**: [specification.md](specification.md)

**Status**: Draft

## 1. Context

Sleepless currently focuses on goal/task management, but it is structured as a Cobra-based CLI with a modular `cmd/` layout and small helper packages. Adding additional top-level utility commands is architecturally straightforward (new Cobra command + package helpers) as long as they remain local-first and keep output predictable.

Timezone conversion in Go can rely on the standard library `time` package, which supports loading IANA locations via `time.LoadLocation`. The main constraints are (a) enumerating available IANA zones for `when zones`, and (b) mapping user-friendly tokens (city names, abbreviations, aliases) to IANA zones.

This design also must preserve the "CLI is scriptable" expectation: stable exit codes, JSON output, and deterministic behavior when flags are used.

## 2. Goals and Non-Goals

### Goals

- Implement `regimen when` command group as defined in the specification.
- Provide a clear zone-resolution pipeline (alias → IANA → abbreviation → city).
- Provide JSON output with a stable schema.
- Ensure DST and date-crossing conversions behave correctly.

### Non-Goals

- Implementing a complete natural-language date parser (e.g., "next Tuesday at 3").
- Network lookups (e.g., geocoding cities).
- Supporting every possible timezone abbreviation used worldwide.

## 3. Options Considered

### Option 1: Ship a static IANA zone list

Embed a generated list of zones into the binary (or store it in the repo).

**Advantages**:
- `when zones` works consistently across platforms.
- No dependency on host filesystem layout.

**Disadvantages**:
- Requires regeneration when tzdata updates.
- Slight binary size increase.

**Complexity**: Medium

### Option 2: Enumerate zones from the system zoneinfo directory

Walk the zoneinfo directory (e.g., `/usr/share/zoneinfo`, respecting `ZONEINFO` env if set), collect zone names, and filter.

**Advantages**:
- Always uses host tzdata.
- No embedded assets.

**Disadvantages**:
- Platform- and distro-dependent paths.
- Some environments may not have zoneinfo files.

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 2 (enumerate zones from system zoneinfo)

**Rationale**: The project is already local-first and Go’s `time` behavior depends on system tzdata in many deployments. Enumerating zones from zoneinfo provides a good-enough UX while avoiding the maintenance overhead of embedding tzdata or a generated list.

**Key Factors**:
1. No new build-time generation step.
2. Keeps dependencies minimal.
3. Matches the system’s timezone capabilities.

## 5. Detailed Design

### Architecture Overview

- `internal/regimen/when.go` (new): root for the `when` command group.
- `internal/regimen/when_*.go` (new): subcommands (`add`, `remove`, `zones`, `diff`, `until`, `overlap`).
- `pkg/when/` (new): parsing, zone resolution, conversion, overlap math.
- Storage: a small config file for aliases/configured zones.

### Component Design

#### Zone Resolver

A resolver component accepts a string token and returns a resolved IANA zone string + `*time.Location`.

- Alias resolution uses the on-disk config mapping.
- IANA resolution checks `time.LoadLocation(token)`.
- Abbreviation and city-token resolution use built-in maps.

Ambiguity: city tokens should map to exactly one IANA zone. If a token maps to multiple zones, resolver returns an explicit ambiguity error with candidates.

#### Config Storage

Store `when` configuration separately from tasks storage to avoid coupling. Proposed format:

```json
{
  "aliases": {"work": "America/New_York"},
  "configured": ["work", "Europe/London", "UTC"]
}
```

Location SHOULD follow XDG (`$XDG_CONFIG_HOME/regimen/when.json`) with a fallback to `~/.config/regimen/when.json`.

### Data Design

No changes to task markdown storage.

### API / CLI Design

Cobra commands mirror the spec. Output formatting uses a shared formatter that accepts:

- resolved reference time
- list of target zones
- format options (`short`, `date`, `12h/24h`, `json`)

### Algorithm Notes

- Conversions: create a `time.Time` in origin location, then format in destination locations.
- `until`: compute next occurrence in local zone by constructing time on today’s date; if <= now, add 24h and re-normalize in local location.
- `overlap`: compute working-hour intervals in both zones for a chosen base date, convert both to a common reference (UTC), intersect intervals, then display in both zones.

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| System zoneinfo enumeration | No embedded assets | Platform variability | Keeps maintenance low and matches host tzdata |
| Limited time parsing | Simpler UX + fewer bugs | Less natural-language support | Keeps feature predictable and testable |

## 7. Cross-Cutting Concerns

### Security

- No network.
- Config file permissions should be user-only.

### Performance

- Zone enumeration may be slow; cache the enumerated list for the process lifetime.

### Reliability

- If enumeration fails, `when zones` errors but other `when` functionality still works.

### Testing

- Unit tests for parsing and zone resolution.
- DST-focused tests using fixed reference times and known zones.

## 8. Implementation Plan

- Phase 1: Add command skeleton + `regimen when` display for configured zones.
- Phase 2: Time parsing and `in`/`to` conversion.
- Phase 3: Config storage for `add/remove` and alias resolution.
- Phase 4: Zone enumeration and `zones` filtering.
- Phase 5: Utilities: `diff`, `until`, `overlap`.
- Phase 6: JSON output and integration tests.

### Migration Strategy

None; new command group.

## 9. Open Questions

- Should configured zones default to an empty list, a small built-in set (e.g., `UTC`), or a list derived from environment?
- How large should the abbreviation/city-token mapping be (minimal set vs. expanded list)?
