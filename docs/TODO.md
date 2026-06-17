## Release Blockers (WAJIB FIX SEBELUM PUBLIC RELEASE)

### Critical

- [ ] Fix cost calculation formula (saat ini over-report ~1000x).
- [ ] Update semua unit test cost calculation (file: pricing_test.go, tracker_test.go, store_test.go).
- [ ] Pindahkan Gemini API key dari URL query parameter ke request header (file: gemini.go).
- [ ] Tambah pricing untuk semua model yang didukung.
- [ ] Normalisasi error response ke client.
- [ ] Hilangkan internal/provider error leakage.

### High

- [ ] Tambah request body size limit (1MB).
- [ ] Tambah SQLite WAL mode.
- [ ] Tambah SQLite busy timeout.
- [ ] Set SQLite MaxOpenConns(1).
- [ ] Gunakan latency_threshold_ms dari config.
- [ ] Gunakan rate_limits dari config.
- [ ] Tambah endpoint GET /health.
- [ ] Fix Dashboard status agar tidak hardcoded RUNNING.
- [ ] Fix Dashboard port agar tidak hardcoded.

---

## Features To Add

### High Priority

#### Reliability

- [ ] GET /health endpoint.
- [ ] Startup provider validation.
- [ ] Request body size limit middleware.
- [ ] Provider health monitoring.
- [ ] SQLite WAL mode.
- [ ] SQLite connection tuning.

#### Developer Experience

- [ ] `tendr health`
- [ ] `tendr test --model <alias>`
- [ ] `tendr doctor`
- [ ] `tendr logs`
- [ ] `tendr start --dry-run`

#### Cost Visibility

- [ ] `tendr cost --explain`
- [ ] Daily cost warning threshold.
- [ ] Monthly spend projection.
- [ ] CSV export.

#### Setup Experience

- [ ] Interactive setup wizard.
- [ ] Provider auto-detection.
- [ ] API key validation during setup.

---

### Medium Priority

#### Cost Intelligence

- [ ] Cost comparison per request.
- [ ] Alternative model suggestions.
- [ ] Model recommendation system.
- [ ] Per-alias cost breakdown.

#### Productivity

- [ ] Request replay.
- [ ] Cache hit rate per alias.
- [ ] Session token budget tracking.
- [ ] Auto alias suggestion.

#### TUI

- [ ] Provider Health panel.
- [ ] Pretty log viewer.
- [ ] Live health status.
- [ ] Live uptime tracking.

---

### Low Priority

- [ ] Config hot reload.
- [ ] Shell completion.
- [ ] Request inspector.
- [ ] Advanced budget alerts.
- [ ] Historical analytics dashboard.

---

## Features To Modify

### Cost System

- [ ] Fix token pricing calculation.
- [ ] Expand pricing coverage.
- [ ] Improve cost reporting transparency.
- [ ] Add explainable cost breakdown.

### Configuration

- [ ] Write config to `~/.tendr/config.yaml`.
- [ ] Use permission `0600`.
- [ ] Remove hardcoded runtime values.
- [ ] Ensure all documented config fields actually work.

### Gateway

- [ ] Normalize all provider errors.
- [ ] Improve fallback explanations.
- [ ] Add request limits.
- [ ] Improve observability.

### Cache

- [ ] Replace pseudo-LRU eviction.
- [ ] Add background cleanup.
- [ ] Improve cache metrics.

### Logging

- [ ] Use RFC3339 timestamps.
- [ ] Match documentation format.
- [ ] Improve structured logging.

### TUI

- [ ] Remove hardcoded values.
- [ ] Implement documented shortcuts.
- [ ] Show real provider status.
- [ ] Improve log readability.

---

## Features To Remove

### Remove From Near-Term Scope

- [ ] Semantic cache (defer to V3+).
- [ ] Embedding-based cache.
- [ ] Enterprise-oriented features.
- [ ] Premature analytics features.

### Simplify

- [ ] Reduce TUI complexity until data is reliable.
- [ ] Remove unused pricing_snapshots table (or implement properly).
- [ ] Remove dead code in gateway.
- [ ] Remove dead code in CLI.

---

## Technical Debt

### High Priority

- [ ] Fix ConfigView dependency on Viper global state.
- [ ] Remove hardcoded latency threshold.
- [ ] Remove hardcoded rate limiter values.
- [ ] Remove hardcoded dashboard values.
- [ ] Fix duplicated cache code path.

### Medium Priority

- [ ] Implement proper LRU cache.
- [ ] Add cache concurrency tests.
- [ ] Add Gemini mock server tests.
- [ ] Improve flaky tests.

### Low Priority

- [ ] Cleanup unused code.
- [ ] Improve internal comments.
- [ ] Refactor TUI rendering.

---

## PLAN.md Audit

| Item | Decision | Reason |
|--------|--------|--------|
| Smart Routing | KEEP | Core differentiator |
| Cost Tracking | KEEP | Core value proposition |
| Unified API | KEEP | Foundation feature |
| Fallback Routing | KEEP | Reliability feature |
| SQLite Store | KEEP | Good enough for target users |
| TUI Dashboard | KEEP | Useful if data becomes accurate |
| Semantic Cache | DEFER | Complexity too high for MVP |
| Embedding Cache | DEFER | Not needed by solo developers |
| Advanced Analytics | DEFER | No user validation yet |
| Team Features | REMOVE | Not target audience |
| Enterprise Controls | REMOVE | Outside product focus |
| Pricing Snapshots Table | REVIEW | Currently unused |
| Multi-user Support | REMOVE | Premature |
| Web Dashboard | DEFER | CLI-first product |

---

## V2 Roadmap

### Goal

Create a clear differentiator versus LiteLLM and OpenRouter.

### Must Have

- [ ] Smart cost routing.
- [ ] Cost comparison per request.
- [ ] `tendr compare`.
- [ ] Interactive setup wizard.
- [ ] `tendr doctor`.
- [ ] Health endpoint.
- [ ] Provider health monitoring.
- [ ] Daily/monthly cost projections.
- [ ] CSV export.

### Success Metric

User can answer:

> "Model mana yang paling murah untuk workload saya?"

Dalam kurang dari 5 menit setelah install.

---

## V3 Roadmap

### Goal

Optimize real workflows based on user feedback.

### Candidate Features

- [ ] Request replay.
- [ ] Request inspector.
- [ ] Latency benchmarking.
- [ ] Budget enforcement.
- [ ] Advanced reporting.
- [ ] Per-alias analytics.
- [ ] Semantic cache (only if requested by users).

### Entry Requirement

Minimal:

- 50+ active users.
- Repeated feature requests.
- Real usage data.

---

## Future Ideas

### Potential

- [ ] Prometheus metrics.
- [ ] OpenTelemetry support.
- [ ] Ollama integration.
- [ ] Local model routing.
- [ ] Embedded web dashboard.
- [ ] Plugin system.

### Rule

Jangan implement sebelum ada demand nyata dari user.

---

## Product Principles

### Always Prioritize

1. Reliability.
2. Cost transparency.
3. Simplicity.
4. Local-first experience.
5. Fast onboarding.

### Avoid

1. Enterprise feature chasing.
2. Premature optimization.
3. Feature parity competition.
4. Complex setup.
5. Features without user demand.

---

## North Star

TENDR harus menjadi:

> "Tool yang membuat developer tahu persis berapa biaya AI mereka dan otomatis memilih provider termurah yang masih memenuhi kebutuhan mereka."

Bukan:

> "AI Gateway dengan fitur paling banyak."

<!--  must used -->
Before returning ANY code block, execute this mental check:
1. Scan all imported packages in the file.
2. For each package, verify if it is explicitly referenced in the functional code below.
3. If an import is NOT used (e.g., standard "embed" or utility packages), REMOVE the import line immediately.
4. If this rule is violated, the code compiler will FAIL..