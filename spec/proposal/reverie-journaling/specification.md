# reverie-journaling

**Depends on**: none
**Affected files**: cmd/reverie/main.go, internal/reverie/, pkg/journal/, pkg/markdown/

## Abstract

This specification defines a new standalone executable `reverie` for personal journaling, idea capture, and note management. It provides commands for quick note-taking, daily journaling, full-text search across entries, tag-based organization, random idea surfacing, and standup generation. All data is stored as Markdown files compatible with personal wiki systems.

## 1. Introduction

Knowledge workers frequently need to capture fleeting thoughts, maintain daily logs, record meeting notes, and organize ideas. The current landscape requires switching between multiple tools: note-taking apps for quick captures, separate journaling apps for daily logs, wiki systems for organized knowledge, and productivity tools for standups.

Most note-taking tools either lock data in proprietary formats, require cloud synchronization, or provide poor terminal integration. Developers who spend their day in the terminal want a fast, keyboard-driven way to capture and retrieve thoughts without context-switching to a GUI.

This specification defines `reverie`, a terminal-first journaling and note-taking tool that stores everything as Markdown files. The name "reverie" captures the dreamlike flow of thoughts and ideas that the tool is designed to capture and surface. It provides quick capture for fleeting thoughts, structured journaling for daily logs, tagging for organization, and full-text search for retrieval.

The tool integrates with the sleepless ecosystem's philosophy: terminal-first, local-first, human-readable storage, and scriptable interfaces.

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`Entry`: A single journal entry or note with a timestamp, content, and optional tags.

`Journal`: A time-based collection of entries organized by date (daily logs, reflections).

`Note`: A standalone entry not necessarily tied to a specific date (ideas, meeting notes, references).

`Tag`: A label attached to an entry for categorization and filtering (e.g., `#work`, `#idea`, `#meeting`).

`Daily Entry`: An entry associated with a specific date, stored in a date-based file structure.

`Floating Note`: A standalone note with a unique ID, stored independently of date structure.

`Journal Root`: The directory where reverie stores all entries (default: `~/journal`).

`Template`: A predefined structure for specific entry types (standup, meeting, reflection).

`Random Surfacing`: The process of retrieving random past entries to surface forgotten ideas.

## 4. Concepts

### 4.1. Dual Entry Model

Reverie supports two types of entries:
1. **Daily Entries**: Time-based logs tied to specific dates, useful for journaling and daily logs
2. **Floating Notes**: Standalone notes with unique IDs, useful for ideas and references that aren't date-specific

This dual model accommodates both chronological journaling and timeless note-taking.

### 4.2. Markdown Storage

All entries are stored as Markdown files with frontmatter metadata. This ensures:
- Human readability and editability
- Compatibility with other tools (Obsidian, vim, VS Code)
- Version control friendliness
- Future-proof format

### 4.3. Tag-Based Organization

Tags provide lightweight organization without rigid hierarchies. Entries can have multiple tags, and tags can be filtered and searched.

### 4.4. Idea Surfacing

Random retrieval of past entries helps surface forgotten ideas and make unexpected connections, mimicking the way memories resurface during actual reveries.

## 5. Requirements

### 5.1. Core Executable

1. The project MUST provide a standalone executable named `reverie`.
2. The executable MUST be independent of other sleepless executables.
3. All commands MUST return exit code 0 on success and non-zero on failure.

### 5.2. Entry Commands

1. `reverie add <text>` MUST create a new entry with the provided text.
2. `reverie add` (without text) MUST open the user's default editor to write an entry.
3. `reverie add --date <date>` MUST create an entry for a specific date (YYYY-MM-DD format).
4. `reverie add --tags <tags>` MUST attach comma-separated tags to the entry.
5. `reverie add --template <name>` MUST create an entry from a predefined template.
6. `reverie add --floating` MUST create a floating note instead of a daily entry.
7. `reverie edit <id>` MUST open the specified entry in the user's editor.
8. `reverie delete <id>` MUST remove the specified entry.
9. Entry IDs MUST be short prefixes (6-8 characters) that uniquely identify entries.

### 5.3. Viewing Commands

1. `reverie today` MUST display all entries from today.
2. `reverie week` MUST display a summary of the past 7 days' entries.
3. `reverie month` MUST display a summary of the current month's entries.
4. `reverie show <id>` MUST display a specific entry.
5. `reverie list` MUST list recent entries with timestamps and previews.
6. `reverie list --limit <n>` MUST limit output to n entries (default: 20).
7. `reverie list --tags <tags>` MUST filter entries by tags (comma-separated, OR logic).
8. `reverie list --date <date>` MUST show entries from a specific date.
9. `reverie list --range <start> <end>` MUST show entries within a date range.

### 5.4. Search Commands

1. `reverie search <query>` MUST search entry text and titles case-insensitively.
2. Search MUST support multiple search terms (AND logic by default).
3. `reverie search --or <terms>` MUST use OR logic for multiple terms.
4. `reverie search --tags <tags>` MUST limit search to specific tags.
5. `reverie search --after <date>` MUST limit search to entries after a date.
6. `reverie search --before <date>` MUST limit search to entries before a date.
7. Search results MUST include entry ID, date, matched context, and tags.

### 5.5. Tag Commands

1. `reverie tags` MUST list all tags with entry counts.
2. `reverie tags <tag>` MUST show all entries with the specified tag.
3. `reverie tag <id> <tags>` MUST add tags to an existing entry.
4. `reverie untag <id> <tags>` MUST remove tags from an existing entry.

### 5.6. Special Commands

1. `reverie random` MUST display a random past entry.
2. `reverie random --tag <tag>` MUST display a random entry with the specified tag.
3. `reverie random --count <n>` MUST display n random entries.
4. `reverie standup` MUST generate a standup report from recent entries.
5. `reverie standup --days <n>` MUST include entries from the past n days (default: 1).
6. The standup report MUST use entries tagged with `#work` or `#progress` by default.
7. `reverie stats` MUST display statistics (entry count, tags, entries per week, streak).

### 5.7. Template Commands

1. `reverie template list` MUST list available templates.
2. `reverie template create <name>` MUST create a new template.
3. Templates MUST support placeholders for date, time, and custom prompts.
4. Built-in templates MUST include: `standup`, `meeting`, `reflection`, `idea`.

### 5.8. Storage Format

1. Daily entries MUST be stored in `YYYY/MM/DD.md` structure.
2. Floating notes MUST be stored in `notes/<id>.md`.
3. Each entry MUST have YAML frontmatter with:
   - `id`: unique identifier (short hex string)
   - `created`: RFC3339 timestamp
   - `updated`: RFC3339 timestamp (if modified)
   - `tags`: array of strings
   - `type`: "daily" or "floating"
4. The Journal Root MUST default to `~/journal`.
5. The system MUST support `REVERIE_JOURNAL_DIR` environment variable to override.

### 5.9. Output Formats

1. All list and search commands MUST support `--json` for machine-readable output.
2. Text output MUST be human-readable with syntax highlighting for Markdown.
3. `--compact` MUST produce single-line summaries for piping to other tools.

## 6. Interface

### 6.1. Commands

```bash
# Add entries
reverie add <text> [--date <date>] [--tags <tags>] [--template <name>] [--floating]
reverie add  # Opens editor
reverie edit <id>
reverie delete <id>

# View entries
reverie today
reverie week
reverie month
reverie show <id>
reverie list [--limit <n>] [--tags <tags>] [--date <date>] [--range <start> <end>]

# Search
reverie search <query> [--or] [--tags <tags>] [--after <date>] [--before <date>]

# Tags
reverie tags
reverie tags <tag>
reverie tag <id> <tags>
reverie untag <id> <tags>

# Special commands
reverie random [--tag <tag>] [--count <n>]
reverie standup [--days <n>]
reverie stats

# Templates
reverie template list
reverie template create <name>
```

### 6.2. Flags

**Global Flags:**
- `--journal <path>`: Override journal directory
- `--json`: Output in JSON format
- `--compact`: Single-line summaries

**Add Flags:**
- `--date <YYYY-MM-DD>`: Specific date for entry
- `--tags <tag1,tag2>`: Comma-separated tags
- `--template <name>`: Use template
- `--floating`: Create floating note

**List/Search Flags:**
- `--limit <n>`: Maximum entries to show
- `--tags <tags>`: Filter by tags
- `--date <date>`: Filter by specific date
- `--range <start> <end>`: Date range filter
- `--after <date>`: Entries after date
- `--before <date>`: Entries before date
- `--or`: OR logic for search terms

### 6.3. Entry File Format

Daily entry (`2026/01/18.md`):
```markdown
---
id: a1b2c3d4
created: 2026-01-18T10:30:00Z
updated: 2026-01-18T14:20:00Z
tags: [work, meeting, project-x]
type: daily
---

# 2026-01-18

## 10:30 - Team Standup

Discussed the new API endpoint design. John raised concerns about rate limiting.

## 14:00 - Code Review

Reviewed PR #42. Suggested refactoring the validation logic into a separate function.

#work #progress
```

Floating note (`notes/e5f6g7h8.md`):
```markdown
---
id: e5f6g7h8
created: 2026-01-18T16:45:00Z
tags: [idea, architecture]
type: floating
---

# API Gateway Caching Strategy

Consider implementing a distributed cache layer using Redis for frequently accessed endpoints. This could reduce load on the database by ~40%.

Key considerations:
- Cache invalidation strategy
- TTL policies per endpoint
- Monitoring cache hit rates

#idea #architecture
```

### 6.4. Template Format

Template file (`.reverie/templates/meeting.md`):
```markdown
---
tags: [meeting]
---

# Meeting: {{PROMPT:Meeting Title}}

**Date:** {{DATE}}
**Attendees:** {{PROMPT:Attendees}}

## Agenda
- 

## Notes


## Action Items
- [ ] 

#meeting
```

## 7. Behavior

### 7.1. Entry Creation

1. When creating an entry without text, reverie MUST:
   - Create a temporary file with frontmatter template
   - Open it in the user's editor (via `$EDITOR` environment variable)
   - Parse the result when the editor closes
   - Create the entry if the content is non-empty
   - Discard if the user exits without saving or content is empty

2. Entry IDs MUST be generated as 8-character lowercase hex strings.

3. When using `--date` with a past or future date, the entry MUST be placed in the appropriate YYYY/MM/DD structure.

### 7.2. Tag Processing

1. Tags specified via `--tags` MUST be parsed as comma-separated values.
2. Leading/trailing whitespace MUST be trimmed from each tag.
3. Tags MUST be stored in lowercase for consistent searching.
4. Tags embedded in entry content (e.g., `#work`) MUST also be extracted and stored in frontmatter.
5. When displaying entries, both frontmatter tags and inline tags MUST be shown.

### 7.3. Search Behavior

1. Search MUST be case-insensitive.
2. Search MUST match against both entry content and tags.
3. By default, multiple search terms MUST use AND logic (all terms must match).
4. With `--or`, any matching term is sufficient.
5. Search results MUST show a context snippet around each match (±50 characters).

### 7.4. Standup Generation

1. Standup MUST aggregate entries from the specified time period.
2. The report MUST group entries by day.
3. Only entries with work-related tags (`#work`, `#progress`, `#meeting`) MUST be included by default.
4. The output MUST be formatted as:
   ```
   Standup Report (Last 1 day)
   
   ## Yesterday (2026-01-17)
   - Completed API endpoint implementation
   - Code review for PR #42
   
   ## Today (2026-01-18)
   - Team standup meeting
   - Working on authentication middleware
   ```

### 7.5. Random Surfacing

1. Random selection MUST be uniformly distributed across all eligible entries.
2. When `--tag` is specified, only entries with that tag are eligible.
3. Random entries MUST be displayed in full (not just previews).

## 8. Error Handling

1. If the Journal Root does not exist, reverie MUST create it on first use.
2. If `$EDITOR` is not set, reverie MUST default to `vim` on Unix systems and `notepad` on Windows.
3. If an entry ID prefix is ambiguous (matches multiple entries), reverie MUST list all matches and exit with code 1.
4. If an entry ID cannot be resolved, reverie MUST exit with code 1 and suggest using `list` or `search`.
5. If a template does not exist, reverie MUST exit with code 1 and list available templates.
6. If date parsing fails, reverie MUST exit with code 1 and show the expected format.
7. Malformed frontmatter in existing entries MUST be reported as warnings but not prevent other operations.

## 9. Examples

```bash
# Quick note
reverie add "Had an idea for improving the login flow"

# Add with tags
reverie add "Sprint planning meeting" --tags "meeting,work,planning"

# Open editor for detailed entry
reverie add

# Create floating note
reverie add --floating "Research: Distributed caching strategies"

# View today's entries
reverie today

# View weekly summary
reverie week

# Search for entries
reverie search "API design"
reverie search "authentication" --after 2026-01-01

# Search with tags
reverie search "meeting" --tags "work"

# List entries with specific tag
reverie tags meeting

# Add tags to existing entry
reverie tag a1b2c3 "important,follow-up"

# Random idea surfacing
reverie random
reverie random --tag idea --count 3

# Generate standup
reverie standup
reverie standup --days 3

# View statistics
reverie stats

# Use template
reverie add --template meeting

# Create custom template
reverie template create retrospective
```

## 10. Security Considerations

1. Reverie MUST NOT transmit journal data over the network.
2. Journal files MAY contain sensitive personal or work information. Documentation MUST warn users to:
   - Set appropriate file permissions on the Journal Root
   - Avoid committing sensitive journals to public repositories
   - Use encryption if storing on cloud-synced directories
3. Entry IDs MUST be generated using a cryptographically secure random source to prevent collisions.
4. When opening editors, reverie MUST validate `$EDITOR` to prevent command injection.
5. Template processing MUST NOT execute arbitrary code from template files.

## 11. Testing Considerations

Test scenarios SHOULD include:

- Entry creation (inline text vs editor)
- Daily vs floating entry storage structure
- Tag parsing from both flags and content
- Search with AND/OR logic
- Date range filtering
- ID prefix resolution (unique and ambiguous)
- Random selection distribution
- Standup generation with various time ranges
- Template processing with prompts
- Handling malformed frontmatter in existing entries
- Large journal directories (performance)
- Concurrent access (file locking if applicable)
- Entry editing and update timestamp handling
- Week/month summary generation
- Statistics calculation accuracy

## 12. References

### Normative References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997, https://www.rfc-editor.org/rfc/rfc2119

[RFC3339] Klyne, G. and C. Newman, "Date and Time on the Internet: Timestamps", July 2002, https://www.rfc-editor.org/rfc/rfc3339

[YAML] YAML Specification, https://yaml.org/spec/

### Informative References

[OBSIDIAN] Obsidian, https://obsidian.md/

[JRNL] jrnl - Command Line Journal, https://jrnl.sh/

[COMMONMARK] CommonMark Specification, https://spec.commonmark.org/
