# Graph Report - .  (2026-06-03)

## Corpus Check
- Corpus is ~15,409 words - fits in a single context window. You may not need a graph.

## Summary
- 48 nodes · 45 edges · 10 communities (6 shown, 4 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 7 edges (avg confidence: 0.85)
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
- [[_COMMUNITY_CCC Skill Integration|CCC Skill Integration]]

## God Nodes (most connected - your core abstractions)
1. `Server` - 5 edges
2. `runStart()` - 5 edges
3. `TENDR CLI Entry Point` - 4 edges
4. `Provider` - 3 edges
5. `OpenAIProvider` - 3 edges
6. `Config` - 3 edges
7. `Load()` - 3 edges
8. `NewServer()` - 3 edges
9. `NewOpenAIProvider()` - 2 edges
10. `validate()` - 2 edges

## Surprising Connections (you probably didn't know these)
- `Config` --conceptually_related_to--> `Default Configuration`  [INFERRED]
  internal/config/config.go → config.yaml
- `runStart()` --calls--> `NewOpenAIProvider()`  [INFERRED]
  cmd/tendr/main.go → internal/provider/openai/openai.go
- `runStart()` --calls--> `Load()`  [INFERRED]
  cmd/tendr/main.go → internal/config/config.go
- `runStart()` --calls--> `NewServer()`  [INFERRED]
  cmd/tendr/main.go → internal/gateway/gateway.go
- `runStart()` --calls--> `Init()`  [INFERRED]
  cmd/tendr/main.go → internal/logger/logger.go

## Hyperedges (group relationships)
- **TENDR Core Architecture** — tendr_main, gateway_gateway, provider_provider, config_config [EXTRACTED 1.00]

## Communities (10 total, 4 thin omitted)

### Community 0 - "CLI Core and OpenAI Adapter"
Cohesion: 0.18
Nodes (4): Init(), NewOpenAIProvider(), OpenAIProvider, runStart()

### Community 1 - "Configuration Management"
Cohesion: 0.33
Nodes (6): Load(), validate(), ModelAliasConfig, ProvidersConfig, ProviderSettings, ServerConfig

### Community 2 - "Core System Components"
Cohesion: 0.33
Nodes (7): Config, Default Configuration, HTTP Gateway Server, Structured Logger, OpenAI Provider Adapter, Provider, TENDR CLI Entry Point

### Community 4 - "AI Provider Interfaces"
Cohesion: 0.33
Nodes (5): Choice, CompletionRequest, CompletionResponse, Message, Usage

### Community 5 - "Project Documentation and Planning"
Cohesion: 0.67
Nodes (3): AI Agent Operational Contract, Implementation Plan, Project README

## Knowledge Gaps
- **12 isolated node(s):** `CompletionRequest`, `Message`, `CompletionResponse`, `Choice`, `Usage` (+7 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `runStart()` connect `CLI Core and OpenAI Adapter` to `Configuration Management`, `HTTP Gateway Implementation`?**
  _High betweenness centrality (0.408) - this node is a cross-community bridge._
- **Why does `Load()` connect `Configuration Management` to `CLI Core and OpenAI Adapter`?**
  _High betweenness centrality (0.316) - this node is a cross-community bridge._
- **Why does `Config` connect `Core System Components` to `Configuration Management`?**
  _High betweenness centrality (0.288) - this node is a cross-community bridge._
- **Are the 4 inferred relationships involving `runStart()` (e.g. with `Load()` and `Init()`) actually correct?**
  _`runStart()` has 4 INFERRED edges - model-reasoned connections that need verification._
- **What connects `CompletionRequest`, `Message`, `CompletionResponse` to the rest of the system?**
  _17 weakly-connected nodes found - possible documentation gaps or missing edges._