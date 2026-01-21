# Implementation: reverie-journaling

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Implement `regimen note` command group within the existing `regimen` executable with flat file storage: daily entries as `YYYY-MM-DD.md` and floating notes as `<id>.md` in `~/wiki/notes/`. Uses Markdown with YAML frontmatter for all entries. Provides full-text search, tag-based filtering, random surfacing, report generation, and wiki-wide encryption/decryption. Notes can reference tasks using `@task:<id>` syntax.

Daily notes are addressed by date and represent the entire day; multiple `note add <text>` calls append timestamped sections to the same daily file.

## Prerequisites

- Existing `regimen` executable and command structure
- YAML parsing from gopkg.in/yaml.v3 (already in use)
- Cobra CLI framework (already in use)

## Phases

### Phase 1: Core Infrastructure

**Goal**: Establish note command structure, entry model, and storage abstraction.

- [x] 1.1 Add global `--wiki-dir` flag + env resolution helper
- [x] 1.2 Add wiki encryption gate helper (detect `<wikiDir>/.encrypted`)
- [x] 1.3 Wire wiki-dir + encryption gate into existing wiki-data command groups (goals, recipes)
- [x] 1.4 Create `internal/regimen/note.go` with note command group
- [x] 1.5 Implement `pkg/notes/entry.go` entry model
- [x] 1.6 Update `pkg/markdown/frontmatter.go` to support note metadata (if needed)
- [x] 1.7 Implement `pkg/notes/store.go` storage abstraction with wiki directory configuration
- [x] 1.8 Implement ID generation using crypto/rand
- [x] 1.9 Implement `.encrypted` marker file check in notes store
- [x] 1.10 Add unit tests for entry parsing and serialization
- [x] 1.11 Add unit tests for frontmatter parsing

**Milestone**: Wiki directory resolution exists; wiki-data commands are blocked when encrypted; can parse/serialize notes with frontmatter.

### Phase 2: Entry Creation and Editing

**Goal**: Implement add, edit, and delete commands with flat file structure.

- [x] 2.1 Implement `pkg/notes/daily.go` for daily file operations (YYYY-MM-DD.md)
- [x] 2.2 Implement `pkg/notes/floating.go` for floating note operations (<id>.md)
- [x] 2.3 Implement `regimen note add <text>` command (daily append with timestamp heading)
- [x] 2.4 Implement `regimen note add` (editor mode; opens daily note)
- [x] 2.5 Implement `--date`, `--tags`, `--floating`, `--daily` flags
- [x] 2.6 Implement `regimen note edit <id>` command (floating)
- [x] 2.7 Implement `regimen note edit --date <date>` command (daily)
- [x] 2.8 Implement `regimen note delete <id>` command (floating)
- [x] 2.9 Implement `regimen note delete --date <date>` command (daily)
- [x] 2.10 Implement ID prefix resolution (floating notes)
- [x] 2.11 Implement inline tag extraction from content
- [x] 2.12 Implement `@task:<id>` reference parsing in `pkg/notes/taskref.go`
- [x] 2.13 Add unit tests for daily file operations
- [x] 2.14 Add unit tests for floating note operations
- [x] 2.15 Add unit tests for task reference parsing
- [x] 2.16 Add integration tests for add/edit/delete workflows

**Milestone**: Daily add appends timestamp blocks; floating notes have stable IDs; edit/delete support both modes.

### Phase 3: Viewing Commands

**Goal**: Implement commands for viewing entries.

- [x] 3.1 Implement `regimen note today` command
- [x] 3.2 Implement `regimen note show <id>` (floating) and `regimen note show --date` (daily)
- [x] 3.3 Implement `regimen note list` command with preview
- [x] 3.4 Implement `--limit`, `--tags`, `--date`, `--range` flags
- [x] 3.5 Implement `regimen note week` command
- [x] 3.6 Implement `regimen note month` command
- [x] 3.7 Implement `--json` and `--compact` output formats
- [x] 3.8 Add unit tests for date range calculations
- [x] 3.9 Add integration tests for viewing commands

**Milestone**: `regimen note today` shows today's entries; `regimen note list --tags work` filters by tag; `regimen note week` shows weekly summary.

### Phase 4: Search and Tags

**Goal**: Implement full-text search and tag management.

- [x] 4.1 Implement `pkg/notes/index.go` search indexing
- [x] 4.2 Implement `regimen note search <query>` command
- [x] 4.3 Implement `--or`, `--tags`, `--after`, `--before` flags
- [x] 4.4 Implement context snippets in search results
- [x] 4.5 Implement `regimen note tags` command (list all tags)
- [x] 4.6 Implement `regimen note tags <tag>` command (entries with tag)
- [x] 4.7 Implement `regimen note tag <id> <tags>` command
- [x] 4.8 Implement `regimen note untag <id> <tags>` command
- [x] 4.9 Add task reference search support
- [x] 4.10 Add unit tests for search matching
- [x] 4.11 Add unit tests for tag operations
- [ ] 4.12 Add integration tests for search

**Milestone**: `regimen note search "API"` finds matching entries with context; `regimen note tags` lists all tags with counts; can search for `@task:<id>` references.

### Phase 5: Special Commands

**Goal**: Implement random surfacing, report generation, and statistics.

- [ ] 5.1 Implement `regimen note random` command
- [ ] 5.2 Implement `--tag` and `--count` flags for random
- [ ] 5.3 Implement `regimen note report` command
- [ ] 5.4 Implement `--days` flag for report
- [ ] 5.5 Implement work-tag filtering for report
- [ ] 5.6 Implement `regimen note stats` command
- [ ] 5.7 Calculate entry counts, tag distribution, streaks
- [ ] 5.8 Add unit tests for random selection
- [ ] 5.9 Add unit tests for report generation
- [ ] 5.10 Add unit tests for statistics calculation

**Milestone**: `regimen note random` surfaces a random entry; `regimen note report` generates report from recent work entries; `regimen note stats` shows note statistics.

### Phase 6: Encryption

**Goal**: Implement wiki-wide encryption/decryption commands.

- [ ] 6.1 Add `pkg/crypto/wiki.go` with encrypt/decrypt functions (AEAD + KDF)
- [ ] 6.2 Define encrypted file format (REGIMENENC header + nonce + ciphertext)
- [ ] 6.3 Define `.encrypted` marker JSON (salt + argon2 params)
- [ ] 6.4 Implement `.encrypted` marker file creation/removal
- [ ] 6.5 Implement `regimen encrypt` command with passphrase handling
- [ ] 6.6 Implement `regimen decrypt` command with passphrase handling
- [ ] 6.7 Support `--passphrase-stdin` flag and interactive prompt (confirm on encrypt)
- [ ] 6.8 Encrypt eligible file types: `.md` and `.json` (skip `.git/`, skip symlinks)
- [ ] 6.9 Best-effort operation with per-file errors and summary
- [ ] 6.10 Add encryption state checks to all note/task commands
- [ ] 6.11 Ensure commands fail with helpful error when wiki is encrypted
- [ ] 6.12 Add unit tests for crypto format round-trips
- [ ] 6.13 Add unit tests for authentication failure behavior
- [ ] 6.14 Add integration tests for encrypt/decrypt workflows
- [ ] 6.15 Add tests for command blocking when encrypted

**Milestone**: `regimen encrypt` encrypts all .md files; `regimen decrypt` decrypts them; all commands block when `.encrypted` marker exists.

### Phase 7: Templates

**Goal**: Implement template system for structured entries.

- [ ] 7.1 Create template storage directory (`~/wiki/.templates/`)
- [ ] 7.2 Implement `regimen note template list` command
- [ ] 7.3 Implement `regimen note template create <name>` command
- [ ] 7.4 Implement `regimen note add --template <name>` integration
- [ ] 7.5 Implement template placeholders (`{{DATE}}`, `{{PROMPT:...}}`)
- [ ] 7.6 Create built-in templates: meeting, reflection, idea, report
- [ ] 7.7 Add unit tests for template parsing
- [ ] 7.8 Add integration tests for template usage

**Milestone**: `regimen note add --template meeting` creates entry from meeting template with prompts.

## Testing Plan

### Unit Tests

- Entry parsing: valid frontmatter, missing fields, malformed YAML
- Entry serialization: round-trip consistency
- Daily file operations: create/read/update/delete entry, handle missing file
- Floating note operations: create, read, update, delete
- Task reference parsing: extract @task:<id>, multiple references, invalid formats
- ID generation: uniqueness, format validation
- ID prefix resolution: unique match, ambiguous match, no match
- Tag extraction: inline tags, frontmatter tags, deduplication
- Search matching: case-insensitivity, multiple terms, AND/OR logic
- Date calculations: today, week boundaries, month boundaries
- Report generation: filtering by tags, grouping by day
- Encryption/decryption: full wiki encrypt/decrypt, marker file, passphrase handling
- Command blocking: all commands fail when .encrypted marker exists
- Template processing: placeholder substitution, prompts
- Random selection: distribution, tag filtering
- Statistics: counting, streak calculation

### Integration Tests

- Full add workflow: add → verify file created → verify content
- Edit workflow: add → edit → verify changes persisted
- Delete workflow: add → delete → verify file removed
- Search workflow: add multiple → search → verify results
- Cross-day entries: add entries across days → verify file structure (one file per day)
- Editor integration: mock editor, verify entry parsing
- Encryption workflow: encrypt → verify .enc files → decrypt → verify .md files
- Encrypted state blocking: encrypt → try note add → verify error message

## Rollback Plan

- Remove note command files from `internal/regimen/note*.go`
- Remove `pkg/notes/` directory
- Remove encryption command files `internal/regimen/encrypt.go` and `decrypt.go`
- Remove `pkg/crypto/wiki.go` if not used elsewhere
- Remove template directory `~/wiki/.templates/` (user data - warn first)

## Open Questions

- Should templates support conditional sections?
- Should there be a `regimen note import` command to import from jrnl or other tools?
- Should there be a `regimen note export` command for backup/sharing?
- When should `@task:<id>` become bidirectional (show notes referencing a task)?
- Should encryption support external tools (age, gpg) or custom format only?
- Should there be a `regimen note graph` command to visualize task references?
