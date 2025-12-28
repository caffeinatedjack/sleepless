# Implementation: goals (core goal/task management)

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

This proposal documents the existing implementation of Sleepless’s core goal/task management functionality. The codebase already contains the implementation; this document exists to make the proposal "complete" for documentation and baseline purposes.

## Prerequisites

None (existing behavior).

## Phases

### Phase 1: Core model (`pkg/task`)

**Goal**: Represent goals/subtasks and core metadata.

- [x] Implement `Task` model (`id`, `title`, `status`, `priority`, `due`, `tags`, `notes`, `created`, `completed`, `subtasks`, `topic`).
- [x] Implement ID generation and Short ID display.
- [x] Implement filtering and grouping helpers.

**Milestone**: Tasks can be created in-memory and filtered/grouped deterministically.

### Phase 2: Markdown parser/serializer (`pkg/parser`)

**Goal**: Round-trip tasks to/from a constrained Markdown format.

- [x] Parse `- [ ]` / `- [x]` task lines with `{#id}` tokens.
- [x] Parse metadata (`priority`, `due`, `tags`, `created`, `completed`) and notes.
- [x] Serialize tasks (including nested subtasks) back to Markdown.

**Milestone**: Saving then reloading preserves IDs, nesting, and core metadata.

### Phase 3: Storage layer (`pkg/storage`)

**Goal**: Persist tasks as Markdown under `~/wiki/tasks` with index + history.

- [x] Ensure store structure exists (`~/wiki/tasks`, `inbox.md`, `.task-meta.json`).
- [x] Create topic files on demand and update `index.md`.
- [x] Implement load/save/move/remove.
- [x] Implement archiving to `archived.md`.
- [x] Implement history log append and retrieval.

**Milestone**: CRUD operations persist correctly to disk and history records are visible.

### Phase 4: CLI commands (`cmd/`)

**Goal**: Provide the `sleepless goals` command surface.

- [x] Implement `sleepless goals add`.
- [x] Implement `sleepless goals list` with filters.
- [x] Implement `sleepless goals done` (with subtask confirmation and optional auto-archive).
- [x] Implement `sleepless goals edit`.
- [x] Implement `sleepless goals remove`.
- [x] Implement `sleepless goals move`.
- [x] Implement `sleepless goals search`.
- [x] Implement `sleepless goals remind` (including `--install`).
- [x] Implement `sleepless goals archive [goal-id]`.
- [x] Implement `sleepless goals history [goal-id]`.
- [x] Implement `sleepless goals view` (`tree`, `progress`, `deps`, `calendar`).

**Milestone**: All commands run end-to-end against the Markdown store.

## Testing Plan

### Unit Tests

Recommended unit tests (may be partial/absent depending on current repo state):

- Parser round-trip tests (parse → write → parse).
- Priority/tag parsing tests.
- Overdue and reminder date boundary tests.
- Prefix match behavior tests.

### Integration Tests

- Create tasks in multiple topics and verify `index.md` updates.
- Move tasks between topics.
- Archive tasks and verify they appear at top of `archived.md`.
- Verify `.task-meta.json` records actions.

## Rollback Plan

Not applicable (existing baseline functionality).

## Open Questions

- Should this proposal be moved to a completed-spec location (if your nocturnal workflow uses `spec/specs/` for completed items)?
- Do you want to document/standardize non-zero exit codes for validation failures as a follow-on proposal?
