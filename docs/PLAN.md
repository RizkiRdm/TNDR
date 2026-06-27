# Implementation Plan

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
4. If this rule is violated, the code compiler will FAIL..