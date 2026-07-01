package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RizkiRdm/TNDR/internal/cache"
	"github.com/RizkiRdm/TNDR/internal/config"
	"github.com/RizkiRdm/TNDR/internal/provider"
	"github.com/RizkiRdm/TNDR/internal/ratelimit"
	"github.com/RizkiRdm/TNDR/internal/router"
	"github.com/RizkiRdm/TNDR/internal/store"
)

// MockProvider implements provider.Provider for testing
type MockProvider struct {
	NameFunc     func() string
	CompleteFunc func(context.Context, *provider.CompletionRequest) (*provider.CompletionResponse, error)
}

func (m *MockProvider) Name() string { return m.NameFunc() }
func (m *MockProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	return m.CompleteFunc(ctx, req)
}
func (m *MockProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	return nil, nil
}
func (m *MockProvider) Validate(ctx context.Context) error { return nil }
func (m *MockProvider) Health(ctx context.Context) error { return nil }

func setupGateway(t *testing.T, providers []provider.Provider, models []config.ModelAliasConfig) (*Server, *store.Store) {
	tmpDB := t.TempDir() + "/tendr.db"
	st, err := store.New(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	cfg := &config.Config{Models: models}
	provMap := make(map[string]provider.Provider)
	for _, p := range providers {
		provMap[p.Name()] = p
	}

	r := router.NewRouter(cfg, provMap, nil)
	c := cache.NewExact(1 * time.Minute)
	serverCfg := &config.ServerConfig{Port: 0, LogLevel: "debug"}

	return NewServer(0, r, c, st, nil, serverCfg), st
}

func TestGateway_HappyPath(t *testing.T) {
	mockP := &MockProvider{
		NameFunc: func() string { return "mock" },
		CompleteFunc: func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
			return &provider.CompletionResponse{Choices: []provider.Choice{{Message: provider.Message{Content: "ok"}}}}, nil
		},
	}
	models := []config.ModelAliasConfig{{Alias: "my-model", Providers: []string{"mock"}}}
	gw, _ := setupGateway(t, []provider.Provider{mockP}, models)

	reqBody, _ := json.Marshal(provider.CompletionRequest{Model: "my-model"})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	gw.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-Tendr-Cache") != "MISS" {
		t.Errorf("expected MISS, got %s", w.Header().Get("X-Tendr-Cache"))
	}
}

func TestGateway_CacheHit(t *testing.T) {
	mockP := &MockProvider{
		NameFunc: func() string { return "mock" },
		CompleteFunc: func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
			return &provider.CompletionResponse{Choices: []provider.Choice{{Message: provider.Message{Content: "ok"}}}}, nil
		},
	}
	models := []config.ModelAliasConfig{{Alias: "my-model", Providers: []string{"mock"}}}
	gw, _ := setupGateway(t, []provider.Provider{mockP}, models)

	reqBody, _ := json.Marshal(provider.CompletionRequest{Model: "my-model"})
	
	// First request (MISS)
	req1 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
	w1 := httptest.NewRecorder()
	gw.router.ServeHTTP(w1, req1)
	
	// Second request (HIT)
	req2 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
	w2 := httptest.NewRecorder()
	gw.router.ServeHTTP(w2, req2)

	if w2.Header().Get("X-Tendr-Cache") != "HIT" {
		t.Errorf("expected HIT, got %s", w2.Header().Get("X-Tendr-Cache"))
	}
}

func TestGateway_UnknownModelAlias(t *testing.T) {
	gw, _ := setupGateway(t, nil, nil)
	reqBody, _ := json.Marshal(provider.CompletionRequest{Model: "unknown"})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	gw.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", w.Code)
	}
}

func TestGateway_ProviderFallback(t *testing.T) {
	p1 := &MockProvider{
		NameFunc: func() string { return "p1" },
		CompleteFunc: func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
			return nil, provider.ErrProviderDown
		},
	}
	p2 := &MockProvider{
		NameFunc: func() string { return "p2" },
		CompleteFunc: func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
			return &provider.CompletionResponse{Choices: []provider.Choice{{Message: provider.Message{Content: "ok"}}}}, nil
		},
	}
	
	models := []config.ModelAliasConfig{{Alias: "my-model", Providers: []string{"p1", "p2"}, FallbackMode: "reliable"}}
	gw, _ := setupGateway(t, []provider.Provider{p1, p2}, models)

	reqBody, _ := json.Marshal(provider.CompletionRequest{Model: "my-model"})
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	gw.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGateway_RateLimit(t *testing.T) {
	models := []config.ModelAliasConfig{{Alias: "my-model", Providers: []string{"mock"}}}
	
	tmpDB := t.TempDir() + "/tendr.db"
	st, _ := store.New(tmpDB)
	
	cfg := &config.Config{Models: models}
	provMap := map[string]provider.Provider{"mock": &MockProvider{
		NameFunc: func() string { return "mock" },
		CompleteFunc: func(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
			return &provider.CompletionResponse{Choices: []provider.Choice{{Message: provider.Message{Content: "ok"}}}}, nil
		},
	}}

	r := router.NewRouter(cfg, provMap, nil)
	c := cache.NewExact(1 * time.Minute)
	
	// Limiter burst=1
	limiters := map[string]*ratelimit.Limiter{"my-model": ratelimit.NewLimiter(1000, 1)}
	
	gw := NewServer(0, r, c, st, limiters, &config.ServerConfig{Port: 0, LogLevel: "debug"})

	reqBody, _ := json.Marshal(provider.CompletionRequest{Model: "my-model"})
	
	// 1st request (ok)
	w1 := httptest.NewRecorder()
	gw.router.ServeHTTP(w1, httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody)))
	
	// 2nd request (rate limited)
	w2 := httptest.NewRecorder()
	gw.router.ServeHTTP(w2, httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody)))

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w2.Code)
	}
}

func TestGateway_InvalidRequestBody(t *testing.T) {
	gw, _ := setupGateway(t, nil, nil)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer([]byte("{invalid")))
	w := httptest.NewRecorder()
	gw.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRateLimit(t *testing.T) {
	c := cache.NewExact(5 * time.Minute)
	s, _ := store.New(":memory:")
	limiters := map[string]*ratelimit.Limiter{
		"test-model": ratelimit.NewLimiter(0.1, 1),
	}
	r := router.NewRouter(&config.Config{}, nil, nil)
	srv := NewServer(0, r, c, s, limiters, &config.ServerConfig{Port: 0, LogLevel: "debug"})

	reqBody, _ := json.Marshal(provider.CompletionRequest{Model: "test-model"})
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
		srv.router.ServeHTTP(rec, req)

		if i > 0 && rec.Code == http.StatusTooManyRequests {
			return
		}
	}
	t.Error("expected rate limit")
}
