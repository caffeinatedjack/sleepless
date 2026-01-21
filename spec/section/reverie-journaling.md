# reverie-journaling

**Depends on**: none
**Affected files**: internal/regimen/note*.go, pkg/notes/, pkg/markdown/, pkg/crypto/

## Abstract

This specification defines a new `note` command group within the `regimen` executable for personal journaling, idea capture, and note management. It provides commands for quick note-taking, daily journaling, full-text search across entries, tag-based organization, random idea surfacing, and report generation. All data is stored as Markdown files in the wiki directory (`~/wiki/notes/`) alongside task files. The wiki directory can be encrypted/decrypted via `regimen encrypt` and `regimen decrypt` commands.

## Overview

### What

- Add `regimen note` commands for daily notes (date-addressed) and floating notes (ID-addressed)
- Store notes in a Vimwiki-compatible directory alongside existing tasks
- Add explicit wiki-wide `regimen encrypt` / `regimen decrypt` commands to protect local data at rest

### Why

- Keep tasks and notes local-first and human-editable
- Reduce context switching by capturing ideas and journal entries from the same CLI that manages tasks
- Provide a simple, explicit encryption workflow that is compatible with "offline" and "single-user" assumptions

### Who

- Primary users: terminal-first users maintaining a `~/wiki/` directory for tasks and knowledge
- Affected systems: the `regimen` CLI and on-disk wiki directory structure

### Success Criteria

The proposal is complete when:

1. Notes can be created/edited/viewed/searched using `regimen note ...` and are persisted in `~/wiki/notes/`.
2. `regimen encrypt` converts eligible files to `.enc`, writes `.encrypted`, and blocks wiki-data commands until decrypted.
3. `regimen decrypt` reverses encryption and restores original filenames.
4. All commands produce actionable errors when the wiki is encrypted or partially encrypted.
5. Unit and integration tests cover the main workflows (notes + encryption) with deterministic fixtures.

### Out of Scope

- Automatic on-the-fly decryption/encryption during normal commands
- Bidirectional task-note linking (notes may contain `@task:<id>` syntax only)
- Syncing/backups/cloud features
- Attachments/media
- Cross-device conflict resolution

## DEPENDENCIES

### Proposal Dependencies

- none

### Third-Party Dependencies

- github.com/spf13/cobra: CLI command parsing and flag handling
- gopkg.in/yaml.v3: YAML frontmatter parsing for notes and templates
- golang.org/x/crypto/argon2: Argon2id KDF for deriving encryption key from passphrase
- golang.org/x/term: no-echo passphrase prompting in terminal

### Missing Third-Party Docs

The repository currently does not include `spec/third/` documentation.
Implementation SHOULD add third-party docs for the dependencies above before coding begins.

## 1. Introduction

Knowledge workers frequently need to capture fleeting thoughts, maintain daily logs, record meeting notes, and organize ideas. The current landscape requires switching between multiple tools: note-taking apps for quick captures, separate journaling apps for daily logs, wiki systems for organized knowledge, and productivity tools for standups.

Most note-taking tools either lock data in proprietary formats, require cloud synchronization, or provide poor terminal integration. Developers who spend their day in the terminal want a fast, keyboard-driven way to capture and retrieve thoughts without context-switching to a GUI.

This specification extends the `regimen` executable with a `note` command group for terminal-first journaling and note-taking. Notes are stored as Markdown files in `~/wiki/notes/` alongside task management files. This provides quick capture for fleeting thoughts, structured journaling for daily logs, tagging for organization, and full-text search for retrieval.

The tool integrates with the sleepless ecosystem's philosophy: terminal-first, local-first, human-readable storage, and scriptable interfaces. Notes can reference tasks using `@task:<id>` syntax for future integration.

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`Note`: A Markdown file managed by `regimen note`.

`Daily Note`: A note associated with a calendar date, stored as `YYYY-MM-DD.md` in the notes directory.

`Floating Note`: A note not associated with a calendar date, stored as `<id>.md` in the notes directory.

`Tag`: A label attached to an entry for categorization and filtering (e.g., `#work`, `#idea`, `#meeting`).

`Daily Entry`: An entry associated with a specific date, stored in a date-based file structure.

`Floating Note`: A standalone note with a unique ID, stored independently of date structure.

`Notes Directory`: The directory where notes are stored (default: `~/wiki/notes/`).

`Template`: A predefined Markdown structure for specific note types (meeting, reflection, idea, report).

`Random Surfacing`: The process of retrieving random past entries to surface forgotten ideas.

## 4. Concepts

### 4.1. Dual Note Model

The system supports two types of notes:
1. **Daily Notes**: One file per date, useful for journaling and daily logs
2. **Floating Notes**: Standalone files with unique IDs, useful for ideas and references that aren't date-specific

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

### 5.1. Core Command Group

1. The `regimen` executable MUST provide a `note` command group.
2. The `note` commands MUST be integrated into the existing `regimen` CLI.
3. All commands MUST return exit code 0 on success and non-zero on failure.
4. The wiki directory MUST be configurable via `--wiki-dir` flag, `REGIMEN_WIKI_DIR` environment variable, or default to `~/wiki`.
5. Configuration precedence MUST be: flag > environment variable > default.

### 5.2. Note Commands

1. `regimen note add <text>` MUST append a new timestamped section to the target daily note.
2. `regimen note add` (without text) MUST open the target daily note in the user's default editor.
3. `regimen note add --date <date>` MUST target a specific daily note (YYYY-MM-DD format).
4. `regimen note add --tags <tags>` MUST attach comma-separated tags to the entry.
5. `regimen note add --template <name>` MUST create an entry from a predefined template.
6. `regimen note add --floating` MUST create a floating note instead of a daily note.
7. `regimen note add --daily` MUST explicitly target a daily note (default behavior).
8. `regimen note edit <id>` MUST open the specified floating note in the user's editor.
9. `regimen note edit --date <date>` MUST open the specified daily note in the user's editor.
10. `regimen note delete <id>` MUST delete the specified floating note.
11. `regimen note delete --date <date>` MUST delete the specified daily note.
12. Floating note IDs MUST be 8-character lowercase hex strings and be addressable via unique prefixes (6-8 characters).
13. Notes MAY reference tasks using `@task:<id>` syntax in their content.

### 5.3. Viewing Commands

1. `regimen note today` MUST display today's daily note.
2. `regimen note week` MUST display a list of daily notes from the past 7 days with previews.
3. `regimen note month` MUST display a list of daily notes from the current month with previews.
4. `regimen note show <id>` MUST display a specific floating note.
5. `regimen note show --date <date>` MUST display a specific daily note.
6. `regimen note list` MUST list recent notes with timestamps and previews.
7. `regimen note list --limit <n>` MUST limit output to n notes (default: 20).
8. `regimen note list --tags <tags>` MUST filter notes by tags (comma-separated, OR logic).
9. `regimen note list --date <date>` MUST show the daily note for a specific date.
10. `regimen note list --range <start> <end>` MUST list daily notes within a date range.

### 5.4. Search Commands

1. `regimen note search <query>` MUST search note text and titles case-insensitively.
2. Search MUST support multiple search terms (AND logic by default).
3. `regimen note search --or <terms>` MUST use OR logic for multiple terms.
4. `regimen note search --tags <tags>` MUST limit search to specific tags.
5. `regimen note search --after <date>` MUST limit search to entries after a date.
6. `regimen note search --before <date>` MUST limit search to entries before a date.
7. Search results MUST include note identifier (date for daily or ID for floating), matched context, and tags.

### 5.5. Tag Commands

1. `regimen note tags` MUST list all tags with entry counts.
2. `regimen note tags <tag>` MUST show all entries with the specified tag.
3. `regimen note tag <id> <tags>` MUST add tags to an existing entry.
4. `regimen note untag <id> <tags>` MUST remove tags from an existing entry.

### 5.6. Special Commands

1. `regimen note random` MUST display a random note (daily or floating).
2. `regimen note random --tag <tag>` MUST display a random entry with the specified tag.
3. `regimen note random --count <n>` MUST display n random entries.
4. `regimen note report` MUST generate a report from recent daily notes.
5. `regimen note report --days <n>` MUST include entries from the past n days (default: 1).
6. The report MUST use entries tagged with `#work` or `#progress` by default.
7. `regimen note stats` MUST display statistics (entry count, tags, entries per week, streak).

### 5.7. Template Commands

1. `regimen note template list` MUST list available templates.
2. `regimen note template create <name>` MUST create a new template.
3. Templates MUST support placeholders for date, time, and custom prompts.
4. Built-in templates MUST include: `meeting`, `reflection`, `idea`, `report`.
5. Templates MUST be stored in `~/wiki/.templates/` by default.

### 5.8. Storage Format

1. Daily notes MUST be stored as `~/wiki/notes/YYYY-MM-DD.md`.
2. Floating notes MUST be stored as `~/wiki/notes/<id>.md`.
3. Daily notes MUST have YAML frontmatter with:
   - `date`: YYYY-MM-DD
   - `created`: RFC3339 timestamp
   - `updated`: RFC3339 timestamp (if modified)
   - `tags`: array of strings
   - `type`: "daily"
4. Floating notes MUST have YAML frontmatter with:
   - `id`: unique identifier (8-char lowercase hex)
   - `created`: RFC3339 timestamp
   - `updated`: RFC3339 timestamp (if modified)
   - `tags`: array of strings
   - `type`: "floating"
5. The Notes Directory MUST default to `~/wiki/notes/`.
6. The wiki directory MUST be configurable via `--wiki-dir` flag or `REGIMEN_WIKI_DIR` environment variable.
7. Notes MAY contain `@task:<id>` references to link to task files (syntax reserved for future use).

### 5.9. Output Formats

1. All list and search commands MUST support `--json` for machine-readable output.
2. Text output MUST be human-readable with syntax highlighting for Markdown.
3. `--compact` MUST produce single-line summaries for piping to other tools.

### 5.10. Encryption

1. `regimen encrypt` MUST encrypt all `.md` and `.json` files in the wiki directory recursively.
2. `regimen decrypt` MUST decrypt all `.enc` files in the wiki directory recursively.
3. Encrypted files MUST use the `.enc` extension (e.g., `file.md` becomes `file.md.enc`).
4. A `.encrypted` marker file MUST be created in the wiki root when encryption is active.
5. The passphrase MUST be provided via `--passphrase-stdin` flag OR interactive prompt.
6. When passphrase is entered interactively for `regimen encrypt`, the passphrase MUST be confirmed by entering it twice.
7. The passphrase MUST NOT be written to disk.
8. If the `.encrypted` marker exists, any command that reads or writes the wiki directory MUST exit with error: "Wiki is encrypted. Run 'regimen decrypt' first."
9. The encryption scheme MUST use authenticated encryption (AEAD) and a password-based key derivation function (Argon2id).
10. The encryption implementation MUST skip `wiki/.git/` entirely.
11. The encryption implementation MUST NOT follow symlinks while scanning the wiki directory.
12. The encryption implementation MUST be best-effort: it MUST attempt all eligible files, then exit non-zero if any failures occurred.
13. `regimen encrypt` MUST fail if wiki is already encrypted (marker file exists).
14. `regimen decrypt` MUST fail if wiki is not encrypted (no marker file).
15. After successful decryption, the `.encrypted` marker file MUST be removed.

### 5.11. Encryption File Format

1. The `.encrypted` marker file MUST be plaintext and MUST be JSON.
2. The `.encrypted` marker JSON MUST include:
   - `version` (integer)
   - `created` (RFC3339 timestamp)
   - `salt_b64` (base64, 16+ bytes)
   - `argon2_time` (integer)
   - `argon2_memory_kib` (integer)
   - `argon2_threads` (integer)
3. The marker file MUST use these default Argon2id parameters ("Balanced"):
   - `argon2_time`: 3
   - `argon2_memory_kib`: 131072 (128 MiB)
   - `argon2_threads`: 4
4. The passphrase-derived key MUST be derived once per `encrypt`/`decrypt` run using the marker file salt.
5. Each encrypted file MUST contain:
   - Magic bytes: `REGIMENENC` (ASCII)
   - Version: 1 byte
   - Nonce: 12 bytes (for AES-GCM)
   - Ciphertext: remaining bytes
6. The plaintext encrypted MUST be the original file bytes.
7. The encryption MUST use AES-256-GCM.
8. The AEAD associated data MUST include the original relative path (from wiki root) to detect file swapping.
9. Each file encryption MUST use a random nonce.
10. Encryption MUST write to a temp file and rename atomically to avoid partial writes.

## 6. Interface

### 6.1. Commands

```bash
# Add notes
regimen note add <text> [--date <date>] [--tags <tags>] [--template <name>] [--floating|--daily]
regimen note add  # Opens editor
regimen note edit <id>
regimen note edit --date <date>
regimen note delete <id>
regimen note delete --date <date>

# View entries
regimen note today
regimen note week
regimen note month
regimen note show <id>
regimen note show --date <date>
regimen note list [--limit <n>] [--tags <tags>] [--date <date>] [--range <start> <end>]

# Search
regimen note search <query> [--or] [--tags <tags>] [--after <date>] [--before <date>]

# Tags
regimen note tags
regimen note tags <tag>
regimen note tag <id> <tags>
regimen note untag <id> <tags>

# Special commands
regimen note random [--tag <tag>] [--count <n>]
regimen note report [--days <n>]
regimen note stats

# Templates
regimen note template list
regimen note template create <name>

# Encryption (wiki-wide, affects all files)
regimen encrypt [--passphrase-stdin]
regimen decrypt [--passphrase-stdin]
```

### 6.2. Flags

**Global Flags:**
- `--wiki-dir <path>`: Override wiki directory (default: ~/wiki)
- `--json`: Output in JSON format
- `--compact`: Single-line summaries

**Add Flags:**
- `--date <YYYY-MM-DD>`: Specific date for entry
- `--tags <tag1,tag2>`: Comma-separated tags
- `--template <name>`: Use template
- `--floating`: Create floating note
- `--daily`: Create daily entry (default behavior)

**Show/Edit/Delete Flags:**
- `--date <YYYY-MM-DD>`: Target a daily note by date

**Encryption Flags:**
- `--passphrase-stdin`: Read passphrase from stdin instead of interactive prompt

**List/Search Flags:**
- `--limit <n>`: Maximum entries to show
- `--tags <tags>`: Filter by tags
- `--date <date>`: Filter by specific date
- `--range <start> <end>`: Date range filter
- `--after <date>`: Entries after date
- `--before <date>`: Entries before date
- `--or`: OR logic for search terms

### 6.3. Note File Format

Daily note (`~/wiki/notes/2026-01-18.md`):
```markdown
---
date: 2026-01-18
created: 2026-01-18T10:30:00Z
updated: 2026-01-18T14:20:00Z
tags: [work, meeting, project-x]
type: daily
---

# 2026-01-18

## 10:30

Discussed the new API endpoint design. John raised concerns about rate limiting.

## 14:00

Reviewed PR #42. Suggested refactoring the validation logic into a separate function.

#work #progress
```

Floating note (`~/wiki/notes/e5f6g7h8.md`):
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

Template file (`~/wiki/.templates/meeting.md`):
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

### 7.1. Note Creation

1. When creating a daily note without text (`regimen note add`), the system MUST:
   - Create the daily note file if missing
   - Open the daily note file in the user's editor (via `$EDITOR`)
   - Update the `updated` timestamp when the editor exits successfully

2. When creating a daily note with inline text (`regimen note add <text>`), the system MUST:
   - Create the daily note file if missing
   - Append a timestamp heading `## HH:MM` using local time
   - Append the provided text
   - Append inline tags for any `--tags` values (as `#tag` tokens)
   - Merge tags into frontmatter `tags`
   - Update the `updated` timestamp

3. When creating a floating note (`--floating`), the system MUST:
   - Generate a new 8-character lowercase hex ID
   - Write a new `<id>.md` file
   - If `<text>` is omitted, open a temporary file in the editor and write it out if non-empty

4. Floating note IDs MUST be generated as 8-character lowercase hex strings.

5. When using `--date` with a past or future date, the daily note MUST be placed in the appropriate `YYYY-MM-DD.md` file.

6. Daily notes MUST be stored as `YYYY-MM-DD.md` in the notes directory.

7. Floating notes MUST be stored as `<id>.md` in the notes directory.

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

### 7.4. Template Application

1. Templates MUST be loaded from `~/wiki/.templates/` (or the configured wiki directory).
2. `regimen note template list` MUST list template names derived from `*.md` filenames.
3. `regimen note add --template <name>` MUST resolve the template file and apply it to the target note.
4. Template placeholders MUST include:
   - `{{DATE}}` replaced with the target date (YYYY-MM-DD)
   - `{{TIME}}` replaced with the current local time (HH:MM)
   - `{{PROMPT:...}}` which MUST prompt the user for a value (stdin)
5. For daily notes, if the target day file does not exist, `regimen note add --template <name>` MUST create it from the template and open it in the editor.
6. For daily notes, if the target day file already exists, `regimen note add --template <name>` MUST open the editor with the existing content plus the rendered template appended at the end.
7. For floating notes, `regimen note add --floating --template <name>` MUST create a new floating note from the rendered template and open it in the editor.

### 7.5. Report Generation

1. The report command MUST aggregate daily notes from the specified time period.
2. The report MUST group output by day.
3. Only notes with work-related tags (`#work`, `#progress`, `#meeting`) MUST be included by default.
4. The output MUST be formatted as:
   ```
   Report (Last 1 day)
   
   ## Yesterday (2026-01-17)
   - Completed API endpoint implementation
   - Code review for PR #42
   
   ## Today (2026-01-18)
   - Team meeting
   - Working on authentication middleware
   ```

### 7.6. Random Surfacing

1. Random selection MUST be uniformly distributed across all eligible entries.
2. When `--tag` is specified, only entries with that tag are eligible.
3. Random entries MUST be displayed in full (not just previews).

### 7.7. Preview and Listing

1. A "preview" MUST be derived from the first non-empty, non-frontmatter line of the note body.
2. For daily notes, previews SHOULD skip the top-level `# YYYY-MM-DD` title if present and use the first section content line.
3. `regimen note week` and `regimen note month` MUST output one line per day in chronological order.
4. Each line MUST include:
   - The date
   - A preview snippet (trimmed to a reasonable terminal width, e.g., ~80 characters)
5. `regimen note list` MUST output most-recently-updated notes first unless a range/date filter is provided.

## 8. Error Handling

1. If the Notes Directory does not exist, it MUST be created on first use.
2. If `$EDITOR` is not set, the system MUST default to `vim` on Unix systems and `notepad` on Windows.
3. If an entry ID prefix is ambiguous (matches multiple entries), the system MUST list all matches and exit with code 1.
4. If an entry ID cannot be resolved, the system MUST exit with code 1 and suggest using `list` or `search`.
5. If a template does not exist, the system MUST exit with code 1 and list available templates.
6. If date parsing fails, the system MUST exit with code 1 and show the expected format.
7. Malformed frontmatter in existing entries MUST be reported as warnings but not prevent other operations.
8. If the `.encrypted` marker file exists, any command that reads or writes the wiki directory MUST exit with: "Wiki is encrypted. Run 'regimen decrypt' first."
9. If `regimen encrypt` is run on an already encrypted wiki, it MUST exit with error.
10. If `regimen decrypt` is run on an unencrypted wiki, it MUST exit with error.
11. If decryption fails due to wrong passphrase, the system MUST exit with: "Decryption failed: incorrect passphrase or corrupted file."

## 9. Examples

```bash
# Quick note
regimen note add "Had an idea for improving the login flow"

# Add with tags
regimen note add "Sprint planning meeting" --tags "meeting,work,planning"

# Open editor for detailed entry
regimen note add

# Create floating note
regimen note add --floating "Research: Distributed caching strategies"

# Reference a task
regimen note add "Discussed @task:a1b2c3 in standup today" --tags "work"

# View today's entries
regimen note today

# View weekly summary
regimen note week

# Search for entries
regimen note search "API design"
regimen note search "authentication" --after 2026-01-01

# Search with tags
regimen note search "meeting" --tags "work"

# List entries with specific tag
regimen note tags meeting

# Add tags to existing entry
regimen note tag a1b2c3 "important,follow-up"

# Random idea surfacing
regimen note random
regimen note random --tag idea --count 3

# Generate report
regimen note report
regimen note report --days 3

# View statistics
regimen note stats

# Use template
regimen note add --template meeting

# Create custom template
regimen note template create retrospective

# Encrypt all `.md` and `.json` files in wiki
regimen encrypt
# Or with passphrase from stdin
echo "my-passphrase" | regimen encrypt --passphrase-stdin

# Decrypt entire wiki
regimen decrypt

# Commands fail when encrypted
regimen note add "test"  # Error: Wiki is encrypted. Run 'regimen decrypt' first.
```

## 10. Security Considerations

1. The system MUST NOT transmit wiki data over the network.
2. Wiki files MAY contain sensitive personal or work information. Documentation MUST warn users to:
   - Set appropriate file permissions on the wiki directory
   - Avoid committing sensitive data to public repositories
   - Use encryption when storing on cloud-synced directories
3. Encryption MUST use authenticated encryption (AES-256-GCM) and a modern KDF (Argon2id).
4. Decryption MUST fail completely if authentication tag verification fails.
5. Passphrases provided via stdin MUST be securely cleared from memory after use.
6. Interactive passphrase prompts MUST NOT echo characters to the terminal.
7. Entry IDs MUST be generated using a cryptographically secure random source to prevent collisions.
8. When opening editors, the system MUST validate `$EDITOR` to prevent command injection.
9. Template processing MUST NOT execute arbitrary code from template files.
10. The `.encrypted` marker file prevents accidental operations on encrypted data.

## 11. Testing Considerations

Test scenarios SHOULD include:

- Entry creation (inline text vs editor)
- Daily vs floating entry storage structure
- Tag parsing from both flags and content
- Search with AND/OR logic
- Date range filtering
- ID prefix resolution (unique and ambiguous)
- Random selection distribution
- Report generation with various time ranges
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
