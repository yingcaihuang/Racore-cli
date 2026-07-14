# Implementation Plan: Racore CLI Full API Coverage

## Overview

Extend the Racore CLI to cover the full Racore Cloud CDN API (~60 operations) organized into six command groups: domain, cache, stats, cert, workorder, and log. Each command group lives in a single file, shares a `newAuthenticatedClient()` helper, and registers ACP tools in `cmd/serve.go`.

## Tasks

- [x] 1. Extend API client and create shared helper
  - [x] 1.1 Add `Put` and `Delete` methods to `internal/api/client.go`
    - Add `Put(endpoint string, body interface{}) (json.RawMessage, error)` following the same `doWithRetry` pattern as `Post`
    - Add `Delete(endpoint string, body interface{}) (json.RawMessage, error)` — note: Delete takes a JSON body, not query params
    - Both methods must include `Authorization: Bearer <token>` and `Content-Type: application/json` headers
    - Both methods must support the 401 retry flow
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

  - [x] 1.2 Create `cmd/helpers.go` with `newAuthenticatedClient()` and shared formatting utilities
    - Implement `newAuthenticatedClient() (*api.Client, error)` that loads credentials, checks nil, creates auth manager, sets cache, returns ready client
    - Implement `formatTable(headers []string, rows [][]string) string` using tabwriter
    - Implement `formatKeyValue(pairs map[string]string) string` for single-resource display
    - Define shared `apiResponse` struct with Code, Data (json.RawMessage), Message fields
    - _Requirements: 26.1, 27.1, 27.2_

- [x] 2. Implement domain command group (`cmd/domain.go`)
  - [x] 2.1 Create `cmd/domain.go` with all domain subcommands and execute functions
    - Implement parent `domainCmd` with `Use: "domain"`
    - Implement `domain list` (GET /API/cdn/domain) with `--filter` flag — reuse `executeDomains` logic from serve.go
    - Implement `domain create` (POST /API/cdn/domain) with `--domain`, `--type`, `--source` flags
    - Implement `domain delete` (DELETE /API/cdn/domain) with `--domain` flag — body is JSON
    - Implement `domain enable` (PUT /API/cdn/domain/state/open) and `domain disable` (PUT /API/cdn/domain/state/close)
    - Implement two-level nested subcommands for: `source`, `ssl`, `enforce-https`, `ip-filter`, `referer-filter`, `ua-filter`, `origin-protocol`, `http2`, `http3`, `tls-version`, `compress`, `ipv6`, `cache-policy`, `origin-host`, `origin-timeout`, `geo-restriction`, `request-headers`, `response-headers`, `request-header-policy`, `response-header-policy` — each with `get` and `set` subcommands
    - Each get subcommand takes `--domain` flag and calls the corresponding GET/API endpoint
    - Each set subcommand takes `--domain` plus configuration flags/JSON and calls corresponding PUT/POST endpoint
    - Register all commands via `init()` with proper nesting
    - _Requirements: 2.1, 2.8, 3.1–3.5, 4.1–4.3, 5.1–5.2, 6.1–6.4, 7.1–7.6, 8.1–8.12, 9.1–9.2, 10.1–10.4, 11.1–11.2, 12.1–12.4, 13.1–13.4_

  - [ ]* 2.2 Write unit tests for domain execute functions
    - Test `executeDomainList` with empty and filtered results
    - Test `executeDomainEnable`/`executeDomainDisable` success and error paths
    - Test nil credentials returns "not logged in" error
    - _Requirements: 3.1, 4.3, 26.1_

- [x] 3. Implement cache command group (`cmd/cache.go`)
  - [x] 3.1 Create `cmd/cache.go` with all cache subcommands and execute functions
    - Implement parent `cacheCmd` with `Use: "cache"`
    - Implement `cache purge` (POST /API/cdn/purge) with `--urls` flag (comma-separated)
    - Implement `cache purge-status` (GET /API/cdn/purge/detail) with `--task-id` flag
    - Implement `cache prefetch` (POST /API/cdn/prefetch) with `--urls`, `--region`, `--country` flags
    - Implement `cache prefetch-status` (GET /API/cdn/prefetch/detail) with `--task-id` flag
    - Implement `cache prewarm-regions` (GET /API/aws/prewarm/get/region) with `--url` flag
    - Implement `cache prewarm-pop` (GET /API/aws/prewarm/pop) with `--region` flag
    - Implement `cache list-policies` (GET /API/cdn/list/cache/policies)
    - Implement `cache list-origin-request-policies` (GET /API/cdn/aws/origin/request/policies)
    - Implement `cache list-response-header-policies` (GET /API/cdn/aws/response/headers/policies)
    - Register all commands via `init()`
    - _Requirements: 2.2, 14.1–14.3, 15.1–15.4, 16.1–16.2, 22.1–22.3_

  - [ ]* 3.2 Write unit tests for cache execute functions
    - Test `executeCachePurge` returns task ID on success
    - Test `executeCachePrefetch` with region/country params
    - _Requirements: 14.3, 15.4_

- [x] 4. Implement statistics command group (`cmd/stats.go`)
  - [x] 4.1 Create `cmd/stats.go` with all stats subcommands and execute functions
    - Implement parent `statsCmd` with `Use: "stats"`
    - Implement `stats flow` (POST /API/cdn/statistics/flow) with `--start-time`, `--end-time`, `--domains`, `--interval` flags
    - Implement `stats request` (POST /API/cdn/statistics/request) with same time/domain flags
    - Implement `stats hit-flow` (POST /API/cdn/statistics/hit/flow)
    - Implement `stats hit-request` (POST /API/cdn/statistics/hit/request)
    - Implement `stats http-code` (POST /API/cdn/statistics/http/code)
    - Implement `stats http-code-detail` (POST /API/cdn/statistics/http/code/detail)
    - Implement `stats district` (POST /API/cdn/statistics/district)
    - Implement `stats iso-country` (GET /API/cdn/statistics/iso/country)
    - Implement `stats top-domain` (POST /API/cdn/statistics/top/domain)
    - Implement `stats top-url` (POST /API/cdn/domain/top/url)
    - Implement `stats top-referer` (POST /API/cdn/domain/top/referer)
    - Implement `stats top-ua` (POST /API/cdn/domain/top/ua)
    - Register all commands via `init()`
    - _Requirements: 2.3, 17.1–17.4, 18.1–18.2, 19.1–19.2, 20.1–20.4_

- [x] 5. Implement certificate command group (`cmd/cert.go`)
  - [x] 5.1 Create `cmd/cert.go` with all cert subcommands and execute functions
    - Implement parent `certCmd` with `Use: "cert"`
    - Implement `cert list` (GET /API/cdn/sslcert) — display name, domain, expiry in table
    - Implement `cert upload` (POST /API/cdn/sslcert) with `--name`, `--cert-file`, `--key-file` flags — read file contents
    - Implement `cert update` (PUT /API/cdn/sslcert) with `--id`, `--cert-file`, `--key-file` flags
    - Implement `cert apply-aws` (POST /API/cdn/sslcert/apply) with `--domain` flag
    - Implement `cert validation-info` (GET /API/cdn/sslcert/validation/options) with `--id` flag
    - Register all commands via `init()`
    - _Requirements: 2.4, 21.1–21.5_

- [x] 6. Implement workorder command group (`cmd/workorder.go`)
  - [x] 6.1 Create `cmd/workorder.go` with all workorder subcommands and execute functions
    - Implement parent `workorderCmd` with `Use: "workorder"`
    - Implement `workorder list` (GET /API/user/workorder) — display ID, title, status, created_at in table
    - Implement `workorder create` (POST /API/user/workorder) with `--title`, `--description`, `--type` flags
    - Implement `workorder reopen` (PUT /API/user/workorder) with `--id` flag
    - Implement `workorder delete` (DELETE /API/user/workorder) with `--id` flag — body is JSON
    - Implement `workorder cancel` (PUT /API/user/workorder/cancel) with `--id` flag
    - Implement `workorder close` (PUT /API/user/workorder/close) with `--id` flag
    - Implement `workorder types` (GET /API/user/workorder/category)
    - Implement `workorder log` (GET /API/user/workorder/log) with `--id` flag
    - Implement `workorder send-message` (POST /API/user/workorder/log) with `--id`, `--message` flags
    - Register all commands via `init()`
    - _Requirements: 2.5, 23.1–23.9_

- [x] 7. Implement log command group (`cmd/log.go`)
  - [x] 7.1 Create `cmd/log.go` with log subcommands and execute functions
    - Implement parent `logCmd` with `Use: "log"`
    - Implement `log list` (GET /API/cdn/domain/log) with `--domain`, `--start-date`, `--end-date` flags
    - Display log files with download URLs in table format
    - Register commands via `init()`
    - _Requirements: 2.6, 24.1–24.2_

- [x] 8. Checkpoint - Verify CLI builds and commands register
  - Ensure `go build ./...` succeeds
  - Ensure `go vet ./...` passes
  - Ask the user if questions arise.

- [x] 9. Register ACP tools and clean up
  - [x] 9.1 Update `cmd/serve.go` to register ACP tools for all new command groups
    - Register tools for all domain subcommands (domain-list, domain-create, domain-delete, domain-enable, domain-disable, domain-source-get, domain-source-set, domain-ssl-get, domain-ssl-set, etc.)
    - Register tools for all cache subcommands (cache-purge, cache-purge-status, cache-prefetch, cache-prefetch-status, cache-prewarm-regions, cache-prewarm-pop, cache-list-policies, cache-list-origin-request-policies, cache-list-response-header-policies)
    - Register tools for all stats subcommands (stats-flow, stats-request, stats-hit-flow, stats-hit-request, stats-http-code, stats-http-code-detail, stats-district, stats-iso-country, stats-top-domain, stats-top-url, stats-top-referer, stats-top-ua)
    - Register tools for all cert subcommands (cert-list, cert-upload, cert-update, cert-apply-aws, cert-validation-info)
    - Register tools for all workorder subcommands (workorder-list, workorder-create, workorder-reopen, workorder-delete, workorder-cancel, workorder-close, workorder-types, workorder-log, workorder-send-message)
    - Register tool for log subcommands (log-list)
    - Each tool must have descriptive name, description, and valid JSON Schema inputSchema
    - Each handler calls the corresponding execute function and returns text content block
    - Retain existing login, whoami, logout tools
    - Remove the old `domains` tool registration (replaced by domain-list)
    - _Requirements: 25.1–25.5_

  - [x] 9.2 Delete `cmd/domains.go` and update `executeDomains` reference in serve.go
    - Remove `cmd/domains.go` file
    - The `domain list` execute function in `cmd/domain.go` replaces the old `executeDomains` in serve.go
    - Remove the old `executeDomains` function from serve.go (logic moves to domain.go)
    - _Requirements: 2.7_

- [x] 10. Final checkpoint - Build, vet, and test
  - Run `go build ./...` — must compile cleanly
  - Run `go vet ./...` — must pass
  - Run `go test ./...` — all tests must pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each command group file follows the same pattern: parent command → subcommands → execute functions → init() registration
- The `newAuthenticatedClient()` helper eliminates credential-loading boilerplate across 60+ execute functions
- The `Delete` method takes a JSON body (not query params) per the Racore API contract
- Two-level nesting (e.g., `domain ssl get`) uses intermediate Cobra commands with no `RunE`
- ACP tools call the same execute functions as CLI commands, ensuring logic equivalence (Property 9)
- All stat commands use POST with a JSON body (time range + domains), except `iso-country` which is GET

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2"] },
    { "id": 1, "tasks": ["2.1", "3.1", "4.1", "5.1", "6.1", "7.1"] },
    { "id": 2, "tasks": ["2.2", "3.2"] },
    { "id": 3, "tasks": ["9.1", "9.2"] },
    { "id": 4, "tasks": [] }
  ]
}
```
