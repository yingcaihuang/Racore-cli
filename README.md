# Racore CLI

Racore Cloud CDN 管理命令行工具，支持 MCP 协议，可被 AI Agent（如 Kiro、AWS Q Developer、Claude Desktop 等）直接调用。

## 功能特性

- 完整的 Racore Cloud CDN API 覆盖（60+ 操作）
- 支持 MCP (Model Context Protocol) 协议
- 可被 AI Agent 直接调用（Kiro、AWS Q Developer、Claude Desktop 等）
- 6 大命令组：domain、cache、stats、cert、workorder、log
- 自动 Token 管理（401 重试、过期自动刷新）
- 凭证安全存储（~/.racore/credentials，0600 权限）

## 安装

### 从源码编译

```bash
git clone <repo-url>
cd Racore-cli
make build
```

### 安装到 PATH

```bash
make install  # 复制到 $GOPATH/bin/
```

## 快速开始

### 1. 登录认证

```bash
racore-cli login --access-key YOUR_KEY --secret-key YOUR_SECRET
```

### 2. 查看认证状态

```bash
racore-cli whoami
```

### 3. 使用命令

```bash
racore-cli domain list
racore-cli stats flow --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com
```

## 在 AI Agent 中使用（MCP 协议）

### 协议说明

racore-cli 实现了 MCP (Model Context Protocol) 协议（基于 JSON-RPC 2.0），通过 `racore-cli serve` 命令启动 stdio 模式的 MCP Server。支持的方法：

- `initialize` — 初始化握手
- `tools/list` — 列出所有可用工具
- `tools/call` — 调用指定工具

响应格式遵循 MCP 规范：

- 成功：`{"content": [{"type": "text", "text": "..."}]}`
- 业务错误：`{"content": [{"type": "text", "text": "error message"}], "isError": true}`
- 协议错误：JSON-RPC error object

### 在 Kiro 中配置

在项目根目录或用户级别创建 `.kiro/settings/mcp.json`：

```json
{
  "mcpServers": {
    "racore": {
      "command": "/path/to/racore-cli",
      "args": ["serve"],
      "env": {},
      "disabled": false
    }
  }
}
```

或者如果 racore-cli 在 PATH 中：

```json
{
  "mcpServers": {
    "racore": {
      "command": "racore-cli",
      "args": ["serve"],
      "disabled": false
    }
  }
}
```

### 在 AWS Q Developer CLI 中配置

编辑 `~/.aws/amazonq/mcp.json`：

```json
{
  "mcpServers": {
    "racore": {
      "command": "/path/to/racore-cli",
      "args": ["serve"],
      "env": {}
    }
  }
}
```

### 在 Claude Desktop 中配置

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`（macOS）：

```json
{
  "mcpServers": {
    "racore": {
      "command": "/path/to/racore-cli",
      "args": ["serve"]
    }
  }
}
```

### 在 VS Code (Copilot/Continue) 中配置

在项目 `.vscode/mcp.json` 中：

```json
{
  "servers": {
    "racore": {
      "type": "stdio",
      "command": "/path/to/racore-cli",
      "args": ["serve"]
    }
  }
}
```

### 前提条件

使用 MCP 模式前，需要先完成登录认证：

```bash
racore-cli login --access-key YOUR_KEY --secret-key YOUR_SECRET
```

或通过环境变量：

```bash
export RACORE_ACCESS_KEY=your_access_key
export RACORE_SECRET_KEY=your_secret_key
racore-cli login
```

认证信息保存在 `~/.racore/credentials`，MCP 模式会自动读取。

### 手动测试 MCP 模式

```bash
# 启动 serve 模式
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./racore-cli serve

# 列出所有工具
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./racore-cli serve

# 调用工具
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"domain-list","arguments":{}}}' | ./racore-cli serve
```

## 可用的 MCP 工具列表

### 认证工具

| 工具名 | 描述 |
|--------|------|
| login | 使用 access_key 和 secret_key 认证 |
| whoami | 查看当前认证状态 |
| logout | 清除本地凭证 |

### 域名管理 (domain)

| 工具名 | 描述 |
|--------|------|
| domain-list | 列出所有 CDN 域名 |
| domain-create | 创建新域名 |
| domain-delete | 删除域名 |
| domain-enable | 启用域名 |
| domain-disable | 停用域名 |
| domain-source-get/set | 回源配置 |
| domain-ssl-get/set | HTTPS/SSL 配置 |
| domain-enforce-https-get/set | 强制 HTTPS 跳转 |
| domain-ip-filter-get/set | IP 黑白名单 |
| domain-referer-filter-get/set | Referer 防盗链 |
| domain-ua-filter-get/set | UA 黑白名单 |
| domain-origin-protocol-get/set | 回源协议 |
| domain-http2-get/set | HTTP/2 配置 |
| domain-http3-get/set | HTTP/3 配置 |
| domain-tls-version-get/set | 最低 TLS 版本 |
| domain-compress-get/set | 智能压缩 |
| domain-ipv6-get/set | IPv6 配置 |
| domain-cache-policy-get/set | 缓存策略 |
| domain-origin-host-get/set | 回源 Host |
| domain-origin-timeout-get/set | 回源超时 |
| domain-geo-restriction-get/set | 地域访问控制 |
| domain-request-headers-get/set | 请求头配置 |
| domain-response-headers-get/set | 响应头配置 |
| domain-request-header-policy-get/set | 请求头策略 |
| domain-response-header-policy-get/set | 响应头策略 |

### 缓存管理 (cache)

| 工具名 | 描述 |
|--------|------|
| cache-purge | 刷新 CDN 缓存 |
| cache-purge-status | 查询刷新任务状态 |
| cache-prefetch | 预热内容 |
| cache-prefetch-status | 查询预热任务状态 |
| cache-prewarm-regions | 获取可用预热区域 |
| cache-prewarm-pop | 获取 PoP 节点 |
| cache-list-policies | 列出缓存策略 |
| cache-list-origin-request-policies | 列出回源请求策略 |
| cache-list-response-header-policies | 列出响应头策略 |

### 数据统计 (stats)

| 工具名 | 描述 |
|--------|------|
| stats-flow | 流量统计 |
| stats-request | 请求数统计 |
| stats-hit-flow | 命中流量 |
| stats-hit-request | 命中请求数 |
| stats-http-code | HTTP 状态码统计 |
| stats-http-code-detail | HTTP 状态码明细 |
| stats-district | 地区流量分布 |
| stats-iso-country | ISO 国家/地区代码 |
| stats-top-domain | TOP 域名 |
| stats-top-url | TOP URL |
| stats-top-referer | TOP Referer |
| stats-top-ua | TOP User-Agent |

### 证书管理 (cert)

| 工具名 | 描述 |
|--------|------|
| cert-list | 列出所有证书 |
| cert-upload | 上传证书 |
| cert-update | 更新证书 |
| cert-apply-aws | 申请 AWS 证书 |
| cert-validation-info | 证书验证信息 |

### 工单管理 (workorder)

| 工具名 | 描述 |
|--------|------|
| workorder-list | 列出工单 |
| workorder-create | 创建工单 |
| workorder-reopen | 重新打开工单 |
| workorder-delete | 删除工单 |
| workorder-cancel | 取消工单 |
| workorder-close | 关闭工单 |
| workorder-types | 工单类型列表 |
| workorder-log | 工单沟通记录 |
| workorder-send-message | 发送工单消息 |

### 日志管理 (log)

| 工具名 | 描述 |
|--------|------|
| log-list | 日志下载列表 |

## CLI 命令参考

### 域名管理

```bash
racore-cli domain list [--filter name]
racore-cli domain create --domain example.com --type web --source origin.example.com
racore-cli domain delete --domain example.com
racore-cli domain enable --domain example.com
racore-cli domain disable --domain example.com
racore-cli domain ssl get --domain example.com
racore-cli domain ssl set --domain example.com --config '{"cert_id":"xxx"}'
```

### 缓存管理

```bash
racore-cli cache purge --urls "https://example.com/path1,https://example.com/path2"
racore-cli cache purge-status --task-id abc123
racore-cli cache prefetch --urls "https://example.com/file.zip"
racore-cli cache list-policies --domain example.com [--type managed|custom]
```

### 数据统计

```bash
racore-cli stats flow --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com
racore-cli stats top-url --scope yesterday
racore-cli stats top-referer --scope month --sorted referer_count
racore-cli stats iso-country
```

### 证书管理

```bash
racore-cli cert list
racore-cli cert upload --name "My Cert" --cert-file ./cert.pem --key-file ./key.pem
racore-cli cert apply-aws --domain example.com
```

### 工单管理

```bash
racore-cli workorder list
racore-cli workorder types
racore-cli workorder create --title "问题描述" --description "详细说明" --type "技术支持"
```

### 日志下载

```bash
racore-cli log list --domain example.com [--start-date 2026-07-01] [--end-date 2026-07-14]
```

## 项目结构

```
racore-cli/
├── main.go                     # 入口
├── Makefile                    # 构建脚本
├── cmd/
│   ├── root.go                # 根命令
│   ├── serve.go               # MCP Server + 工具注册
│   ├── helpers.go             # 共享辅助函数
│   ├── domain.go              # 域名命令组
│   ├── cache.go               # 缓存命令组
│   ├── stats.go               # 统计命令组
│   ├── cert.go                # 证书命令组
│   ├── workorder.go           # 工单命令组
│   ├── log.go                 # 日志命令组
│   ├── login.go               # 登录命令
│   ├── logout.go              # 登出命令
│   ├── whoami.go              # 认证状态命令
│   └── errors.go              # 错误类型定义
├── internal/
│   ├── mcp/
│   │   ├── server.go          # MCP JSON-RPC Server
│   │   └── types.go           # 协议类型定义
│   ├── api/
│   │   └── client.go          # HTTP API Client (GET/POST/PUT/DELETE + 401 重试)
│   ├── auth/
│   │   └── auth.go            # HMAC-SHA512 签名认证
│   └── credential/
│       ├── store.go           # 凭证存储
│       └── resolve.go         # 凭证解析（flag > env > prompt）
└── test_all_apis.sh           # 集成测试脚本
```

## 安全说明

- 凭证文件存储在 `~/.racore/credentials`，权限为 0600
- Secret Key 在内存中使用后会被清零
- Token 自动刷新，无需手动管理
- 支持环境变量传入凭证，避免命令行历史泄露

## 开发

```bash
make build    # 编译
make test     # 运行测试
make lint     # 代码检查
make clean    # 清理
```

## 许可证

MIT License
