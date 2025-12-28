# Design: goals (core goal/task management)
**Specification Reference**: [specification.md](specification.md)

**Status**: Draft (documentation of existing implementation)

## 1. Context

The existing Sleepless implementation is a Cobra-based CLI where the primary feature set is the `sleepless goals` command group. Persistence is file-based and local-first: tasks are stored as human-editable Markdown files under a wiki directory and are parsed/serialized by an internal parser.

The design prioritizes simplicity, manual editability, and low coupling. Topics are represented as separate Markdown files to keep files small, enable manual browsing, and allow the CLI to load only what it needs (topic-by-topic) when desired.

Non-functional constraints include: predictable and fast ID lookup via prefix matching, safe-ish concurrent access via file locks (where used), and a pleasant terminal UX via formatting helpers.

## 2. Goals and Non-Goals

### Goals

- Provide a terminal-first CRUD experience for tasks/goals.
- Store tasks in a Markdown format intended for human editing.
- Support hierarchical subtasks.
- Support grouping by topic, and moving tasks between topics.
- Provide convenience views: tree, progress, deps, calendar.
- Provide a lightweight audit log of changes.

### Non-Goals

- Multi-user sync/collaboration.
- A full project management system (dependencies beyond parent-child).
- A database-backed store.
- Perfect preservation of arbitrary hand-edited Markdown (the parser expects a specific structure).

## 3. Options Considered

### Option 1: Markdown files per topic (implemented)

Store each topic in a separate `<topic>.md` file with a strict-ish task list structure.

**Advantages**:
- Human-editable.
- Works naturally with wiki tooling.
- Easy to version control.
- Limits file size per topic.

**Disadvantages**:
- Parsing/writing is more complex than JSON.
- Manual edits can break parsing if the structure is not followed.

**Complexity**: Medium

### Option 2: Structured data store (JSON/SQLite)

Store tasks in a structured database or JSON file and generate views.

**Advantages**:
- Strong schema and easier querying.
- Easier to guarantee data integrity.

**Disadvantages**:
- Not naturally wiki-integrated.
- Harder to manually edit.

**Complexity**: Medium–High

## 4. Decision

**Chosen Option**: Option 1 (Markdown per topic)

**Rationale**: The project’s core value is wiki integration and human-editable storage. Markdown-per-topic provides a good balance of editability, navigability, and separation.

**Key Factors**:
1. Human-readable storage.
2. Simple local-first setup.
3. Natural integration with Markdown/wiki tooling.

## 5. Detailed Design

### Architecture Overview

- Cobra command tree rooted at `sleepless`.
- `cmd/` defines subcommands under `goals` (`add`, `list`, `done`, `edit`, `remove`, `move`, `search`, `remind`, `archive`, `history`, `view`).
- `pkg/task` defines the in-memory model and filtering helpers.
- `pkg/parser` parses Markdown into tasks and serializes tasks back to Markdown.
- `pkg/storage` manages filesystem structure, topic files, index updates, and history.
- `pkg/ui` handles output formatting and styling.

### Component Design

#### `pkg/task`

- Defines `Task` with nested subtasks.
- Generates IDs using crypto randomness and hex encoding.
- Provides filter/group helpers for the CLI.

#### `pkg/parser`

- Parses a constrained Markdown subset:
  - Task line: `- [ ] Title {#id}` (or `[x]`)
  - Metadata: `- key: value`
  - Notes: `- Note: ...`
- Uses indentation to build parent-child relationships.
- Writes a normalized Markdown representation.

Known implementation detail worth documenting:
- `created` is written only for top-level tasks in the current serializer. Subtasks do not include `created` unless manually added.

#### `pkg/storage`

- Default store path: `~/wiki/tasks`.
- Ensures `inbox.md` and `.task-meta.json` exist.
- Creates topic files on demand and updates `index.md`.
- Uses file locks (`flock`) for writing topic files and the meta file.
- Implements `ArchiveTask` by moving completed tasks into `archived.md`.

### Data Design

#### Task files

- One Markdown file per topic, named `<topic>.md`.
- A special `archived.md` for completed tasks.
- `index.md` linking to topics.

#### History

- `.task-meta.json` holds a bounded array of history entries (max 1000).

### API Design (CLI)

- Cobra subcommands map directly to user actions.
- Commands generally:
  1. Ensure store structure
  2. Load tasks (topic or all)
  3. Resolve ID via prefix
  4. Mutate in-memory
  5. Persist to Markdown
  6. Append history entry

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| Markdown storage | Human editing + wiki integration | Parsing constraints | Core product goal |
| Prefix IDs | Fast referencing | Ambiguity handling required | Improves UX in terminal |
| Topic files | Smaller files, clearer organization | Cross-topic queries require scanning | Acceptable for local CLI |

## 7. Cross-Cutting Concerns

### Security

- Local-first; no networking.
- Users may store sensitive data; avoid leaking task contents in unexpected logs.

### Performance

- `FindAllByPrefix` loads all topics to find matches; acceptable for moderate task counts.
- Views and searches scan in-memory lists.

### Reliability

- File locking reduces risk of corruption during concurrent writes.
- Manual edits can break parsing; CLI rewrites normalize formatting.

### Testing

- Parser round-trip tests are the highest value.
- Storage behavior tests (archive/move/index update/history append).

## 8. Implementation Plan

This is existing functionality; the plan here documents how the system is (already) structured:

- Phase 1: Implement task model, ID generation, and filters.
- Phase 2: Implement Markdown parser and serializer.
- Phase 3: Implement storage layer with topic files + index + archive.
- Phase 4: Implement goals command group (CRUD + search + reminders + views).
- Phase 5: Add history log.

### Migration Strategy

None.

## 9. Open Questions

- Should subtasks persist `created` timestamps (currently only top-level tasks do)?
- Should all commands return non-zero exit codes on validation failures (currently many errors are printed but the process may still exit 0)?
- Tree view connector rendering has a TODO for last-child formatting.
