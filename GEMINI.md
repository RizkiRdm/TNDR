# AI Agent Operational Contract

Document Version: 1.0.0
Project: TENDR — AI Gateway Binary
Purpose: Single source of truth for AI agents.

---

## Project Context

TENDR: self-hosted single-binary AI gateway in Go.
Run local, proxy multiple AI providers.
NOT SaaS. NOT web app. Terminal tool.

Interface: CLI + TUI (Bubble Tea)
Config: YAML
Storage: SQLite (local, `~/.tendr/tendr.db`)
Logging: JSONL, rotation.

---

## Tech Stack

| Layer | Technology | Version |
|---|---|---|
| Language | Go | 1.22+ |
| TUI | Bubble Tea | latest |
| TUI Styling | Lip Gloss | latest |
| HTTP Router | chi | v5 |
| Config | viper | latest |
| SQLite | modernc/sqlite | latest (NO CGO) |
| Disk Cache | bbolt | latest |
| Logger | zerolog | latest |
| Log Rotation | lumberjack | v2 |
| Build | goreleaser | latest |

---

## Architecture Overview

```
cmd/tendr/main.go   → CLI flags, launch GW/TUI
internal/gateway    → HTTP server, request lifecycle
internal/router     → provider selection, fallback, modes
internal/provider/* → per-provider adapters
internal/cache      → exact + semantic cache, hit/miss
internal/cost       → token count, pricing, cost record
internal/ratelimit  → token bucket limiter
internal/store      → SQLite access, migrations, queries
internal/logger     → zerolog init, entry construction
internal/config     → YAML parse, validation, pricing fetch
internal/tui        → TUI models, views, handlers
```

Full architecture: `docs/ARCHITECTURE.md`.

---

## Project Structure

```
tendr/
├── cmd/tendr/main.go
├── internal/
│   ├── config/
│   ├── gateway/
│   ├── router/
│   ├── provider/
│   │   ├── openai/
│   │   ├── anthropic/
│   │   ├── gemini/
│   │   └── groq/
│   ├── cache/
│   ├── cost/
│   ├── ratelimit/
│   ├── store/
│   │   └── migrations/
│   ├── logger/
│   └── tui/
│       ├── tabs/
│       └── styles/
├── pricing.json
├── config.example.yaml
├── .goreleaser.yaml
├── Makefile
├── go.mod
└── docs/
```

---

## Key Commands

```bash
make build
go run ./cmd/tendr start
make test
make lint
make release
go run ./cmd/tendr init
go run ./cmd/tendr cost
go run ./cmd/tendr cache clear
```

---

## Coding Conventions

### General

- Exported fns MUST have godoc.
- Wrap errors: `fmt.Errorf("component: action: %w", err)`.
- NO `panic()`, `log.Fatal()`, `os.Exit()` outside `main.go`.
- `context.Context` first param on all I/O fns.
- Respect cancellation in provider calls.

### Error Handling

```go
// CORRECT
result, err := someOperation(ctx)
if err != nil {
    return fmt.Errorf("gateway: handle request: %w", err)
}

// WRONG
result, err := someOperation(ctx)
if err != nil {
    log.Fatal(err)
}
```

### Typed Errors

Use `internal/provider/provider.go` errors: `ErrRateLimit`, `ErrTimeout`, `ErrProviderDown`, `ErrInvalidKey`. Check with `errors.Is()`.

### Interfaces

Define subsystem as interface before impl.

### CGO

PROHIBITED. Deps MUST be pure Go.
Verify: `CGO_ENABLED=0 go build ./...`

---

## API Conventions

### Endpoint

`POST /v1/chat/completions`. OpenAI-compatible request only. Model field maps to TENDR alias.

### Request Headers

`Content-Type: application/json`. `X-Request-ID: <uuid>` (auto-injected).

### Success Response

Append `x-tendr` field to every response.

### Error Response

Error codes: `rate_limit_exceeded`, `provider_unavailable`, `invalid_model_alias`, `bad_request`.

---

## Database Conventions

- Raw SQL only — NO ORM.
- Queries in `internal/store/`.
- Migrations in `internal/store/migrations/` as `.sql`.
- Auto-run migrations at startup.
- NO modify existing migrations.
- Table/Column: `snake_case`.
- PK: TEXT uuid. Timestamps: RFC3339. Bool: INTEGER (0/1). Money: REAL (USD).

---

## Logging Conventions

Schema: `ts`, `level`, `component`, `event`, `request_id`, `user_id`, `metadata`.

### Rules

- `component` match package name.
- `event` is `snake_case`.
- `user_id` is `"anonymous"` (MVP).
- NEVER log API keys or raw bodies.
- NO `fmt.Println` — use zerolog.

---

## Config Conventions

- Loc: `~/.tendr/config.yaml`.
- Validate on load — fail fast.
- Design for hot-reload.
- Pass as dep, NO global vars.
- Keys from config only.

---

## Security Rules

- NO log API keys. Mask in TUI: `prefix••••••••`.
- NO raw provider errors in client response.
- Validate pricing URLs.
- NO raw bodies in SQLite.
- Request IDs: uuid v4.

---

## Testing Rules

- Providers MUST have unit tests + mock server.
- Cost/Router/Config: table-driven tests.
- Integration: real SQLite.
- NO TUI unit tests — manual verify.
- Files: `*_test.go`.

---

## Git Conventions

### Branch Naming

`feat/stage-1-foundation`, `fix/openai-streaming`, etc.

### Commit Messages

Terse, imperative: `feat(provider): add anthropic adapter`.

### PR Rules

- One stage = one PR.
- Pass lint + tests before merge.
- Update `PLAN.md` checklist.

---

## AI Agent Rules

### MUST

- Read `docs/ARCHITECTURE.md`, `docs/PLAN.md` before work.
- Follow package boundaries.
- Use Provider interface.
- Structured log every event.
- Typed errors for failures.
- Pure Go only.

### MUST NOT

- Add deps without instruction.
- Implement features outside current stage.
- Modify existing migrations.
- Add providers beyond MVP list.
- Use `SELECT *`.
- Use global vars for config/logger.
- Implement SaaS/auth/multi-user.

### When Stuck

- Check `docs/ARCHITECTURE.md`, `docs/DESIGN.md`.
- NO new patterns or abstractions.
- Ask for clarification.

---

## Context Hints

- Binary: `tendr`. Port: `4821`.
- Config: `~/.tendr/config.yaml`. DB: `~/.tendr/tendr.db`.
- Pricing embedded: `pricing.json`.
- `user_id` always `"anonymous"`.
- TUI: Bubble Tea — Msg/Cmd pattern.

---

## Out of Scope

NO Auth, Multi-user, Web UI, Billing, Prometheus, OpenTelemetry, Plugin, Agent builder, Docker Compose.
