# Requirements Document

## Introduction

This feature extends the Racore CLI to provide full coverage of the Racore Cloud CDN API. The CLI will expose all 60+ operations organized into nested subcommand groups: domain, cache, stats, cert, workorder, and log. Each CLI command will also be registered as an ACP tool in the serve command. The existing `cmd/domains.go` will be refactored into the new `cmd/domain.go` with subcommands. The API client will be extended with `Put` and `Delete` methods following the same 401-retry pattern.

## Glossary

- **CLI**: The `racore-cli` command-line interface built with Cobra
- **ACP_Server**: The Agent Communication Protocol server started by `racore-cli serve`, exposing tools over JSON-RPC 2.0
- **API_Client**: The Go HTTP client in `internal/api/client.go` that handles Bearer token auth and 401 retry logic
- **Domain_Group**: The `racore-cli domain` subcommand group for CDN domain management operations
- **Cache_Group**: The `racore-cli cache` subcommand group for purge and prefetch operations
- **Stats_Group**: The `racore-cli stats` subcommand group for statistics and analytics queries
- **Cert_Group**: The `racore-cli cert` subcommand group for SSL certificate management
- **Workorder_Group**: The `racore-cli workorder` subcommand group for support ticket management
- **Log_Group**: The `racore-cli log` subcommand group for log download operations
- **Credential_Store**: The local file-based credential storage used for authentication state
- **Auth_Manager**: The authentication manager that handles token acquisition and caching

## Requirements

### Requirement 1: API Client Extension

**User Story:** As a developer, I want the API client to support PUT and DELETE HTTP methods, so that the CLI can interact with all Racore Cloud API endpoints.

#### Acceptance Criteria

1. THE API_Client SHALL provide a `Put` method that accepts an endpoint string and a body interface and returns a json.RawMessage response.
2. THE API_Client SHALL provide a `Delete` method that accepts an endpoint string and query parameters and returns a json.RawMessage response.
3. WHEN the API_Client sends a PUT request, THE API_Client SHALL include an `Authorization: Bearer <token>` header and a `Content-Type: application/json` header.
4. WHEN the API_Client sends a DELETE request, THE API_Client SHALL include an `Authorization: Bearer <token>` header.
5. WHEN the API_Client receives a 401 response on a PUT request, THE API_Client SHALL clear the token cache, re-authenticate, and retry the request once.
6. WHEN the API_Client receives a 401 response on a DELETE request, THE API_Client SHALL clear the token cache, re-authenticate, and retry the request once.
7. IF the retry attempt also returns a 401, THEN THE API_Client SHALL return an AuthError indicating re-authentication is needed.

### Requirement 2: Command File Organization

**User Story:** As a developer, I want commands organized into one file per group with nested subcommands, so that the codebase remains maintainable as coverage grows.

#### Acceptance Criteria

1. THE CLI SHALL organize domain commands in `cmd/domain.go` as nested subcommands under `racore-cli domain`.
2. THE CLI SHALL organize cache commands in `cmd/cache.go` as nested subcommands under `racore-cli cache`.
3. THE CLI SHALL organize statistics commands in `cmd/stats.go` as nested subcommands under `racore-cli stats`.
4. THE CLI SHALL organize certificate commands in `cmd/cert.go` as nested subcommands under `racore-cli cert`.
5. THE CLI SHALL organize work order commands in `cmd/workorder.go` as nested subcommands under `racore-cli workorder`.
6. THE CLI SHALL organize log commands in `cmd/log.go` as nested subcommands under `racore-cli log`.
7. WHEN the new `cmd/domain.go` is implemented, THE CLI SHALL remove the existing `cmd/domains.go` file.
8. THE CLI SHALL support further nesting for domain sub-features (e.g., `racore-cli domain ssl get`, `racore-cli domain ssl set`).

### Requirement 3: Domain CRUD Operations

**User Story:** As a CDN operator, I want to create, list, and delete domains via the CLI, so that I can manage my CDN configuration without using the web portal.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain list`, THE Domain_Group SHALL call `GET /API/cdn/domain` and display domain name, CNAME, type, and status in a table.
2. WHEN a user provides a `--filter` flag to `domain list`, THE Domain_Group SHALL pass the filter value as the `domain` query parameter.
3. WHEN a user runs `racore-cli domain create` with required parameters, THE Domain_Group SHALL call `POST /API/cdn/domain` with the provided domain configuration.
4. WHEN a user runs `racore-cli domain delete` with a domain identifier, THE Domain_Group SHALL call `DELETE /API/cdn/domain` with the specified domain.
5. IF the API returns a non-success code, THEN THE Domain_Group SHALL display the error message from the API response and exit with a non-zero status code.

### Requirement 4: Domain State Management

**User Story:** As a CDN operator, I want to enable and disable domains, so that I can control CDN traffic without deleting configurations.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain enable` with a domain name, THE Domain_Group SHALL call `PUT /API/cdn/domain/state/open` with the domain parameter.
2. WHEN a user runs `racore-cli domain disable` with a domain name, THE Domain_Group SHALL call `PUT /API/cdn/domain/state/close` with the domain parameter.
3. WHEN the enable or disable operation succeeds, THE Domain_Group SHALL display a confirmation message with the domain name and new state.

### Requirement 5: Domain Origin Server Configuration

**User Story:** As a CDN operator, I want to query and set origin server configurations, so that I can control where the CDN fetches content from.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain source get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/source` and display the origin server configuration.
2. WHEN a user runs `racore-cli domain source set` with domain and origin parameters, THE Domain_Group SHALL call `PUT /API/cdn/domain/source` with the provided configuration.

### Requirement 6: Domain SSL/HTTPS Configuration

**User Story:** As a CDN operator, I want to manage SSL/HTTPS settings per domain, so that I can control encryption and certificate behavior.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain ssl get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/ssl` and display the HTTPS configuration.
2. WHEN a user runs `racore-cli domain ssl set` with domain and SSL parameters, THE Domain_Group SHALL call `PUT /API/cdn/domain/ssl` with the provided settings.
3. WHEN a user runs `racore-cli domain enforce-https get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/enforce/https` and display the force HTTPS redirect status.
4. WHEN a user runs `racore-cli domain enforce-https set` with domain and redirect parameters, THE Domain_Group SHALL call `PUT /API/cdn/domain/enforce/https` with the provided settings.

### Requirement 7: Domain Access Control Filters

**User Story:** As a CDN operator, I want to manage IP, Referer, and User-Agent filters, so that I can restrict access to my CDN content.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain ip-filter get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/ip/filter` and display the IP blacklist/whitelist.
2. WHEN a user runs `racore-cli domain ip-filter set` with domain and filter parameters, THE Domain_Group SHALL call `POST /API/cdn/domain/ip/filter` with the provided configuration.
3. WHEN a user runs `racore-cli domain referer-filter get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/referer/filter` and display the Referer blacklist/whitelist.
4. WHEN a user runs `racore-cli domain referer-filter set` with domain and filter parameters, THE Domain_Group SHALL call `POST /API/cdn/domain/referer/filter` with the provided configuration.
5. WHEN a user runs `racore-cli domain ua-filter get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/user/agent/filter` and display the UA blacklist/whitelist.
6. WHEN a user runs `racore-cli domain ua-filter set` with domain and filter parameters, THE Domain_Group SHALL call `POST /API/cdn/domain/user/agent/filter` with the provided configuration.

### Requirement 8: Domain Protocol and Performance Settings

**User Story:** As a CDN operator, I want to manage HTTP/2, HTTP/3, origin protocol, TLS version, compression, and IPv6 settings, so that I can optimize delivery performance.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain origin-protocol get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/origin/protocol/policy` and display the origin protocol setting.
2. WHEN a user runs `racore-cli domain origin-protocol set` with domain and protocol parameters, THE Domain_Group SHALL call `PUT /API/cdn/domain/origin/protocol/policy` with the provided setting.
3. WHEN a user runs `racore-cli domain http2 get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/http2` and display the HTTP/2 status.
4. WHEN a user runs `racore-cli domain http2 set` with domain and enable/disable flag, THE Domain_Group SHALL call `PUT /API/cdn/domain/http2` with the provided setting.
5. WHEN a user runs `racore-cli domain http3 get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/http3` and display the HTTP/3 status.
6. WHEN a user runs `racore-cli domain http3 set` with domain and enable/disable flag, THE Domain_Group SHALL call `PUT /API/cdn/domain/http3` with the provided setting.
7. WHEN a user runs `racore-cli domain tls-version get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/min/tls/version` and display the minimum TLS version.
8. WHEN a user runs `racore-cli domain tls-version set` with domain and version parameter, THE Domain_Group SHALL call `PUT /API/cdn/domain/min/tls/version` with the provided version.
9. WHEN a user runs `racore-cli domain compress get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/page/compress` and display the compression status.
10. WHEN a user runs `racore-cli domain compress set` with domain and enable/disable flag, THE Domain_Group SHALL call `PUT /API/cdn/domain/page/compress` with the provided setting.
11. WHEN a user runs `racore-cli domain ipv6 get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/ipv6` and display the IPv6 status.
12. WHEN a user runs `racore-cli domain ipv6 set` with domain and enable/disable flag, THE Domain_Group SHALL call `PUT /API/cdn/domain/ipv6` with the provided setting.

### Requirement 9: Domain Cache Policy

**User Story:** As a CDN operator, I want to query and configure cache policies per domain, so that I can control content freshness and caching behavior.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain cache-policy get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/cache/conf` and display the cache policy configuration.
2. WHEN a user runs `racore-cli domain cache-policy set` with domain and policy parameters, THE Domain_Group SHALL call `PUT /API/cdn/domain/cache/conf` with the provided configuration.

### Requirement 10: Domain Origin Host and Timeout

**User Story:** As a CDN operator, I want to configure the origin host header and connection timeout, so that I can control how the CDN communicates with my origin servers.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain origin-host get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/origin/host` and display the origin host setting.
2. WHEN a user runs `racore-cli domain origin-host set` with domain and host parameters, THE Domain_Group SHALL call `PUT /API/cdn/domain/origin/host` with the provided host value.
3. WHEN a user runs `racore-cli domain origin-timeout get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/origin/connection/policy` and display the origin connection timeout.
4. WHEN a user runs `racore-cli domain origin-timeout set` with domain and timeout parameters, THE Domain_Group SHALL call `PUT /API/cdn/domain/origin/connection/policy` with the provided timeout value.

### Requirement 11: Domain Geographic Restriction (AWS Only)

**User Story:** As a CDN operator using AWS channels, I want to configure geographic access controls, so that I can restrict content delivery by country.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain geo-restriction get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/geo/restriction` and display the geographic restriction configuration.
2. WHEN a user runs `racore-cli domain geo-restriction set` with domain and restriction parameters, THE Domain_Group SHALL call `PUT /API/cdn/domain/geo/restriction` with the provided configuration.

### Requirement 12: Domain HTTP Request/Response Headers

**User Story:** As a CDN operator, I want to manage custom HTTP request and response headers, so that I can add, modify, or remove headers at the CDN edge.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain request-headers get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/http/request/headers` and display the configured request headers.
2. WHEN a user runs `racore-cli domain request-headers set` with domain and header parameters, THE Domain_Group SHALL call `POST /API/cdn/domain/http/request/headers` with the provided headers.
3. WHEN a user runs `racore-cli domain response-headers get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/http/response/headers` and display the configured response headers.
4. WHEN a user runs `racore-cli domain response-headers set` with domain and header parameters, THE Domain_Group SHALL call `POST /API/cdn/domain/http/response/headers` with the provided headers.

### Requirement 13: Domain AWS Policy Configuration

**User Story:** As a CDN operator using AWS channels, I want to manage AWS-specific request and response header policies, so that I can leverage CloudFront behaviors.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli domain request-header-policy get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/request/header/policy` and display the AWS origin request header policy.
2. WHEN a user runs `racore-cli domain request-header-policy set` with domain and policy parameters, THE Domain_Group SHALL call `PUT /API/cdn/domain/request/header/policy` with the provided policy.
3. WHEN a user runs `racore-cli domain response-header-policy get` with a domain name, THE Domain_Group SHALL call `GET /API/cdn/domain/response/header/policy` and display the AWS response header policy.
4. WHEN a user runs `racore-cli domain response-header-policy set` with domain and policy parameters, THE Domain_Group SHALL call `PUT /API/cdn/domain/response/header/policy` with the provided policy.

### Requirement 14: Cache Purge Operations

**User Story:** As a CDN operator, I want to purge cached content by file URL or directory, so that I can invalidate stale content immediately.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli cache purge` with a list of URLs or directories, THE Cache_Group SHALL call `POST /API/cdn/purge` with the specified targets.
2. WHEN a user runs `racore-cli cache purge-status` with a task ID, THE Cache_Group SHALL call `GET /API/cdn/purge/detail` and display the purge task status.
3. WHEN the purge request succeeds, THE Cache_Group SHALL display the task ID for status tracking.

### Requirement 15: Cache Prefetch Operations

**User Story:** As a CDN operator, I want to prefetch content to CDN edge nodes, so that I can warm the cache before users request the content.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli cache prefetch` with a list of URLs, THE Cache_Group SHALL call `POST /API/cdn/prefetch` with the specified URLs.
2. WHEN a user provides `--region` or `--country` flags to `cache prefetch`, THE Cache_Group SHALL include the region and country parameters in the prefetch request.
3. WHEN a user runs `racore-cli cache prefetch-status` with a task ID, THE Cache_Group SHALL call `GET /API/cdn/prefetch/detail` and display the prefetch task status.
4. WHEN the prefetch request succeeds, THE Cache_Group SHALL display the task ID for status tracking.

### Requirement 16: Cache Prewarm Regions

**User Story:** As a CDN operator, I want to query available prewarm regions and POP points, so that I can target specific geographic locations for prefetch.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli cache prewarm-regions` with a URL, THE Cache_Group SHALL call `GET /API/aws/prewarm/get/region` and display available regions and countries.
2. WHEN a user runs `racore-cli cache prewarm-pop` with a region parameter, THE Cache_Group SHALL call `GET /API/aws/prewarm/pop` and display available POP points for the specified region.

### Requirement 17: Statistics - Flow and Request Count

**User Story:** As a CDN operator, I want to query bandwidth and request count statistics, so that I can monitor CDN usage and performance.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli stats flow` with required time range and domain parameters, THE Stats_Group SHALL call `POST /API/cdn/statistics/flow` and display consumption details.
2. WHEN a user runs `racore-cli stats request` with required time range and domain parameters, THE Stats_Group SHALL call `POST /API/cdn/statistics/request` and display request count data.
3. WHEN a user runs `racore-cli stats hit-flow` with required time range and domain parameters, THE Stats_Group SHALL call `POST /API/cdn/statistics/hit/flow` and display hit traffic data.
4. WHEN a user runs `racore-cli stats hit-request` with required time range and domain parameters, THE Stats_Group SHALL call `POST /API/cdn/statistics/hit/request` and display hit request count data.

### Requirement 18: Statistics - HTTP Status Codes

**User Story:** As a CDN operator, I want to query HTTP status code distributions, so that I can identify error patterns and optimize configurations.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli stats http-code` with required time range and domain parameters, THE Stats_Group SHALL call `POST /API/cdn/statistics/http/code` and display status code summary data.
2. WHEN a user runs `racore-cli stats http-code-detail` with required time range and domain parameters, THE Stats_Group SHALL call `POST /API/cdn/statistics/http/code/detail` and display detailed status code data.

### Requirement 19: Statistics - Geographic Distribution

**User Story:** As a CDN operator, I want to query traffic distribution by country and region, so that I can understand geographic usage patterns.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli stats district` with required time range and domain parameters, THE Stats_Group SHALL call `POST /API/cdn/statistics/district` and display country/region consumption details.
2. WHEN a user runs `racore-cli stats iso-country` with required parameters, THE Stats_Group SHALL call `GET /API/cdn/statistics/iso/country` and display ISO country/region data.

### Requirement 20: Statistics - Top Rankings

**User Story:** As a CDN operator, I want to query top domains, URLs, referers, and user agents, so that I can identify high-traffic resources and access patterns.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli stats top-domain` with required time range parameters, THE Stats_Group SHALL call `POST /API/cdn/statistics/top/domain` and display top domains by traffic.
2. WHEN a user runs `racore-cli stats top-url` with required time range and domain parameters, THE Stats_Group SHALL call `POST /API/cdn/domain/top/url` and display top URLs by traffic.
3. WHEN a user runs `racore-cli stats top-referer` with required time range and domain parameters, THE Stats_Group SHALL call `POST /API/cdn/domain/top/referer` and display top referers.
4. WHEN a user runs `racore-cli stats top-ua` with required time range and domain parameters, THE Stats_Group SHALL call `POST /API/cdn/domain/top/ua` and display top user agents.

### Requirement 21: Certificate Management

**User Story:** As a CDN operator, I want to upload, update, and list SSL certificates, so that I can manage HTTPS encryption for my domains.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli cert list`, THE Cert_Group SHALL call `GET /API/cdn/sslcert` and display the certificate list with name, domain, and expiry information.
2. WHEN a user runs `racore-cli cert upload` with certificate and key file paths, THE Cert_Group SHALL read the files and call `POST /API/cdn/sslcert` with the certificate data.
3. WHEN a user runs `racore-cli cert update` with certificate ID and new certificate/key data, THE Cert_Group SHALL call `PUT /API/cdn/sslcert` with the updated data.
4. WHEN a user runs `racore-cli cert apply-aws` with required parameters, THE Cert_Group SHALL call `POST /API/cdn/sslcert/apply` to apply for an AWS certificate.
5. WHEN a user runs `racore-cli cert validation-info` with a certificate identifier, THE Cert_Group SHALL call `GET /API/cdn/sslcert/validation/options` and display the validation information.

### Requirement 22: AWS Policy Lists

**User Story:** As a CDN operator, I want to list available AWS cache and header policies, so that I can reference valid policy IDs when configuring domains.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli cache list-policies`, THE Cache_Group SHALL call `GET /API/cdn/list/cache/policies` and display available AWS cache policies.
2. WHEN a user runs `racore-cli cache list-origin-request-policies`, THE Cache_Group SHALL call `GET /API/cdn/aws/origin/request/policies` and display available AWS origin request policies.
3. WHEN a user runs `racore-cli cache list-response-header-policies`, THE Cache_Group SHALL call `GET /API/cdn/aws/response/headers/policies` and display available AWS response header policies.

### Requirement 23: Work Order Management

**User Story:** As a CDN operator, I want to create, view, update, and manage support work orders, so that I can communicate issues to the Racore support team.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli workorder list`, THE Workorder_Group SHALL call `GET /API/user/workorder` and display work orders with ID, title, status, and creation date.
2. WHEN a user runs `racore-cli workorder create` with title and description, THE Workorder_Group SHALL call `POST /API/user/workorder` with the provided details.
3. WHEN a user runs `racore-cli workorder reopen` with a work order ID, THE Workorder_Group SHALL call `PUT /API/user/workorder` to reopen the specified work order.
4. WHEN a user runs `racore-cli workorder delete` with a work order ID, THE Workorder_Group SHALL call `DELETE /API/user/workorder` to delete the specified work order.
5. WHEN a user runs `racore-cli workorder cancel` with a work order ID, THE Workorder_Group SHALL call `PUT /API/user/workorder/cancel` to cancel the specified work order.
6. WHEN a user runs `racore-cli workorder close` with a work order ID, THE Workorder_Group SHALL call `PUT /API/user/workorder/close` to close the specified work order.
7. WHEN a user runs `racore-cli workorder types`, THE Workorder_Group SHALL call `GET /API/user/workorder/category` and display available work order types.
8. WHEN a user runs `racore-cli workorder log` with a work order ID, THE Workorder_Group SHALL call `GET /API/user/workorder/log` and display communication records.
9. WHEN a user runs `racore-cli workorder send-message` with a work order ID and message content, THE Workorder_Group SHALL call `POST /API/user/workorder/log` with the message.

### Requirement 24: Log Download

**User Story:** As a CDN operator, I want to retrieve CDN log download links, so that I can analyze raw access logs for troubleshooting and reporting.

#### Acceptance Criteria

1. WHEN a user runs `racore-cli log list` with a domain and date range, THE Log_Group SHALL call `GET /API/cdn/domain/log` and display available log files with download URLs.
2. WHEN the log list request includes a `--domain` flag, THE Log_Group SHALL pass the domain as a query parameter.

### Requirement 25: ACP Tool Registration

**User Story:** As an AI agent developer, I want every CLI command to be available as an ACP tool, so that agents can invoke any Racore Cloud operation programmatically.

#### Acceptance Criteria

1. WHEN the ACP_Server starts, THE ACP_Server SHALL register one tool for each CLI subcommand across all six command groups.
2. THE ACP_Server SHALL define each tool with a descriptive name, description, and JSON Schema for input parameters.
3. WHEN an ACP tool is called, THE ACP_Server SHALL execute the same logic as the corresponding CLI command and return the result as a text content block.
4. IF an ACP tool call fails, THEN THE ACP_Server SHALL return a JSON-RPC error with a descriptive message.
5. THE ACP_Server SHALL retain the existing login, whoami, and logout tools alongside the new command tools.

### Requirement 26: Authentication Prerequisite

**User Story:** As a CDN operator, I want commands to verify authentication before making API calls, so that I receive clear errors when not logged in.

#### Acceptance Criteria

1. WHILE a user is not logged in, THE CLI SHALL display an error message stating "Not logged in. Please run 'racore-cli login' first" and exit with a non-zero status when any API command is invoked.
2. WHEN the Credential_Store contains expired credentials and the Auth_Manager cannot refresh the token, THE CLI SHALL display an error message indicating re-authentication is required.
3. IF an API call fails with an AuthError after retry, THEN THE CLI SHALL display the authentication error message and exit with status code 1.

### Requirement 27: Consistent Output Formatting

**User Story:** As a CDN operator, I want consistent and readable output across all commands, so that I can parse and understand results predictably.

#### Acceptance Criteria

1. THE CLI SHALL display list data in tabwriter-formatted tables with column headers.
2. THE CLI SHALL display single-resource details as key-value pairs.
3. THE CLI SHALL display success confirmations with the operation performed and affected resource.
4. IF an API returns an error, THEN THE CLI SHALL write the error message to stderr and exit with a non-zero status code.
5. WHEN a command runs in ACP mode, THE CLI SHALL return results as plain text in the ACP content block format.
