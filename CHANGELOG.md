# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-06-14

### Added
- Multi-provider support (OpenAI, Anthropic, Gemini, Groq)
- Automatic fallback engine with 3 modes (reliable, fast, smart)
- SQLite-based cost tracking and usage analytics
- Exact request caching for zero-cost repeated queries
- Rate limiting per model alias and provider
- Interactive TUI dashboard with real-time monitoring
- Cross-platform build system via GoReleaser
