# AI Agent Operational Contract

Document Version: 1.0.0
Project: TENDR — AI Gateway Binary
Purpose: Single source of truth for AI coding agents (Claude Code, Gemini CLI, Codex, etc.)

---

## Project Context

TENDR is a self-hosted, single-binary AI gateway written in Go.
It runs locally on a developer's machine and proxies requests to multiple AI providers.
It is NOT a SaaS. It is NOT a web application. It is a terminal tool.

Primary interface: CLI + TUI (Bubble Tea)
Config: YAML
Storage: SQLite (local, `~/.tendr/tendr.db`)
Logging: JSONL with log rotation

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
cmd/tendr/main.go
  → parses CLI flags
  → launches HTTP gateway (tendr start)
  → launches TUI (tendr / tendr monitor)

internal/gateway    → HTTP server, request lifecycle
internal/router     → provider selection, fallback logic
internal/provider/* → per-provider adapters (openai, anthropic, gemini, groq)
internal/cache      → exact + semantic cache
internal/cost       → cost calculation, pricing resolution
internal/ratelimit  → token bucket rate limiter
internal/store      → SQLite access layer
internal/logger     → zerolog JSONL logger
internal/config     → YAML config parsing + validation
internal/tui        → Bubble Tea dashboard
```

Full architecture in `docs/ARCHITECTURE.md`.

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
# Build
make build

# Run in development
go run ./cmd/tendr start

# Run tests
make test

# Lint
make lint

# Release build (all platforms)
make release

# Generate default config
go run ./cmd/tendr init

# View cost summary
go run ./cmd/tendr cost

# Clear cache
go run ./cmd/tendr cache clear
```

---

## Coding Conventions

### General

- ALL exported functions MUST have godoc comments
- ALL errors MUST be wrapped with context: `fmt.Errorf("component: action: %w", err)`
- NEVER use `panic()` outside of `main.go` initialization
- NEVER use `log.Fatal()` outside of `main.go`
- NEVER use `os.Exit()` outside of `main.go`
- Use `context.Context` as first parameter on all I/O functions
- Respect context cancellation in all provider calls

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

Provider errors MUST use typed errors defined in `internal/provider/provider.go`:

```go
var (
    ErrRateLimit    = errors.New("rate_limit")
    ErrTimeout      = errors.New("timeout")
    ErrProviderDown = errors.New("provider_down")
    ErrInvalidKey   = errors.New("invalid_key")
)
```

Check with `errors.Is()`, never string comparison.

### Interfaces

Every subsystem MUST be defined as an interface before implementation:

```go
// internal/provider/provider.go
type Provider interface {
    Name() string
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    EstimateTokens(req *CompletionRequest) int
}
```

### CGO

PROHIBITED. Every dependency MUST be pure Go.
Verify with: `CGO_ENABLED=0 go build ./...`

---

## API Conventions

### Endpoint

```
POST /v1/chat/completions
```

OpenAI-compatible request schema ONLY.
Model field maps to TENDR model alias (configured in YAML).

### Request Headers

```
Content-Type: application/json
X-Request-ID: <uuid>   (injected by gateway if not present)
```

### Success Response

```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "gpt-4o",
  "choices": [...],
  "usage": {
    "prompt_tokens": 150,
    "completion_tokens": 80,
    "total_tokens": 230
  },
  "x-tendr": {
    "provider": "openai",
    "fallback_used": false,
    "cache_hit": false,
    "cost_usd": 0.000425,
    "latency_ms": 142
  }
}
```

`x-tendr` field MUST be appended to every response.

### Error Response

```json
{
  "error": {
    "code": "provider_unavailable",
    "message": "All providers failed. Last error: timeout after 30s",
    "request_id": "req_abc123"
  }
}
```

Error codes: `rate_limit_exceeded`, `provider_unavailable`, `invalid_model_alias`, `bad_request`

---

## Database Conventions

- ALL SQL is raw — NO ORM
- ALL queries in `internal/store/` package only
- ALL migrations in `internal/store/migrations/` as numbered `.sql` files
- ALL migrations run at startup automatically
- NEVER modify existing migration files — add new ones only
- Table names: `snake_case`, plural
- Column names: `snake_case`
- Primary keys: TEXT uuid (use `github.com/google/uuid`)
- Timestamps: TEXT in RFC3339 format
- Booleans: INTEGER (0/1)
- Money: REAL (USD)

### Query Pattern

```go
// CORRECT — explicit column list
row := db.QueryRowContext(ctx,
    `SELECT id, ts, provider, cost_usd FROM requests WHERE id = ?`, id)

// WRONG — never SELECT *
row := db.QueryRowContext(ctx, `SELECT * FROM requests WHERE id = ?`, id)
```

---

## Logging Conventions

ALL log entries MUST follow this schema:

```json
{
  "ts": "2026-05-30T11:00:00Z",
  "level": "INFO",
  "component": "gateway",
  "event": "request_received",
  "request_id": "req_abc123",
  "user_id": "anonymous",
  "metadata": {}
}
```

### Rules

- `component` MUST match package name: `gateway`, `router`, `provider`, `cache`, `cost`, `ratelimit`, `config`
- `event` MUST be `snake_case`
- `user_id` MUST be `"anonymous"` in MVP (no auth)
- `metadata` MUST contain all event-specific fields
- NEVER log raw API keys, even partially
- NEVER log raw request/response bodies
- NEVER use `fmt.Println` for logging — use zerolog only

### Using the Logger

```go
// CORRECT
logger.Info().
    Str("component", "gateway").
    Str("event", "request_received").
    Str("request_id", reqID).
    Str("user_id", "anonymous").
    Dict("metadata", zerolog.Dict().
        Str("model_alias", alias).
        Str("provider", provider)).
    Msg("")

// WRONG
log.Printf("request received: %s", reqID)
```

---

## Config Conventions

- Config file location: `~/.tendr/config.yaml`
- Config MUST be validated on load — fail fast with line reference on error
- Config MUST be hot-reloadable in future (design for it: no global config vars)
- Pass config as dependency, never read from global state
- API keys MUST be read from config only — no env var override in MVP

---

## Security Rules

- NEVER log API keys in any form
- NEVER expose API keys in TUI (mask to: `prefix••••••••`)
- NEVER include raw provider error messages in client responses
- NEVER allow config to specify arbitrary URLs for pricing fetch (validate against allowlist)
- NEVER store request/response bodies in SQLite
- Request IDs MUST be uuid v4, generated server-side if not provided

---

## Testing Rules

- ALL `internal/provider/*` adapters MUST have unit tests with mock HTTP server
- ALL `internal/cost` calculations MUST have table-driven unit tests
- ALL `internal/router` fallback modes MUST have unit tests
- ALL `internal/config` validation MUST have unit tests covering invalid configs
- Integration tests MUST use real SQLite (not mocked)
- NEVER test TUI rendering with unit tests — use manual verification
- Test files: `*_test.go` in same package
- Table-driven tests preferred for all calculation logic

```go
// preferred test pattern
func TestCostCalculation(t *testing.T) {
    tests := []struct {
        name         string
        inputTokens  int
        outputTokens int
        pricing      Pricing
        wantCost     float64
    }{
        {"gpt-4o basic", 1000, 500, gpt4oPricing, 0.0075},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Calculate(tt.inputTokens, tt.outputTokens, tt.pricing)
            if got != tt.wantCost {
                t.Errorf("got %f, want %f", got, tt.wantCost)
            }
        })
    }
}
```

---

## Git Conventions

### Branch Naming

```
feat/stage-1-foundation
feat/stage-2-providers
fix/openai-streaming
chore/goreleaser-config
```

### Commit Messages

```
feat(provider): add anthropic adapter
fix(router): handle nil response from groq
chore(config): add validation for timeout range
docs(readme): add quickstart instructions
test(cost): add table tests for pricing calculation
```

### PR Rules

- One stage = one PR (or series of small PRs within stage)
- MUST pass lint and tests before merge
- MUST update PLAN.md checklist on merge

---

## AI Agent Rules

These rules are for AI coding agents working on this codebase:

### MUST

- Read `docs/ARCHITECTURE.md` before implementing any component
- Read `docs/PLAN.md` before starting any work — check current stage
- Follow package boundaries in ARCHITECTURE.md — no cross-layer imports
- Use the Provider interface for all provider implementations
- Emit structured log entries for every meaningful event
- Use typed errors for all provider failures
- Write pure Go only — no CGO

### MUST NOT

- Add new dependencies without explicit instruction
- Implement features not in current stage's checklist
- Modify migration files that already exist
- Add new providers beyond OpenAI, Anthropic, Gemini, Groq without instruction
- Use `SELECT *` in any SQL query
- Use global variables for config or logger
- Implement any SaaS or web UI features
- Add authentication or multi-user features (not in MVP)

### When Stuck

- Check `docs/ARCHITECTURE.md` for component contracts
- Check `docs/DESIGN.md` for TUI specifications
- Do NOT invent new patterns not described in docs
- Do NOT add abstraction layers not in ARCHITECTURE.md
- Ask for clarification rather than improvising

---

## Context Hints

- The binary is called `tendr`
- Default port: `4821`
- Default config: `~/.tendr/config.yaml`
- Default DB: `~/.tendr/tendr.db`
- Default log dir: `~/.tendr/logs/`
- Pricing file embedded in binary: `pricing.json` (go:embed)
- The `x-tendr` response field is custom — always append it
- `user_id` is always `"anonymous"` in MVP — no auth system
- TUI is built with Bubble Tea — all updates via Msg/Cmd pattern, never direct mutation
- All monetary values in USD as float64

---

## Known Issues / Constraints

- Semantic cache is optional and off by default — do not implement unless config explicitly enables it
- Streaming responses must be pass-through — do not buffer full streaming response for cost calculation; use provider's usage field in final chunk
- Gemini API is not OpenAI-compatible natively — adapter handles full translation
- Anthropic uses different token counting — adapter normalizes to prompt_tokens/completion_tokens

---

## Out of Scope (Do Not Implement)

- Authentication / API key management for users
- Multi-user support
- Web UI or browser dashboard
- Billing / payments
- Prometheus metrics endpoint
- OpenTelemetry export
- Plugin system
- Fine-tuning
- Image generation
- Agent builder
- Workflow automation
- Docker Compose setup
- Kubernetes configuration
- SaaS features of any kind

Respond terse like smart caveman. All technical substance stay. Only fluff die.

Rules:
- Drop: articles (a/an/the), filler (just/really/basically), pleasantries, hedging
- Fragments OK. Short synonyms. Technical terms exact. Code unchanged.
- Pattern: [thing] [action] [reason]. [next step].
- Not: "Sure! I'd be happy to help you with that."
- Yes: "Bug in auth middleware. Fix:"

Switch level: /caveman lite|full|ultra|wenyan
Stop: "stop caveman" or "normal mode"

Auto-Clarity: drop caveman for security warnings, irreversible actions, user confused. Resume after.

Boundaries: code/commits/PRs written normal.
