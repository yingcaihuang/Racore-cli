# Design Document: racore-cli-acp

## Overview

racore-cli 是一个 Go CLI 工具，为 Racore Cloud CDN 平台提供命令行管理能力，并通过 ACP（Agent Communication Protocol）协议允许 AI Agent 以 JSON-RPC 方式调用 CLI 能力。项目采用 Go 1.22+、Cobra 框架、标准库网络与加密模块。

---

## Architecture

### 系统分层

```
┌──────────────────────────────────────────────────────────────┐
│                      Entry Points                            │
│  ┌─────────────┐      ┌──────────────────────────────────┐  │
│  │  CLI (Cobra) │      │  ACP Server (stdin/stdout)       │  │
│  └──────┬──────┘      └──────────────┬───────────────────┘  │
│         │                            │                       │
├─────────┼────────────────────────────┼───────────────────────┤
│         │       Command Layer        │                       │
│  ┌──────▼──────────────────────────▼─────────────────────┐  │
│  │  cmd/login  cmd/whoami  cmd/logout  cmd/domains        │  │
│  └──────────────────────┬────────────────────────────────┘  │
│                         │                                    │
├─────────────────────────┼────────────────────────────────────┤
│                   Core Services                              │
│  ┌──────────────┐ ┌──────────────┐ ┌─────────────────────┐  │
│  │ Auth Manager │ │  API Client  │ │  Credential Store   │  │
│  └──────┬───────┘ └──────┬───────┘ └─────────┬───────────┘  │
│         │                │                    │              │
├─────────┼────────────────┼────────────────────┼──────────────┤
│         │           Platform                  │              │
│  ┌──────▼───────────────▼─────────────────────▼───────────┐  │
│  │  crypto/hmac    net/http    os (filesystem)             │  │
│  └─────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

### 模块职责

| 模块 | 职责 |
|------|------|
| `cmd/` | Cobra 命令定义，参数解析，输出格式化 |
| `internal/auth` | HMAC-SHA512 签名计算，Token 获取 |
| `internal/credential` | 凭证文件读写，权限管理，内存清理 |
| `internal/api` | HTTP 客户端封装，Bearer 鉴权，401 重试 |
| `internal/acp` | JSON-RPC 协议解析，请求路由，工具注册 |

---

## Components

### 1. Credential Store (`internal/credential`)

负责凭证的持久化存储与安全读写。

```go
package credential

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

// Credentials 表示持久化的凭证数据
type Credentials struct {
    AccessKey string `json:"access_key"`
    SecretKey string `json:"secret_key"`
    Token     string `json:"token,omitempty"`
    Expire    int64  `json:"expire,omitempty"`
}

// Store 管理凭证文件的读写
type Store struct {
    dir  string // ~/.racore/
    file string // ~/.racore/credentials
}

// NewStore 创建凭证存储实例
func NewStore() (*Store, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("cannot determine home directory: %w", err)
    }
    dir := filepath.Join(home, ".racore")
    return &Store{
        dir:  dir,
        file: filepath.Join(dir, "credentials"),
    }, nil
}

// Save 持久化凭证到磁盘（0700 目录 + 0600 文件）
func (s *Store) Save(creds *Credentials) error

// Load 从磁盘加载凭证，文件不存在返回 (nil, nil)
func (s *Store) Load() (*Credentials, error)

// Delete 删除凭证文件
func (s *Store) Delete() error

// CheckPermissions 检查文件权限，超过 0600 返回 warning
func (s *Store) CheckPermissions() (warning string, err error)

// ClearSensitive 将 Credentials 中敏感字段对应的字节切片清零
func ClearSensitive(creds *Credentials)
```

### 2. Auth Manager (`internal/auth`)

负责 HMAC-SHA512 签名计算和 Token 生命周期管理。

```go
package auth

import (
    "crypto/hmac"
    "crypto/sha512"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

const (
    AuthEndpoint          = "https://portal.racorecloud.com/API/OAuth/token"
    TokenRefreshThreshold = 300 // seconds
    RequestTimeout        = 30 * time.Second
)

// ComputeSignature 计算 HMAC-SHA512 签名
// message = dateStr + accessKey + secretKey
// key = secretKey
// 返回小写十六进制字符串
func ComputeSignature(accessKey, secretKey, dateStr string) string {
    message := dateStr + accessKey + secretKey
    mac := hmac.New(sha512.New, []byte(secretKey))
    mac.Write([]byte(message))
    return hex.EncodeToString(mac.Sum(nil))
}

// FormatRFC1123 将时间格式化为 RFC1123 (HTTP date) 字符串
func FormatRFC1123(t time.Time) string {
    return t.UTC().Format(time.RFC1123)
}

// TokenResponse 表示 OAuth 接口返回的数据结构
type TokenResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        Token  string `json:"token"`
        Expire int64  `json:"expire"`
    } `json:"data"`
}

// Manager 管理 Token 获取与缓存
type Manager struct {
    accessKey string
    secretKey string
    token     string
    expire    int64
    client    *http.Client
}

// NewManager 创建认证管理器
func NewManager(accessKey, secretKey string) *Manager

// SetCache 从已存储的凭证恢复缓存状态
func (m *Manager) SetCache(token string, expire int64)

// GetValidToken 获取有效 Token，必要时自动刷新
func (m *Manager) GetValidToken() (string, error)

// IsTokenValid 判断缓存 Token 是否有效（距过期 >= 300s）
func (m *Manager) IsTokenValid() bool

// Authenticate 执行认证请求获取新 Token
func (m *Manager) Authenticate() (*TokenResponse, error)

// ClearCache 清除 Token 缓存
func (m *Manager) ClearCache()
```

### 3. API Client (`internal/api`)

封装 Racore Cloud API 调用，处理 Bearer Token 鉴权与 401 重试。

```go
package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "racore-cli/internal/auth"
)

const RequestTimeout = 30 * time.Second

// Client 封装 API 调用
type Client struct {
    authMgr  *auth.Manager
    baseURL  string
    client   *http.Client
}

// NewClient 创建 API 客户端
func NewClient(authMgr *auth.Manager) *Client

// Get 发送 GET 请求，自动处理 Bearer 鉴权和 401 重试
func (c *Client) Get(endpoint string, queryParams map[string]string) (json.RawMessage, error)

// Post 发送 POST 请求，自动处理 Bearer 鉴权和 401 重试
func (c *Client) Post(endpoint string, body interface{}) (json.RawMessage, error)
```

**401 重试流程：**

```
GET/POST request
    ↓
response.status == 401?
    ├── No → parse response → return
    └── Yes → ClearCache() → GetValidToken() → retry request
                  ↓
           retry.status == 401?
               ├── No → parse response → return
               └── Yes → return AuthError
```

### 4. ACP Server (`internal/acp`)

通过 stdin/stdout 进行 JSON-RPC 2.0 通信，实现 ACP 协议。

```go
package acp

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "os"
)

// JSON-RPC 2.0 结构
type Request struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      interface{}     `json:"id"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// 标准 JSON-RPC 错误码
const (
    ParseError     = -32700
    InvalidRequest = -32600
    MethodNotFound = -32601
    InvalidParams  = -32602
    InternalError  = -32603
)

// ToolDefinition 工具定义
type ToolDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    InputSchema json.RawMessage `json:"inputSchema"`
}

// ToolHandler 工具执行函数签名
type ToolHandler func(params json.RawMessage) (interface{}, error)

// Server ACP 服务器
type Server struct {
    tools    map[string]ToolDefinition
    handlers map[string]ToolHandler
    reader   *bufio.Scanner
    writer   io.Writer
    logger   io.Writer // stderr
}

// NewServer 创建 ACP 服务器（stdin 读取，stdout 写入，stderr 日志）
func NewServer(reader io.Reader, writer io.Writer, logger io.Writer) *Server

// RegisterTool 注册一个 ACP 工具
func (s *Server) RegisterTool(def ToolDefinition, handler ToolHandler)

// Serve 启动主循环，逐行读取 JSON-RPC 请求并处理
func (s *Server) Serve() error

// handleRequest 路由并执行请求
func (s *Server) handleRequest(req *Request) *Response

// sendResponse 序列化并写入响应到 stdout
func (s *Server) sendResponse(resp *Response) error
```

### 5. Credential Input Resolver (`internal/credential`)

处理凭证来源的优先级 fallback。

```go
// ResolveInput 按优先级解析凭证来源
// 优先级: flags > env vars > interactive prompt
type InputSource struct {
    FlagAccessKey string
    FlagSecretKey string
}

// Resolve 返回最终的 accessKey 和 secretKey
func Resolve(src InputSource) (accessKey, secretKey string, err error)
```

逻辑：
1. 如果 `FlagAccessKey` 和 `FlagSecretKey` 非空 → 直接返回
2. 如果环境变量 `RACORE_ACCESS_KEY` 和 `RACORE_SECRET_KEY` 非空 → 返回环境变量值
3. 否则 → 交互式提示用户输入（secret_key 输入时关闭回显）

---

## Data Models

### 凭证文件格式 (`~/.racore/credentials`)

```json
{
    "access_key": "your-access-key",
    "secret_key": "your-secret-key",
    "token": "jwt-token-string",
    "expire": 1719000000
}
```

### OAuth 请求体

```json
{
    "access_key": "your-access-key",
    "signature": "hmac-sha512-hex-string"
}
```

### OAuth 响应体

```json
{
    "code": 1,
    "message": "success",
    "data": {
        "token": "jwt-token-string",
        "expire": 1719000000
    }
}
```

### JSON-RPC 请求（ACP）

```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "domains",
        "arguments": {
            "filter": "example.com"
        }
    }
}
```

### JSON-RPC 响应（ACP）

```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "content": [
            {
                "type": "text",
                "text": "Domain list:\n..."
            }
        ]
    }
}
```

### JSON-RPC 错误响应

```json
{
    "jsonrpc": "2.0",
    "id": null,
    "error": {
        "code": -32700,
        "message": "Parse error"
    }
}
```

---

## Interfaces & Data Flow

### login 命令流程

```
User → racore-cli login [--access-key X --secret-key Y]
    │
    ▼
Resolve credentials (flags → env → prompt)
    │
    ▼
auth.ComputeSignature(accessKey, secretKey, dateStr)
    │
    ▼
POST /API/OAuth/token { access_key, signature }
    │
    ├─ code == 1 → credential.Store.Save({accessKey, secretKey, token, expire})
    │                → stdout: "Login successful. Token expires at ..."
    │                → exit 0
    │
    └─ code != 1 → stderr: "Error: <message>"
                   → exit 1
```

### whoami 命令流程

```
User → racore-cli whoami
    │
    ▼
credential.Store.Load()
    │
    ├─ nil → stderr: "Error: Not logged in" → exit 1
    │
    └─ credentials found
        │
        ▼
    Display masked access_key (first4 + "****" + last4)
    Check expire vs now:
        ├─ expire - now >= 300 → "Token Status: Valid (Xh Ym remaining)"
        └─ expire - now < 300  → "Token Status: Expired/Expiring"
```

### domains 命令流程

```
User → racore-cli domains [--filter <domain>]
    │
    ▼
credential.Store.Load() → auth.Manager.SetCache(token, expire)
    │
    ▼
api.Client.Get("/API/cdn/domain", {domain: filter})
    │
    ├─ 401 → retry (see 401 flow above)
    ├─ success (code==1) → format table → stdout
    └─ error → stderr: "Error: <desc>" → exit 1/2
```

### ACP serve 模式流程

```
racore-cli serve
    │
    ▼
acp.NewServer(os.Stdin, os.Stdout, os.Stderr)
    │
    ▼
Register tools: login, whoami, logout, domains
    │
    ▼
server.Serve() — 主循环:
    │
    ├── Read line from stdin
    │     ├─ Invalid JSON → respond {error: -32700}
    │     └─ Valid JSON → parse Request
    │
    ├── Route by method:
    │     ├─ "initialize" → return capabilities
    │     ├─ "tools/list" → return tool definitions
    │     ├─ "tools/call" → execute handler → return result
    │     └─ unknown → respond {error: -32601}
    │
    └── Write Response to stdout (one JSON object per line)
```

---

## Error Handling

### 退出码规范

| Exit Code | 含义 | 触发场景 |
|-----------|------|----------|
| 0 | 成功 | 命令正常完成 |
| 1 | 认证错误 | Token 获取失败、401 重试失败 |
| 2 | 网络错误 | 连接超时、DNS 解析失败 |
| 3 | 输入错误 | 缺少参数、参数格式无效 |

### 错误输出格式

所有错误输出到 stderr，格式统一为：
```
Error: <描述信息>
```

### Panic 恢复

在 `main.go` 入口设置全局 recover：

```go
func main() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Fprintf(os.Stderr, "Error: unexpected panic: %v\n", r)
            debug.PrintStack()
            os.Exit(1)
        }
    }()
    cmd.Execute()
}
```

### ACP Server 错误处理

ACP 模式下错误不会导致进程退出，而是返回 JSON-RPC 错误响应：

| 错误场景 | JSON-RPC Error Code |
|----------|---------------------|
| 无法解析 JSON | -32700 (Parse error) |
| 请求格式不合法 | -32600 (Invalid Request) |
| 未知方法 | -32601 (Method not found) |
| 参数不合法 | -32602 (Invalid params) |
| 工具执行内部错误 | -32603 (Internal error) |

---

## Security Design

### 文件权限

```go
// 创建目录
os.MkdirAll(dir, 0700)

// 写入文件
os.WriteFile(file, data, 0600)
```

### 权限检查

```go
func (s *Store) CheckPermissions() (string, error) {
    info, err := os.Stat(s.file)
    if err != nil {
        return "", err
    }
    mode := info.Mode().Perm()
    if mode&0077 != 0 { // 存在 group/other 权限
        return fmt.Sprintf("WARNING: credentials file has permission %04o, expected 0600", mode), nil
    }
    return "", nil
}
```

### 内存安全

```go
// ClearSensitive 清除敏感数据
func ClearSensitive(creds *Credentials) {
    clear := func(s *string) {
        b := []byte(*s)
        for i := range b {
            b[i] = 0
        }
        *s = ""
    }
    clear(&creds.SecretKey)
    clear(&creds.Token)
}
```

### Access Key 掩码

```go
// MaskAccessKey 掩码显示 access_key
func MaskAccessKey(key string) string {
    if len(key) <= 8 {
        return "****"
    }
    return key[:4] + "****" + key[len(key)-4:]
}
```

---

## Project Structure

```
racore-cli/
├── go.mod
├── go.sum
├── main.go                        # 入口，panic recovery
├── cmd/
│   ├── root.go                    # Cobra root 命令
│   ├── login.go                   # login 子命令
│   ├── whoami.go                  # whoami 子命令
│   ├── logout.go                  # logout 子命令
│   ├── domains.go                 # domains 子命令
│   └── serve.go                   # serve 子命令 (ACP 模式)
├── internal/
│   ├── auth/
│   │   ├── auth.go                # 签名计算 & Token 管理
│   │   └── auth_test.go
│   ├── api/
│   │   ├── client.go              # HTTP 客户端
│   │   └── client_test.go
│   ├── credential/
│   │   ├── store.go               # 凭证存储
│   │   ├── resolve.go             # 凭证来源解析
│   │   └── store_test.go
│   └── acp/
│       ├── server.go              # ACP JSON-RPC 服务器
│       ├── types.go               # JSON-RPC 类型定义
│       └── server_test.go
└── Makefile
```

---

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system — essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Credential Source Priority

*For any* combination of credential sources (flags, environment variables, interactive input), the Resolve function SHALL return the value from the highest-priority non-empty source, following the order: flags > environment variables > interactive prompt.

**Validates: Requirements 1.1, 1.2, 1.5**

### Property 2: HMAC-SHA512 Signature Correctness

*For any* valid accessKey, secretKey, and dateStr strings, ComputeSignature SHALL produce a 128-character lowercase hexadecimal string equal to HMAC-SHA512(key=secretKey, message=dateStr+accessKey+secretKey).

**Validates: Requirements 2.1, 2.2**

### Property 3: RFC1123 Date Format

*For any* time.Time value, FormatRFC1123 SHALL produce a string matching the pattern `[A-Z][a-z]{2}, [0-9]{2} [A-Z][a-z]{2} [0-9]{4} [0-9]{2}:[0-9]{2}:[0-9]{2} GMT`.

**Validates: Requirements 2.4**

### Property 4: Token Validity Threshold

*For any* cached token with expire timestamp, IsTokenValid() SHALL return true if and only if (expire - current_unix_time) >= 300.

**Validates: Requirements 3.2, 3.3**

### Property 5: Credential Serialization Round-Trip

*For any* valid Credentials struct, saving to the Store and then loading SHALL produce a Credentials struct equal to the original.

**Validates: Requirements 4.1, 4.4, 4.5**

### Property 6: Access Key Masking

*For any* access_key string of length > 8, MaskAccessKey SHALL return a string that starts with the first 4 characters of the input, ends with the last 4 characters, and contains only asterisks in between. For strings of length <= 8, it SHALL return "****".

**Validates: Requirements 6.1, 11.5**

### Property 7: Token Status Display Consistency

*For any* expire timestamp, the whoami display SHALL show "Valid" if (expire - now) >= 300, and "Expired/Expiring" if (expire - now) < 300.

**Validates: Requirements 6.2, 6.3**

### Property 8: Domain Filter Query Parameter

*For any* non-empty filter string passed to the domains command, the API request URL SHALL contain the query parameter `domain=<filter>` exactly.

**Validates: Requirements 8.2**

### Property 9: Domain List Formatted Output

*For any* list of domain objects returned by the API (each having name, cname, type, status), the CLI formatted output SHALL contain every domain's name, cname, type, and status values as substrings.

**Validates: Requirements 8.3**

### Property 10: JSON-RPC Response Validity

*For any* input received on stdin by the ACP server, the stdout output SHALL consist exclusively of valid JSON objects (one per line), and no log/diagnostic text SHALL appear on stdout.

**Validates: Requirements 9.1, 9.7**

### Property 11: Malformed JSON Yields Parse Error

*For any* byte sequence that is not valid JSON, when received by the ACP server, the response SHALL have error code -32700.

**Validates: Requirements 9.5**

### Property 12: Unknown Method Yields Method Not Found

*For any* method string not in the set of registered methods (initialize, tools/list, tools/call, notifications/initialized), the ACP server response SHALL have error code -32601.

**Validates: Requirements 9.6**

### Property 13: Error Message Format

*For any* error scenario that outputs to stderr, the message SHALL match the pattern `Error: <non-empty description>`.

**Validates: Requirements 10.5**

### Property 14: File Permission Security Warning

*For any* credentials file whose Unix permission mode has group or other bits set (mode & 0077 != 0), the Store.CheckPermissions function SHALL return a non-empty warning string.

**Validates: Requirements 11.3**
