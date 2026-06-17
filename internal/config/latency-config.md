# Latency Threshold Implementation Plan

## Goal
Use `latency_threshold_ms` from `config.yaml` instead of hardcoded 500ms in `router.go`.

## Plan
1. Update `internal/config/config.go` to include `LatencyThresholdMs` in `ServerConfig`.
2. Update `internal/config/config.go` to set a default (e.g., 500ms).
3. Update `cmd/tendr/main.go` or where `NewRouter` is called to pass the new config value.
4. Update `internal/router/router.go` to use `cfg.Server.LatencyThresholdMs`.

## Verification
- Run `make test`
- Verify `config.yaml` can override threshold.
