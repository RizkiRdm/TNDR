# Graph Report - .  (2026-06-04)

## Corpus Check
- Corpus is ~16,449 words - fits in a single context window. You may not need a graph.

## Summary
- 461 nodes · 463 edges · 31 communities (24 shown, 7 thin omitted)
- Extraction: 98% EXTRACTED · 2% INFERRED · 0% AMBIGUOUS · INFERRED: 9 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_CLI Core and OpenAI Adapter|CLI Core and OpenAI Adapter]]
- [[_COMMUNITY_Configuration Management|Configuration Management]]
- [[_COMMUNITY_Core System Components|Core System Components]]
- [[_COMMUNITY_HTTP Gateway Implementation|HTTP Gateway Implementation]]
- [[_COMMUNITY_AI Provider Interfaces|AI Provider Interfaces]]
- [[_COMMUNITY_Project Documentation and Planning|Project Documentation and Planning]]
- [[_COMMUNITY_Gemini CLI Settings|Gemini CLI Settings]]
- [[_COMMUNITY_Architecture and PRD|Architecture and PRD]]
- [[_COMMUNITY_Dependency Management|Dependency Management]]
- [[_COMMUNITY_CCC Skill Integration|CCC Skill Integration]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]

## God Nodes (most connected - your core abstractions)
1. `AI Agent Operational Contract` - 18 edges
2. `AI Agent Operational Contract` - 18 edges
3. `AI Agent Operational Contract` - 17 edges
4. `Engineering Blueprint` - 15 edges
5. `Architecture Overview` - 14 edges
6. `Tokenized Engine for Networked Distributed External Routing` - 12 edges
7. `Terminal UI Design System` - 11 edges
8. `5. Core Capabilities (MVP)` - 10 edges
9. `Version A — MVP Execution Plan` - 9 edges
10. `Implementation Plan` - 8 edges

## Surprising Connections (you probably didn't know these)
- `Config` --conceptually_related_to--> `Default Configuration`  [INFERRED]
  internal/config/config.go → config.yaml
- `runStart()` --calls--> `NewOpenAIProvider()`  [INFERRED]
  cmd/tendr/main.go → internal/provider/openai/openai.go
- `runStart()` --calls--> `NewServer()`  [INFERRED]
  cmd/tendr/main.go → internal/gateway/gateway.go
- `runStart()` --calls--> `Init()`  [INFERRED]
  cmd/tendr/main.go → internal/logger/logger.go
- `runStart()` --calls--> `Load()`  [INFERRED]
  cmd/tendr/main.go → internal/config/config.go

## Communities (31 total, 7 thin omitted)

### Community 0 - "CLI Core and OpenAI Adapter"
Cohesion: 0.05
Nodes (42): AI Agent Operational Contract, AI Agent Rules, API Conventions, Architecture Overview, CGO, code:block1 (cmd/tendr/main.go), code:go (// CORRECT — explicit column list), code:go (// preferred test pattern) (+34 more)

### Community 1 - "Configuration Management"
Cohesion: 0.05
Nodes (41): code:block2 (tendr/), Project Structure, 10. Cache Contract, 11. Rate Limit Contract, 12. Project Structure, 13. Component Responsibility Matrix, 14. Verification Rules, 1. Context Lock (+33 more)

### Community 2 - "Core System Components"
Cohesion: 0.05
Nodes (39): 10. Forbidden Patterns, 1. Design Philosophy, 2. Color System, 3. Typography, 4. Layout System, 7. Keyboard Navigation, 8. Motion Rules, 9. CLI Output Style (+31 more)

### Community 3 - "HTTP Gateway Implementation"
Cohesion: 0.06
Nodes (38): AI Agent Operational Contract, API Conventions, CGO, code:go (// CORRECT — explicit column list), code:go (// preferred test pattern), code:bash (# Build), code:go (// CORRECT), code:go (// internal/provider/provider.go) (+30 more)

### Community 4 - "AI Provider Interfaces"
Cohesion: 0.05
Nodes (37): AI Agent Operational Contract, AI Agent Rules, API Conventions, Architecture Overview, Branch Naming, CGO, code:block1 (cmd/tendr/main.go   → CLI flags, launch GW/TUI), code:block2 (tendr/) (+29 more)

### Community 5 - "Project Documentation and Planning"
Cohesion: 0.06
Nodes (29): How to Use This File, Ideas That Are NOT In Scope, Parked Ideas, PARKING_LOT.md — TENDR, Implementation Plan, B1 — Observability, B2 — Semantic Cache, B3 — Hardening (+21 more)

### Community 6 - "Gemini CLI Settings"
Cohesion: 0.07
Nodes (21): Config, Load(), validate(), ModelAliasConfig, ProvidersConfig, ProviderSettings, ServerConfig, Default Configuration (+13 more)

### Community 7 - "Architecture and PRD"
Cohesion: 0.07
Nodes (27): 10. Technical Constraints, 11. Anti-Mangkrak Rules (Project Discipline), 1. Project Summary, 2. Target Users, 3. Problem Statement, 4. Success Metrics, 5.1 Unified Proxy Endpoint, 5.2 Multi-Provider Normalization (+19 more)

### Community 8 - "Dependency Management"
Cohesion: 0.10
Nodes (20): Authoring `guides.yml` Interactively, ccc - Semantic Code Search & Indexing, code:bash (ccc search <query terms>), code:bash (ccc search database connection pooling), code:bash (ccc search --lang python --lang markdown database schema), code:bash (ccc search --path 'src/api/*' request validation), code:bash (ccc search --offset 5 --limit 5 database schema), code:bash (ccc describe src/auth/session.py        # one file) (+12 more)

### Community 9 - "CCC Skill Integration"
Cohesion: 0.14
Nodes (15): ccc Management, Checking Project Status, Cleanup, code:bash (pipx install 'cocoindex-code[full]'      # batteries include), code:bash (pipx upgrade cocoindex-code), code:bash (ccc init), code:bash (ccc init --litellm-model openai/text-embedding-3-small), code:bash (ccc doctor) (+7 more)

### Community 10 - "Community 10"
Cohesion: 0.20
Nodes (14): code:block1 (cmd/tendr/main.go), cmd/tendr, internal/cache, internal/config, internal/cost, internal/gateway, internal/logger, internal/provider (+6 more)

### Community 11 - "Community 11"
Cohesion: 0.18
Nodes (9): code:bash (go build -o tendr ./cmd/tendr), code:bash (./tendr init), code:bash (./tendr start), code:bash (curl -X POST http://localhost:4821/v1/chat/completions \), Development, Installation, Quickstart, Stage 1: Foundation (Current) (+1 more)

### Community 12 - "Community 12"
Cohesion: 0.18
Nodes (8): 5. Data Model Contract, `cache_entries`, Constraint Policy, Index Policy, Normalization Level: 3NF, `pricing_snapshots`, `requests`, Tables

### Community 13 - "Community 13"
Cohesion: 0.18
Nodes (11): 5. Tab Specifications, code:block5 (┌─ GATEWAY STATUS ──────────────────────────────────────────), code:block6 (┌─ COST SUMMARY ────────────────────────────────────────────), code:block7 (┌─ CACHE STATUS ────────────────────────────────────────────), code:block8 (┌─ ACTIVE CONFIG ───────────────────────────────────────────), code:block9 (┌─ REQUEST LOG ─────────────────────────────────────────────), Tab 1 — Dashboard, Tab 2 — Cost (+3 more)

### Community 14 - "Community 14"
Cohesion: 0.22
Nodes (10): ccc Settings, code:yaml (embedding:), code:bash (ccc reset && ccc index), code:yaml (include_patterns:), Editing Tips, Embedding Model Examples, Fields, Important (+2 more)

### Community 15 - "Community 15"
Cohesion: 0.20
Nodes (10): 6. Component Patterns, Border Boxes, code:block10 (● RUNNING     (ColorSuccess + bold)), code:block11 (● ok      (ColorSuccess)), code:block12 (OpenAI   ████████████░░░░  67%), code:block13 (COLUMN_A    COLUMN_B    COLUMN_C), Inline Bar Chart (Cost/Usage), Provider Health Badge (+2 more)

### Community 16 - "Community 16"
Cohesion: 0.22
Nodes (8): Agent Guidelines, Coding Conventions, Development Workflow, graphify, Key Commands, Project Context, Tech Stack, TENDR — AI Gateway Binary

### Community 17 - "Community 17"
Cohesion: 0.29
Nodes (8): 7. User Flows, code:block1 (Developer downloads binary (brew / curl / go install)), code:block2 (Request arrives at TENDR), code:block4 (Developer opens TUI: tendr), Flow 1 — Installation and First Request, Flow 2 — Fallback Triggered, Flow 3 — Cache Hit, Flow 4 — Cost Review in TUI

### Community 18 - "Community 18"
Cohesion: 0.29
Nodes (7): 9. Logging Contract, code:json ({), code:block15 (max_size_mb: 50        (per file)), JSONL Schema (every log entry), Level Policy, Log Rotation Policy, Required Events

### Community 19 - "Community 19"
Cohesion: 0.29
Nodes (7): 8. Cost Tracking Contract, code:block11 (1. Check YAML override for provider + model → use if present), code:block12 (input_cost  = (input_tokens  / 1_000_000) * pricing.input_pe), code:block13 (MUST fetch on startup if fetch_on_startup: true), Cost Calculation, Pricing Fetch Policy, Pricing Resolution Order

### Community 21 - "Community 21"
Cohesion: 0.33
Nodes (6): Branch Naming, code:block15 (feat/stage-1-foundation), code:block16 (feat(provider): add anthropic adapter), Commit Messages, Git Conventions, PR Rules

### Community 22 - "Community 22"
Cohesion: 0.33
Nodes (6): Branch Naming, code:block15 (feat/stage-1-foundation), code:block16 (feat(provider): add anthropic adapter), Commit Messages, Git Conventions, PR Rules

### Community 23 - "Community 23"
Cohesion: 0.50
Nodes (4): Architectural Layers, TENDR — AI Gateway Binary, TENDR — AI Gateway Binary, TENDR — AI Gateway Binary

## Knowledge Gaps
- **241 isolated node(s):** `CompletionRequest`, `Message`, `CompletionResponse`, `Choice`, `Usage` (+236 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `AI Agent Operational Contract` connect `HTTP Gateway Implementation` to `Configuration Management`, `Community 10`, `Community 11`, `Community 21`?**
  _High betweenness centrality (0.100) - this node is a cross-community bridge._
- **Why does `Engineering Blueprint` connect `Configuration Management` to `Community 18`, `Community 19`, `Community 12`?**
  _High betweenness centrality (0.079) - this node is a cross-community bridge._
- **Why does `Project Structure` connect `Configuration Management` to `HTTP Gateway Implementation`?**
  _High betweenness centrality (0.063) - this node is a cross-community bridge._
- **What connects `CompletionRequest`, `Message`, `CompletionResponse` to the rest of the system?**
  _245 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `CLI Core and OpenAI Adapter` be split into smaller, more focused modules?**
  _Cohesion score 0.048625792811839326 - nodes in this community are weakly interconnected._
- **Should `Configuration Management` be split into smaller, more focused modules?**
  _Cohesion score 0.047619047619047616 - nodes in this community are weakly interconnected._
- **Should `Core System Components` be split into smaller, more focused modules?**
  _Cohesion score 0.05 - nodes in this community are weakly interconnected._