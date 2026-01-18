# Design: lucid-http-testing
**Specification Reference**: [specification.md](specification.md)
Status: Draft

## 1. Context

Lucid is a new standalone executable for HTTP/API testing, combining request execution, workflow chaining, and mock servers in a single tool. It fills a gap in the sleepless ecosystem: while nightwatch handles security utilities and regimen handles productivity, developers frequently need HTTP testing capabilities that are more ergonomic than curl but more scriptable than Postman.

The Go standard library provides excellent HTTP client and server primitives, making implementation straightforward. The main design decisions involve CLI ergonomics, chain file format, and mock server configuration.

Key constraints:
- Must be usable offline (mock servers enable local testing)
- Must support JSON output for all commands (CI/CD integration)
- Must handle large response bodies without memory issues
- Chain files must be human-readable and version-control friendly

## 2. Goals and Non-Goals

### Goals

- Provide ergonomic HTTP request commands with intuitive flag syntax
- Support request chaining with variable extraction (JSONPath) and substitution
- Enable local mock servers with configurable routes and response templating
- Generate detailed timing information for performance debugging
- Output machine-readable JSON for automation and scripting

### Non-Goals

- WebSocket support (may be added in future)
- GraphQL-specific features (use generic POST with JSON body)
- Browser automation or JavaScript execution
- Persistent sessions or cookie jars across invocations (chain files handle this)
- Load testing or benchmarking (use dedicated tools like k6, wrk)

## 3. Options Considered

### Option 1: Flat Command Structure

All commands at the top level: `lucid get`, `lucid post`, `lucid chain`, `lucid mock`.

**Advantages**:
- Minimal typing for common operations
- Intuitive command discovery
- Matches httpie ergonomics

**Disadvantages**:
- Less room for future command expansion
- Method shorthands (get, post) feel different from other commands

**Complexity**: Low

### Option 2: Grouped Command Structure

Commands grouped by function: `lucid request get`, `lucid server mock`, `lucid workflow chain`.

**Advantages**:
- Clear organization as features grow
- Consistent naming pattern

**Disadvantages**:
- More typing for common operations
- Feels bureaucratic for simple requests

**Complexity**: Low

## 4. Decision

**Chosen Option**: Option 1 (Flat Command Structure)

**Rationale**: HTTP requests are the primary use case. Making `lucid get <url>` as fast as possible encourages adoption. The mock server and chain commands are distinct enough that top-level placement is natural.

**Key Factors**:
1. Ergonomics for the most common operation (single requests)
2. Discoverability via `lucid --help`
3. Consistency with tools like httpie

## 5. Detailed Design

### Architecture Overview

```
cmd/lucid/main.go           Entry point
internal/lucid/
  root.go                   Root command setup
  request.go                HTTP request commands (get, post, put, etc.)
  chain.go                  Chain execution command
  mock.go                   Mock server command
  inspect.go                URL inspection command
  curl.go                   Curl command generation
pkg/httpclient/
  client.go                 HTTP client wrapper with timing
  request.go                Request building from flags
  response.go               Response formatting and display
pkg/httpserver/
  mock.go                   Mock server implementation
  router.go                 Route matching logic
  template.go               Response body templating
pkg/chain/
  parser.go                 Chain file parsing (YAML/JSON)
  executor.go               Chain execution with variable handling
  extractor.go              JSONPath extraction
```

### Component Design

**HTTP Client (`pkg/httpclient`)**:
- Wraps `net/http` with timing instrumentation
- Captures DNS, TCP, TLS, and transfer timing
- Handles redirect following with configurable limits
- Supports TLS configuration (insecure mode)

**Request Builder**:
- Constructs requests from CLI flags
- Handles header parsing (`Key:Value` format)
- URL-encodes query parameters and form data
- Validates JSON bodies before sending

**Response Formatter**:
- Pretty-prints JSON bodies with syntax highlighting
- Displays headers in readable format
- Shows timing breakdown in verbose mode
- Supports raw output mode

**Chain Executor (`pkg/chain`)**:
- Parses YAML/JSON chain files
- Maintains variable scope across requests
- Extracts values using JSONPath expressions
- Substitutes variables in URLs, headers, and bodies
- Handles conditional execution and expectations

**Mock Server (`pkg/httpserver`)**:
- Matches incoming requests to configured routes
- Supports path parameters (`:id` syntax)
- Templates response bodies with request data
- Simulates latency with configurable delays
- Logs all requests to stdout

### Data Design

**Chain File Schema**:
```yaml
name: string              # Workflow name
variables:                # Initial variables
  key: value
requests:
  - name: string          # Request identifier
    method: string        # HTTP method
    url: string           # URL with variable substitution
    headers:              # Optional headers
      key: value
    json: object          # JSON body (sets Content-Type)
    form:                 # Form body
      key: value
    extract:              # JSONPath extractions
      varname: "$.path"
    expect:               # Optional assertions
      status: int
```

**Mock Configuration Schema**:
```yaml
routes:
  - path: string          # Path pattern (supports :param)
    method: string        # HTTP method (default: GET)
    response:
      status: int         # Status code (default: 200)
      headers:            # Response headers
        key: value
      body: any           # Static or templated body
      delay: duration     # Optional delay (e.g., "100ms")
```

### API Design

Not applicable (CLI-only).

## 6. Trade-offs

| Trade-off | Gain | Sacrifice | Justification |
|-----------|------|-----------|---------------|
| JSONPath only (no XPath) | Simpler implementation | XML response handling | JSON dominates modern APIs |
| No persistent cookies | Stateless commands | Session continuity | Chain files handle sessions explicitly |
| Localhost-only mock by default | Security | Remote mock testing | Can add `--bind` flag later |
| YAML chain files | Readability | JSON-only tooling | YAML is more human-friendly for workflows |

## 7. Cross-Cutting Concerns

### Security

- Redact `Authorization` and `Cookie` headers in non-verbose output
- Warn when `--insecure` is used
- Mock server binds to localhost only by default
- Document risks of secrets in command-line arguments (visible in process lists)
- Support environment variable substitution in chain files (`${ENV:VAR_NAME}`)

### Performance

- Stream large response bodies to avoid memory exhaustion
- Use connection pooling for chain execution
- Mock server uses standard library's efficient HTTP server

### Reliability

- Clear error messages for network failures (DNS, connection, timeout)
- Chain execution stops on first error by default (configurable)
- Mock server gracefully handles malformed requests

### Testing

- Unit tests for request building from flags
- Unit tests for JSONPath extraction
- Unit tests for response formatting
- Unit tests for mock route matching
- Integration tests for full request/response cycles
- Integration tests for chain execution

## 8. Implementation Plan

See [implementation.md](implementation.md) for phased task breakdown.

### Migration Strategy

Not applicable (new executable).

## 9. Open Questions

- Should `lucid chain` support parallel request execution for independent requests?
- Should mock server support recording mode (capture requests for replay)?
- Should we support HAR file import/export for interoperability with browser DevTools?
- Should response body size limits be configurable (default: unlimited)?
