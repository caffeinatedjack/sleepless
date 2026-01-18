# lucid-http-testing

**Depends on**: none
**Affected files**: cmd/lucid/main.go, internal/lucid/, pkg/httpclient/, pkg/httpserver/

## Abstract

This specification defines a new standalone executable `lucid` for HTTP/API testing, debugging, and mocking. It provides commands for making HTTP requests with full control over headers and body, running request chains/workflows from files, starting mock HTTP servers with configurable responses, and inspecting HTTP responses with detailed timing and TLS information.

## 1. Introduction

Developers frequently need to test HTTP APIs, debug request/response cycles, and create mock servers for development and testing. Existing tools like curl are powerful but have cryptic syntax; tools like Postman provide great UIs but lack scriptability; tools like httpie improve on curl's ergonomics but don't provide mock server capabilities.

Most API testing workflows involve multiple tools: curl/httpie for requests, custom scripts for sequences, separate mock server frameworks, and browser DevTools for inspection. This fragmentation slows development and makes automation difficult.

This specification defines `lucid`, a unified HTTP testing tool that combines the scriptability of curl, the ergonomics of httpie, the workflow capabilities of Postman collections, and the mock server functionality of tools like json-server. The name "lucid" reflects the clarity it brings to HTTP debugging—like a lucid dream where everything is visible and controllable.

The tool is designed to be offline-capable (mock servers run locally), scriptable (all commands support JSON output), and integration-friendly (can be used in CI/CD pipelines and automated tests).

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Terminology

`Request`: An HTTP request with method, URL, headers, and optional body.

`Chain`: A sequence of HTTP requests executed in order, where later requests can reference data from earlier responses.

`Mock Server`: A local HTTP server that returns predefined responses based on request matching rules.

`Mock Route`: A configured endpoint on a mock server with request matching rules and response definition.

`Request Template`: A YAML/JSON definition of an HTTP request that can be executed by lucid.

`Response Extraction`: The process of extracting values from HTTP responses using JSONPath or regex for use in subsequent requests.

`Timing Information`: Detailed breakdown of request phases (DNS lookup, TCP handshake, TLS negotiation, first byte, total time).

## 4. Concepts

### 4.1. Request Execution

Lucid provides a simple CLI for making HTTP requests with intuitive syntax. Headers, query parameters, and request bodies are specified via flags, making the commands more readable than curl while maintaining scriptability.

### 4.2. Request Chaining

Complex API workflows often require multiple requests where later requests depend on earlier responses (e.g., login to get a token, then use that token in subsequent requests). Lucid supports defining these workflows in YAML/JSON files with variable extraction and substitution.

### 4.3. Mock Servers

Mock servers allow developers to test clients against controlled environments. Lucid's mock server supports:
- Static responses defined in configuration files
- Request matching by path, method, headers, and body patterns
- Response templating with request data
- Logging all incoming requests for debugging

### 4.4. Offline-First Design

All functionality works without internet access. Mock servers run locally, and request chains can be tested against local services.

## 5. Requirements

### 5.1. Core Executable

1. The project MUST provide a standalone executable named `lucid`.
2. The executable MUST be independent of other sleepless executables.
3. All commands MUST return exit code 0 on success and non-zero on failure.

### 5.2. Request Commands

1. `lucid request <method> <url>` MUST make an HTTP request and display the response.
2. Supported methods MUST include: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS.
3. `lucid get <url>` MUST be provided as a shorthand for `lucid request GET <url>`.
4. Similar shorthands MUST be provided for POST, PUT, PATCH, and DELETE.
5. The request command MUST support `--header <key>:<value>` to add request headers (repeatable).
6. The request command MUST support `--data <body>` to set the request body.
7. The request command MUST support `--json <json>` to set a JSON body and automatically set `Content-Type: application/json`.
8. The request command MUST support `--form <key>=<value>` for form-encoded data (repeatable).
9. The request command MUST support `--query <key>=<value>` to add query parameters (repeatable).
10. The request command MUST support `--auth <user>:<pass>` for basic authentication.
11. The request command MUST support `--bearer <token>` to set an `Authorization: Bearer <token>` header.
12. The request command MUST support `--timeout <duration>` to set request timeout (default: 30s).
13. The request command MUST support `--follow` to follow redirects (default: false).
14. The request command MUST support `--max-redirects <n>` when following redirects (default: 10).

### 5.3. Response Display

1. By default, lucid MUST display both response headers and body.
2. `--headers-only` MUST display only response headers.
3. `--body-only` MUST display only response body.
4. `--verbose` MUST display full request details, timing information, and TLS details.
5. Response bodies MUST be pretty-printed if they are valid JSON.
6. `--raw` MUST disable response formatting and output the body exactly as received.

### 5.4. Chain Commands

1. `lucid chain <file>` MUST execute a sequence of requests defined in a YAML/JSON file.
2. Chain files MUST support defining multiple requests with names.
3. Chain files MUST support variable extraction from responses using JSONPath.
4. Chain files MUST support variable substitution in subsequent requests using `${variable}` syntax.
5. Chain files MUST support conditional execution based on response status codes.
6. Chain execution MUST stop on first failure unless `--continue-on-error` is provided.
7. Chain execution MUST support `--var <key>=<value>` to provide initial variables (repeatable).
8. Chain execution MUST output results for each request in the chain.

### 5.5. Mock Server Commands

1. `lucid mock <port>` MUST start a mock HTTP server on the specified port.
2. `lucid mock <port> --config <file>` MUST load mock routes from a configuration file.
3. If no configuration file is provided, the mock server MUST start in echo mode (returns request details).
4. The mock server MUST log all incoming requests to stdout.
5. The mock server MUST support `--log-level <level>` (debug, info, warn, error).
6. The mock server MUST run until terminated with Ctrl+C.
7. Mock configuration MUST support defining routes with:
   - Path pattern (exact match or wildcard)
   - HTTP method
   - Response status code
   - Response headers
   - Response body (static or templated)
   - Delay (simulate network latency)

### 5.6. Inspection Commands

1. `lucid inspect <url>` MUST display detailed information about the endpoint without making a full request.
2. Inspection MUST include:
   - DNS resolution results
   - TLS certificate details (if HTTPS)
   - Response headers from a HEAD request
   - Server software detection
3. `lucid curl <method> <url> [flags]` MUST generate an equivalent curl command from the provided lucid options.

### 5.7. Output Formats

1. All commands MUST support `--json` for machine-readable output.
2. JSON output MUST include request details, response details, and timing information.
3. Text output MUST use clear formatting with syntax highlighting for JSON bodies.

## 6. Interface

### 6.1. Commands

```bash
# Make requests
lucid request GET <url> [flags]
lucid get <url> [flags]
lucid post <url> [flags]
lucid put <url> [flags]
lucid patch <url> [flags]
lucid delete <url> [flags]

# Run request chains
lucid chain <file> [flags]

# Mock servers
lucid mock <port> [--config <file>] [flags]

# Inspection
lucid inspect <url>
lucid curl <method> <url> [flags]
```

### 6.2. Request Flags

```
--header <key>:<value>    Add request header (repeatable)
--data <body>             Request body
--json <json>             JSON request body (sets Content-Type)
--form <key>=<value>      Form-encoded data (repeatable)
--query <key>=<value>     Query parameter (repeatable)
--auth <user>:<pass>      Basic authentication
--bearer <token>          Bearer token authentication
--timeout <duration>      Request timeout (default: 30s)
--follow                  Follow redirects
--max-redirects <n>       Maximum redirects (default: 10)
--insecure                Skip TLS certificate verification
--headers-only            Display only response headers
--body-only               Display only response body
--verbose                 Show detailed request/response info
--raw                     Disable response formatting
--json                    Output in JSON format
```

### 6.3. Chain File Format

```yaml
name: "User API Workflow"
variables:
  base_url: "https://api.example.com"

requests:
  - name: "Login"
    method: POST
    url: "${base_url}/auth/login"
    json:
      username: "user@example.com"
      password: "secret"
    extract:
      token: "$.access_token"
  
  - name: "Get User Profile"
    method: GET
    url: "${base_url}/users/me"
    headers:
      Authorization: "Bearer ${token}"
    extract:
      user_id: "$.id"
  
  - name: "Update Profile"
    method: PATCH
    url: "${base_url}/users/${user_id}"
    headers:
      Authorization: "Bearer ${token}"
    json:
      display_name: "New Name"
    expect:
      status: 200
```

### 6.4. Mock Server Configuration Format

```yaml
routes:
  - path: "/api/users"
    method: GET
    response:
      status: 200
      headers:
        Content-Type: "application/json"
      body:
        users:
          - id: 1
            name: "Alice"
          - id: 2
            name: "Bob"
  
  - path: "/api/users/:id"
    method: GET
    response:
      status: 200
      headers:
        Content-Type: "application/json"
      body:
        id: "{{.Params.id}}"
        name: "User {{.Params.id}}"
      delay: 100ms
  
  - path: "/api/users"
    method: POST
    response:
      status: 201
      headers:
        Content-Type: "application/json"
      body:
        id: "{{randomInt 1000 9999}}"
        created_at: "{{now}}"
```

## 7. Behavior

### 7.1. Request Execution

1. Headers specified via `--header` MUST be added to the request exactly as provided.
2. When `--json` is used, lucid MUST:
   - Set `Content-Type: application/json` header unless explicitly overridden
   - Validate that the provided JSON is well-formed
   - Send the JSON as the request body
3. When `--form` is used, lucid MUST:
   - Set `Content-Type: application/x-www-form-urlencoded` header
   - Properly URL-encode form values
4. Query parameters specified via `--query` MUST be properly URL-encoded and appended to the URL.

### 7.2. Chain Execution

1. Requests in a chain MUST be executed in the order defined in the file.
2. Variable extraction MUST be performed immediately after a successful response.
3. If extraction fails, the chain MUST stop unless `--continue-on-error` is provided.
4. Variables MUST be substituted in URLs, headers, and request bodies before execution.
5. If a variable reference cannot be resolved, lucid MUST report an error and stop execution.

### 7.3. Mock Server

1. The mock server MUST match incoming requests to routes in order of definition.
2. The first matching route MUST be used to generate the response.
3. If no route matches, the server MUST return 404 Not Found.
4. Response body templates MUST support:
   - Path parameters via `{{.Params.name}}`
   - Query parameters via `{{.Query.name}}`
   - Request headers via `{{.Headers.name}}`
   - Request body via `{{.Body}}`
   - Built-in functions: `{{randomInt min max}}`, `{{uuid}}`, `{{now}}`
5. If a delay is specified, the server MUST wait before sending the response.

### 7.4. TLS Handling

1. By default, lucid MUST verify TLS certificates.
2. When `--insecure` is provided, lucid MUST skip certificate verification.
3. In verbose mode, lucid MUST display certificate information including:
   - Issuer
   - Subject
   - Validity period
   - SANs (Subject Alternative Names)

## 8. Error Handling

1. If a URL is malformed, lucid MUST exit with code 1 and report the parsing error.
2. If a request times out, lucid MUST exit with code 1 and report the timeout.
3. If a DNS lookup fails, lucid MUST exit with code 1 and report the failure.
4. If a connection cannot be established, lucid MUST exit with code 1 and report the connection error.
5. If a chain file cannot be parsed, lucid MUST exit with code 1 and report the parsing error with line number.
6. If a mock server cannot bind to the specified port, lucid MUST exit with code 1 and report the error.
7. HTTP error status codes (4xx, 5xx) MUST NOT cause a non-zero exit code unless explicitly checked via chain expectations.

## 9. Examples

```bash
# Simple GET request
lucid get https://api.example.com/users

# POST with JSON body
lucid post https://api.example.com/users \
  --json '{"name":"Alice","email":"alice@example.com"}'

# Request with headers and auth
lucid get https://api.example.com/protected \
  --bearer "eyJhbGc..." \
  --header "X-Custom-Header:value"

# Form submission
lucid post https://api.example.com/login \
  --form "username=user" \
  --form "password=secret"

# Verbose output with timing
lucid get https://api.example.com/data --verbose

# Run a request chain
lucid chain workflows/user-signup.yml

# Run chain with custom variables
lucid chain workflows/api-test.yml \
  --var "base_url=https://staging.example.com" \
  --var "api_key=abc123"

# Start mock server
lucid mock 8080 --config mocks/api.yml

# Start echo server (no config)
lucid mock 8080

# Inspect endpoint
lucid inspect https://api.example.com

# Generate curl command
lucid curl POST https://api.example.com/users \
  --json '{"name":"Alice"}' \
  --bearer "token123"
```

## 10. Security Considerations

1. Lucid MUST NOT transmit request/response data to external services.
2. When displaying request details, lucid SHOULD redact sensitive headers (Authorization, Cookie, Set-Cookie) unless `--verbose` is used.
3. Chain files and mock configurations MAY contain sensitive data (tokens, API keys). Users MUST be warned in documentation about version control considerations.
4. The `--insecure` flag MUST be used cautiously and SHOULD emit a warning when used.
5. Mock servers MUST bind to localhost by default to prevent external access. A `--bind <address>` flag MAY be added for advanced use cases.
6. Request bodies specified via command-line arguments are visible in process lists. Documentation SHOULD recommend using chain files for sensitive data.
7. Chain files SHOULD support environment variable substitution (e.g., `${ENV:API_KEY}`) to avoid hardcoding secrets.

## 11. Testing Considerations

Test scenarios SHOULD include:

- Requests to various HTTP methods and status codes
- JSON, form, and binary request bodies
- Header parsing and URL encoding
- TLS certificate verification and `--insecure` mode
- Chain execution with variable extraction and substitution
- Chain failure handling with `--continue-on-error`
- Mock server route matching (exact, wildcard, parameter extraction)
- Mock server response templating
- Timeout handling for slow servers
- Redirect following with `--follow`
- Large response bodies
- Invalid JSON response handling
- DNS resolution failures
- Connection refused scenarios

## 12. References

### Normative References

[RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997, https://www.rfc-editor.org/rfc/rfc2119

[RFC7230] Fielding, R. and J. Reschke, "Hypertext Transfer Protocol (HTTP/1.1): Message Syntax and Routing", June 2014, https://www.rfc-editor.org/rfc/rfc7230

[RFC7231] Fielding, R. and J. Reschke, "Hypertext Transfer Protocol (HTTP/1.1): Semantics and Content", June 2014, https://www.rfc-editor.org/rfc/rfc7231

### Informative References

[CURL] curl project, https://curl.se/

[HTTPIE] HTTPie, https://httpie.io/

[POSTMAN] Postman, https://www.postman.com/

[JSONPATH] JSONPath specification, https://goessner.net/articles/JsonPath/
