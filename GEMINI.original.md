# TENDR — AI Gateway Binary

TENDR is a self-hosted, single-binary AI gateway written in Go. It proxies requests to multiple AI providers (OpenAI, Anthropic, Gemini, Groq) with local cost tracking, caching, and a TUI dashboard.

## Project Context

- **Active Goal**: Stage 2 of `docs/PLAN.md` (Multi-Provider + Normalization).
- **Architecture**: Modular Monolith as defined in `docs/ARCHITECTURE.md`.
- **Interface**: CLI and TUI (Bubble Tea).
- **Storage**: SQLite (`modernc.org/sqlite` — Pure Go) + `bbolt` for disk cache.

### Tech Stack
- **Language**: Go 1.22+ (CGO_ENABLED=0 REQUIRED)
- **HTTP Router**: [chi v5](https://github.com/go-chi/chi)
- **TUI**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Logging**: `zerolog` (JSONL with rotation via lumberjack)
- **Config**: `yaml.v3` + `viper`

## Development Workflow

### Key Commands
- `go run ./cmd/tendr init` — Generate default config
- `go run ./cmd/tendr start` — Start HTTP gateway
- `make build` — Build local binary
- `make test` — Run unit and integration tests

### Coding Conventions
- **Error Handling**: Wrap all errors with context: `fmt.Errorf("package: action: %w", err)`.
- **I/O**: Every I/O function MUST accept `context.Context` as the first parameter.
- **Safety**: NEVER use `panic()`, `log.Fatal()`, or `os.Exit()` outside of `main.go`.
- **Typed Errors**: Use provider-specific errors (`ErrRateLimit`, `ErrTimeout`) from `internal/provider`.
- **Logging**: Use structured `zerolog` entries per `AGENTS.md` schema.
- **Database**: Raw SQL only (No ORM). Migrations in `internal/store/migrations/`.
- **CGO**: PROHIBITED. All dependencies must be pure Go.

## Agent Guidelines
- Follow `docs/PLAN.md` strictly. Do not skip stages.
- Adhere to the `AI Agent Rules` in `AGENTS.md`.
- Use `graphify` for codebase navigation and relationship mapping.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review.
- After modifying code, run `graphify update .` to keep the graph current.
