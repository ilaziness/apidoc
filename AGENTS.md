# AGENTS.md

Instructions for AI coding agents working on **ApiDoc**.  
Format: [AGENTS.md](https://agents.md/) open format (Markdown; no required schema).

## Project overview

ApiDoc is a **local desktop** API doc manager + HTTP debugger (Postman-like subset).

- Module path: **`github.com/ilaziness/apidoc`** (must match GitHub remote `ilaziness/apidoc`).
- **No** login, teams, sync, or collaboration server in v1.
- Top-level entity (UI: **项目 / Project**): code & SQLite still use **collection**. No workspace/team layer in v1.
- UI: [go-gui](https://github.com/go-gui-org/go-gui) (no Electron / browser shell).
- Data: SQLite via [Bun](https://bun.uptrace.dev); driver **must** be pure Go (`modernc.org/sqlite`). **Never** use `mattn/go-sqlite3`.
- Schema changes: Bun migrations under `internal/store` (no `CREATE TABLE IF NOT EXISTS` as the migration strategy).
- Protocols: implement **HTTP** only in `internal/protocol/httpapi`; reserve `grpc` / `tcp` / `udp` / `websocket` on nodes.
- License: MIT.

Canonical product scope: **`docs/PRD.md`**. If code and PRD disagree, **fix the docs first**, then implement.

## Documentation sync (mandatory)

**Feature changes MUST stay in sync with project documentation.**

1. **Before** implementing a new capability, behavior change, or scope cut: update `docs/PRD.md` (and any other `docs/*` that describe that feature). Write confirmed behavior into confirmed sections and add a revision-log line.
2. **In the same change set** as the code (same PR / same agent turn): keep docs and code aligned. Do not ship code-only feature diffs.
3. If the user asks for a feature **not** in the PRD: ask or update the PRD first; do not invent scope.
4. Touch at least: `docs/PRD.md` for product behavior; README / AGENTS.md when build, layout, or agent rules change.
5. Prefer small doc edits that state the new rule clearly over leaving “TBD” after the feature already exists in code.

Agents that change behavior without updating docs are doing incomplete work.

## Source of truth

| Doc | Use for |
| --- | --- |
| `docs/PRD.md` | Product scope, UX rules, non-goals, tech constraints |
| `README.md` | Human quick start |
| `AGENTS.md` | Agent workflow, commands, conventions |
| `LICENSE` | MIT |

## Repository layout

```text
cmd/apidoc/                 # main: config → store → migrate → ui
internal/
  ui/                       # go-gui only; files by window region (run/app/layout/home/sidebar/editor/http/panel/dialog/colors)
  service/                  # use cases (was "app"; avoid vague name)
  protocol/                 # Kind constants (http/grpc/…)
  protocol/httpapi/         # HTTP send (package httpapi, NOT http)
  store/                    # Bun + SQLite + Migrate()
  store/migrations/         # migration scripts (embed later)
  openapi/                  # import only (2.0 / 3.0 / 3.1)
  config/                   # OS app-config dir (not inside data dir)
  i18n/                     # en / zh-Hans / zh-Hant UI strings
docs/                       # PRD and design docs
```

Do not put business SQL in `ui`. Do not put send logic in view functions—use `httpapi` (or future protocol packages).

Do **not** name a package `http` under this module; it collides with `net/http` in the same files.

## Setup and build

`go.mod` uses:

```text
go 1.27
toolchain go1.27.3
```

Track the latest stable Go; do not lower the language version below go-gui’s **1.26+** floor.

```bash
go build -o apidoc ./cmd/apidoc
go test ./...
go vet ./...
golangci-lint run   # requires golangci-lint v2; config: .golangci.yml
```

Editor defaults: `.editorconfig` (Go uses tabs to match `gofmt`).

On macOS, go-gui’s Metal backend needs a C toolchain / CGO. Windows and Linux UI builds can use `CGO_ENABLED=0` for go-gui, but **SQLite must still be modernc**, not mattn—even when CGO is on.

Do not add `github.com/mattn/go-sqlite3`. Open SQLite with `modernc.org/sqlite` (or an equivalent that cannot fall back to mattn). Avoid Bun `sqliteshim` if it can select mattn under CGO.

## Code style

- Standard Go: `gofmt`, meaningful package docs, small interfaces at use site.
- Errors: wrap with `%w` where useful; prefer `errors.New` / `fmt.Errorf` over custom string error types.
- Comments: explain non-obvious constraints (especially PRD rules), not obvious code.
- Prefer explicit names over abbreviations in exported APIs.
- Typed constants for config enums (`config.LangEN` / `LangZHHans` / `LangZHHant`, `ThemeSystem`, `ProxyOff`, …).
- UI i18n: only `en` / `zh-Hans` / `zh-Hant`; first launch detects OS locale and **persists** one of these (no lasting “follow OS” setting).

## Testing

- Prefer unit tests next to packages (`*_test.go`).
- UI automation / go-gui headless tests are **out of v1 scope** unless the PRD is updated.
- After substantive changes, run `go test ./...` and `go vet ./...` before finishing.

## Security and data

- Secrets in environments are **plaintext in SQLite** in v1 (PRD). Do not silently add encryption.
- Never commit user DB files, `.env`, or real credentials.
- Default data dir: `{HOME}/.apidoc`. App config lives in OS conventional dirs (see PRD §6.8).

## Git / PR habits

- Do not create commits or PRs unless the user asks.
- Do not invent features outside `docs/PRD.md`.
- When the user asks to commit, keep messages focused on *why*; include doc updates in the same commit when behavior changed.

## Non-goals (v1) — do not implement unless PRD changes

Login, teams/workspace UI, sync server, OAuth2, request history, OpenAPI export, Postman import/export, SOCKS5, Cookie jar, client certs, HTTP/2 streaming, search, duplicate-node, Mock, codegen, go-edit highlighting, WASM UI.

## When stuck

1. Read `docs/PRD.md` for the decision.
2. If still unspecified: **ask the user**; do not invent.
3. Record the answer in the PRD, then code.
