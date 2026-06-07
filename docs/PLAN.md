# PLAN.md — TENDR
# Implementation Plan

Document Version: 1.0.0
Project: TENDR
Last Updated: 2026-05-30

---

## HARD RULES

These rules exist because the developer has a documented pattern of scope creep and project abandonment. They are non-negotiable.

```
RULE 1: Each stage MUST be independently shippable.
RULE 2: Each stage MUST be completable in ≤ 1 week.
RULE 3: Do NOT start Stage N+1 before Stage N is pushed to GitHub.
RULE 4: New ideas go to PARKING_LOT section below — NOT into code.
RULE 5: No feature is valid unless it exists in this PLAN.md.
RULE 6: If a stage takes > 1 week, SCOPE MUST BE CUT — not extended.
RULE 7: Do NOT add providers beyond the MVP list until Stage 5 is complete.
RULE 8: Do NOT build TUI before gateway is functional (Stage 3 minimum).
```

---

## Daily Status Template

Update this at end of each session:

```
Date: 2026-06-04
Stage: Stage 2 — Multi-Provider + Normalization
Completed today: Started Stage 2, verified Stage 1 foundation.
Blocked on: None.
Next session starts at: Implementing Anthropic provider.
```

---

## Version A — MVP Execution Plan

---

### Stage 1 — Foundation
**Goal:** Repo exists, binary compiles, config loads, single provider works.
**Shippable:** `tendr start` proxies a request to OpenAI.
**Time limit:** 1 week

- [x] Initialize Go module (`go mod init github.com/RizkiRdm/TNDR`)
- [x] Set up project directory structure per ARCHITECTURE.md
- [x] Implement `internal/config` — YAML parsing + validation
- [x] Implement `internal/logger` — zerolog init, JSONL output, lumberjack rotation
- [x] Implement `internal/gateway` — HTTP server on configurable port
- [x] Implement `/v1/chat/completions` handler (pass-through only)
- [x] Implement `internal/provider/openai` — adapter, normalize request/response
- [x] Wire: gateway → openai provider
- [x] Implement `cmd/tendr` — `start` and `stop` commands
- [x] Implement `tendr init` — generate default config.yaml
- [x] Write `config.example.yaml`
- [x] Verify: `curl localhost:4821/v1/chat/completions` returns valid response
- [x] Push to GitHub with README: install + quickstart instructions

**Checkpoint:** Binary works. One provider works. Config loads. Logs emit JSONL.

---

### Stage 2 — Multi-Provider + Normalization
**Goal:** All 4 MVP providers work through the same endpoint.
**Shippable:** Users can switch provider via model alias in config.
**Time limit:** 1 week

- [x] Implement `internal/provider/anthropic` adapter
- [x] Implement `internal/provider/gemini` adapter
- [x] Implement `internal/provider/groq` adapter
- [x] Implement `internal/router` — basic provider selection by model alias
- [x] Define typed provider errors: `ErrRateLimit`, `ErrTimeout`, `ErrProviderDown`, `ErrInvalidKey`
- [x] Wire: gateway → router → provider selection
- [x] Test: each provider returns normalized OpenAI-compatible response
- [x] Verify streaming responses pass through correctly
- [x] Push to GitHub, update README with multi-provider config example

**Checkpoint:** 4 providers work. Model alias routes correctly. Streaming works.

---

### Stage 3 — Fallback Engine
**Goal:** Automatic provider failover works in all 3 modes.
**Shippable:** Configure fallback chain, gateway handles provider failures gracefully.
**Time limit:** 1 week

- [x] Implement `internal/router/fallback.go` — provider chain execution
- [x] Implement `reliable` mode — fallback on hard errors
- [x] Implement `fast` mode — fallback on latency threshold
- [x] Implement `smart` mode — combined conditions
- [x] Add per-provider timeout configuration
- [x] Log fallback events: `fallback_triggered`, `fallback_from`, `fallback_reason`
- [x] Return HTTP 502 with clear error if all providers exhausted
- [x] Test: kill primary provider → fallback triggers → secondary responds
- [x] Test: simulate slow provider → fast mode fallback triggers
- [x] Push to GitHub

**Checkpoint:** Fallback works. All 3 modes tested. Provider failures are handled gracefully.

---

### Stage 4 — Cost Tracking
**Goal:** Every request has an accurate cost record.
**Shippable:** `tendr cost` shows spending breakdown.
**Time limit:** 1 week

- [x] Implement `internal/store` — SQLite init, migrations, connection pool
- [x] Create migration `001_init.sql` — requests + cache_entries + pricing_snapshots tables
- [x] Implement `internal/cost/tracker.go` — cost calculation logic
- [x] Implement `internal/cost/pricing.go` — pricing resolution (override → fetched → hardcoded)
- [x] Implement pricing fetch from GitHub on startup
- [x] Bundle `pricing.json` in binary (go:embed)
- [x] Record cost entry per request in SQLite
- [x] Log `pricing_source` on every cost record
- [x] Implement CLI: `tendr cost` — today / week / month / all-time summary
- [x] Implement CLI: `tendr cost --provider openai`
- [x] Implement CLI: `tendr cost --json` for script-friendly output
- [x] Push to GitHub

**Checkpoint:** Cost is recorded accurately. Pricing source is visible. CLI queries work.

---

### Stage 5 — Caching
**Goal:** Repeated identical requests return cached responses at $0.00.
**Shippable:** `tendr cache` shows hit rate and entries.
**Time limit:** 1 week

- [x] Implement `internal/cache/exact.go` — SHA256 hash, in-memory LRU
- [x] Implement cache TTL support
- [x] Implement cache hit path in gateway handler
- [x] Record cache hits in SQLite (cache_entries table)
- [x] Record cost as $0.00 on cache hit
- [x] Implement CLI: `tendr cache` — hit rate, entry count, cost saved
- [x] Implement CLI: `tendr cache clear`
- [x] Implement CLI: `tendr cache clear --alias <name>` (Parked: schema doesn't support alias tracking)
- [x] Optional: implement `internal/cache/disk.go` — bbolt persistence
- [ ] Push to GitHub

**Checkpoint:** Cache works. Hit rate is measurable. CLI commands work.

---

### Stage 6 — Rate Limiting
**Goal:** Protect provider quotas from accidental abuse.
**Shippable:** Rate limits configurable in YAML, enforced by gateway.
**Time limit:** 1 week

- [x] Implement `internal/ratelimit/limiter.go` — token bucket, in-memory
- [x] Implement per-model-alias rate limit
- [x] Implement per-provider rate limit
- [x] Return HTTP 429 with `retry_after_ms` on limit exceeded
- [x] Log `rate_limit_hit` event
- [x] Test: exceed limit → 429 returned → retry after cooldown works
- [x] Push to GitHub

**Checkpoint:** Rate limits enforced. 429 responses correct. Logs capture limit hits.

---

### Stage 7 — TUI
**Goal:** Interactive terminal dashboard works for monitoring and ops.
**Shippable:** `tendr` launches TUI with all 5 tabs functional.
**Time limit:** 1 week

- [ ] Implement `internal/tui/tui.go` — root Bubble Tea model, tab routing
- [ ] Implement `styles/styles.go` — all Lip Gloss definitions per DESIGN.md
- [ ] Implement Tab 1: Dashboard — gateway status, provider health, last 10 requests
- [ ] Implement Tab 2: Cost — summary cards, provider breakdown, inline bar chart
- [ ] Implement Tab 3: Cache — status, entry list, cache controls
- [ ] Implement Tab 4: Config — read-only view, masked API keys, open in $EDITOR
- [ ] Implement Tab 5: Logs — live log stream, pause/resume, level filter
- [ ] Implement all keyboard shortcuts per DESIGN.md
- [ ] Implement statusbar with contextual key hints
- [ ] Handle minimum terminal size (80×24) gracefully
- [ ] Wire: `tendr` (no args) → launch TUI
- [ ] Wire: `tendr monitor` → launch TUI on Dashboard tab
- [ ] Wire: `tendr cost` (standalone) → launch TUI on Cost tab
- [ ] Push to GitHub

**Checkpoint:** TUI works. All tabs render correctly. Keyboard navigation works. Live data updates.

---

### Stage 8 — Polish + Release
**Goal:** Binary is distributable and installable by others.
**Shippable:** v0.1.0 GitHub release with binaries for macOS/Linux/Windows.
**Time limit:** 1 week

- [ ] Write `.goreleaser.yaml` — cross-platform builds
- [ ] Write `Makefile` — build, test, lint, release targets
- [ ] Add `--version` flag to binary
- [ ] Write comprehensive README.md — install, quickstart, config reference
- [ ] Write CHANGELOG.md
- [ ] Add GitHub Actions CI — lint + test on push
- [ ] Test install flow: fresh machine, install binary, `tendr init`, first request
- [ ] Tag v0.1.0
- [ ] Create GitHub release with goreleaser
- [ ] Post to relevant communities

**Checkpoint:** v0.1.0 released. Anyone can install and use in < 5 minutes.

---

## Version B — Production Foundation

After v0.1.0 is shipped and has users.

### B1 — Observability
- [ ] Prometheus metrics endpoint `/metrics`
- [ ] OpenTelemetry trace export
- [ ] Health check endpoint `/health`

### B2 — Semantic Cache
- [ ] Embedding-based similarity cache
- [ ] Configurable similarity threshold
- [ ] Requires embedding provider (local ollama or API)

### B3 — Hardening
- [ ] Config hot-reload without restart
- [ ] Graceful shutdown with in-flight request drain
- [ ] Persistent rate limit state across restarts

### B4 — Developer Experience
- [ ] Shell completion (bash, zsh, fish)
- [ ] `tendr doctor` — diagnose config and connectivity issues
- [ ] `tendr benchmark` — latency test per provider

---

## Version C — Post-MVP Evolution

Only consider after consistent usage data from real users.

### C1 — Team Gateway
- [ ] Multi-user API key management
- [ ] Per-key usage limits and tracking
- [ ] Shared gateway mode (listen on network, not just localhost)

### C2 — Advanced Routing
- [ ] Cost-based routing (always route to cheapest that meets latency SLA)
- [ ] Load balancing across multiple keys per provider
- [ ] Circuit breaker per provider

### C3 — Local Model Support
- [ ] ollama provider adapter
- [ ] LM Studio provider adapter
- [ ] Hybrid routing: local first → cloud fallback

### C4 — Web Dashboard (Optional)
- [ ] Embedded web UI served by binary
- [ ] Read-only dashboard for sharing with non-terminal users

---

## PARKING LOT

Ideas that came up during planning but are NOT in scope. Review after v0.1.0.

| Idea | Why Parked |
|---|---|
| SaaS hosted version | Not the product direction |
| Prompt compression | Research-level complexity |
| Agent builder | Out of scope entirely |
| Plugin system | Premature abstraction |
| Fine-tuning support | Out of scope |
| Mobile companion app | Out of scope |
| Subscription billing | Not a SaaS |
| Multi-modal support | Defer post v0.1.0 |


<!--  must used -->
Before returning ANY code block, execute this mental check:
1. Scan all imported packages in the file.
2. For each package, verify if it is explicitly referenced in the functional code below.
3. If an import is NOT used (e.g., standard "embed" or utility packages), REMOVE the import line immediately.
4. If this rule is violated, the code compiler will FAIL.