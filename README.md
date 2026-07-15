# Racore CLI

Racore Cloud CDN 管理命令行工具，支持 MCP (Model Context Protocol) 协议，可被 AI Agent（如 Kiro、AWS Q Developer、Claude Desktop 等）直接调用。

## 功能特性

- 完整的 Racore Cloud CDN API 覆盖（60+ 操作）
- 支持 MCP (Model Context Protocol) 协议，AI Agent 可直接调用
- 6 大命令组：domain、cache、stats、cert、workorder、log
- 自动 Token 管理（401 重试、过期自动刷新）
- 凭证安全存储（系统钥匙串 / Credential Manager）
- 跨平台：macOS、Linux、Windows 全支持
- 创建域名时自动匹配泛域 SSL 证书

## 安装

所有安装方式安装后均可直接使用 `racore-cli` 命令，**无需手动配置 PATH**。

### macOS

| 方式 | 命令 | PATH |
|------|------|------|
| Homebrew（推荐） | `brew tap yingcaihuang/tap && brew install racore-cli` | ✅ 自动 |
| 一键脚本 | `curl -fsSL https://raw.githubusercontent.com/yingcaihuang/Racore-cli/main/install.sh \| bash` | ✅ 安装到 /usr/local/bin |

```bash
# Homebrew 安装（推荐）
brew tap yingcaihuang/tap
brew install racore-cli

# 升级
brew upgrade racore-cli

# 一键安装（适合没有 Homebrew 的环境）
curl -fsSL https://raw.githubusercontent.com/yingcaihuang/Racore-cli/main/install.sh | bash
```

### Linux

| 方式 | 适用系统 | PATH |
|------|---------|------|
| .deb 包 | Ubuntu / Debian | ✅ 安装到 /usr/local/bin |
| .rpm 包 | CentOS / RHEL / Fedora | ✅ 安装到 /usr/local/bin |
| 一键脚本 | 所有 Linux | ✅ 安装到 /usr/local/bin |

```bash
# Ubuntu / Debian (.deb)
VERSION=$(curl -fsSL https://api.github.com/repos/yingcaihuang/Racore-cli/releases/latest | grep tag_name | cut -d '"' -f4 | tr -d 'v')
curl -LO "https://github.com/yingcaihuang/Racore-cli/releases/download/v${VERSION}/racore-cli_${VERSION}_linux_amd64.deb"
sudo dpkg -i "racore-cli_${VERSION}_linux_amd64.deb"

# CentOS / RHEL / Fedora (.rpm)
VERSION=$(curl -fsSL https://api.github.com/repos/yingcaihuang/Racore-cli/releases/latest | grep tag_name | cut -d '"' -f4 | tr -d 'v')
curl -LO "https://github.com/yingcaihuang/Racore-cli/releases/download/v${VERSION}/racore-cli_${VERSION}_linux_amd64.rpm"
sudo rpm -i "racore-cli_${VERSION}_linux_amd64.rpm"

# ARM64 架构请将 amd64 替换为 arm64

# 一键安装（自动检测架构）
curl -fsSL https://raw.githubusercontent.com/yingcaihuang/Racore-cli/main/install.sh | bash
```

### Windows

| 方式 | 说明 | PATH |
|------|------|------|
| MSI 安装包（推荐） | 双击安装，安装到 Program Files | ✅ 自动添加到系统 PATH |
| Scoop | 包管理器安装 | ✅ Scoop 自动管理 |
| ZIP 解压 | 手动解压 | ❌ 需要手动添加到 PATH |

```powershell
# MSI 安装包（推荐，自动加 PATH）
# 从 https://github.com/yingcaihuang/Racore-cli/releases 下载 racore-cli_xxx_windows_amd64.msi
# 双击安装即可，安装后打开新终端即可使用

# Scoop（自动加 PATH）
scoop bucket add racore https://github.com/yingcaihuang/Racore-cli
scoop install racore-cli

# 升级
scoop update racore-cli
```

### 从源码编译

```bash
git clone https://github.com/yingcaihuang/Racore-cli.git
cd Racore-cli
make build      # 编译到当前目录
make install    # 安装到 $GOPATH/bin/
```

### 验证安装

```bash
racore-cli --version
# 输出: racore-cli version 0.2.1 (commit: xxx, built: 2026-07-15T...)
```

## 快速开始

### 1. 登录认证

```bash
# 交互式（会提示输入）
racore-cli login

# 通过参数
racore-cli login --access-key YOUR_KEY --secret-key YOUR_SECRET

# 通过环境变量
export RACORE_ACCESS_KEY=your_key
export RACORE_SECRET_KEY=your_secret
racore-cli login
```

### 2. 查看认证状态

```bash
racore-cli whoami
# Access Key: AKID****ABCD
# Storage: system keyring
# Token Status: Valid (23h 45m remaining)
```

### 3. 使用命令

```bash
racore-cli domain list
racore-cli domain list --filter bbc           # 模糊搜索
racore-cli stats flow --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com
```

## CLI 命令参考

### 域名管理

```bash
# 列出域名（支持模糊搜索）
racore-cli domain list
racore-cli domain list --filter example.com
racore-cli domain list --filter bbc            # 部分匹配

# 创建域名（自动匹配泛域 SSL 证书）
# type: oversea, live, video, dynamic, static, download
racore-cli domain create --domain cdn.example.com --type oversea --source origin.example.com
racore-cli domain create --domain cdn.example.com --type static --source origin.example.com --cert-id 123

# 启用/停用/删除
racore-cli domain enable --domain cdn.example.com
racore-cli domain disable --domain cdn.example.com
racore-cli domain delete --domain cdn.example.com    # 必须先停用
```

### 域名配置（get/set）

每个配置项都有 `get`（查询）和 `set`（设置）两个子命令：

```bash
# HTTP/2
racore-cli domain http2 get --domain example.com
racore-cli domain http2 set --domain example.com --config '{"enable":"on"}'
racore-cli domain http2 set --domain example.com --config '{"enable":"off"}'

# HTTP/3
racore-cli domain http3 get --domain example.com
racore-cli domain http3 set --domain example.com --config '{"enable":"on"}'

# 强制 HTTPS 跳转（需先开启 SSL）
racore-cli domain enforce-https get --domain example.com
racore-cli domain enforce-https set --domain example.com --config '{"https_redirect":"on"}'

# 智能压缩
racore-cli domain compress get --domain example.com
racore-cli domain compress set --domain example.com --config '{"enable":"on"}'

# IPv6
racore-cli domain ipv6 get --domain example.com
racore-cli domain ipv6 set --domain example.com --config '{"enable":"1"}'   # 1=开 0=关

# 最低 TLS 版本
# 可选值: SSLv3, TLSv1, TLSv1_2016, TLSv1.1_2016, TLSv1.2_2018, TLSv1.2_2019, TLSv1.2_2021
racore-cli domain tls-version get --domain example.com
racore-cli domain tls-version set --domain example.com --config '{"min_tls_version":"TLSv1.2_2021"}'

# 回源协议 (match-viewer, http-only, https-only)
racore-cli domain origin-protocol get --domain example.com
racore-cli domain origin-protocol set --domain example.com --config '{"origin_protocol_policy":"https-only","origin_protocol_http_port":"80","origin_protocol_https_port":"443"}'

# 回源 Host (type: 1=源站域名, 2=加速域名, 3=自定义)
racore-cli domain origin-host get --domain example.com
racore-cli domain origin-host set --domain example.com --config '{"origin_host_type":"3","origin_host":"custom.example.com"}'

# SSL 证书绑定
racore-cli domain ssl get --domain example.com
racore-cli domain ssl set --domain example.com --config '{"is_ssl":"1","cert_id":"123"}'

# 回源配置
racore-cli domain source get --domain example.com
racore-cli domain source set --domain example.com --config '{"source_type":"2","source_conf":[{"source":"origin.example.com","type":"1"}]}'

# IP 黑白名单 (state: on/off, type: white/black)
racore-cli domain ip-filter get --domain example.com
racore-cli domain ip-filter set --domain example.com --config '{"state":"on","type":"white","value":["1.2.3.4","5.6.7.8"]}'
racore-cli domain ip-filter set --domain example.com --config '{"state":"off","value":[]}'

# Referer 防盗链 (state: on/off, type: white/black/off)
racore-cli domain referer-filter get --domain example.com
racore-cli domain referer-filter set --domain example.com --config '{"state":"on","type":"white","values":["example.com","*.example.com"],"allow_empty":"on"}'
racore-cli domain referer-filter set --domain example.com --config '{"state":"off","type":"off","values":["example.com"],"allow_empty":"on"}'

# UA 黑白名单
racore-cli domain ua-filter get --domain example.com
racore-cli domain ua-filter set --domain example.com --config '{"state":"on","type":"black","values":["curl/*","wget/*"]}'

# 缓存策略 (AWS)
racore-cli domain cache-policy get --domain example.com
racore-cli domain cache-policy set --domain example.com --config '{"cache_policy_id":"658327ea-f89d-4fab-a63d-7e88639e58f6"}'

# 回源超时 (AWS)
racore-cli domain origin-timeout get --domain example.com
racore-cli domain origin-timeout set --domain example.com --config '{"connection_timeout":"10","response_timeout":"30"}'

# 地域访问控制 (AWS)
racore-cli domain geo-restriction get --domain example.com
racore-cli domain geo-restriction set --domain example.com --config '{"restriction_type":"whitelist","items":["CN","US","JP"]}'
```

### 缓存管理

```bash
# 刷新缓存
racore-cli cache purge --urls "https://example.com/path1,https://example.com/path2"
racore-cli cache purge-status --task-id abc123

# 预热内容
racore-cli cache prefetch --urls "https://example.com/bigfile.zip"
racore-cli cache prefetch-status --task-id abc123

# 预热区域和节点
racore-cli cache prewarm-regions --url "https://example.com/"
racore-cli cache prewarm-pop --region us-east-1

# 策略列表 (type: managed 或 custom)
racore-cli cache list-policies --domain example.com
racore-cli cache list-policies --domain example.com --type custom
racore-cli cache list-origin-request-policies --domain example.com
racore-cli cache list-response-header-policies --domain example.com
```

### 数据统计

```bash
# 流量/请求统计（支持 --domains 和 --interval）
racore-cli stats flow --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com
racore-cli stats request --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com
racore-cli stats hit-flow --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com
racore-cli stats hit-request --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com

# HTTP 状态码
racore-cli stats http-code --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com
racore-cli stats http-code-detail --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com

# 地区分布
racore-cli stats district --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com

# TOP 排行（支持 --scope: today, yesterday, week, month, last_month）
racore-cli stats top-domain --start-time "2026-07-01" --end-time "2026-07-14"
racore-cli stats top-url --scope yesterday
racore-cli stats top-referer --scope month --sorted referer_count
racore-cli stats top-ua --scope yesterday

# ISO 国家代码
racore-cli stats iso-country
```

### 证书管理

```bash
racore-cli cert list
racore-cli cert upload --name "My Cert" --cert-file ./cert.pem --key-file ./key.pem
racore-cli cert update --id 123 --cert-file ./new-cert.pem --key-file ./new-key.pem
racore-cli cert apply-aws --domain example.com
racore-cli cert validation-info --id 123
```

### 工单管理

```bash
racore-cli workorder list
racore-cli workorder types
racore-cli workorder create --title "问题描述" --description "详细说明" --type "技术支持"
racore-cli workorder log --id 123
racore-cli workorder send-message --id 123 --message "补充信息"
racore-cli workorder close --id 123
racore-cli workorder cancel --id 123
racore-cli workorder reopen --id 123
racore-cli workorder delete --id 123
```

### 日志下载

```bash
racore-cli log list --domain example.com
racore-cli log list --domain example.com --start-date 2026-07-01 --end-date 2026-07-14
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

### 认证配置

racore-cli 支持三种认证方式（按优先级排序），且配置一次即可永久生效：

> **自动持久化**：通过 MCP 配置传入的凭证会在首次启动时自动保存到系统安全存储中（macOS Keychain / Windows Credential Manager / Linux Keyring）。之后在终端直接使用 `racore-cli` 命令也无需再登录。

**方式 1：环境变量（推荐）**

```json
{
  "mcpServers": {
    "racore": {
      "command": "racore-cli",
      "args": ["serve"],
      "env": {
        "RACORE_ACCESS_KEY": "your_access_key",
        "RACORE_SECRET_KEY": "your_secret_key"
      }
    }
  }
}
```

**方式 2：命令行参数**

```json
{
  "mcpServers": {
    "racore": {
      "command": "racore-cli",
      "args": ["serve", "--access-key", "your_access_key", "--secret-key", "your_secret_key"]
    }
  }
}
```

**方式 3：手动登录（不通过 MCP）**

```bash
racore-cli login --access-key YOUR_KEY --secret-key YOUR_SECRET
```

### 各平台 MCP 配置文件位置

| AI Agent | 配置文件路径 |
|----------|------------|
| Kiro | `.kiro/settings/mcp.json`（项目级）或 `~/.kiro/settings/mcp.json`（用户级） |
| AWS Q Developer | `~/.aws/amazonq/mcp.json` |
| Claude Desktop (macOS) | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Claude Desktop (Windows) | `%APPDATA%\Claude\claude_desktop_config.json` |
| VS Code (Copilot/Continue) | `.vscode/mcp.json`（项目级） |
| Cursor | `~/.cursor/mcp.json` |

### 配置示例

以下示例使用环境变量方式（推荐），也可替换为 args 方式。

**Kiro** — `.kiro/settings/mcp.json`：
```json
{
  "mcpServers": {
    "racore": {
      "command": "racore-cli",
      "args": ["serve"],
      "env": {
        "RACORE_ACCESS_KEY": "your_access_key",
        "RACORE_SECRET_KEY": "your_secret_key"
      }
    }
  }
}
```

**AWS Q Developer** — `~/.aws/amazonq/mcp.json`：
```json
{
  "mcpServers": {
    "racore": {
      "command": "racore-cli",
      "args": ["serve"],
      "env": {
        "RACORE_ACCESS_KEY": "your_access_key",
        "RACORE_SECRET_KEY": "your_secret_key"
      }
    }
  }
}
```

**Claude Desktop** — macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`，Windows: `%APPDATA%\Claude\claude_desktop_config.json`：
```json
{
  "mcpServers": {
    "racore": {
      "command": "racore-cli",
      "args": ["serve"],
      "env": {
        "RACORE_ACCESS_KEY": "your_access_key",
        "RACORE_SECRET_KEY": "your_secret_key"
      }
    }
  }
}
```

**VS Code / Cursor** — `.vscode/mcp.json`：
```json
{
  "servers": {
    "racore": {
      "type": "stdio",
      "command": "racore-cli",
      "args": ["serve"],
      "env": {
        "RACORE_ACCESS_KEY": "your_access_key",
        "RACORE_SECRET_KEY": "your_secret_key"
      }
    }
  }
}
```

**Windows 使用 args 方式**（如果 env 方式不生效）：
```json
{
  "mcpServers": {
    "racore": {
      "command": "racore-cli",
      "args": ["serve", "--access-key", "your_access_key", "--secret-key", "your_secret_key"]
    }
  }
}
```

> **提示**：Windows 上如果 racore-cli 不在 PATH 中，`command` 需要写完整路径，如 `"C:\\Program Files\\Racore CLI\\racore-cli.exe"`

### 手动测试 MCP 模式

```bash
# 列出所有可用工具
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | racore-cli serve

# 调用工具
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"domain-list","arguments":{}}}' | racore-cli serve
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
| domain-list | 列出所有 CDN 域名（支持模糊搜索） |
| domain-create | 创建新域名（自动匹配泛域证书） |
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
│       ├── store.go           # 凭证安全存储（系统钥匙串 + 文件降级）
│       └── resolve.go         # 凭证解析（env > flag > prompt）
└── test_all_apis.sh           # 集成测试脚本
```

## 安全说明

- 凭证优先存储在操作系统安全凭证管理器中：
  - **macOS**: Keychain（钥匙串）
  - **Windows**: Credential Manager（凭据管理器）
  - **Linux**: Secret Service（GNOME Keyring / KDE Wallet）
- 无桌面环境时自动降级到文件存储（`~/.racore/credentials`，权限 0600）
- Secret Key 在内存中使用后会被清零
- Token 自动刷新，无需手动管理
- 支持环境变量传入凭证（`RACORE_ACCESS_KEY` / `RACORE_SECRET_KEY`），避免持久化存储
- `whoami` 命令可查看当前使用的存储后端

## 开发

```bash
make build    # 编译
make test     # 运行测试
make lint     # 代码检查
make clean    # 清理
```

## 许可证

MIT License
