# Implementation: reverie-journaling

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Implement `reverie` as a new standalone executable with hybrid storage: daily entries in `YYYY/MM/DD.md` files and floating notes in `notes/<id>.md`. Uses Markdown with YAML frontmatter for all entries. Provides full-text search, tag-based filtering, random surfacing, and standup generation.

## Prerequisites

- None (new executable, no dependencies on existing sleepless code)
- Uses existing YAML parsing from gopkg.in/yaml.v3

## Phases

### Phase 1: Core Infrastructure

**Goal**: Establish executable structure, entry model, and storage abstraction.

- [ ] 1.1 Create `cmd/reverie/main.go` entry point
- [ ] 1.2 Create `internal/reverie/root.go` with Cobra root command
- [ ] 1.3 Implement `pkg/journal/entry.go` entry model
- [ ] 1.4 Implement `pkg/markdown/frontmatter.go` for YAML frontmatter parsing
- [ ] 1.5 Implement `pkg/journal/store.go` storage abstraction
- [ ] 1.6 Implement ID generation using crypto/rand
- [ ] 1.7 Add unit tests for entry parsing and serialization
- [ ] 1.8 Add unit tests for frontmatter parsing

**Milestone**: Can parse and serialize entries with frontmatter.

### Phase 2: Entry Creation and Editing

**Goal**: Implement add, edit, and delete commands.

- [ ] 2.1 Implement `pkg/journal/daily.go` for daily file operations
- [ ] 2.2 Implement `pkg/journal/floating.go` for floating note operations
- [ ] 2.3 Implement `reverie add <text>` command (inline text)
- [ ] 2.4 Implement `reverie add` (editor mode)
- [ ] 2.5 Implement `--date`, `--tags`, `--floating` flags
- [ ] 2.6 Implement `reverie edit <id>` command
- [ ] 2.7 Implement `reverie delete <id>` command
- [ ] 2.8 Implement ID prefix resolution
- [ ] 2.9 Implement inline tag extraction from content
- [ ] 2.10 Add unit tests for daily file operations
- [ ] 2.11 Add unit tests for floating note operations
- [ ] 2.12 Add integration tests for add/edit/delete workflows

**Milestone**: `reverie add "thought"` creates entry; `reverie add --floating` creates floating note; `reverie edit <id>` opens editor.

### Phase 3: Viewing Commands

**Goal**: Implement commands for viewing entries.

- [ ] 3.1 Implement `reverie today` command
- [ ] 3.2 Implement `reverie show <id>` command
- [ ] 3.3 Implement `reverie list` command with preview
- [ ] 3.4 Implement `--limit`, `--tags`, `--date`, `--range` flags
- [ ] 3.5 Implement `reverie week` command
- [ ] 3.6 Implement `reverie month` command
- [ ] 3.7 Implement `--json` and `--compact` output formats
- [ ] 3.8 Add unit tests for date range calculations
- [ ] 3.9 Add integration tests for viewing commands

**Milestone**: `reverie today` shows today's entries; `reverie list --tags work` filters by tag; `reverie week` shows weekly summary.

### Phase 4: Search and Tags

**Goal**: Implement full-text search and tag management.

- [ ] 4.1 Implement `pkg/journal/index.go` search indexing
- [ ] 4.2 Implement `reverie search <query>` command
- [ ] 4.3 Implement `--or`, `--tags`, `--after`, `--before` flags
- [ ] 4.4 Implement context snippets in search results
- [ ] 4.5 Implement `reverie tags` command (list all tags)
- [ ] 4.6 Implement `reverie tags <tag>` command (entries with tag)
- [ ] 4.7 Implement `reverie tag <id> <tags>` command
- [ ] 4.8 Implement `reverie untag <id> <tags>` command
- [ ] 4.9 Add unit tests for search matching
- [ ] 4.10 Add unit tests for tag operations
- [ ] 4.11 Add integration tests for search

**Milestone**: `reverie search "API"` finds matching entries with context; `reverie tags` lists all tags with counts.

### Phase 5: Special Commands

**Goal**: Implement random surfacing, standup, and statistics.

- [ ] 5.1 Implement `reverie random` command
- [ ] 5.2 Implement `--tag` and `--count` flags for random
- [ ] 5.3 Implement `reverie standup` command
- [ ] 5.4 Implement `--days` flag for standup
- [ ] 5.5 Implement work-tag filtering for standup
- [ ] 5.6 Implement `reverie stats` command
- [ ] 5.7 Calculate entry counts, tag distribution, streaks
- [ ] 5.8 Add unit tests for random selection
- [ ] 5.9 Add unit tests for standup generation
- [ ] 5.10 Add unit tests for statistics calculation

**Milestone**: `reverie random` surfaces a random entry; `reverie standup` generates report from recent work entries; `reverie stats` shows journal statistics.

### Phase 6: Templates

**Goal**: Implement template system for structured entries.

- [ ] 6.1 Create template storage directory (`~/.config/reverie/templates/`)
- [ ] 6.2 Implement `reverie template list` command
- [ ] 6.3 Implement `reverie template create <name>` command
- [ ] 6.4 Implement `reverie add --template <name>` integration
- [ ] 6.5 Implement template placeholders (`{{DATE}}`, `{{PROMPT:...}}`)
- [ ] 6.6 Create built-in templates: standup, meeting, reflection, idea
- [ ] 6.7 Add unit tests for template parsing
- [ ] 6.8 Add integration tests for template usage

**Milestone**: `reverie add --template meeting` creates entry from meeting template with prompts.

## Testing Plan

### Unit Tests

- Entry parsing: valid frontmatter, missing fields, malformed YAML
- Entry serialization: round-trip consistency
- Daily file operations: append entry, parse multiple entries, handle empty file
- Floating note operations: create, read, update, delete
- ID generation: uniqueness, format validation
- ID prefix resolution: unique match, ambiguous match, no match
- Tag extraction: inline tags, frontmatter tags, deduplication
- Search matching: case-insensitivity, multiple terms, AND/OR logic
- Date calculations: today, week boundaries, month boundaries
- Standup generation: filtering by tags, grouping by day
- Random selection: distribution, tag filtering
- Statistics: counting, streak calculation

### Integration Tests

- Full add workflow: add → verify file created → verify content
- Edit workflow: add → edit → verify changes persisted
- Delete workflow: add → delete → verify file removed/updated
- Search workflow: add multiple → search → verify results
- Daily file accumulation: add multiple entries same day → verify single file
- Cross-day entries: add entries across days → verify file structure
- Editor integration: mock editor, verify entry parsing

## Rollback Plan

- Remove `cmd/reverie/`, `internal/reverie/`, `pkg/journal/`, `pkg/markdown/`
- Remove `reverie` from Makefile build targets

## Open Questions

- Should templates support conditional sections?
- Should `reverie import` support jrnl format for migration?
- Should there be a `reverie export` command for backup/sharing?
