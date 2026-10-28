# ApiDoc

本机桌面端 API 文档管理与 HTTP 调试工具（Postman 能力子集）。

- **许可证：** MIT
- **产品需求：** [docs/PRD.md](docs/PRD.md)
- **Agent 约定：** [AGENTS.md](AGENTS.md)
- **模块路径：** `github.com/ilaziness/apidoc`（与 Git remote 一致）

## 环境要求

- Go **1.27**（`go.mod`；推荐工具链 **go1.27.3**）。go-gui 语言下限为 1.26+。
- Windows / macOS / Linux（amd64、arm64）

## 快速开始

```bash
go build -o apidoc ./cmd/apidoc
./apidoc
```

当前会加载默认配置、打开数据目录下的库路径并跑迁移占位，随后进入 UI 占位（go-gui / Bun 尚未接入）。

## 目录结构

| 路径 | 职责 |
| --- | --- |
| `cmd/apidoc` | 进程入口与启动编排 |
| `internal/ui` | go-gui 界面（禁止直接 SQL / net/http 发送） |
| `internal/service` | 用例层（文档集、树、环境、导入、发送） |
| `internal/protocol` | 协议枚举与注册表 |
| `internal/protocol/httpapi` | HTTP 发送实现（避免包名 `http` 与标准库冲突） |
| `internal/store` | Bun + SQLite；迁移逻辑同包，`migrations/` 放脚本 |
| `internal/openapi` | 仅 OpenAPI 导入 |
| `internal/config` | 应用配置（在 OS 配置目录，不在数据目录内） |
| `docs/` | 产品文档（功能变更必须同步） |

## 数据位置

- **业务数据（默认）：** `{HOME}/.apidoc`
- **应用配置：** Windows `%APPDATA%\apidoc`，macOS `~/Library/Application Support/apidoc`，Linux `~/.config/apidoc`
