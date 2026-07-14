# Design Document

## Introduction

This design extends the Racore CLI to cover the full Racore Cloud CDN API (~60 operations) organized into six command groups. The architecture follows the existing patterns: a shared `execute*` function layer consumed by both Cobra CLI commands and ACP tool handlers, the `internal/api` client for HTTP communication with 401 retry, and tabwriter-based output formatting.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Entry Points                          │
│  ┌─────────────┐              ┌──────────────────────┐  │
│  │  Cobra CLI  │              │  ACP Server (serve)  │  │
│  │  commands   │              │  JSON-RPC handlers   │  │
│  └──────┬──────┘              └──────────┬───────────┘  │
│         │                                │              │
│         ▼                                ▼              │
│  ┌──────────────────────────────────────────────────┐   │
│  │         execute* functions (shared logic)         │   │
│  │  executeDomainList, executeDomainCreate, ...      │   │
│  └────────────────────────┬─────────────────────────┘   │
│                           │                             │
│                           ▼                             │
│  ┌──────────────────────────────────────────────────┐   │
│  │           internal/api.Client                     │   │
│  │     Get | Post | Put | Delete                     │   │
│  │     doWithRetry (401 → re-auth → retry once)     │   │
│  └────────────────────────┬─────────────────────────┘   │
│                           │                             │
│                           ▼                             │
│  ┌──────────────────────────────────────────────────┐   │
│  │           internal/auth.Manager                   │   │
│  │     GetValidToken | Authenticate | ClearCache     │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

## Components

### 1. API Client Extension (`internal/api/client.go`)

Add `Put` and `Delete` methods that follow the same `doWithRetry` pattern as `Get` and `Post`.

```go
// Put sends a PUT request with Bearer auth and 401 retry.
func (c *Client) Put(endpoint string, body interface{}) (json.RawMessage, error) {
    return c.doWithRetry(func(token string) (*http.Response, error) {
        url := c.baseURL + endpoint
        bodyBytes, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal request body: %w", err)
        }
        req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(bodyBytes))
        if err != nil {
            return nil, fmt.Errorf("failed to create request: %w", err)
        }
        req.Header.Set("Authorization", "Bearer "+token)
        req.Header.Set("Content-Type", "application/json")
        return c.client.Do(req)
    })
}

// Delete sends a DELETE request with a JSON body, Bearer auth, and 401 retry.
func (c *Client) Delete(endpoint string, body interface{}) (json.RawMessage, error) {
    return c.doWithRetry(func(token string) (*http.Response, error) {
        url := c.baseURL + endpoint
        bodyBytes, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal request body: %w", err)
        }
        req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(bodyBytes))
        if err != nil {
            return nil, fmt.Errorf("failed to create request: %w", err)
        }
        req.Header.Set("Authorization", "Bearer "+token)
        req.Header.Set("Content-Type", "application/json")
        return c.client.Do(req)
    })
}
```

The `Delete` method accepts a `body interface{}` (not query params) because the Racore API sends JSON bodies for DELETE requests.

### 2. Command Group Files

Each command group occupies a single file with a parent command and nested subcommands:

| File | Parent Command | Subcommands |
|------|---------------|-------------|
| `cmd/domain.go` | `racore-cli domain` | `list`, `create`, `delete`, `enable`, `disable`, `source get/set`, `ssl get/set`, `enforce-https get/set`, `ip-filter get/set`, `referer-filter get/set`, `ua-filter get/set`, `origin-protocol get/set`, `http2 get/set`, `http3 get/set`, `tls-version get/set`, `compress get/set`, `ipv6 get/set`, `cache-policy get/set`, `origin-host get/set`, `origin-timeout get/set`, `geo-restriction get/set`, `request-headers get/set`, `response-headers get/set`, `request-header-policy get/set`, `response-header-policy get/set` |
| `cmd/cache.go` | `racore-cli cache` | `purge`, `purge-status`, `prefetch`, `prefetch-status`, `prewarm-regions`, `prewarm-pop`, `list-policies`, `list-origin-request-policies`, `list-response-header-policies` |
| `cmd/stats.go` | `racore-cli stats` | `flow`, `request`, `hit-flow`, `hit-request`, `http-code`, `http-code-detail`, `district`, `iso-country`, `top-domain`, `top-url`, `top-referer`, `top-ua` |
| `cmd/cert.go` | `racore-cli cert` | `list`, `upload`, `update`, `apply-aws`, `validation-info` |
| `cmd/workorder.go` | `racore-cli workorder` | `list`, `create`, `reopen`, `delete`, `cancel`, `close`, `types`, `log`, `send-message` |
| `cmd/log.go` | `racore-cli log` | `list` |

The existing `cmd/domains.go` is removed once `cmd/domain.go` replaces it.

### 3. Shared Execute Functions Pattern

Each command's logic lives in an `execute*` function that:
1. Loads credentials from the credential store
2. Checks for nil credentials (returns error if not logged in)
3. Creates an auth manager and sets the cached token
4. Creates an API client
5. Calls the appropriate API endpoint
6. Parses and formats the response
7. Returns a formatted string or error

This pattern is already established in `cmd/serve.go` with `executeDomains`, `executeLogin`, etc.

```go
// Example: executeDomainEnable
func executeDomainEnable(domain string) (string, error) {
    client, err := newAuthenticatedClient()
    if err != nil {
        return "", err
    }

    body := map[string]string{"domain": domain}
    rawResp, err := client.Put("/API/cdn/domain/state/open", body)
    if err != nil {
        return "", fmt.Errorf("API request failed: %w", err)
    }

    var resp apiResponse
    if err := json.Unmarshal(rawResp, &resp); err != nil {
        return "", fmt.Errorf("failed to parse API response: %w", err)
    }
    if resp.Code != 1 {
        return "", fmt.Errorf("API error: %s", resp.Message)
    }

    return fmt.Sprintf("Domain %s enabled successfully.", domain), nil
}
```

### 4. Authentication Helper

To avoid duplicating credential-loading boilerplate across 60+ execute functions, introduce a helper:

```go
// newAuthenticatedClient loads credentials and returns a ready-to-use API client.
func newAuthenticatedClient() (*api.Client, error) {
    store, err := credential.NewStore()
    if err != nil {
        return nil, fmt.Errorf("cannot initialize credential store: %w", err)
    }

    creds, err := store.Load()
    if err != nil {
        return nil, fmt.Errorf("cannot load credentials: %w", err)
    }

    if creds == nil {
        return nil, fmt.Errorf("not logged in. Please run 'racore-cli login' first")
    }

    mgr := auth.NewManager(creds.AccessKey, creds.SecretKey)
    mgr.SetCache(creds.Token, creds.Expire)

    return api.NewClient(mgr), nil
}
```

### 5. Command Registration Pattern (Cobra)

Each file uses `init()` to register the parent command and subcommands:

```go
var domainCmd = &cobra.Command{
    Use:   "domain",
    Short: "Manage CDN domains",
}

var domainListCmd = &cobra.Command{
    Use:   "list",
    Short: "List CDN domains",
    RunE:  runDomainList,
}

var domainEnableCmd = &cobra.Command{
    Use:   "enable",
    Short: "Enable a CDN domain",
    RunE:  runDomainEnable,
}

func init() {
    domainCmd.AddCommand(domainListCmd)
    domainCmd.AddCommand(domainEnableCmd)
    // ... more subcommands
    rootCmd.AddCommand(domainCmd)
}
```

For two-level nesting (e.g., `domain ssl get`):

```go
var domainSSLCmd = &cobra.Command{
    Use:   "ssl",
    Short: "Manage SSL/HTTPS configuration",
}

var domainSSLGetCmd = &cobra.Command{
    Use:   "get",
    Short: "Get SSL configuration for a domain",
    RunE:  runDomainSSLGet,
}

func init() {
    domainSSLCmd.AddCommand(domainSSLGetCmd)
    domainSSLCmd.AddCommand(domainSSLSetCmd)
    domainCmd.AddCommand(domainSSLCmd)
}
```

### 6. ACP Tool Registration Pattern

All tools are registered in `cmd/serve.go`. Each tool handler follows the same pattern:

```go
server.RegisterTool(acp.ToolDefinition{
    Name:        "domain-enable",
    Description: "Enable a CDN domain",
    InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "domain": {"type": "string", "description": "Domain name to enable"}
        },
        "required": ["domain"]
    }`),
}, domainEnableHandler)

func domainEnableHandler(params json.RawMessage) (interface{}, error) {
    var input struct {
        Domain string `json:"domain"`
    }
    if err := json.Unmarshal(params, &input); err != nil {
        return nil, fmt.Errorf("invalid parameters: %w", err)
    }
    if input.Domain == "" {
        return nil, fmt.Errorf("domain is required")
    }

    output, err := executeDomainEnable(input.Domain)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "content": []map[string]interface{}{
            {"type": "text", "text": output},
        },
    }, nil
}
```

### 7. Response Formatting

**List data** — tabwriter table:
```go
func formatTable(headers []string, rows [][]string) string {
    var buf bytes.Buffer
    w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
    fmt.Fprintf(w, "%s\n", strings.Join(headers, "\t"))
    for _, row := range rows {
        fmt.Fprintf(w, "%s\n", strings.Join(row, "\t"))
    }
    w.Flush()
    return buf.String()
}
```

**Single resource** — key-value pairs:
```go
func formatKeyValue(pairs map[string]string) string {
    var buf strings.Builder
    for k, v := range pairs {
        buf.WriteString(fmt.Sprintf("%s: %s\n", k, v))
    }
    return buf.String()
}
```

**Success confirmation**:
```
Domain example.com enabled successfully.
```

**Error output** — written to stderr via the error return path.

## Interfaces

### API Client Interface

```go
type Client struct { ... }

func (c *Client) Get(endpoint string, queryParams map[string]string) (json.RawMessage, error)
func (c *Client) Post(endpoint string, body interface{}) (json.RawMessage, error)
func (c *Client) Put(endpoint string, body interface{}) (json.RawMessage, error)
func (c *Client) Delete(endpoint string, body interface{}) (json.RawMessage, error)
```

### Execute Function Signatures

All execute functions follow one of these patterns:

```go
// Query (GET) - returns formatted string
func executeDomainList(filter string) (string, error)
func executeDomainSSLGet(domain string) (string, error)

// Mutation (POST/PUT/DELETE) - returns confirmation string
func executeDomainCreate(params DomainCreateParams) (string, error)
func executeDomainEnable(domain string) (string, error)
func executeDomainDelete(domain string) (string, error)

// Complex query (POST with body for stats)
func executeStatsFlow(params StatsFlowParams) (string, error)
```

### ACP Handler Signature

All handlers conform to:
```go
type ToolHandler func(params json.RawMessage) (interface{}, error)
```

Return format on success:
```json
{
  "content": [{"type": "text", "text": "<formatted output>"}]
}
```

## Data Models

### Common API Response Envelope

```go
// apiResponse is the standard envelope for all Racore API responses.
type apiResponse struct {
    Code    int             `json:"code"`
    Data    json.RawMessage `json:"data"`
    Message string          `json:"message"`
}
```

### Domain-Specific Types

```go
type DomainInfo struct {
    Name   string `json:"name"`
    Cname  string `json:"cname"`
    Type   string `json:"type"`
    Status string `json:"status"`
}

type DomainCreateParams struct {
    Domain string `json:"domain"`
    Type   string `json:"type"`
    Source string `json:"source"`
}
```

### Stats Parameters

```go
type StatsQueryParams struct {
    StartTime string   `json:"start_time"`
    EndTime   string   `json:"end_time"`
    Domains   []string `json:"domains"`
    Interval  string   `json:"interval,omitempty"`
}
```

### Certificate Types

```go
type CertInfo struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Domain string `json:"domain"`
    Expiry string `json:"expiry"`
}

type CertUploadParams struct {
    Name        string `json:"name"`
    Certificate string `json:"certificate"`
    PrivateKey  string `json:"private_key"`
}
```

### Workorder Types

```go
type WorkorderInfo struct {
    ID        string `json:"id"`
    Title     string `json:"title"`
    Status    string `json:"status"`
    CreatedAt string `json:"created_at"`
}

type WorkorderCreateParams struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Type        string `json:"type"`
}
```

## Error Handling

All errors flow through the same mechanism:

1. **API Client layer**: Returns `*api.AuthError` for double-401, wraps network errors with `fmt.Errorf`
2. **Execute function layer**: Wraps API errors, checks `resp.Code != 1` for business logic errors
3. **CLI layer** (Cobra `RunE`): Returns the error to `Execute()` in `cmd/root.go`, which maps error types to exit codes:
   - `*AuthError` / `*api.AuthError` → exit 1
   - `*NetworkError` → exit 2
   - `*InputError` → exit 3
   - Other → exit 1
4. **ACP layer**: Returns the error directly as a JSON-RPC InternalError (-32603)

The `newAuthenticatedClient()` helper returns a clear "not logged in" message when credentials are nil, ensuring requirement 26.1 is met uniformly.

## File Changes Summary

| Action | File | Description |
|--------|------|-------------|
| Modify | `internal/api/client.go` | Add `Put` and `Delete` methods |
| Create | `cmd/domain.go` | All domain subcommands + execute functions |
| Create | `cmd/cache.go` | Cache purge/prefetch/policy commands |
| Create | `cmd/stats.go` | Statistics query commands |
| Create | `cmd/cert.go` | Certificate management commands |
| Create | `cmd/workorder.go` | Work order management commands |
| Create | `cmd/log.go` | Log download commands |
| Modify | `cmd/serve.go` | Register all new ACP tools |
| Delete | `cmd/domains.go` | Replaced by `cmd/domain.go` |

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system — essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: HTTP Method Response Contract

*For any* valid endpoint string and marshallable body value, calling `Put` or `Delete` on the API client SHALL return either a non-nil `json.RawMessage` with a nil error, or a nil `json.RawMessage` with a non-nil error — never both nil and never both non-nil error with non-nil response.

**Validates: Requirements 1.1, 1.2**

### Property 2: Request Authorization Headers

*For any* request sent by the API client (Get, Post, Put, or Delete), the HTTP request SHALL include an `Authorization: Bearer <token>` header where token is non-empty. Additionally, for Post, Put, and Delete requests with a body, the request SHALL include a `Content-Type: application/json` header.

**Validates: Requirements 1.3, 1.4**

### Property 3: Single Retry on 401

*For any* API request (Put or Delete) that receives a 401 response on the first attempt, the client SHALL clear the token cache, re-authenticate, and retry exactly once. If the retry succeeds (non-401), the successful response SHALL be returned.

**Validates: Requirements 1.5, 1.6**

### Property 4: Double 401 Produces AuthError

*For any* API endpoint and request parameters, if both the initial request and the retry attempt return HTTP 401, the client SHALL return an error of type `*AuthError`.

**Validates: Requirements 1.7**

### Property 5: List Output Completeness

*For any* list of items returned by the API (domains, certificates, work orders, etc.), the formatted output string SHALL contain every item's primary identifier (name, ID, title) from the response data.

**Validates: Requirements 3.1, 21.1, 23.1, 24.1**

### Property 6: API Error Display

*For any* API response with `code != 1` and a non-empty message string, the execute function SHALL return an error whose message contains the API's error message text.

**Validates: Requirements 3.5, 27.4**

### Property 7: Success Confirmation Content

*For any* successful mutation operation (create, enable, disable, delete, purge, etc.), the returned confirmation string SHALL contain the affected resource's identifier (domain name, task ID, work order ID, etc.).

**Validates: Requirements 4.3, 14.3, 27.3**

### Property 8: ACP Tool Registration Completeness

*For any* CLI subcommand that calls a Racore API endpoint, there SHALL exist a corresponding ACP tool registration with a non-empty name, a non-empty description, and a valid JSON Schema for inputSchema.

**Validates: Requirements 25.1, 25.2**

### Property 9: ACP-CLI Logic Equivalence

*For any* valid set of input parameters to an ACP tool handler, the text content returned by the handler SHALL be identical to the string returned by calling the corresponding execute function with the same parameters.

**Validates: Requirements 25.3**

### Property 10: ACP Response Format

*For any* ACP tool call that succeeds, the result SHALL be a map containing a "content" key with an array of maps each having "type": "text" and a non-empty "text" value. *For any* ACP tool call that fails, the server SHALL return a JSON-RPC error response with a non-empty message.

**Validates: Requirements 25.4, 27.5**

### Property 11: Authentication Prerequisite

*For any* execute function that calls the API client, if the credential store returns nil credentials, the function SHALL return an error containing the text "not logged in" (case-insensitive).

**Validates: Requirements 26.1**

### Property 12: Table Output Format

*For any* list command that returns multiple items, the formatted output SHALL begin with a header line containing column names separated by whitespace, followed by one data line per item with the same column structure.

**Validates: Requirements 27.1**
