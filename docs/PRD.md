# PRD.md — TENDR
# Tokenized Engine for Networked Distributed External Routing

Document Version: 1.0.0
Status: MVP Definition
Type: Open Source CLI + TUI Tool

---

## 1. Project Summary

### Overview
TENDR is a self-hosted, single-binary AI gateway written in Go. It runs locally on a developer's machine and acts as a unified proxy between developer tools and multiple AI providers.

### Objective
Give developers a single local endpoint that handles provider routing, fallback, caching, and cost tracking — without managing multiple API keys, SDKs, or provider-specific configurations in each project.

### Value Proposition
Install once. Point all your tools at `localhost:4821`. Never touch provider config again.

---

## 2. Target Users

### Primary — Solo Developer / Indie Hacker
- Builds personal projects, freelance systems, or indie SaaS
- Uses AI heavily in development workflow (Cursor, Continue.dev, Cline, etc.)
- Budget-sensitive — wants to know exactly what they are spending
- Does not want operational overhead
- Installs tools via single binary or package manager

**Frustrations:**
- Paying for multiple provider subscriptions
- Not knowing how much a session actually cost
- Manually switching providers when one goes down
- Re-configuring API keys in every project

### Secondary — Small Engineering Team
- 2–10 engineers sharing AI tooling budget
- Needs shared gateway for consistent provider access
- Wants centralized cost visibility

---

## 3. Problem Statement

Developers using AI tools face three friction points:

**Reliability friction** — A single provider going down, rate-limiting, or timing out breaks the entire workflow. There is no automatic fallback.

**Cost opacity** — Most tools either show no cost tracking or show inaccurate costs based on stale pricing tables. Developers are surprised by actual bills.

**Integration friction** — Every project needs its own provider configuration. Switching providers requires code changes.

TENDR solves all three by acting as a local smart proxy: one endpoint, automatic fallback, accurate cost tracking.

---

## 4. Success Metrics

| Metric | Target |
|---|---|
| Time from install to first successful proxied request | < 5 minutes |
| Binary size | < 30MB |
| Gateway latency overhead | < 50ms p95 |
| Cache hit reduces cost to | $0.00 |
| Cost tracking accuracy vs provider dashboard | < 2% variance |
| TUI startup time | < 500ms |
| Fallback trigger to next provider | < 3s |
| Config parse error surfaced to user | 100% with line reference |

---

## 5. Core Capabilities (MVP)

### 5.1 Unified Proxy Endpoint
Single local HTTP endpoint, OpenAI-compatible.
All AI tools that support custom base URL work without modification.

### 5.2 Multi-Provider Normalization
Normalize request/response format across:
- OpenAI
- Anthropic
- Google Gemini
- Groq

Each provider has an adapter. Input always follows OpenAI schema. Output always follows OpenAI schema.

### 5.3 Fallback Engine
Three modes, user-configured per model alias:
- `reliable` — fallback on hard errors (5xx, timeout, 429)
- `fast` — fallback on latency threshold exceeded
- `smart` — fallback on both conditions

Provider chain is ordered. TENDR tries providers in order until one succeeds or all fail.

### 5.4 Caching Layer
Two cache types:
- **Exact cache** — identical request body hash → return stored response
- **Semantic cache** — embedding similarity above threshold → return closest stored response

Cache backend: in-memory (default) + optional disk persistence (bbolt).
Cache is always opt-in per model alias.

### 5.5 Cost Tracking
Per-request cost calculation using:
- token counts from provider response
- pricing table (hardcoded default + GitHub-fetched updates + user YAML override)

Stored in SQLite. Queryable via TUI and CLI.
Every cost entry records which pricing source was used.

### 5.6 Rate Limiting
Per-API-key and per-provider rate limiting.
Prevents hitting provider limits accidentally.
Configurable in YAML.

### 5.7 JSONL Request Logging
Every request logged to JSONL file with rotation.
Schema compatible with OpenTelemetry, Grafana Loki, Datadog.

### 5.8 TUI (Terminal UI)
Interactive dashboard via Bubble Tea.
Tabs: Dashboard | Cost | Cache | Config | Logs

### 5.9 CLI
All TUI operations accessible as CLI commands for scripting and quick ops.

---

## 6. Non-MVP Features

The following are EXPLICITLY OUT OF SCOPE for v1:

| Feature | Reason |
|---|---|
| Web UI / browser dashboard | Binary-first, terminal-native tool |
| Auth / API key management for users | Single-user local tool in MVP |
| Billing / payments | Not a SaaS |
| Multi-user / team shared gateway | v2 feature |
| Plugin system | Premature abstraction |
| Agent builder | Out of scope |
| Image generation / multi-modal | Complexity, defer |
| Fine-tuning | Out of scope |
| Kubernetes / Docker orchestration | Self-hosted binary only |
| Prometheus metrics endpoint | v2 feature |
| OpenTelemetry exporter | v2 feature |
| Workflow automation | Out of scope |

---

## 7. User Flows

### Flow 1 — Installation and First Request

```
Developer downloads binary (brew / curl / go install)
  ↓
tendr init → generates config.yaml with guided prompts
  ↓
Developer adds API keys to config.yaml
  ↓
tendr start → gateway running on localhost:4821
  ↓
Developer points Cursor / app to http://localhost:4821/v1
  ↓
First request proxied successfully
  ↓
tendr → opens TUI, shows first request in logs
```

### Flow 2 — Fallback Triggered

```
Request arrives at TENDR
  ↓
Route to primary provider (OpenAI)
  ↓
Provider returns 429 (rate limit)
  ↓
Fallback engine detects trigger condition
  ↓
Route to secondary provider (Anthropic)
  ↓
Success → response returned to client
  ↓
Log entry records: fallback_used: true, fallback_reason: "429"
```

### Flow 3 — Cache Hit

```
Request arrives at TENDR
  ↓
Hash request body
  ↓
Exact match found in cache
  ↓
Return cached response immediately
  ↓
Cost recorded as $0.00, cache_hit: true
  ↓
No provider request made
```

### Flow 4 — Cost Review in TUI

```
Developer opens TUI: tendr
  ↓
Navigates to Cost tab
  ↓
Sees: today $0.042 | this week $0.31 | this month $1.24
  ↓
Drills into model breakdown
  ↓
Sees pricing source for each entry
  ↓
Exports to CSV if needed
```

---

## 8. Configuration Schema (YAML)

```yaml
tendr:
  port: 4821
  log_level: info
  log_file: ~/.tendr/logs/tendr.jsonl
  log_rotation:
    max_size_mb: 50
    max_backups: 5

providers:
  openai:
    api_key: "sk-..."
    timeout_ms: 30000
  anthropic:
    api_key: "sk-ant-..."
    timeout_ms: 30000
  gemini:
    api_key: "AI..."
    timeout_ms: 30000
  groq:
    api_key: "gsk_..."
    timeout_ms: 10000

models:
  - alias: "default"
    fallback_mode: reliable
    providers:
      - provider: openai
        model: gpt-4o
        priority: 1
      - provider: anthropic
        model: claude-3-5-sonnet-20241022
        priority: 2

  - alias: "fast"
    fallback_mode: fast
    latency_threshold_ms: 5000
    providers:
      - provider: groq
        model: llama-3.1-70b-versatile
        priority: 1
      - provider: openai
        model: gpt-4o-mini
        priority: 2

  - alias: "coding"
    fallback_mode: smart
    latency_threshold_ms: 10000
    cache:
      enabled: true
      type: exact
      ttl_minutes: 60
    providers:
      - provider: openai
        model: gpt-4o
        priority: 1
      - provider: anthropic
        model: claude-3-5-sonnet-20241022
        priority: 2

pricing:
  fetch_on_startup: true
  pricing_url: "https://raw.githubusercontent.com/username/tendr/main/pricing.json"
  override:
    openai:
      gpt-4o:
        input_per_1m: 2.50
        output_per_1m: 10.00

rate_limits:
  per_minute: 60
  per_provider:
    openai: 500
    groq: 100
```

---

## 9. Tech Stack

| Layer | Technology | Rationale |
|---|---|---|
| Language | Go 1.22+ | Single binary, fast, strong concurrency |
| TUI | Bubble Tea + Lip Gloss | De facto standard Go TUI framework |
| HTTP Server | net/http + chi router | Stdlib-first, minimal dependency |
| Config | viper + YAML | Standard Go config library |
| Database | SQLite via modernc/sqlite | No CGO, pure Go, zero dependency |
| Cache (memory) | sync.Map + custom LRU | Zero dependency |
| Cache (disk) | bbolt | Pure Go embedded KV store |
| Logging | zerolog | JSONL output, structured, fast |
| Log rotation | lumberjack | Battle-tested, integrates with zerolog |
| Embeddings (semantic cache) | local via ollama OR provider API | Optional feature |
| Build | goreleaser | Cross-platform binary releases |

---

## 10. Technical Constraints

- MUST compile to single binary, no runtime dependencies
- MUST NOT use CGO (ensures cross-platform builds)
- MUST run on macOS, Linux, Windows
- MUST NOT require Docker or any container runtime
- MUST NOT require network access after initial pricing fetch
- SQLite database stored at `~/.tendr/tendr.db`
- Config stored at `~/.tendr/config.yaml`
- Logs stored at `~/.tendr/logs/`

---

## 11. Anti-Mangkrak Rules (Project Discipline)

These rules exist because the developer has identified a personal pattern of abandoning projects. They are part of the PRD intentionally.

1. Each stage MUST be independently shippable
2. Each stage MUST be completable within 1 week
3. A new stage MUST NOT start before current stage is pushed to GitHub
4. New ideas MUST go to PARKING_LOT.md — not into code
5. No feature is valid unless it exists in PLAN.md
6. If a stage takes longer than 1 week, scope MUST be cut — not extended