# regimen specification

This document consolidates the completed specifications for the **`regimen`** executable.

It replaces the prior per-feature standalone documents under `spec/section/` for regimen.

## Contents

- [goals (core goal/task management)](#goals-core-goaltask-management)
- [decide (random choice utilities)](#decide-random-choice-utilities)
- [when (world clock and time conversion)](#when-world-clock-and-time-conversion)
- [recipes (vimwiki recipe management cli)](#recipes-vimwiki-recipe-management-cli)

---

## goals (core goal/task management)

> Note: This section was originally specified as `sleepless goals`. In this repository, the
> implemented executable and command surface is **`regimen goals`**.

**Depends on**: none

### Abstract

This specification defines the existing `regimen goals` command group for managing goals and subtasks stored as human-editable Markdown in a local wiki directory. It specifies the task data model, storage format, and the CLI commands for creating, listing, editing, completing, moving, searching, viewing, archiving, and auditing history.

### 1. Introduction

Regimen is a terminal-first goal/task manager designed to integrate with a Markdown-based personal wiki. The core workflow is centered on quickly adding tasks, organizing them by topic, and keeping storage in a format that is both scriptable and manually editable.

The system stores tasks in topic-based Markdown files and uses stable IDs for referencing tasks from the CLI. This specification documents the current behavior of the existing implementation so it can be treated as complete and used as a baseline for future proposals.

### 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

### 3. Terminology

`Goal`: A top-level task.

`Subtask`: A task nested under a goal.

`Topic`: A category name. Topics map 1:1 to Markdown files.

`Task ID`: A 32-character lowercase hexadecimal identifier.

`Short ID`: The first 6 characters of a Task ID, used for display and prefix lookup.

`Archive`: The special topic `archived` stored in `archived.md`.

`Task Store`: The on-disk directory containing topic Markdown files, `index.md`, `archived.md`, and `.task-meta.json`.

### 4. Concepts

#### 4.1. Local-first Markdown Storage

Tasks are stored in a per-topic Markdown file named `<topic>.md` under the Task Store directory. Tasks are intended to be editable by humans (in an editor) and by the CLI.

#### 4.2. ID Prefix Resolution

The CLI resolves a user-provided task identifier by prefix matching against Task IDs.

- If exactly one task matches the prefix, it is selected.
- If zero tasks match, the operation does not proceed.
- If multiple tasks match, the operation does not proceed and the ambiguity is reported.

#### 4.3. Topic Index

The Task Store maintains an `index.md` file containing links to topics. When a new topic file is created, the index is updated to include it (excluding `inbox` and `archived`).

#### 4.4. History Log

The Task Store maintains `.task-meta.json` containing a bounded history list of operations (save/remove/move/archive). This provides auditability and basic change tracking.

### 5. Requirements

#### 5.1. CLI Command Group

1. The system MUST provide a top-level command group `regimen goals`.
2. The `regimen goals` command group MUST ensure the Task Store directory structure exists before executing any subcommand.

#### 5.2. Task Store Location

1. The default Task Store directory MUST be `~/wiki/tasks`.

#### 5.3. Task Data Model

A task MUST include:

- `id` (Task ID): 32-character lowercase hex string
- `title`: non-empty string
- `status`: `open` or `complete`
- `priority`: `low`, `medium`, or `high`
- `topic`: string (defaults to `inbox`)

A task MAY include:

- `due`: date (`YYYY-MM-DD`)
- `tags`: list of strings
- `notes`: list of strings
- `created`: RFC3339 timestamp
- `completed`: RFC3339 timestamp
- `subtasks`: list of child tasks

#### 5.4. Storage Format

1. Each topic MUST be stored as a Markdown file `/<topic>.md` within the Task Store.
2. Each topic file MUST begin with a level-1 heading representing the topic title.
3. Each task MUST be represented by a checkbox list item with an embedded `{#id}` token.
4. Task metadata MUST be represented as indented bullet items under the task.

#### 5.5. Core Commands

The following commands MUST exist:

- `regimen goals add <title>`
- `regimen goals list`
- `regimen goals done <goal-id>`
- `regimen goals edit <goal-id>`
- `regimen goals remove <goal-id>`
- `regimen goals move <goal-id> <new-topic>`
- `regimen goals search <query>`
- `regimen goals remind`
- `regimen goals archive [goal-id]`
- `regimen goals history [goal-id]`
- `regimen goals view <view>` where `<view>` is one of `tree`, `progress`, `deps`, `calendar`

#### 5.6. Inbox Aging Warning

1. The system SHOULD warn when there are open tasks in topic `inbox` with `created` older than 7 days.
2. The warning SHOULD show up to 5 example tasks.

### 6. Interface

#### 6.1. Command: add

```bash
regimen goals add <title> [--topic <topic>] [--priority low|medium|high] [--due YYYY-MM-DD] [--tags a,b,c] [--parent <goal-id>]
```

- `--topic` default MUST be `inbox`.
- `--priority` default MUST be `medium`.
- If `--parent` is provided, the new task MUST be added as a subtask of the resolved parent.

#### 6.2. Command: list

```bash
regimen goals list [--topic <topic>] [--priority low|medium|high] [--status open|complete] [--overdue] [--tags <tag|tag,tag>]
```

#### 6.3. Command: done

```bash
regimen goals done <goal-id> [--auto-archive]
```

#### 6.4. Command: edit

```bash
regimen goals edit <goal-id> [--title <title>] [--priority low|medium|high] [--due YYYY-MM-DD] [--tags a,b,c] [--note <text>]
```

#### 6.5. Command: remove

```bash
regimen goals remove <goal-id> [--force]
```

#### 6.6. Command: move

```bash
regimen goals move <goal-id> <new-topic>
```

#### 6.7. Command: search

```bash
regimen goals search <query> [--topic <topic>] [--status open|complete]
```

#### 6.8. Command: remind

```bash
regimen goals remind [--days <n>] [--install]
```

- `--days` default MUST be 7.
- `--install` MUST print a cron snippet that runs reminders daily at 9 AM.

#### 6.9. Command: archive

```bash
regimen goals archive [goal-id]
```

- If `goal-id` is provided, the specified goal MUST be archived if complete.
- If `goal-id` is omitted, all completed goals not already in `archived` MUST be archived.

#### 6.10. Command: history

```bash
regimen goals history [goal-id] [--limit <n>]
```

- `--limit` default MUST be 20.
- If `goal-id` is provided, history MUST be filtered by Task ID prefix.

#### 6.11. Command: view

```bash
regimen goals view tree [--topic <topic>] [--expand-completed]
regimen goals view progress
regimen goals view deps
regimen goals view calendar [--days <n>]
```

- `view calendar --days` default MUST be 14.

### 7. Behavior

#### 7.1. Completion

1. `done` MUST set the task status to `complete`.
2. If the task has incomplete subtasks, the system MUST prompt for confirmation before completing those subtasks.
3. If `--auto-archive` is set, the system MUST archive the task after completion.

#### 7.2. Archiving

1. Archiving MUST remove a completed task from its current topic file.
2. Archiving MUST add the task to the top of `archived.md`.
3. After archiving, the task topic MUST be `archived`.

#### 7.3. Search

1. Search MUST match query substrings case-insensitively.
2. Search MUST search across title, tags, and notes.
3. Search MUST include subtasks.

#### 7.4. Views

- `tree` MUST display goals grouped by topic and show nested subtasks.
- `progress` MUST display progress summaries by topic and by priority.
- `deps` MUST visualize parent-child relationships.
- `calendar` MUST list overdue tasks and upcoming due dates.

### 8. Error Handling

1. Invalid due date values (not `YYYY-MM-DD`) MUST be rejected with an error message and MUST NOT modify stored data.
2. If an ID prefix resolves to multiple tasks, the command MUST report all matches and MUST NOT modify stored data.
3. If an ID prefix resolves to zero tasks, the command MUST report "No task found" and MUST NOT modify stored data.
4. If the Task Store cannot be read or written, the command MUST report the failure.

### 9. Examples

```bash
# Add goals
regimen goals add "Buy groceries"
regimen goals add "Finish report" --topic work --priority high --due 2025-01-15 --tags urgent,writing

# Add subtask
regimen goals add "Review section 1" --parent a1b2c3

# List
regimen goals list
regimen goals list --topic work --priority high
regimen goals list --overdue

# Complete and archive
regimen goals done a1b2c3 --auto-archive

# Edit
regimen goals edit a1b2c3 --note "Remember to check X"

# Move
regimen goals move a1b2c3 home

# Search
regimen goals search "urgent" --topic work

# Reminders
regimen goals remind --days 14
regimen goals remind --install

# Views
regimen goals view tree --topic work
regimen goals view progress
regimen goals view calendar --days 30

# History
regimen goals history --limit 50
regimen goals history a1b2c3
```

### 10. Security Considerations

1. The system MUST store all task data locally and MUST NOT require network access.
2. Because tasks may include sensitive personal/work information, the system SHOULD avoid printing file paths or raw file contents in error messages unless necessary.
3. The `.task-meta.json` history MAY reveal task titles; users SHOULD treat the Task Store as sensitive data.

### 11. Testing Considerations

- Add/list/edit/done/remove/move flows across multiple topics.
- Prefix resolution: unique, ambiguous, and no-match.
- Markdown round-trip: parse then write preserves IDs, nesting, and metadata.
- Archive: task is removed from topic and prepended to archive.
- Reminders: overdue and upcoming logic with date boundaries.

### 12. References

[VIMWIKI] Vimwiki project, https://github.com/vimwiki/vimwiki

---

## decide (random choice utilities)

> Note: This section was originally specified as `sleepless decide`. In this repository, the
> implemented executable and command surface is **`regimen decide`**.

**Depends on**: none

### Abstract

This specification adds a `regimen decide` command group for random choice utilities to reduce decision paralysis. It defines CLI modes for choosing items, picking N unique items, shuffling, dice rolling, coin flips, random numbers, and weighted selection. It also defines reproducibility controls (`--seed`) and machine-readable output (`--json`).

### 1. Introduction

Users often need small, quick randomness utilities from the terminal: choosing between options, rolling dice, shuffling lists, or generating a random number in a range.

While many tools exist, they vary across platforms and often have inconsistent syntax. Adding a small, consistent command group to Regimen improves usability and scripting, and keeps common utilities in one CLI.

The proposed interface is centered around `regimen decide` with clear subcommands and flags, with default behavior of picking a single option.

### 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

### 3. Terminology

`Option`: A single candidate value (string) to be selected or permuted.

`Run`: One execution of a selection operation that yields one or more outputs.

`Seeded Mode`: Random generation performed using a user-provided seed such that output is deterministic for the same inputs.

`Unseeded Mode`: Random generation performed without a user-provided seed.

### 4. Concepts

#### 4.1. Input Sources

Options may be provided either:

- As positional arguments, or
- Via `--from <file>`.

If both are provided, the system MUST treat this as an error (to avoid accidental mixing).

#### 4.2. Repetition

Some operations may be repeated using `--count N`. Each repetition is a separate `Run`.

### 5. Requirements

#### 5.1. Command Group

1. The CLI MUST provide a top-level command `regimen decide`.

#### 5.2. Modes

1. `regimen decide <options...>` MUST select exactly one option uniformly at random.
2. `regimen decide pick <n> <options...>` MUST select `n` unique options uniformly without replacement.
3. `regimen decide shuffle <options...>` MUST output the input options in a uniformly random order.
4. `regimen decide roll <dice>` MUST support dice notation `NdM` where `N` and `M` are positive integers.
5. `regimen decide coin` MUST output one of `heads` or `tails`.
6. `regimen decide number <min> <max>` MUST output a uniformly random integer in the inclusive range `[min, max]`.
7. `regimen decide weighted <option:weight...>` MUST select one option where selection probability is proportional to weight.

#### 5.3. Flags

1. `--from <path>` MUST read options from a UTF-8 text file.
2. `--count <n>` MUST repeat the operation `n` times.
3. `--seed <int>` MUST enable Seeded Mode.
4. `--json` MUST output machine-readable JSON.

#### 5.4. File Parsing

1. When `--from` is provided, the system MUST treat each non-empty line as one option.
2. Leading/trailing whitespace on each line SHOULD be trimmed.
3. Lines beginning with `#` MAY be treated as comments.

### 6. Interface

#### 6.1. Commands

```bash
regimen decide <options...>
regimen decide pick <n> <options...>
regimen decide shuffle <options...>
regimen decide roll <NdM>
regimen decide coin
regimen decide number <min> <max>
regimen decide weighted "a:3" "b:1"
```

#### 6.2. Flags

- `--from <file>`
- `--count <n>`
- `--seed <int>`
- `--json`

#### 6.3. JSON Output

When `--json` is provided, stdout MUST contain a single JSON object with:

- `mode`: one of `pickOne`, `pickN`, `shuffle`, `roll`, `coin`, `number`, `weighted`
- `seed`: integer or null
- `count`: integer
- `results`: array

Each result element MUST be:

- for `pickOne`/`weighted`/`coin`: string
- for `pickN`/`shuffle`: array of strings
- for `number`: integer
- for `roll`: object with `notation`, `rolls` (array of ints), `total` (int)

### 7. Behavior

1. In Seeded Mode, the system MUST produce the same results for the same inputs, mode, and seed.
2. In Unseeded Mode, the system SHOULD use a randomness source appropriate for interactive use.
3. `pick <n>` MUST error if `n` is less than 1 or greater than the number of available options.
4. `number <min> <max>` MUST error if `min > max`.
5. For `weighted`, weights MUST be positive integers. Zero or negative weights MUST error.

### 8. Error Handling

1. Missing required arguments MUST produce a non-zero exit code and usage help.
2. Invalid integers (e.g., `n`, `seed`, `min/max`, weights) MUST produce a non-zero exit code.
3. Invalid dice notation MUST produce a non-zero exit code.
4. Empty option sets MUST produce a non-zero exit code.

### 9. Examples

```bash
regimen decide pizza tacos sushi
regimen decide pick 2 alice bob charlie
regimen decide shuffle a b c d e
regimen decide roll 2d6
regimen decide coin
regimen decide number 1 100
regimen decide weighted "a:3" "b:1"

# From file
regimen decide --from options.txt
regimen decide pick 3 --from people.txt

# Reproducible
regimen decide --seed 42 pizza tacos sushi
regimen decide pick 2 --seed 42 alice bob charlie
```

### 10. Security Considerations

1. The `decide` command group MUST NOT be represented as a secure randomness source for cryptographic purposes.
2. If Unseeded Mode uses a non-cryptographic RNG, documentation and help text SHOULD avoid implying security.

### 11. Testing Considerations

- Determinism tests for seeded mode.
- Distribution sanity checks (basic statistical tests) for uniform modes.
- Dice parsing table tests.
- File parsing tests including comments/blank lines.
- Error code tests for invalid inputs.

### 12. References

[REPRO] "Reproducible random number generation" (general background), https://en.wikipedia.org/wiki/Random_seed

---

## when (world clock and time conversion)

**Depends on**: none

### Abstract

This specification adds a new `regimen when` command group that provides timezone-aware world clock output, time conversion between zones, and small timezone utilities (difference, countdown, overlap). It defines CLI inputs, supported timezone identifiers, output formats (including JSON), and error handling so the feature is testable and script-friendly.

### 1. Introduction

Regimen is a CLI-first tool. Users frequently coordinate across multiple time zones and need a quick way to (a) see the current time in a set of commonly used zones, and (b) convert a time expressed in one zone into another.

Existing solutions (web sites, OS widgets, ad-hoc `date` invocations) are either not scriptable, not cross-platform in the same way, or do not integrate with a consistent CLI UX. This feature adds a single, predictable command surface in Regimen.

The solution is a `when` command group that supports displaying the current time in configured zones, converting a provided time with optional `in`/`to` clauses, and managing user-defined zone aliases.

### 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

### 3. Terminology

`Zone`: A timezone identifier that can be resolved to a location.

`IANA Zone`: A timezone in IANA/Olson form (e.g., `America/New_York`).

`Abbreviation`: A short timezone identifier (e.g., `UTC`, `GMT`, `PST`) that maps to a `Zone`.

`City Token`: A user-friendly name (e.g., `London`, `Tokyo`, `NYC`) that maps to a `Zone`.

`Alias`: A user-defined name (e.g., `work`, `home`) that maps to a `Zone`.

`Configured Zone`: A `Zone` that the user has added to the `when` configuration for display in `regimen when`.

`Local Zone`: The system local timezone.

### 4. Concepts

#### 4.1. Zone Resolution

The system resolves a zone token in the following order:

1. Alias
2. IANA Zone
3. Abbreviation
4. City Token

If resolution fails, the token is treated as invalid.

If a token can resolve to multiple zones (e.g., ambiguous city tokens), the system MUST return an error describing the ambiguity and MUST NOT guess.

#### 4.2. Time Expression

A time expression is one of:

- `now`
- A 12-hour time such as `3pm`, `3:30pm`
- A 24-hour time such as `17:00`, `05:30`

A time expression MAY be combined with a date (via a flag; see Interface).

### 5. Requirements

#### 5.1. Core Command Group

1. The CLI MUST provide a top-level command `regimen when`.
2. `regimen when` (with no positional arguments) MUST display the current local time and the current time in each configured zone.
3. `regimen when <time-expr>` MUST interpret `<time-expr>` in the local zone and MUST display the equivalent time in each configured zone.

#### 5.2. Conversion Clauses

1. `regimen when <time-expr> in <zone>` MUST interpret `<time-expr>` in `<zone>` and MUST display the equivalent time in each configured zone.
2. `regimen when <time-expr> in <zone> to <zone>` MUST output only the direct conversion result between the origin and destination zones.

#### 5.3. Zone Management

1. `regimen when add <alias> <zone>` MUST add or update an alias mapping and MUST include the aliased zone in the configured-zone set.
2. `regimen when remove <alias>` MUST remove the alias mapping.
3. `regimen when zones` MUST list available IANA zones.
4. `regimen when zones <query>` MUST filter the available IANA zone list by substring match on the zone name.

#### 5.4. Utilities

1. `regimen when diff <zone-a> <zone-b>` MUST output a human-readable difference such as "`<zone-a>` is 5 hours ahead of `<zone-b>`".
2. `regimen when until <time-expr>` MUST output the duration until the next occurrence of `<time-expr>` in the local zone.
3. `regimen when overlap <zone-a> <zone-b>` MUST output at least one overlapping window of working hours between the two zones.

Working hours MUST default to 09:00-17:00 local time in each respective zone. The CLI MAY later add flags to customize this window.

#### 5.5. Output Formats

1. The command group MUST support a `--json` flag for machine-readable output.
2. The command group MUST support a `--short` flag that produces a compact single-line output.
3. The command group MUST support a time format selection (`--24h` and `--12h`). If both are omitted, the default MUST be `--24h`.
4. The command group MUST support `--date` to include the date component in formatted times.

### 6. Interface

#### 6.1. Commands

The interface uses the following forms (whitespace is significant; tokens in angle brackets are positional arguments):

```bash
regimen when
regimen when <time-expr>
regimen when <time-expr> in <zone>
regimen when <time-expr> in <zone> to <zone>

regimen when add <alias> <zone>
regimen when remove <alias>

regimen when zones [query]

regimen when diff <zone-a> <zone-b>
regimen when until <time-expr>
regimen when overlap <zone-a> <zone-b>
```

#### 6.2. Flags

- `--json` (boolean): Output JSON.
- `--short` (boolean): Compact text output.
- `--24h` (boolean): Format times in 24-hour form.
- `--12h` (boolean): Format times in 12-hour form.
- `--date` (boolean): Include date.

#### 6.3. JSON Output (normative)

When `--json` is provided, the command MUST output a single JSON object to stdout.

For `regimen when` and `regimen when <time-expr> [...]` the object MUST contain:

- `reference`: an object with `time` (RFC3339) and `zone` (resolved zone string)
- `zones`: an array of objects with:
  - `label` (string): the display label (alias if used, else the resolved zone string)
  - `zone` (string): resolved zone string
  - `time` (RFC3339)

For single-result conversions (`... to <zone>`) the object MUST contain `from` and `to` objects each with `zone` and `time`.

### 7. Behavior

1. The system MUST use the system local timezone as `Local Zone`.
2. If `--date` is not set, text output MUST still compute conversions with full date context (e.g., crossing midnight) and MUST clearly indicate date differences when they occur (e.g., `(+1d)` or by printing the date when different from the reference date).
3. For `until`, the system MUST compute until the next occurrence of the provided time in the local zone (i.e., if the time has already passed today, it uses tomorrow).
4. For `overlap`, the system MUST compute overlap based on the same calendar date in each respective zone, and MUST indicate if overlap occurs on different local dates.

### 8. Error Handling

1. Invalid time expressions MUST produce a non-zero exit code and a message including the invalid token.
2. Unknown or unresolvable zones MUST produce a non-zero exit code and a message including the token.
3. Ambiguous city tokens MUST produce a non-zero exit code and MUST list at least two possible matches.
4. For `zones`, if the system cannot enumerate zones, it MUST return a non-zero exit code and a message describing the limitation.

### 9. Examples

```bash
# Current time across configured zones
regimen when

# Convert local 3pm to all zones
regimen when 3pm

# Convert 3pm London to all zones
regimen when 3pm in London

# Direct conversion
regimen when 3pm in London to Tokyo

# Manage zones
regimen when add work America/New_York
regimen when remove work

# List and search zones
regimen when zones
regimen when zones europe

# Utilities
regimen when diff London NYC
regimen when until 17:00
regimen when overlap London Tokyo
```

### 10. Security Considerations

1. The `when` command group MUST NOT transmit data over the network.
2. If configuration is stored on disk, the system SHOULD store it with user-only permissions (e.g., not world-readable) because aliases may reveal travel/work patterns.

### 11. Testing Considerations

Test scenarios SHOULD include:

- Zone resolution for IANA, abbreviation, city token, and alias
- DST boundaries (spring forward / fall back) for `in` and `to` conversions
- Conversions crossing midnight and date display behavior
- `until` when target time is before/after current time
- `overlap` for non-overlapping zones and near-24h-offset zones
- JSON schema validation

### 12. References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997, https://www.rfc-editor.org/rfc/rfc2119

[IANA-TZ] IANA, "Time Zone Database", https://www.iana.org/time-zones

---

## recipes (vimwiki recipe management cli)

**Depends on**: none

### Abstract

This specification adds recipe management capabilities to the `regimen` executable by defining file-backed recipes stored as Markdown files in a Vimwiki-compatible directory (default `~/wiki/recipes`). It defines a CLI interface for listing/viewing/searching/creating recipes, category support via a wiki index file (`~/wiki/Recipes.md`), and a shopping list generator that consolidates ingredient entries from one or more recipes.

### 1. Introduction

Regimen focuses on managing goals in a Vimwiki-backed Markdown layout. Recipes in the same wiki are typically managed by manually creating and editing files, which makes quick retrieval (list/search/category browsing) awkward from the terminal.

This specification defines a CLI surface in `regimen` for basic recipe operations while keeping recipes as plain Markdown files on disk. It also defines a category/index mechanism driven by an existing wiki page (`Recipes.md`) and a shopping list generator that extracts ingredient bullet lines, enabling a “plan meals → generate list → shop” workflow.

### 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

### 3. Terminology

`RecipeRoot`: The filesystem directory containing recipe files. Default: `~/wiki/recipes`.

`Recipe File`: A Markdown file in `RecipeRoot` with extension `.md`.

`Recipe Title`: The human name of the recipe, derived from the first Markdown H1 heading in a recipe file.

`Recipe Slug`: A filename-safe identifier derived from a recipe title (used for new recipe creation).

`Ingredient Entry`: A single unordered list item line under an `Ingredients` section.

`Recipe Index`: A Markdown file that defines categories and links to recipe files. Default: `~/wiki/Recipes.md`.

`Category`: A label derived from a heading in the Recipe Index.

`Recipe Tags`: Comma-separated labels on a `Tags:` line in a recipe file, used for filtering and organization.

### 4. Concepts

#### 4.1. File-backed recipes

Recipes are stored as Markdown files in `RecipeRoot`. The tool does not require a database and MUST treat the filesystem as the source of truth.

#### 4.2. Minimal parsing

To support list/search/shopping-list behavior without imposing a strict schema, recipe parsing is intentionally minimal:

- Recipe Title is derived from the first H1 (`# ...`) heading.
- Recipe Tags are derived from a line matching `Tags: <comma-separated values>` (case-insensitive key).
- Ingredient Entries are derived from unordered list item lines under a section heading `## Ingredients` (case-insensitive).

Additional sections (e.g. method, notes, servings) MAY exist and MUST be preserved if the CLI writes files.

#### 4.3. Categories via index

Categories are defined in a Recipe Index Markdown file (example: `~/wiki/Recipes.md`). The CLI uses this file as a source of category→recipe mappings, without requiring any front-matter or metadata inside recipe files.

### 5. Requirements

#### 5.1. Storage location

1. The CLI MUST default `RecipeRoot` to `~/wiki/recipes`.
2. A `--dir` flag MUST override the default `RecipeRoot` for a single invocation.
3. An environment variable `REGIMEN_RECIPES_DIR` SHOULD be supported as an override.

#### 5.2. Recipe file discovery

1. The CLI MUST discover recipes by scanning `RecipeRoot` for files matching `*.md`.
2. The CLI MUST ignore subdirectories (non-recursive scan).
3. The CLI MUST ignore non-regular files.

#### 5.3. Recipe title parsing

1. The CLI MUST parse the Recipe Title as the first line matching the Markdown H1 pattern `^#\s+(.+)$`.
2. If a recipe file contains no H1 heading, the CLI MUST treat the recipe as invalid and surface an error (or omit it from listing with a warning; see Section 8).

#### 5.4. Ingredient parsing

1. The CLI MUST treat a Markdown heading matching `## Ingredients` (case-insensitive) as the start of an ingredient section.
2. Within the ingredient section, the CLI MUST treat unordered list items (`- ` or `* `) as Ingredient Entries.
3. The CLI MUST include Ingredient Entries nested under subheadings (e.g. `### Main`) as Ingredient Entries.
4. Ingredient parsing MUST stop at the next heading of level 2 or higher (i.e. a line beginning with `## `).

#### 5.5. Tag parsing

1. The CLI MUST parse Recipe Tags from a line matching the pattern `Tags:\s*(.+)` (case-insensitive).
2. The tag line MAY appear anywhere in the recipe file before the first H2 heading.
3. Tags MUST be parsed as comma-separated values, with leading/trailing whitespace trimmed from each tag.
4. Tag matching MUST be case-insensitive.
5. If no `Tags:` line is present, the recipe MUST be treated as having an empty tag list.

#### 5.6. Categories (Recipe Index)

1. The CLI MUST support a Recipe Index file.
2. The CLI MUST default the Recipe Index path to `~/wiki/Recipes.md`.
3. A `--index` flag MUST override the default index path for a single invocation.
4. An environment variable `REGIMEN_RECIPES_INDEX` SHOULD be supported as an override.
5. The Recipe Index MUST be parsed as follows:
   - The CLI MUST treat headings (`#`, `##`, etc.) after the document title as category names.
   - The CLI MUST treat Markdown links in the form `[Text](path)` as recipe references.
   - A recipe reference belongs to the most recent category heading.
   - Link targets that are relative paths MUST be resolved relative to the directory containing the index file.
6. The CLI MUST ignore non-`.md` recipe links.

#### 5.7. CLI operations

The CLI MUST provide the following commands under `regimen recipes`:

- `list`: list recipes.
- `view`: show a recipe.
- `search`: search recipes.
- `new`: create a new recipe file skeleton.
- `shopping-list`: generate a shopping list from recipes.

The CLI MAY provide additional commands in the future.

### 6. Interface

#### 6.1. Command group

The `regimen` executable MUST include a top-level command group:

- `regimen recipes`

#### 6.2. `regimen recipes list`

**Synopsis**:

- `regimen recipes list [--dir <path>] [--category <name>] [--tag <name>] [--index <path>]`

**Behavior**:

- MUST print one recipe title per line.
- Output SHOULD be stable (sorted by title).
- If `--category` is provided, the command MUST list recipes belonging to that category (as defined by the Recipe Index).
- If `--tag` is provided, the command MUST list only recipes that have the specified tag.

#### 6.3. `regimen recipes view`

**Synopsis**:

- `regimen recipes view <recipe> [--dir <path>] [--ingredients]`

**Behavior**:

- By default MUST print the full recipe file contents to stdout.
- With `--ingredients`, MUST print only Ingredient Entries (one per line) in source order.

#### 6.4. `regimen recipes search`

**Synopsis**:

- `regimen recipes search <query> [--dir <path>] [--ingredients-only]`

**Behavior**:

- MUST match case-insensitively.
- MUST search recipe titles.
- MUST search ingredient entries unless `--ingredients-only` is specified.
- Output MUST identify matching recipes (at minimum by title; MAY include filename/path via a future flag).

#### 6.5. `regimen recipes new`

**Synopsis**:

- `regimen recipes new <title> [--dir <path>] [--slug <slug>] [--force]`

**Behavior**:

- MUST create a new Markdown file in `RecipeRoot`.
- Filename MUST be `<slug>.md`.
- If `--slug` is omitted, the CLI MUST derive a slug from `<title>`.
- If the target file already exists, the CLI MUST error unless `--force` is provided.
- The created file MUST be a skeleton intended for manual editing.
- The created file MUST contain the following headings:
  - `# <Recipe Title>`
  - `## Ingredients`
  - `## Method`
  - `## Notes`

#### 6.6. `regimen recipes shopping-list`

**Synopsis**:

- `regimen recipes shopping-list <recipe>... [--dir <path>] [--out <path>]`
- `regimen recipes shopping-list --category <name> [--index <path>] [--out <path>]`
- `regimen recipes shopping-list --tag <name> [--dir <path>] [--out <path>]`

**Behavior**:

- MUST extract Ingredient Entries from each referenced recipe.
- When one or more recipes are provided as positional arguments (`<recipe>...`), the output MUST list recipes in the same order as provided on the command line.
- When `--tag` is provided, the command MUST generate a shopping list from all recipes that have the specified tag.
- The shopping list MUST include ingredient entry text verbatim (including measurements such as `250g`, `1 tsp`, and free-form qualifiers such as `to taste`), after removing the list marker (`- ` or `* `) and trimming surrounding whitespace.
- The shopping list generator MUST NOT attempt to normalize units, parse quantities, deduplicate entries, or sum measurements across recipes.
- MUST output a shopping list as Markdown.
- If `--out` is provided, MUST write the shopping list to that file; otherwise MUST write to stdout.
- Output MUST include:
  - a top-level heading `# Shopping List`
  - for each recipe, a subheading `## <Recipe Title>` followed by bullet items of Ingredient Entries

### 7. Behavior

#### 7.1. Slug derivation (for `new`)

When deriving a slug from a title, the CLI SHOULD:

- lowercase the title
- replace runs of whitespace with `_`
- remove characters not in `[a-z0-9_-]`

#### 7.2. Name resolution

Commands that take a recipe identifier SHOULD accept either:

- an explicit path to a recipe file, or
- a base name / slug that resolves to `${RecipeRoot}/<name>.md`.

If resolution matches multiple candidates, the CLI MUST return an ambiguity error.

#### 7.3. Output streams

- Normal command output MUST go to stdout.
- Errors and warnings MUST go to stderr.

### 8. Error Handling

1. If `RecipeRoot` does not exist, commands MUST return an error describing the missing directory.
2. If `RecipeRoot` exists but is not a directory, commands MUST return an error.
3. If a recipe file cannot be read, the CLI MUST return an error.
4. For `list`, unreadable/invalid recipes SHOULD be skipped with a warning to stderr (and MUST NOT crash the command).
5. For `view` / `search` / `new` / `shopping-list`, errors MUST cause a non-zero exit status.
6. If `--category` is used and the Recipe Index cannot be read or parsed, the command MUST error.
7. If a category name does not exist, the command MUST error and SHOULD suggest available categories.

### 9. Examples

#### 9.1. Example recipe file

File: `~/wiki/recipes/dhal_vhal.md`

```md
# Vaal Dhal

Tags: vegetarian, indian, quick

Estimated recipe from ingredients list: https://hodmedods.co.uk/products/vaal-dhal

## Ingredients
### Main
- Fava Beans (Broad Beans) (45%)
- Water*
- Tomato Puree
- Rapeseed Oil
- Rice Flour
- Sugar
- Salt*

### Herbs and Spices
- Cumin
- Turmeric
- Fenugreek
- Coriander
- Caraway Seeds
- Chilli Powder
- Coriander Leaf
- Mustard seeds
```

#### 9.2. Example Recipe Index

File: `~/wiki/Recipes.md`

```md
# Recipes

# British
[Broad Bean Dhal](recipes/dhal_vhal.md)
```

#### 9.3. Listing recipes

```bash
regimen recipes list --dir ~/wiki/recipes
```

#### 9.4. Listing recipes by category

```bash
regimen recipes list --category British
```

#### 9.5. Listing recipes by tag

```bash
regimen recipes list --tag vegetarian
```

#### 9.6. Generating a shopping list

```bash
regimen recipes shopping-list dhal_vhal --out /tmp/shopping.md
```

#### 9.7. Generating a shopping list by tag

```bash
regimen recipes shopping-list --tag quick --out /tmp/shopping.md
```

### 10. Security Considerations

1. The CLI MUST treat `RecipeRoot` and Recipe Index paths as untrusted input.
2. The CLI MUST clean and validate user-supplied recipe identifiers to prevent path traversal when resolving within `RecipeRoot`.
3. The CLI SHOULD avoid following symlinks that escape `RecipeRoot` when opening files.
4. The CLI MUST NOT execute any content from recipe files.

### 11. Testing Considerations

Tests SHOULD cover:

- title parsing (valid/invalid)
- tag parsing (present, absent, case-insensitivity, whitespace trimming)
- ingredient extraction across subheadings
- search behavior (title matches vs ingredient matches)
- name resolution (path vs slug)
- slug generation for `new`
- index parsing (categories + relative link resolution)
- tag filtering for list and shopping-list commands
- shopping list output structure and determinism

### 12. References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997, https://www.rfc-editor.org/rfc/rfc2119

[CommonMark] https://spec.commonmark.org/

[Vimwiki] https://github.com/vimwiki/vimwiki
