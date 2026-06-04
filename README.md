# TENDR — AI Gateway Binary

TENDR is a self-hosted, single-binary AI gateway written in Go. It proxies requests to multiple AI providers with local cost tracking, caching, and a TUI dashboard.

## Stage 3: Fallback Engine (Current)

- Go 1.25+
- HTTP Gateway on port 4821
- Multi-provider support (OpenAI, Anthropic, Gemini, Groq)
- Provider failover (Reliable, Fast modes)
- Structured Logging (zerolog + lumberjack)
- YAML Configuration (viper)

## Installation

```bash
go build -o tendr ./cmd/tendr
```

## Quickstart

1. Initialize configuration:
   ```bash
   ./tendr init
   ```
2. Edit `config.yaml` and add your API keys.
3. Start the gateway:
   ```bash
   ./tendr start
   ```
4. Test the completions endpoint:
   ```bash
   curl -X POST http://localhost:4821/v1/chat/completions \
     -d '{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hello TENDR"}]}'
```

## Development

Project structure follows the [AGENTS.md](./AGENTS.md) operational contract.

- `cmd/tendr`: Entry point
- `internal/gateway`: HTTP server and routing
- `internal/provider`: AI provider adapters
- `internal/router`: Provider selection and fallback logic
- `internal/config`: Configuration management
- `internal/logger`: Structured logging
