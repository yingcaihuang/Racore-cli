# Implementation Plan: racore-cli-acp

## Overview

使用 Go 1.22+ 和 Cobra 框架实现 racore-cli，一个通过 ACP 协议与 AI Agent 交互的 CDN 管理命令行工具。实现按分层架构推进：项目初始化 → 核心服务层（credential store, auth manager, api client）→ CLI 命令层 → ACP 协议层 → 构建配置。

## Tasks

- [x] 1. 项目初始化与基础结构
  - [x] 1.1 初始化 Go module 和安装依赖
    - 执行 `go mod init racore-cli`
    - 添加依赖: `github.com/spf13/cobra`、`golang.org/x/term`
    - 创建目录结构: `cmd/`, `internal/auth/`, `internal/credential/`, `internal/api/`, `internal/acp/`
    - _Requirements: 全部（项目基础）_

  - [x] 1.2 创建 main.go 入口与 panic recovery
    - 实现 `main()` 函数，包含 `defer recover()` 全局 panic 恢复
    - panic 时输出错误到 stderr 并 `os.Exit(1)`
    - 调用 `cmd.Execute()` 启动 Cobra
    - _Requirements: 10.1, 10.6_

  - [x] 1.3 创建 Cobra root 命令 (`cmd/root.go`)
    - 定义 root 命令 `racore-cli`，设置 version、description
    - 配置 `SilenceUsage: true`, `SilenceErrors: true` 避免 Cobra 默认错误输出干扰 ACP
    - _Requirements: 10.5_

- [x] 2. Credential Store 实现
  - [x] 2.1 实现凭证存储核心 (`internal/credential/store.go`)
    - 实现 `Store` 结构体：`NewStore()`, `Save()`, `Load()`, `Delete()`
    - `Save`: 创建 `~/.racore/` (0700)，写入 credentials JSON (0600)
    - `Load`: 文件不存在返回 `(nil, nil)`，存在则反序列化
    - `Delete`: 删除凭证文件
    - 实现 `CheckPermissions()`: 检查文件权限是否超过 0600，超过返回 warning
    - 实现 `ClearSensitive()`: 将 SecretKey、Token 字节清零
    - 实现 `MaskAccessKey()`: 长度 > 8 显示 first4 + "****" + last4，否则返回 "****"
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 11.1, 11.2, 11.3, 11.4, 11.5_

  - [ ]* 2.2 Write property tests for Credential Store
    - **Property 5: Credential Serialization Round-Trip**
    - **Property 6: Access Key Masking**
    - **Property 14: File Permission Security Warning**
    - **Validates: Requirements 4.1, 4.4, 4.5, 6.1, 11.3, 11.5**

  - [x] 2.3 实现凭证来源解析 (`internal/credential/resolve.go`)
    - 实现 `InputSource` 结构体和 `Resolve()` 函数
    - 优先级: flags → 环境变量 (`RACORE_ACCESS_KEY`, `RACORE_SECRET_KEY`) → 交互式提示
    - 交互式输入 secret_key 时使用 `golang.org/x/term` 关闭回显
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [ ]* 2.4 Write property test for credential source priority
    - **Property 1: Credential Source Priority**
    - **Validates: Requirements 1.1, 1.2, 1.5**

- [x] 3. Auth Manager 实现
  - [x] 3.1 实现认证管理器 (`internal/auth/auth.go`)
    - 实现 `ComputeSignature(accessKey, secretKey, dateStr) string`: HMAC-SHA512, message=dateStr+accessKey+secretKey, key=secretKey, 输出小写 hex
    - 实现 `FormatRFC1123(t time.Time) string`: UTC RFC1123 格式
    - 实现 `Manager` 结构体: `NewManager()`, `SetCache()`, `GetValidToken()`, `IsTokenValid()`, `Authenticate()`, `ClearCache()`
    - `Authenticate()`: POST 到 AuthEndpoint, body=`{access_key, signature}`, headers: `x-request-date` (RFC1123), `Content-Type: application/json`
    - `IsTokenValid()`: expire - now >= 300 返回 true
    - `GetValidToken()`: 有效则复用缓存, 否则重新认证
    - HTTP client 超时 30 秒
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4_

  - [ ]* 3.2 Write property tests for Auth Manager
    - **Property 2: HMAC-SHA512 Signature Correctness**
    - **Property 3: RFC1123 Date Format**
    - **Property 4: Token Validity Threshold**
    - **Validates: Requirements 2.1, 2.2, 2.4, 3.2, 3.3**

- [x] 4. API Client 实现
  - [x] 4.1 实现 API 客户端 (`internal/api/client.go`)
    - 实现 `Client` 结构体: `NewClient(authMgr)`, `Get()`, `Post()`
    - 请求自动添加 `Authorization: Bearer <token>` header
    - 401 响应时: 清除缓存 → 重新获取 token → 重试一次
    - 重试仍 401 返回 AuthError
    - HTTP client 超时 30 秒
    - baseURL: `https://portal.racorecloud.com`
    - _Requirements: 8.1, 8.4, 8.5, 8.6_

  - [ ]* 4.2 Write unit tests for API Client
    - 使用 httptest 模拟 401 重试逻辑
    - 测试 Bearer token 正确传递
    - 测试网络超时错误处理
    - _Requirements: 8.4, 8.5, 8.6_

- [x] 5. Checkpoint - 核心模块验证
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. CLI 命令实现
  - [x] 6.1 实现 login 命令 (`cmd/login.go`)
    - 添加 `--access-key` 和 `--secret-key` flags
    - 调用 `credential.Resolve()` 获取凭证
    - 调用 `auth.Manager.Authenticate()` 执行认证
    - 成功 (code==1): 调用 `store.Save()` 持久化，输出成功信息和过期时间
    - 失败: 输出错误到 stderr，exit 1 (认证错误) 或 exit 2 (网络错误)
    - 完成后调用 `credential.ClearSensitive()` 清除内存敏感数据
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 10.1, 10.2, 10.3, 11.4_

  - [x] 6.2 实现 whoami 命令 (`cmd/whoami.go`)
    - 加载凭证，不存在时输出 "Not logged in" 到 stderr，exit 1
    - 检查文件权限，有 warning 输出到 stderr
    - 显示 masked access_key (first4 + "****" + last4)
    - 根据 expire 判断 token 状态: Valid (显示剩余时间) 或 Expired/Expiring
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 11.3, 11.5_

  - [ ]* 6.3 Write property test for whoami display
    - **Property 7: Token Status Display Consistency**
    - **Validates: Requirements 6.2, 6.3**

  - [x] 6.4 实现 logout 命令 (`cmd/logout.go`)
    - 凭证文件存在: 删除文件，输出成功信息
    - 凭证文件不存在: 输出 "Already logged out"，exit 0
    - 文件系统错误: 输出错误到 stderr，exit 非零
    - _Requirements: 7.1, 7.2, 7.3_

  - [x] 6.5 实现 domains 命令 (`cmd/domains.go`)
    - 添加 `--filter` flag
    - 加载凭证，初始化 auth.Manager 和 api.Client
    - 调用 `api.Client.Get("/API/cdn/domain", queryParams)`
    - 成功 (code==1): 格式化表格输出 domain name, cname, type, status
    - 失败: 输出错误，exit 对应退出码
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6_

  - [ ]* 6.6 Write property tests for domains command
    - **Property 8: Domain Filter Query Parameter**
    - **Property 9: Domain List Formatted Output**
    - **Validates: Requirements 8.2, 8.3**

- [x] 7. ACP Server 实现
  - [x] 7.1 实现 JSON-RPC 类型定义 (`internal/acp/types.go`)
    - 定义 `Request`, `Response`, `RPCError`, `ToolDefinition` 结构体
    - 定义 JSON-RPC 错误码常量: ParseError(-32700), InvalidRequest(-32600), MethodNotFound(-32601), InvalidParams(-32602), InternalError(-32603)
    - _Requirements: 9.5, 9.6_

  - [x] 7.2 实现 ACP Server 核心 (`internal/acp/server.go`)
    - 实现 `Server` 结构体: `NewServer(reader, writer, logger)`, `RegisterTool()`, `Serve()`
    - `Serve()`: 逐行从 stdin 读取 JSON，解析为 Request，路由处理
    - 路由: `initialize` → 返回 capabilities, `tools/list` → 返回工具列表, `tools/call` → 执行 handler
    - 未知方法 → MethodNotFound 错误
    - 无效 JSON → ParseError 错误
    - 日志写入 stderr，stdout 仅输出 JSON-RPC 响应
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 9.7_

  - [ ]* 7.3 Write property tests for ACP Server
    - **Property 10: JSON-RPC Response Validity**
    - **Property 11: Malformed JSON Yields Parse Error**
    - **Property 12: Unknown Method Yields Method Not Found**
    - **Validates: Requirements 9.1, 9.5, 9.6, 9.7**

  - [x] 7.4 实现 serve 命令 (`cmd/serve.go`)
    - 创建 ACP Server 实例 (os.Stdin, os.Stdout, os.Stderr)
    - 注册工具: login, whoami, logout, domains（复用命令逻辑，适配 ToolHandler 签名）
    - 定义每个工具的 inputSchema（JSON Schema）
    - 启动 `server.Serve()` 主循环
    - _Requirements: 9.2, 9.3, 9.4_

- [x] 8. Checkpoint - 全功能验证
  - Ensure all tests pass, ask the user if questions arise.

- [x] 9. 错误处理完善与集成
  - [x] 9.1 统一错误处理与退出码
    - 定义错误类型: `AuthError` (exit 1), `NetworkError` (exit 2), `InputError` (exit 3)
    - 在 `cmd/root.go` 中根据错误类型设置正确退出码
    - 所有错误输出到 stderr，格式: `Error: <description>`
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

  - [ ]* 9.2 Write property test for error message format
    - **Property 13: Error Message Format**
    - **Validates: Requirements 10.5**

- [x] 10. Makefile 与构建配置
  - [x] 10.1 创建 Makefile
    - `build`: `go build -o racore-cli .`
    - `test`: `go test ./...`
    - `lint`: `go vet ./...`
    - `clean`: 删除构建产物
    - `install`: 构建并安装到 `$GOPATH/bin`
    - _Requirements: 全部（构建配置）_

- [x] 11. Final checkpoint - 完整集成验证
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties defined in design
- Unit tests validate specific examples and edge cases
- 项目使用 Go 语言，所有代码遵循 Go 标准项目布局
- ACP Server 的工具 handler 复用命令层逻辑，避免代码重复

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2", "1.3"] },
    { "id": 2, "tasks": ["2.1", "2.3"] },
    { "id": 3, "tasks": ["2.2", "2.4", "3.1"] },
    { "id": 4, "tasks": ["3.2", "4.1"] },
    { "id": 5, "tasks": ["4.2", "6.1", "6.2", "6.4"] },
    { "id": 6, "tasks": ["6.3", "6.5", "7.1"] },
    { "id": 7, "tasks": ["6.6", "7.2"] },
    { "id": 8, "tasks": ["7.3", "7.4"] },
    { "id": 9, "tasks": ["9.1"] },
    { "id": 10, "tasks": ["9.2", "10.1"] }
  ]
}
```
