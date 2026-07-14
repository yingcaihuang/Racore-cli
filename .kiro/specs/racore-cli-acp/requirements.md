# Requirements Document

## Introduction

racore-cli 是一个使用 Go 语言开发的命令行工具，通过 ACP（Agent Communication Protocol）协议与 AI Agent 交互，提供 Racore Cloud 平台的 CDN 管理能力。初始版本实现认证命令（login/whoami/logout）和一个示例 CDN 命令（域名列表），验证端到端流程。CLI 通过 stdin/stdout 进行 JSON 通信，遵循 ACP 协议规范。

## Glossary

- **CLI**: racore-cli 命令行工具主程序，基于 Go + Cobra 框架构建
- **Credential_Store**: 位于 ~/.racore/credentials 的 JSON 文件，存储 access_key、secret_key、缓存的 token 及其过期时间
- **Auth_Manager**: 负责 HMAC-SHA512 签名计算、Token 获取与缓存刷新的认证管理模块
- **API_Client**: 封装 Racore Cloud API HTTP 调用逻辑的客户端模块，处理 Bearer Token 鉴权与 401 重试
- **ACP_Server**: 通过 stdin/stdout 进行 JSON-RPC 通信的 ACP 协议服务端模块
- **Token**: Racore Cloud API 返回的 Bearer 访问令牌，用于后续 API 调用的身份验证
- **Signature**: 使用 HMAC-SHA512 算法对 message（dateStr + accessKey + secretKey）用 secretKey 签名后的十六进制字符串

## Requirements

### Requirement 1: 凭证输入与 Fallback

**User Story:** As a developer, I want to provide credentials through multiple methods, so that I can use the tool flexibly in different environments (interactive, CI/CD, scripted).

#### Acceptance Criteria

1. WHEN the user executes a login command with --access-key and --secret-key flags, THE CLI SHALL use the provided flag values as credentials.
2. WHEN the user executes a login command without credential flags and environment variables RACORE_ACCESS_KEY and RACORE_SECRET_KEY are set, THE CLI SHALL use the environment variable values as credentials.
3. WHEN the user executes a login command without credential flags and without environment variables set, THE CLI SHALL prompt the user interactively for access_key and secret_key via stdin.
4. WHEN the user provides credentials through interactive prompt, THE CLI SHALL mask the secret_key input (display asterisks instead of characters).
5. THE CLI SHALL evaluate credential sources in priority order: command-line flags, environment variables, interactive prompt.

### Requirement 2: HMAC-SHA512 认证签名

**User Story:** As a developer, I want the CLI to compute authentication signatures correctly, so that I can obtain valid API tokens from Racore Cloud.

#### Acceptance Criteria

1. WHEN the Auth_Manager computes a signature, THE Auth_Manager SHALL construct the message as the concatenation of dateStr (RFC1123 format), accessKey, and secretKey in that exact order.
2. WHEN the Auth_Manager computes a signature, THE Auth_Manager SHALL use HMAC-SHA512 with secretKey as the signing key and output the result as a lowercase hexadecimal string.
3. WHEN the Auth_Manager requests a token, THE Auth_Manager SHALL send a POST request to https://portal.racorecloud.com/API/OAuth/token with body containing access_key and signature fields.
4. WHEN the Auth_Manager requests a token, THE Auth_Manager SHALL include the header x-request-date with the current time in RFC1123 format.
5. WHEN the Auth_Manager requests a token, THE Auth_Manager SHALL include the header Content-Type with value application/json.

### Requirement 3: Token 缓存与自动刷新

**User Story:** As a developer, I want the CLI to cache tokens and refresh them automatically, so that I avoid unnecessary authentication requests and token expiry failures.

#### Acceptance Criteria

1. WHEN authentication succeeds (response code equals 1), THE Auth_Manager SHALL store the token and expire (unix timestamp) values in the Credential_Store.
2. WHEN a cached token exists and its expire value minus the current unix timestamp is greater than or equal to 300 seconds, THE Auth_Manager SHALL reuse the cached token without re-authenticating.
3. WHEN a cached token exists and its expire value minus the current unix timestamp is less than 300 seconds, THE Auth_Manager SHALL perform a new authentication request to obtain a fresh token.
4. WHEN the Credential_Store file does not exist, THE Auth_Manager SHALL treat the state as unauthenticated and require a new login.

### Requirement 4: 凭证持久化存储

**User Story:** As a developer, I want my credentials persisted to disk, so that I do not need to re-enter them on every CLI invocation.

#### Acceptance Criteria

1. WHEN login succeeds, THE CLI SHALL write the access_key, secret_key, token, and expire to ~/.racore/credentials in JSON format.
2. THE CLI SHALL create the ~/.racore/ directory with permission mode 0700 if the directory does not exist.
3. THE CLI SHALL write the credentials file with permission mode 0600.
4. WHEN the credentials file already exists, THE CLI SHALL overwrite the file content with the new credentials.
5. WHEN any command requires authentication, THE CLI SHALL read credentials from ~/.racore/credentials if the file exists.

### Requirement 5: login 命令

**User Story:** As a developer, I want a login command, so that I can authenticate with Racore Cloud and store my credentials for subsequent commands.

#### Acceptance Criteria

1. WHEN the user executes "racore-cli login", THE CLI SHALL obtain credentials following the fallback priority (flags → environment variables → interactive prompt).
2. WHEN credentials are obtained, THE CLI SHALL perform HMAC-SHA512 authentication against the Racore Cloud OAuth endpoint.
3. WHEN authentication response code equals 1, THE CLI SHALL persist credentials and token to the Credential_Store and output a success message including the token expiry time.
4. IF the authentication response code does not equal 1, THEN THE CLI SHALL output the error message from the response and exit with a non-zero exit code.
5. IF a network error or timeout (30 seconds) occurs during authentication, THEN THE CLI SHALL output a descriptive error message and exit with a non-zero exit code.

### Requirement 6: whoami 命令

**User Story:** As a developer, I want a whoami command, so that I can verify which account is currently authenticated and check the token status.

#### Acceptance Criteria

1. WHEN the user executes "racore-cli whoami" and valid credentials exist in the Credential_Store, THE CLI SHALL display the access_key (masked: show first 4 and last 4 characters with asterisks in between) and token expiry status.
2. WHEN the user executes "racore-cli whoami" and the cached token is still valid (more than 300 seconds until expiry), THE CLI SHALL display "Token Status: Valid" and the remaining validity duration.
3. WHEN the user executes "racore-cli whoami" and the cached token is expired or within 300 seconds of expiry, THE CLI SHALL display "Token Status: Expired/Expiring" and indicate re-authentication is needed.
4. IF the user executes "racore-cli whoami" and no credentials file exists, THEN THE CLI SHALL output "Not logged in" and exit with a non-zero exit code.

### Requirement 7: logout 命令

**User Story:** As a developer, I want a logout command, so that I can clear stored credentials and revoke local access.

#### Acceptance Criteria

1. WHEN the user executes "racore-cli logout" and a credentials file exists, THE CLI SHALL delete the ~/.racore/credentials file and output a success message.
2. WHEN the user executes "racore-cli logout" and no credentials file exists, THE CLI SHALL output "Already logged out" and exit with exit code 0.
3. IF the credentials file cannot be deleted due to a filesystem error, THEN THE CLI SHALL output the error details and exit with a non-zero exit code.

### Requirement 8: CDN 域名列表命令

**User Story:** As a developer, I want a domains command, so that I can list CDN domains managed by my Racore Cloud account and verify end-to-end API connectivity.

#### Acceptance Criteria

1. WHEN the user executes "racore-cli domains", THE API_Client SHALL send a GET request to /API/cdn/domain with a valid Bearer token in the Authorization header.
2. WHEN the user executes "racore-cli domains --filter <domain>", THE API_Client SHALL include the domain query parameter with the specified filter value.
3. WHEN the API response code equals 1, THE CLI SHALL display the domain list in a formatted table showing domain name, CNAME, type, and status.
4. IF the API returns HTTP 401, THEN THE API_Client SHALL clear the cached token, re-authenticate, and retry the request once.
5. IF the retry also returns HTTP 401, THEN THE CLI SHALL output "Authentication failed: please run 'racore-cli login' to re-authenticate" and exit with a non-zero exit code.
6. IF a network error or timeout (30 seconds) occurs, THEN THE CLI SHALL output a descriptive error message and exit with a non-zero exit code.

### Requirement 9: ACP 协议通信

**User Story:** As an AI agent, I want to communicate with racore-cli via ACP protocol over stdin/stdout, so that I can invoke CLI capabilities programmatically.

#### Acceptance Criteria

1. WHEN the CLI is started with the "serve" subcommand, THE ACP_Server SHALL read JSON-RPC requests from stdin and write JSON-RPC responses to stdout.
2. THE ACP_Server SHALL implement the ACP initialize handshake, responding with server capabilities and supported tool definitions.
3. WHEN the ACP_Server receives a tools/list request, THE ACP_Server SHALL return the list of available tools (login, whoami, logout, domains) with their parameter schemas.
4. WHEN the ACP_Server receives a tools/call request with a valid tool name and parameters, THE ACP_Server SHALL execute the corresponding command logic and return the result as a JSON-RPC response.
5. IF the ACP_Server receives a malformed JSON request, THEN THE ACP_Server SHALL respond with a JSON-RPC error (code -32700: Parse error).
6. IF the ACP_Server receives a request for an unknown method, THEN THE ACP_Server SHALL respond with a JSON-RPC error (code -32601: Method not found).
7. THE ACP_Server SHALL write diagnostic and log messages exclusively to stderr, keeping stdout reserved for JSON-RPC responses.

### Requirement 10: 错误处理与退出码

**User Story:** As a developer, I want consistent error handling and meaningful exit codes, so that I can script the CLI and diagnose failures easily.

#### Acceptance Criteria

1. WHEN a command executes successfully, THE CLI SHALL exit with exit code 0.
2. WHEN a command fails due to authentication errors, THE CLI SHALL exit with exit code 1.
3. WHEN a command fails due to network errors or timeouts, THE CLI SHALL exit with exit code 2.
4. WHEN a command fails due to invalid input or missing arguments, THE CLI SHALL exit with exit code 3.
5. THE CLI SHALL output all error messages to stderr in the format "Error: <description>".
6. IF an unexpected panic occurs, THEN THE CLI SHALL recover the panic, log the stack trace to stderr, and exit with exit code 1.

### Requirement 11: 安全性

**User Story:** As a developer, I want my credentials to be stored securely, so that other users on the system cannot access my Racore Cloud credentials.

#### Acceptance Criteria

1. THE CLI SHALL set the ~/.racore/ directory permission to 0700 (owner read/write/execute only).
2. THE CLI SHALL set the ~/.racore/credentials file permission to 0600 (owner read/write only).
3. THE CLI SHALL validate file permissions on the credentials file before reading; IF the file permissions are more permissive than 0600, THEN THE CLI SHALL output a security warning to stderr.
4. THE CLI SHALL clear sensitive data (secret_key, token) from memory after use by overwriting the corresponding byte slices with zeros.
5. WHEN displaying credentials (e.g., whoami), THE CLI SHALL mask the secret_key entirely and mask the access_key showing only the first 4 and last 4 characters.
