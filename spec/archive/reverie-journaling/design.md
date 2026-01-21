# Design: reverie-journaling
**Specification Reference**: [specification.md](specification.md)
Status: Draft

## 1. Context

The `regimen note` command group extends the existing `regimen` executable with personal journaling and note-taking capabilities. It complements `regimen` task management by focusing on unstructured capture rather than structured tasks. While `regimen` manages goals with defined states (open/complete), notes handle freeform thoughts, daily logs, and ideas.

The sleepless project already uses Markdown with YAML frontmatter for task storage in `~/wiki/tasks/`. The note system adopts the same pattern for consistency, storing files in `~/wiki/notes/`, enabling interoperability with tools like Obsidian, vim, and VS Code.

Key constraints:
- Must work offline without network dependencies
- Storage format must be human-readable Markdown
- Must support both date-based entries (daily notes) and standalone notes (floating ideas)
- Full-text search must be fast for large note collections (thousands of entries)
- Must integrate with user's preferred editor via `$EDITOR`
- Must coexist with existing regimen task management
- Wiki-wide encryption/decryption affects all markdown files (tasks and notes)

## 2. Goals and Non-Goals

### Goals

- Integrate journaling into `regimen` as a `note` command group
- Provide fast capture for fleeting thoughts (`regimen note add "idea"`)
- Support daily journaling with date-based file naming in `~/wiki/notes/`
- Enable floating notes for timeless ideas and references (same directory, ID-based naming)
- Implement full-text search across all entries
- Support tag-based organization and filtering
- Surface random past entries for rediscovery
- Generate reports from recent work entries
- Support wiki-wide encryption/decryption via `regimen encrypt/decrypt`
- Allow notes to reference tasks using `@task:<id>` syntax (for future integration)

### Non-Goals

- Rich text editing (Markdown only)
- Synchronization or cloud backup (use git or syncthing separately)
- Attachments or media embedding (link to external files)
- Mobile app or web interface
- Real-time collaboration
- Automatic/transparent encryption (explicit encrypt/decrypt commands only)
- Per-file or selective encryption (entire wiki is encrypted or not)

## 3. Options Considered

### Option 1: Nested Date Folders

Daily entries in nested folders (`~/wiki/notes/YYYY/MM/DD.md`), floating notes in a separate subdirectory (`~/wiki/notes/floating/<id>.md`).

**Advantages**:
- Clear separation between daily and floating notes
- Traditional journal structure with year/month hierarchy
- Easy to browse by date in file manager

**Disadvantages**:
- Deep directory nesting
- More complex file path handling
- Harder to list all notes together

**Complexity**: Medium

### Option 2: Flat Structure with Naming Convention

All files in `~/wiki/notes/` root using date naming for daily entries (`2026-01-20.md`) and IDs for floating notes (`abc123.md`).

**Advantages**:
- Simple flat directory structure
- All notes in one place
- Easy to list and search
- No deep nesting
- Date-based files naturally sort first

**Disadvantages**:
- All files mixed together
- Relies on naming convention for organization

**Complexity**: Low

### Option 3: Subdirectories for Each Type

Daily notes in `~/wiki/notes/daily/` and floating notes in `~/wiki/notes/ideas/` or similar.

**Advantages**:
- Clear separation of concerns
- Can use different subdirectory names for semantic clarity
- Organized without deep nesting

**Disadvantages**:
- Still requires managing two directories
- More complex path handling

**Complexity**: Medium

## 4. Decision

**Chosen Option**: Option 2 (Flat Structure with Naming Convention)

**Rationale**: The flat structure is simplest and most flexible. All notes live in `~/wiki/notes/` with daily entries using date naming (`YYYY-MM-DD.md`) and floating notes using ID naming (`<id>.md`). This makes listing, searching, and managing all notes straightforward.

**Key Factors**:
1. Simplicity - single directory to manage
2. Flexibility - easy to list all notes or filter by name pattern
3. Natural sorting - date-based files sort chronologically
4. Integration - works well alongside `~/wiki/tasks/` directory
5. User requested flat structure

## 5. Detailed Design

### Architecture Overview

```
internal/regimen/
  note.go                   Note command group setup
  note_add.go               Add/edit/delete commands
  note_view.go              Today/week/month/show/list commands
  note_search.go            Search command
  note_tags.go              Tag management commands
  note_random.go            Random surfacing command
  note_report.go            Report generation command
  note_stats.go             Statistics command
  note_template.go          Template management commands
  encrypt.go                Wiki encryption command
  decrypt.go                Wiki decryption command
pkg/notes/
  store.go                  Note storage abstraction
  entry.go                  Entry model and parsing
  daily.go                  Daily file operations (YYYY-MM-DD.md)
  floating.go               Floating note operations (<id>.md)
  index.go                  Tag and search indexing
  taskref.go                Task reference parsing (@task:<id> syntax)
pkg/crypto/
  wiki.go                   Wiki-wide encryption/decryption (AEAD + KDF)
pkg/markdown/
  frontmatter.go            YAML frontmatter parsing (shared with tasks)
  parser.go                 Markdown parsing for inline tags
```

### Data Flow

**Wiki Directory Resolution**

All commands that read/write the wiki MUST resolve the wiki directory using:

1. `--wiki-dir` persistent flag (highest precedence)
2. `REGIMEN_WIKI_DIR` environment variable
3. Default `~/wiki`

Resolved paths:

- Notes directory: `<wikiDir>/notes/`
- Templates directory: `<wikiDir>/.templates/`
- Tasks directory (existing): `<wikiDir>/tasks/`

**Encrypted Wiki Gate**

```
command invocation
  -> resolve wikiDir
  -> if <wikiDir>/.encrypted exists:
       if command is `decrypt`: allow
       else: return error "Wiki is encrypted..."
```

**Daily Note Add (inline text)**

```
regimen note add "text" [--date D] [--tags ...]
  -> resolve wikiDir
  -> open/create <wikiDir>/notes/D.md
  -> append "## HH:MM" + text + inline tags
  -> re-extract tags; update frontmatter
  -> atomic write
```

**Wiki Encrypt**

```
regimen encrypt
  -> resolve wikiDir
  -> ensure <wikiDir>/.encrypted does not exist
  -> prompt/confirm passphrase (or --passphrase-stdin)
  -> create <wikiDir>/.encrypted (JSON: salt + argon2 params)
  -> derive key once (argon2id)
  -> for each eligible file (.md/.json), skipping .git and symlinks:
       read plaintext
       encrypt bytes using AES-256-GCM (AAD=relative path)
       write temp + rename to <path>.enc
       remove original plaintext file
  -> exit 0 if no failures, else exit non-zero with summary
```

### Component Design

The design separates the system into:

- **Command layer** (`internal/regimen/*`): parse flags, validate inputs, format output, call package APIs.
- **Domain/storage layer** (`pkg/notes/*`): read/write notes, parse frontmatter, tags, search.
- **Crypto layer** (`pkg/crypto/wiki.go`): encrypt/decrypt arbitrary files under wiki.

Commands SHOULD keep file-system details inside `pkg/*` packages.

**Note Store (`pkg/notes/store.go`)**:
- Abstracts file system operations
- Handles wiki directory configuration (--wiki-dir flag > REGIMEN_WIKI_DIR env > ~/wiki)
- Provides unified entry access regardless of storage type
- Checks for `.encrypted` marker and fails if wiki is encrypted

Suggested Go API:

```go
type Store struct {
  WikiDir      string
  NotesDir     string
  TemplatesDir string
}

func NewStore(wikiDir string) (*Store, error)

func (s *Store) EnsureStructure() error
func (s *Store) IsEncrypted() bool

// Daily
func (s *Store) DailyPath(date string) (string, error)
func (s *Store) LoadDaily(date string) (*Note, error)
func (s *Store) SaveDaily(note *Note) error
func (s *Store) DeleteDaily(date string) error

// Floating
func (s *Store) FloatingPath(id string) (string, error)
func (s *Store) LoadFloating(idOrPrefix string) (*Note, error)
func (s *Store) SaveFloating(note *Note) error
func (s *Store) DeleteFloating(idOrPrefix string) error
```

**Error Handling**

The store SHOULD return typed errors so the command layer can format consistent messages:

- `ErrWikiEncrypted`
- `ErrNotFound`
- `ErrAmbiguousID`
- `ErrInvalidDate`
- `ErrInvalidFrontmatter`

**Entry Model (`pkg/notes/entry.go`)**:
- Represents a single entry with ID, content, tags, timestamps
- Parses from and serializes to Markdown with frontmatter
- Extracts inline tags from content
- Identifies and parses `@task:<id>` references (stored but not actioned)

Suggested data model:

```go
type NoteType string

const (
  NoteTypeDaily    NoteType = "daily"
  NoteTypeFloating NoteType = "floating"
)

type Note struct {
  Type    NoteType
  ID      string    // floating only
  Date    string    // daily only (YYYY-MM-DD)
  Created time.Time
  Updated time.Time
  Tags    []string
  Body    string // markdown (without frontmatter)
}
```

Notes MUST be round-trippable: parse -> serialize -> parse yields equivalent fields.

**Daily File Manager (`pkg/notes/daily.go`)**:
- Manages `YYYY-MM-DD.md` files in `~/wiki/notes/`
- One file per day; file content may contain multiple timestamped sections
- Appends `## HH:MM` sections for inline `note add <text>`
- Opens full file in editor for `note add` (no text) and `note edit --date`

**Floating Note Manager (`pkg/notes/floating.go`)**:
- Manages `<id>.md` files in `~/wiki/notes/`
- One entry per file
- Simple CRUD operations

**Search Index (`pkg/notes/index.go`)**:
- Builds in-memory index for fast search
- Indexes entry content, tags, metadata, and task references
- Supports case-insensitive substring matching

**Task Reference Parser (`pkg/notes/taskref.go`)**:
- Parses `@task:<id>` syntax from note content
- Stores references in entry metadata
- Reserved for future bidirectional linking

The parser MUST:

- Match literal `@task:` prefix followed by an ID/prefix token
- Permit IDs that match regimen task IDs (currently hex-ish short IDs)
- Return a de-duplicated list of referenced IDs/prefixes
- Not attempt to resolve or validate the task during parsing

**Wiki Encryption (`pkg/crypto/wiki.go`)**:
- Encrypts all `.md` and `.json` files recursively in wiki directory
- Adds `.enc` extension to encrypted files
- Creates `.encrypted` marker file
- Uses AES-256-GCM with Argon2id KDF
- Passphrase from stdin or interactive prompt
- Best-effort operation with per-file error reporting
- Skips `wiki/.git/` and does not follow symlinks

The crypto package MUST be independent of the note/task domain; it encrypts arbitrary files.

Suggested API:

```go
type Marker struct {
  Version        int    `json:"version"`
  Created        string `json:"created"`
  SaltB64        string `json:"salt_b64"`
  Argon2Time     int    `json:"argon2_time"`
  Argon2MemoryKiB int   `json:"argon2_memory_kib"`
  Argon2Threads  int    `json:"argon2_threads"`
}

func EncryptWiki(wikiDir string, passphrase []byte) (EncryptReport, error)
func DecryptWiki(wikiDir string, passphrase []byte) (DecryptReport, error)
```

Where `EncryptReport` and `DecryptReport` include:

- total files scanned
- files encrypted/decrypted
- list of per-file failures

**Editor Integration**:
- Uses `$EDITOR` environment variable (fallback: `vim` on Unix, `notepad` on Windows)
- Creates temporary file with frontmatter template
- Parses result when editor closes
- Discards empty entries

The command layer SHOULD implement editor integration once and reuse for notes and tasks.

### Edge Cases

- **Date parsing**: reject invalid YYYY-MM-DD, including impossible dates.
- **Timezone**: timestamp headings use local time; frontmatter timestamps use RFC3339 with offset.
- **Tag normalization**: normalize to lowercase; strip leading `#` if present.
- **Daily note missing**: `show --date` should error with not found; `add` should create.
- **Encryption partial state**: if `.encrypted` exists but plaintext files remain, commands still treat wiki as encrypted.
- **Symlinks**: skip symlinked files/dirs to avoid encrypting outside wiki.
- **.enc collisions**: if `<file>.enc` already exists during encrypt, treat as failure and continue.

### Data Design

**Daily Note File Structure** (`~/wiki/notes/2026-01-18.md`):
```markdown
---
date: 2026-01-18
created: 2026-01-18T10:30:00Z
updated: 2026-01-18T14:00:00Z
tags: [work, meeting]
type: daily
---

# 2026-01-18

## 10:30

Team discussion about API design. Following up on @task:abc123 from yesterday.

#work #meeting

## 14:00

Idea: cache invalidation using event sourcing.

#idea
```

**Floating Note File Structure** (`~/wiki/notes/a1b2c3d4.md`):
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
| In-memory search index | Fast search | Memory for large note collections | Personal notes rarely exceed 10K entries |
| Flat file structure | Simplicity, easy search | No visual hierarchy | All in one place; date naming provides natural sort |
| Integration into regimen | Unified tool | Larger executable | Better UX - one tool for tasks and notes |
| YAML frontmatter | Machine-readable metadata | File header noise | Standard pattern (Jekyll, Hugo, Obsidian) |
| Inline tags in content | Natural writing flow | Duplicate data (frontmatter + content) | Sync on save; frontmatter is source of truth |
| Wiki-wide encrypt/decrypt | Simple all-or-nothing security | No selective encryption | Clear mental model; prevents mixed state bugs |
| .enc extension for encrypted files | Easy to identify encrypted state | Filename changes | Clear visual indicator; marker file tracks state |
| Manual decrypt required | Explicit security model | Convenience | Prevents accidental plaintext leakage |
| @task:<id> syntax reserved | Future integration path | No immediate behavior | Design for future; syntax won't break later |

## 7. Cross-Cutting Concerns

### Security

- Wiki files may contain sensitive personal and work information
- Recommend 0700 permissions on wiki root
- Wiki-wide encryption via explicit `regimen encrypt` command
- Encrypted files use `.enc` extension; `.encrypted` marker tracks state
- Passphrase from `--passphrase-stdin` or interactive prompt (never written to disk)
- Use AEAD (AES-256-GCM) with KDF (Argon2id via `golang.org/x/crypto`)
- Fail closed on authentication tag mismatch
- All commands block when `.encrypted` marker exists (must decrypt first)
- Encryption applies to `.md` and `.json` files (including regimen metadata)
- Entry IDs use crypto/rand to prevent collision attacks
- Validate `$EDITOR` to prevent command injection
- No environment variable for passphrase (reduces exposure risk)

### Performance

- In-memory index built on first search, cached for session
- Daily file parsing is fast (one entry per file in flat structure)
- Flat directory structure allows simple file listing (no recursive traversal)
- Stats command scans all files (acceptable for personal use; can optimize with cache if needed)
- Random selection uses pre-built entry list
- Encryption/decryption processes all .md files; acceptable for one-time operation

### Reliability

- Atomic writes using temp file + rename
- Frontmatter validation on load (warn on malformed, don't crash)
- Editor integration handles empty saves gracefully
- Concurrent access not supported (single-user tool)

### Testing

- Unit tests for entry parsing and serialization
- Unit tests for tag extraction from content
- Unit tests for search matching
- Unit tests for report generation
- Integration tests for add/edit/delete workflows
- Integration tests for date range queries

## 8. Implementation Plan

See [implementation.md](implementation.md) for phased task breakdown.

### Migration Strategy

Not applicable (new feature, no existing data to migrate).

## 9. Open Questions

- Should the search index be persisted to disk for faster startup with large note collections?
- Should inline tags (`#tag`) automatically be added to frontmatter, or should they be independent?
- Should `regimen note week` show a calendar view or a list view?
- Should report generation support custom templates?
- Should there be a `regimen note import` command to import from other journaling tools?
- When should `@task:<id>` references become bidirectional links (future phase)?
- Should encrypted files be compatible with standard tools (age, gpg) or use custom format?
