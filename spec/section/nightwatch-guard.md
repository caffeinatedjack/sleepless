# nightwatch-guard

**Depends on**: none
**Affected files**: internal/nightwatch/guard.go, pkg/redact/* (reuse), pkg/guard/*

## Abstract

This specification adds a new `nightwatch guard` command group that scans local code (staged content, working tree, or paths) for likely secrets and PII and fails fast when findings exist. It is designed for offline, scriptable usage in pre-commit hooks and CI pipelines. Output is safe-by-default: the tool MUST NOT print raw secret values.

## 1. Introduction

In large development organizations, one of the most expensive and disruptive classes of mistakes is accidental exposure of secrets (API keys, JWTs, private keys) and sensitive data (emails, phone numbers) in source control. Even when access is limited, secrets that land in history often trigger incident processes: revocation, credential rotation, audit work, and emergency releases.

The `nightwatch` executable already provides `redact` for transforming text, logs, or diffs into a shareable form. However, a redaction tool does not, by itself, prevent risky data from being committed or merged.

This proposal introduces a guardrail command group: `nightwatch guard`. Unlike redaction, `guard` is explicitly an auditing/gating tool: it scans content and produces a report. When findings exist, `guard` returns a non-zero exit code suitable for pre-commit hooks and CI.

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`Finding`: A detected match of a sensitive-data pattern.

`Target`: The data source being scanned (staged snapshot, working tree snapshot, or an explicit filesystem path).

`Fingerprint`: A stable identifier derived from a finding that allows suppressing known findings without storing the raw sensitive value.

`Baseline`: A file containing fingerprints to suppress (ignore) during scanning.

## 4. Concepts

### 4.1. Offline, Local-First Scanning

The `guard` command group is local-only. It reads from the local filesystem and (optionally) uses local git plumbing to obtain staged snapshots. It MUST NOT make network requests.

### 4.2. Pattern Sources

The implementation SHOULD reuse the existing redaction pattern definitions in `pkg/redact` so that `redact` and `guard` stay consistent.

`guard` is a checker: it reports findings and counts. It MUST NOT rewrite files.

### 4.3. Safe Output

`guard` output MUST NOT contain raw secret/PII values.

When human-readable output includes a content excerpt, the excerpt MUST be redacted/masked before printing.

### 4.4. Baselines

A baseline allows a team to adopt scanning incrementally.

A baseline entry SHOULD use a `fingerprint` derived from:

- the finding type, and
- the matched value

The baseline MUST NOT store the raw matched value.

## 5. Requirements

### 5.1. Command Group

1. The CLI MUST provide a top-level command group: `nightwatch guard`.
2. All guard subcommands MUST return exit code 0 when no findings exist.
3. When findings exist (after baseline suppression), guard commands MUST return a non-zero exit code suitable for CI.
4. All guard subcommands MUST support `--json` output.
5. Guard commands MUST NOT modify files.

### 5.2. Targets

1. `nightwatch guard staged` MUST scan the staged snapshot of files that are staged for commit.
2. `nightwatch guard worktree` MUST scan the working tree snapshot of files changed relative to HEAD.
3. `nightwatch guard path <path>` MUST scan the provided file or directory.

### 5.3. Baseline Support

1. Guard subcommands MUST accept `--baseline <path>`.
2. When a baseline is provided, findings whose fingerprint is present in the baseline MUST be suppressed.
3. `nightwatch guard baseline` MUST scan a target (same defaults as `staged`) and emit a baseline file to stdout or to `--out <path>`.

### 5.4. Output Schema

1. JSON output MUST be a single JSON object.
2. JSON output MUST NOT contain raw matched values.
3. The output MUST include:
   - the target type
   - the total finding count
   - counts per finding type
   - a list of findings including file path and 1-indexed line/column

## 6. Interface

### 6.1. Commands

```bash
nightwatch guard staged
nightwatch guard worktree
nightwatch guard path <path>
nightwatch guard baseline [staged|worktree|path <path>]
```

### 6.2. Flags

- `--json` (boolean): machine-readable output
- `--baseline <path>` (optional): suppress known findings

For `baseline`:
- `--out <path>` (optional): write baseline file to disk (default: stdout)

## 7. Behavior

1. `staged` scans content as it would be committed (the staged snapshot), not the working tree.
2. `worktree` scans current working tree file contents.
3. `path` scans UTF-8 text files.
4. Binary files SHOULD be skipped with a warning.
5. The scanner MUST report file-relative line and column numbers (1-indexed).

## 8. Error Handling

1. If the repository is not a git repository, `staged` and `worktree` MUST return a non-zero exit code with a clear error.
2. If a baseline file cannot be read or parsed, the command MUST return a non-zero exit code.
3. If a path cannot be read, the command MUST return a non-zero exit code.

## 9. Examples

### 9.1. Staged scan (CI / pre-commit)

```bash
nightwatch guard staged
```

### 9.2. Staged scan with JSON

```bash
nightwatch guard staged --json
```

### 9.3. Creating a baseline

```bash
nightwatch guard baseline staged --out .nightwatch-guard-baseline.json
```

## 10. Security Considerations

1. The implementation MUST NOT print raw secret values.
2. The implementation MUST NOT transmit scanned content over the network.
3. Baseline files MUST NOT store raw secret values.
4. The implementation SHOULD warn users that passing secrets via CLI args can leak via shell history/process lists.

## 11. Testing Considerations

- Scanning a file with known patterns yields expected findings and counts.
- Baseline suppression removes matching findings.
- `staged` scans staged snapshots (verify by staging different content than the working tree).
- JSON output is stable and contains no raw matches.

## 12. References

### Normative References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997, https://www.rfc-editor.org/rfc/rfc2119

### Informative References

[Git] "git", https://git-scm.com/docs
