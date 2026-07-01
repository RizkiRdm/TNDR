package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RizkiRdm/TNDR/internal/config"
)

// TestLoad_Defaults exercises black-box Load() via temp YAML files.
// Scenarios: Happy Path (minimal YAML gets defaults), Edge Case
// (extreme port + empty fields), Unhappy Path (invalid config errors).
func TestLoad_Defaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		check   func(t *testing.T, cfg *config.Config)
	}{
		{
			name: "happy path — minimal YAML gets all defaults applied",
			yaml: `server:
  port: 4821
providers:
  openai:
    api_key: "sk-test"
models:
  - alias: default
    providers: [openai]
`,
			wantErr: false,
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				// Server defaults
				if cfg.Server.Port != 4821 {
					t.Errorf("Port = %d, want 4821", cfg.Server.Port)
				}
				if cfg.Server.LogLevel != "info" {
					t.Errorf("LogLevel = %q, want info", cfg.Server.LogLevel)
				}
				if cfg.Server.LatencyThresholdMs != 500 {
					t.Errorf("LatencyThresholdMs = %d, want 500", cfg.Server.LatencyThresholdMs)
				}
				if cfg.Server.LogMaxSizeMB != 50 {
					t.Errorf("LogMaxSizeMB = %d, want 50", cfg.Server.LogMaxSizeMB)
				}
				if cfg.Server.LogMaxBackups != 5 {
					t.Errorf("LogMaxBackups = %d, want 5", cfg.Server.LogMaxBackups)
				}
				if cfg.Server.LogMaxAgeDays != 28 {
					t.Errorf("LogMaxAgeDays = %d, want 28", cfg.Server.LogMaxAgeDays)
				}
				// Provider timeout defaults
				if cfg.Providers.OpenAI.Timeout != 30000 {
					t.Errorf("OpenAI Timeout = %d, want 30000", cfg.Providers.OpenAI.Timeout)
				}
				if cfg.Providers.Anthropic.Timeout != 30000 {
					t.Errorf("Anthropic Timeout = %d, want 30000", cfg.Providers.Anthropic.Timeout)
				}
				if cfg.Providers.Gemini.Timeout != 30000 {
					t.Errorf("Gemini Timeout = %d, want 30000", cfg.Providers.Gemini.Timeout)
				}
				if cfg.Providers.Groq.Timeout != 30000 {
					t.Errorf("Groq Timeout = %d, want 30000", cfg.Providers.Groq.Timeout)
				}
				// Model rate limit default
				if len(cfg.Models) > 0 && cfg.Models[0].RateLimit != 10 {
					t.Errorf("RateLimit = %d, want 10", cfg.Models[0].RateLimit)
				}
				// GlobalKey and DailyCostLimit stay zero (not defaulted)
				if cfg.Server.GlobalKey != "" {
					t.Errorf("GlobalKey = %q, want empty string", cfg.Server.GlobalKey)
				}
			},
		},
		{
			name: "edge case — extreme port + explicit overrides not defaulted",
			yaml: `server:
  port: 65535
  log_level: debug
  latency_threshold_ms: 100
  global_key: "secret"
  daily_cost_limit: 5.0
  log_max_size_mb: 100
  log_max_backups: 10
  log_max_age_days: 7
providers:
  openai:
    api_key: "sk-test"
    timeout_ms: 60000
  anthropic:
    api_key: "sk-test2"
  gemini:
    api_key: ""
  groq:
    api_key: ""
models:
  - alias: extreme
    providers: [anthropic]
    rate_limit: 50
  - alias: empty-rate
    providers: [openai]
`,
			wantErr: false,
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				// Explicit overrides preserved
				if cfg.Server.Port != 65535 {
					t.Errorf("Port = %d, want 65535", cfg.Server.Port)
				}
				if cfg.Server.LogLevel != "debug" {
					t.Errorf("LogLevel = %q, want debug", cfg.Server.LogLevel)
				}
				if cfg.Server.LatencyThresholdMs != 100 {
					t.Errorf("LatencyThresholdMs = %d, want 100", cfg.Server.LatencyThresholdMs)
				}
				if cfg.Server.GlobalKey != "secret" {
					t.Errorf("GlobalKey = %q, want secret", cfg.Server.GlobalKey)
				}
				if cfg.Server.DailyCostLimit != 5.0 {
					t.Errorf("DailyCostLimit = %f, want 5.0", cfg.Server.DailyCostLimit)
				}
				if cfg.Server.LogMaxSizeMB != 100 {
					t.Errorf("LogMaxSizeMB = %d, want 100", cfg.Server.LogMaxSizeMB)
				}
				if cfg.Server.LogMaxBackups != 10 {
					t.Errorf("LogMaxBackups = %d, want 10", cfg.Server.LogMaxBackups)
				}
				if cfg.Server.LogMaxAgeDays != 7 {
					t.Errorf("LogMaxAgeDays = %d, want 7", cfg.Server.LogMaxAgeDays)
				}
				// Explicit timeout override
				if cfg.Providers.OpenAI.Timeout != 60000 {
					t.Errorf("OpenAI Timeout = %d, want 60000", cfg.Providers.OpenAI.Timeout)
				}
				// Default timeout for unspecified providers
				if cfg.Providers.Anthropic.Timeout != 30000 {
					t.Errorf("Anthropic Timeout = %d, want 30000", cfg.Providers.Anthropic.Timeout)
				}
				if cfg.Providers.Gemini.Timeout != 30000 {
					t.Errorf("Gemini Timeout = %d, want 30000", cfg.Providers.Gemini.Timeout)
				}
				// Explicit rate limit preserved
				if cfg.Models[0].RateLimit != 50 {
					t.Errorf("Models[0].RateLimit = %d, want 50", cfg.Models[0].RateLimit)
				}
				// Default rate limit for model with 0 value
				if cfg.Models[1].RateLimit != 10 {
					t.Errorf("Models[1].RateLimit = %d, want 10", cfg.Models[1].RateLimit)
				}
			},
		},
		{
			name: "unhappy path — invalid port >65535 triggers validation error",
			yaml: `server:
  port: 99999
providers:
  openai:
    api_key: "sk-test"
models: []
`,
			wantErr: true,
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				// defaults still applied for other fields despite validation failure
				if cfg.Server.LogLevel != "info" {
					t.Errorf("LogLevel = %q, want info (default applied before validate)", cfg.Server.LogLevel)
				}
			},
		},
		{
			name: "unhappy path — empty model alias",
			yaml: `server:
  port: 4821
providers:
  openai:
    api_key: "sk-test"
models:
  - alias: ""
    providers: [openai]
`,
			wantErr: true,
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				if len(cfg.Models) != 1 {
					t.Errorf("len(Models) = %d, want 1", len(cfg.Models))
				}
			},
		},
		{
			name: "unhappy path — no providers for model",
			yaml: `server:
  port: 4821
providers:
  openai:
    api_key: "sk-test"
models:
  - alias: orphan
    providers: []
`,
			wantErr: true,
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				if cfg.Server.LogLevel != "info" {
					t.Errorf("LogLevel = %q, want info (default applied)", cfg.Server.LogLevel)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0600); err != nil {
				t.Fatalf("write temp config: %v", err)
			}

			cfg, err := config.Load(path)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if cfg != nil {
				tt.check(t, cfg)
			}
		})
	}
}
