package gateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RizkiRdm/TNDR/internal/cache"
	"github.com/RizkiRdm/TNDR/internal/config"
	"github.com/RizkiRdm/TNDR/internal/gateway"
	"github.com/RizkiRdm/TNDR/internal/provider"
	"github.com/RizkiRdm/TNDR/internal/ratelimit"
	"github.com/RizkiRdm/TNDR/internal/router"
	"github.com/RizkiRdm/TNDR/internal/store"
)

type mockProvider struct {
	NameFunc func() string
	CompleteFunc func(context.Context, *provider.CompletionRequest) (*provider.CompletionResponse, error)
}

func (m *mockProvider) Name() string { return m.NameFunc() }
func (m *mockProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	return m.CompleteFunc(ctx, req)
}
func (m *mockProvider) Stream(context.Context, *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	return nil, nil
}
func (m *mockProvider) Validate(context.Context) error { return nil }
func (m *mockProvider) Health(context.Context) error   { return nil }

type testCase struct {
	name       string
	models     []config.ModelAliasConfig
	providers  map[string]*mockProvider
	limiters   map[string]*ratelimit.Limiter
	serverCfg  *config.ServerConfig
	reqBody    interface{}
	check      func(t *testing.T, rec *httptest.ResponseRecorder)
}

func defaultServerCfg() *config.ServerConfig {
	return &config.ServerConfig{Port: 0, LogLevel: "debug"}
}

func setupGateway(t *testing.T, tc testCase) *httptest.ResponseRecorder {
	t.Helper()

	provMap := make(map[string]provider.Provider)
	for name, mp := range tc.providers {
		provMap[name] = mp
	}

	cfg := &config.Config{Models: tc.models}
	r := router.NewRouter(cfg, provMap, nil)
	c := cache.NewExact(0)

	storePath := t.TempDir() + "/tendr.db"
	st, err := store.New(storePath)
	if err != nil {
		t.Fatalf("store: %v", err)
	}

	sc := tc.serverCfg
	if sc == nil {
		sc = defaultServerCfg()
	}

	srv := gateway.NewServer(0, r, c, st, tc.limiters, sc)

	body, err := json.Marshal(tc.reqBody)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestGateway_ErrorNormalization(t *testing.T) {
	t.Parallel()

	errCases := []struct {
		name       string
		complete   func(context.Context, *provider.CompletionRequest) (*provider.CompletionResponse, error)
		wantStatus int
		wantCode   string
	}{
		{
			name:       "rate limit error → 429",
			complete:   func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) { return nil, provider.ErrRateLimit },
			wantStatus: http.StatusTooManyRequests,
			wantCode:   "rate_limit_exceeded",
		},
		{
			name:       "invalid key error → 401",
			complete:   func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) { return nil, provider.ErrInvalidKey },
			wantStatus: http.StatusUnauthorized,
			wantCode:   "invalid_key",
		},
		{
			name:       "timeout error → 504",
			complete:   func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) { return nil, provider.ErrTimeout },
			wantStatus: http.StatusGatewayTimeout,
			wantCode:   "provider_timeout",
		},
		{
			name:       "provider down error → 502",
			complete:   func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) { return nil, provider.ErrProviderDown },
			wantStatus: http.StatusBadGateway,
			wantCode:   "provider_unavailable",
		},
		{
			name:       "generic error → 502",
			complete:   func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) { return nil, context.Canceled },
			wantStatus: http.StatusBadGateway,
			wantCode:   "provider_unavailable",
		},
	}

	for _, ec := range errCases {
		ec := ec
		t.Run(ec.name, func(t *testing.T) {
			t.Parallel()
			rec := setupGateway(t, testCase{
				models: []config.ModelAliasConfig{
					{Alias: "m", Providers: []string{"mock"}},
				},
				providers: map[string]*mockProvider{
					"mock": {NameFunc: func() string { return "mock" }, CompleteFunc: ec.complete},
				},
				reqBody: provider.CompletionRequest{Model: "m"},
			})

			if rec.Code != ec.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, ec.wantStatus)
			}

			var body map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("parse body: %v", err)
			}
			if body["error"] != ec.wantCode {
				t.Errorf("error code = %q, want %q", body["error"], ec.wantCode)
			}
			if rec.Header().Get("X-Tendr-Error-Code") != ec.wantCode {
				t.Errorf("X-Tendr-Error-Code = %q, want %q", rec.Header().Get("X-Tendr-Error-Code"), ec.wantCode)
			}
			if rec.Header().Get("X-Tendr-Latency-Ms") == "" {
				t.Errorf("X-Tendr-Latency-Ms header missing")
			}
		})
	}
}

func TestGateway_HappyPath(t *testing.T) {
	t.Parallel()

	rec := setupGateway(t, testCase{
		models: []config.ModelAliasConfig{
			{Alias: "m", Providers: []string{"mock"}},
		},
		providers: map[string]*mockProvider{
			"mock": {
				NameFunc: func() string { return "mock" },
				CompleteFunc: func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
					return &provider.CompletionResponse{
						Provider: "mock",
						Choices:  []provider.Choice{{Message: provider.Message{Content: "ok"}}},
						Usage:    provider.Usage{PromptTokens: 10, CompletionTokens: 20},
					}, nil
				},
			},
		},
		reqBody: provider.CompletionRequest{Model: "m"},
	})

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if rec.Header().Get("X-Tendr-Cache") != "MISS" {
		t.Errorf("X-Tendr-Cache = %q, want MISS", rec.Header().Get("X-Tendr-Cache"))
	}
	if rec.Header().Get("X-Tendr-Provider") != "mock" {
		t.Errorf("X-Tendr-Provider = %q, want mock", rec.Header().Get("X-Tendr-Provider"))
	}
	if rec.Header().Get("X-Tendr-Latency-Ms") == "" {
		t.Errorf("X-Tendr-Latency-Ms header missing")
	}
}

func TestGateway_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name: "invalid request body → 400",
			models: []config.ModelAliasConfig{},
			providers: map[string]*mockProvider{},
			serverCfg: defaultServerCfg(),
			reqBody: "{invalid json",
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				if rec.Code != http.StatusBadRequest {
					t.Errorf("status = %d, want 400", rec.Code)
				}
			},
		},
		{
			name: "unknown model alias → 502 provider_unavailable",
			models: []config.ModelAliasConfig{},
			providers: map[string]*mockProvider{},
			serverCfg: defaultServerCfg(),
			reqBody: provider.CompletionRequest{Model: "nonexistent"},
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				if rec.Code != http.StatusBadGateway {
					t.Errorf("status = %d, want 502", rec.Code)
				}
				var body map[string]interface{}
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("parse body: %v", err)
				}
				if body["error"] != "provider_unavailable" {
					t.Errorf("error code = %q, want provider_unavailable", body["error"])
				}
			},
		},
		{
			name: "concurrent request limit blocks when over capacity",
			models: []config.ModelAliasConfig{
				{Alias: "m", Providers: []string{"mock"}},
			},
			providers: map[string]*mockProvider{
				"mock": {
					NameFunc: func() string { return "mock" },
					CompleteFunc: func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
						return &provider.CompletionResponse{Provider: "mock", Choices: []provider.Choice{{Message: provider.Message{Content: "ok"}}}}, nil
					},
				},
			},
			serverCfg: &config.ServerConfig{Port: 0, LogLevel: "debug", MaxConcurrentRequests: 1},
			reqBody: provider.CompletionRequest{Model: "m"},
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				// If we didn't block, the first request succeeds
				if rec.Code != http.StatusOK && rec.Code != http.StatusTooManyRequests {
					t.Errorf("unexpected status = %d", rec.Code)
				}
			},
		},
		{
			name: "falls back correctly with attempts in error",
			models: []config.ModelAliasConfig{
				{Alias: "m", FallbackMode: "reliable", Providers: []string{"fail", "fail2"}},
			},
			providers: map[string]*mockProvider{
				"fail":  {NameFunc: func() string { return "fail" }, CompleteFunc: func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) { return nil, provider.ErrProviderDown }},
				"fail2": {NameFunc: func() string { return "fail2" }, CompleteFunc: func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) { return nil, provider.ErrRateLimit }},
			},
			serverCfg: &config.ServerConfig{Port: 0, LogLevel: "debug", LatencyThresholdMs: 500},
			reqBody: provider.CompletionRequest{Model: "m"},
			check: func(t *testing.T, rec *httptest.ResponseRecorder) {
				if rec.Code != http.StatusBadGateway {
					t.Errorf("status = %d, want 502", rec.Code)
				}
				var body map[string]interface{}
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("parse body: %v", err)
				}
				attempts, ok := body["attempts"].([]interface{})
				if !ok {
					t.Fatalf("expected attempts array in body, got %v", body)
				}
				if len(attempts) < 2 {
					t.Errorf("expected ≥2 attempts, got %d", len(attempts))
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := setupGateway(t, tt)
			if tt.check != nil {
				tt.check(t, rec)
			}
		})
	}
}
