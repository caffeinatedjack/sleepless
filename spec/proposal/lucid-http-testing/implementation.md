# Implementation: lucid-http-testing

**Specification Reference**: [specification.md](specification.md)
**Design Reference**: [design.md](design.md)

## Overview

Implement `lucid` as a new standalone executable with four main capabilities: HTTP requests (`get`, `post`, etc.), request chaining (`chain`), mock servers (`mock`), and inspection utilities (`inspect`, `curl`). Uses Go's standard library for HTTP client/server with additional packages for JSONPath and YAML parsing.

## Prerequisites

- External dependency: JSONPath library (e.g., `github.com/ohler55/ojg` or `github.com/PaesslerAG/jsonpath`)
- External dependency: YAML parser (already used in project: `gopkg.in/yaml.v3`)

## Phases

### Phase 1: Core HTTP Client

**Goal**: Implement basic HTTP request commands with flag parsing and response display.

- [ ] 1.1 Create `cmd/lucid/main.go` entry point
- [ ] 1.2 Create `internal/lucid/root.go` with Cobra root command
- [ ] 1.3 Implement `pkg/httpclient/client.go` with timing instrumentation
- [ ] 1.4 Implement `pkg/httpclient/request.go` for request building from flags
- [ ] 1.5 Implement `pkg/httpclient/response.go` for response formatting
- [ ] 1.6 Implement `lucid request <method> <url>` command
- [ ] 1.7 Implement method shorthands: `lucid get`, `lucid post`, `lucid put`, `lucid patch`, `lucid delete`
- [ ] 1.8 Implement request flags: `--header`, `--data`, `--json`, `--form`, `--query`
- [ ] 1.9 Implement auth flags: `--auth`, `--bearer`
- [ ] 1.10 Implement output flags: `--headers-only`, `--body-only`, `--verbose`, `--raw`, `--json`
- [ ] 1.11 Implement `--timeout`, `--follow`, `--max-redirects`, `--insecure` flags
- [ ] 1.12 Add JSON pretty-printing for response bodies
- [ ] 1.13 Add unit tests for request building
- [ ] 1.14 Add unit tests for response formatting

**Milestone**: `lucid get https://httpbin.org/get --verbose` displays response with timing breakdown.

### Phase 2: Chain Execution

**Goal**: Implement request chaining with variable extraction and substitution.

- [ ] 2.1 Implement `pkg/chain/parser.go` for YAML/JSON chain file parsing
- [ ] 2.2 Implement `pkg/chain/extractor.go` for JSONPath extraction
- [ ] 2.3 Implement variable substitution (`${variable}` syntax)
- [ ] 2.4 Implement `pkg/chain/executor.go` for sequential request execution
- [ ] 2.5 Implement `lucid chain <file>` command
- [ ] 2.6 Implement `--var <key>=<value>` flag for initial variables
- [ ] 2.7 Implement `--continue-on-error` flag
- [ ] 2.8 Implement expectation checking (`expect.status`)
- [ ] 2.9 Implement environment variable substitution (`${ENV:VAR_NAME}`)
- [ ] 2.10 Add chain execution output (per-request results)
- [ ] 2.11 Add `--json` output for chain results
- [ ] 2.12 Add unit tests for chain parsing
- [ ] 2.13 Add unit tests for JSONPath extraction
- [ ] 2.14 Add unit tests for variable substitution
- [ ] 2.15 Add integration tests for chain execution

**Milestone**: Can execute a multi-step chain that logs in, extracts a token, and uses it in subsequent requests.

### Phase 3: Mock Server

**Goal**: Implement configurable mock HTTP server with route matching and response templating.

- [ ] 3.1 Implement `pkg/httpserver/router.go` for route matching
- [ ] 3.2 Implement path parameter extraction (`:id` syntax)
- [ ] 3.3 Implement `pkg/httpserver/template.go` for response body templating
- [ ] 3.4 Implement `pkg/httpserver/mock.go` mock server main loop
- [ ] 3.5 Implement `lucid mock <port>` command
- [ ] 3.6 Implement echo mode (default when no config provided)
- [ ] 3.7 Implement `--config <file>` flag for route configuration
- [ ] 3.8 Implement response delay simulation
- [ ] 3.9 Implement request logging to stdout
- [ ] 3.10 Implement `--log-level` flag
- [ ] 3.11 Implement template functions: `{{.Params.x}}`, `{{.Query.x}}`, `{{.Headers.x}}`, `{{.Body}}`
- [ ] 3.12 Implement built-in functions: `{{randomInt min max}}`, `{{uuid}}`, `{{now}}`
- [ ] 3.13 Add unit tests for route matching
- [ ] 3.14 Add unit tests for response templating
- [ ] 3.15 Add integration tests for mock server

**Milestone**: `lucid mock 8080 --config api.yml` serves configured routes; `lucid get http://localhost:8080/api/users` returns mocked response.

### Phase 4: Inspection and Utilities

**Goal**: Implement inspection commands and curl generation.

- [ ] 4.1 Implement `lucid inspect <url>` command
- [ ] 4.2 Add DNS resolution display
- [ ] 4.3 Add TLS certificate information display
- [ ] 4.4 Add HEAD request for server headers
- [ ] 4.5 Implement `lucid curl <method> <url>` command
- [ ] 4.6 Generate curl command from lucid flags
- [ ] 4.7 Add unit tests for curl generation
- [ ] 4.8 Add integration tests for inspect command

**Milestone**: `lucid inspect https://example.com` shows DNS, TLS, and header info; `lucid curl post https://api.example.com --json '{"x":1}'` outputs equivalent curl command.

## Testing Plan

### Unit Tests

- Request building: header parsing, body construction, URL encoding
- Response formatting: JSON pretty-printing, header display, timing breakdown
- Chain parsing: valid YAML, invalid YAML, missing fields
- JSONPath extraction: simple paths, nested paths, array access, missing paths
- Variable substitution: simple variables, nested substitution, undefined variables
- Route matching: exact paths, parameterized paths, method matching
- Response templating: parameter substitution, built-in functions
- Curl generation: all flag combinations

### Integration Tests

- Full request cycle: make request to httpbin.org, verify response parsing
- Chain execution: multi-step workflow with variable passing
- Mock server: start server, make requests, verify responses
- TLS handling: verify certificate, test `--insecure` mode
- Timeout handling: test request timeout behavior
- Redirect following: test `--follow` with various redirect codes

## Rollback Plan

- Remove `cmd/lucid/`, `internal/lucid/`, `pkg/httpclient/`, `pkg/httpserver/`, `pkg/chain/`
- Remove `lucid` from Makefile build targets
- Remove external dependencies if no longer needed

## Open Questions

- Should we vendor the JSONPath library or use a well-maintained external dependency?
- Should chain files support TOML in addition to YAML/JSON?
- Should the mock server support HTTPS with self-signed certificates?
