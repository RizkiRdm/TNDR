# TENDR — AI Gateway Binary

TENDR is a self-hosted, single-binary AI gateway written in Go. It proxies requests to multiple AI providers with local cost tracking, caching, rate limiting, and an interactive TUI dashboard.

## Features

- **Multi-Provider Support**: Unified OpenAI-compatible API for OpenAI, Anthropic, Gemini, and Groq.
- **Automatic Fallback**: Intelligent routing with `reliable`, `fast`, and `smart` failover modes.
- **Cost Tracking**: Local SQLite database to record every request and calculate USD cost.
- **Request Caching**: In-memory and disk persistence to save tokens on repeated identical queries.
- **Rate Limiting**: Protect your provider quotas with per-alias and per-provider limits.
- **Interactive TUI**: Real-time monitoring dashboard, cost analytics, and log viewer in your terminal.
- **Zero Dependencies**: Pure Go implementation, no CGO required.

## Installation

### From Source
```bash
make install
```

### Pre-built Binaries
Download the latest release for your platform from the [Releases](https://github.com/RizkiRdm/TNDR/releases) page.

## Quickstart

1. **Initialize configuration**:
   ```bash
   tendr init
   ```
2. **Configure API keys**: Edit `config.yaml` and add your provider keys.
3. **Start the gateway**:
   ```bash
   tendr start
   ```
4. **Proxy a request**:
   ```bash
   curl -X POST http://localhost:4821/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{
       "model": "coding",
       "messages": [{"role": "user", "content": "Hello TENDR"}]
     }'
   ```
5. **Monitor usage**:
   ```bash
   tendr monitor
   ```

## TUI Dashboard

Launch the interactive dashboard to see your gateway in action:
```bash
tendr monitor
```
- **Tab 1**: Live Dashboard (Status, Health, Last Requests)
- **Tab 2**: Cost Analytics (Spending by provider/time)
- **Tab 3**: Cache Stats (Hit rate, storage)
- **Tab 4**: Config Viewer (Masked keys)
- **Tab 5**: Live Logs (Stream filtered logs)

## CLI Reference

- `tendr start`: Run the gateway server.
- `tendr init`: Generate a default `config.yaml`.
- `tendr cost`: Show usage cost summary.
- `tendr monitor`: Launch the TUI dashboard.
- `tendr cache clear`: Purge the request cache.
- `tendr --version`: Show version and build info.

## License

MIT
