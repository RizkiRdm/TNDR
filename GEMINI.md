# TENDR — AI Gateway Binary

TENDR is a self-hosted, single-binary AI gateway written in Go. It proxies requests to multiple AI providers (OpenAI, Anthropic, Gemini, Groq) with local cost tracking, caching, and a TUI dashboard.

## Project Context

- **Active Goal**: Stage 1 of `docs/PLAN.md` (Foundation: Go init, config, gateway start).
- **Architecture**: Layered monolith with package-private implementations and interface-based boundaries.
- **Interface**: CLI and TUI (Bubble Tea).
- **Storage**: SQLite (pure Go, NO CGO) + bbolt for disk cache.

### Tech Stack
- **Language**: Go 1.22+ (CGO_ENABLED=0 required)
- **HTTP Router**: [chi v5](https://github.com/go-chi/chi)
- **TUI**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Storage**: `modernc.org/sqlite` (Pure Go)
- **Caching**: `bbolt`
- **Logging**: `zerolog` (JSONL with rotation via lumberjack)
- **Config**: `yaml.v3` + `viper`

## Development Workflow

### Key Commands
(Note: Project is in early stage; commands below are targets for Stage 1/8)
- `go run ./cmd/tendr init` — Generate default config
- `go run ./cmd/tendr start` — Start HTTP gateway
- `make build` — Build local binary
- `make test` — Run unit and integration tests

### Coding Conventions
- **Error Handling**: Wrap all errors with context: `fmt.Errorf("package: action: %w", err)`.
- **I/O**: Every I/O function MUST accept `context.Context` as the first parameter.
- **Safety**: NEVER use `panic()`, `log.Fatal()`, or `os.Exit()` outside of `main.go`.
- **Typed Errors**: Use provider-specific errors (`ErrRateLimit`, `ErrTimeout`) from `internal/provider`.
- **Logging**: Use structured `zerolog` entries. Metadata MUST be in a nested `metadata` dictionary.
- **Database**: Raw SQL only (No ORM). Migrations in `internal/store/migrations/`.

## Critical Discrepancies
- **Docs Mismatch**: `docs/PRD.md` and `docs/ARCHITECTURE.md` currently describe an unrelated "buruh" task queue project. **Ignore these for TENDR implementation** or update them to match the specifications in `AGENTS.md` and `docs/PLAN.md`.
- **CGO**: Prohibited. All dependencies must be pure Go.

## Agent Guidelines
- Follow `docs/PLAN.md` strictly. Do not skip stages.
- Adhere to the `AI Agent Rules` in `AGENTS.md`.
- Respect `caveman` mode preferences when active (terse technical communication).


## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
