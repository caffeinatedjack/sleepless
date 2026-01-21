# readme-documentation

## Rule

The project README MUST be updated whenever new CLI commands or executables are added to the project. The documentation MUST include:

- A brief description of what the command/executable does
- Basic usage examples
- Key flags or subcommands
- A reference to more detailed documentation if available

## Rationale

The README serves as the primary entry point for users and developers to understand what the project does and how to use it. Without up-to-date documentation of available commands and executables, users cannot discover or effectively use the features that have been built. This leads to:

- Reduced adoption and usability
- Increased support burden from questions about undocumented features
- Confusion about what functionality is actually available
- Wasted development effort on features that users don't know exist

By requiring README updates alongside feature development, we ensure that documentation stays in sync with the codebase and users always have access to current information.

## Scope

This rule applies to:

- **CLI commands**: Any new commands or subcommands added to existing executables (e.g., `regimen note`, `nightwatch login`)
- **New executables**: Any new binaries added to the project (e.g., `doze`, `regimen`, `nightwatch`)
- **Major flag additions**: Significant new flags that change how commands work
- **Breaking changes**: Any changes that modify existing command behavior

This rule applies at the proposal completion stage - before a proposal is marked as complete, the README MUST be updated to reflect the new functionality.

## Exception

Minor internal commands or debugging utilities that are not intended for general use MAY be excluded from README documentation. However, these SHOULD be documented in developer-specific documentation (e.g., `CONTRIBUTING.md` or `docs/development.md`).
