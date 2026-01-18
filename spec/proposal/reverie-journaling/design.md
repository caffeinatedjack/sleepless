# Design: reverie-journaling
**Specification Reference**: [specification.md](specification.md)
Status: Draft

## 1. Context

Reverie is a new standalone executable for personal journaling and note-taking. It complements regimen (task management) by focusing on unstructured capture rather than structured tasks. While regimen manages goals with defined states (open/complete), reverie handles freeform thoughts, daily logs, and ideas.

The sleepless project already uses Markdown with YAML frontmatter for task storage (regimen). Reverie adopts the same pattern for consistency, enabling interoperability with tools like Obsidian, vim, and VS Code.

Key constraints:
- Must work offline without network dependencies
- Storage format must be human-readable Markdown
- Must support both date-based entries (journals) and standalone notes (ideas)
- Full-text search must be fast for large journals (thousands of entries)
- Must integrate with user's preferred editor via `$EDITOR`

## 2. Goals and Non-Goals

### Goals

- Provide fast capture for fleeting thoughts (`reverie add "idea"`)
- Support daily journaling with date-based file organization
- Enable floating notes for timeless ideas and references
- Implement full-text search across all entries
- Support tag-based organization and filtering
- Surface random past entries for rediscovery
- Generate standup reports from recent entries

### Non-Goals

- Rich text editing (Markdown only)
- Synchronization or cloud backup (use git or syncthing separately)
- Encryption (use file-system encryption or dedicated tools)
- Attachments or media embedding (link to external files)
- Mobile app or web interface
- Real-time collaboration

## 3. Options Considered

### Option 1: Single File Per Day

Store all entries for a day in a single Markdown file (`YYYY/MM/DD.md`). Multiple entries on the same day are appended with timestamps.

**Advantages**:
- Matches traditional journaling (one file = one day)
- Easy to read a day's entries as a continuous document
- Simple file structure
- Compatible with daily notes in Obsidian

**Disadvantages**:
- Each entry lacks its own file (harder to link)
- Must parse file to extract individual entries

**Complexity**: Low

### Option 2: One File Per Entry

Store each entry as a separate file with a unique ID. Use directory structure for organization.

**Advantages**:
- Each entry is independently addressable
- Easier to move/reorganize entries
- Cleaner ID-to-file mapping

**Disadvantages**:
- Many small files for active journalers
- File system clutter
- Harder to read a day's entries together

**Complexity**: Medium

### Option 3: Hybrid Approach

Daily entries go into date-based files (`YYYY/MM/DD.md`), floating notes go into `notes/<id>.md`.

**Advantages**:
- Best of both worlds
- Daily journals are cohesive documents
- Standalone ideas are independently addressable
- Matches how people naturally think about journals vs notes

**Disadvantages**:
- Two different storage patterns to maintain
- Slightly more complex ID resolution

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 3 (Hybrid Approach)

**Rationale**: The dual model reflects how people naturally organize thoughts. Daily logs belong together chronologically; ideas and references exist independently of time. This matches the specification's "daily entry" vs "floating note" terminology.

**Key Factors**:
1. Matches user mental model of journals vs notes
2. Daily files are human-readable without tooling
3. Floating notes are easy to link and reference
4. Compatible with Obsidian's daily notes + zettelkasten pattern

## 5. Detailed Design

### Architecture Overview

```
cmd/reverie/main.go         Entry point
internal/reverie/
  root.go                   Root command setup
  add.go                    Add/edit/delete commands
  view.go                   Today/week/month/show/list commands
  search.go                 Search command
  tags.go                   Tag management commands
  random.go                 Random surfacing command
  standup.go                Standup generation command
  stats.go                  Statistics command
  template.go               Template management commands
pkg/journal/
  store.go                  Journal storage abstraction
  entry.go                  Entry model and parsing
  daily.go                  Daily file operations
  floating.go               Floating note operations
  index.go                  Tag and search indexing
pkg/markdown/
  frontmatter.go            YAML frontmatter parsing
  parser.go                 Markdown parsing for inline tags
```

### Component Design

**Journal Store (`pkg/journal/store.go`)**:
- Abstracts file system operations
- Handles journal root configuration
- Provides unified entry access regardless of storage type

**Entry Model (`pkg/journal/entry.go`)**:
- Represents a single entry with ID, content, tags, timestamps
- Parses from and serializes to Markdown with frontmatter
- Extracts inline tags from content

**Daily File Manager (`pkg/journal/daily.go`)**:
- Manages `YYYY/MM/DD.md` files
- Appends entries with timestamps
- Parses multiple entries from a single file

**Floating Note Manager (`pkg/journal/floating.go`)**:
- Manages `notes/<id>.md` files
- One entry per file
- Simple CRUD operations

**Search Index (`pkg/journal/index.go`)**:
- Builds in-memory index for fast search
- Indexes entry content, tags, and metadata
- Supports case-insensitive substring matching

**Editor Integration**:
- Uses `$EDITOR` environment variable (fallback: `vim` on Unix, `notepad` on Windows)
- Creates temporary file with frontmatter template
- Parses result when editor closes
- Discards empty entries

### Data Design

**Daily Entry File Structure** (`~/journal/2026/01/18.md`):
```markdown
---
entries:
  - id: a1b2c3d4
    created: 2026-01-18T10:30:00Z
    tags: [work, meeting]
  - id: e5f6g7h8
    created: 2026-01-18T14:00:00Z
    tags: [idea]
---

# 2026-01-18

## 10:30 {#a1b2c3d4}

Team standup discussion about API design.

#work #meeting

## 14:00 {#e5f6g7h8}

Idea: cache invalidation using event sourcing.

#idea
```

**Floating Note File Structure** (`~/journal/notes/a1b2c3d4.md`):
```markdown
---
id: a1b2c3d4
created: 2026-01-18T16:45:00Z
updated: 2026-01-19T09:00:00Z
tags: [architecture, research]
type: floating
---

# API Gateway Caching Strategy

Research notes on distributed caching approaches...

#architecture #research
```

**ID Generation**:
- 8-character lowercase hexadecimal (32 bits)
- Generated using crypto/rand
- Collision probability negligible for personal journals

### API Design

Not applicable (CLI-only).

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| In-memory search index | Fast search | Memory for large journals | Personal journals rarely exceed 10K entries |
| Hybrid storage | Matches mental model | Implementation complexity | Clear separation of concerns |
| YAML frontmatter | Machine-readable metadata | File header noise | Standard pattern (Jekyll, Hugo, Obsidian) |
| Inline tags in content | Natural writing flow | Duplicate data (frontmatter + content) | Sync on save; frontmatter is source of truth |

## 7. Cross-Cutting Concerns

### Security

- Journal files may contain sensitive personal information
- Recommend 0700 permissions on journal root
- Document risks of cloud sync without encryption
- Entry IDs use crypto/rand to prevent collision attacks
- Validate `$EDITOR` to prevent command injection

### Performance

- In-memory index built on first search, cached for session
- Daily file parsing is fast (single file per day, small content)
- Stats command may need to scan all files (optimize with cached metadata if needed)
- Random selection uses pre-built entry list

### Reliability

- Atomic writes using temp file + rename
- Frontmatter validation on load (warn on malformed, don't crash)
- Editor integration handles empty saves gracefully
- Concurrent access not supported (single-user tool)

### Testing

- Unit tests for entry parsing and serialization
- Unit tests for tag extraction from content
- Unit tests for search matching
- Unit tests for standup generation
- Integration tests for add/edit/delete workflows
- Integration tests for date range queries

## 8. Implementation Plan

See [implementation.md](implementation.md) for phased task breakdown.

### Migration Strategy

Not applicable (new executable).

## 9. Open Questions

- Should the search index be persisted to disk for faster startup with large journals?
- Should inline tags (`#tag`) automatically be added to frontmatter, or should they be independent?
- Should `reverie week` show a calendar view or a list view?
- Should standup generation support custom templates?
- Should there be a `reverie import` command to import from other journaling tools?
