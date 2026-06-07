package router

import (
	"context"
	"testing"
	"time"

	"github.com/RizkiRdm/TNDR/internal/config"
	"github.com/RizkiRdm/TNDR/internal/provider"
)

type mockProvider struct {
	name   string
	err    error
	delay  time.Duration
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.err != nil {
		return nil, m.err
	}
	return &provider.CompletionResponse{Model: m.name}, nil
}
func (m *mockProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	return nil, nil
}

func TestRouter_Fallback(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelAliasConfig{
			{Alias: "fallback-alias", FallbackMode: "reliable", Providers: []string{"fail", "success"}},
		},
	}
	providers := map[string]provider.Provider{
		"fail":    &mockProvider{name: "fail", err: provider.ErrProviderDown},
		"success": &mockProvider{name: "success"},
	}

	r := NewRouter(cfg, providers, nil)
	resp, err := r.Complete(context.Background(), "fallback-alias", &provider.CompletionRequest{})

	if err != nil {
		t.Fatalf("router fail: %v", err)
	}
	if resp.Model != "success" {
		t.Errorf("wrong provider: got %s, want success", resp.Model)
	}
}

func TestRouter_FastModeFallback(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelAliasConfig{
			{Alias: "fast-alias", FallbackMode: "fast", Providers: []string{"slow", "fast"}},
		},
	}
	providers := map[string]provider.Provider{
		"slow": &mockProvider{name: "slow", delay: 600 * time.Millisecond},
		"fast": &mockProvider{name: "fast"},
	}

	r := NewRouter(cfg, providers, nil)
	resp, err := r.Complete(context.Background(), "fast-alias", &provider.CompletionRequest{})

	if err != nil {
		t.Fatalf("router fail: %v", err)
	}
	if resp.Model != "fast" {
		t.Errorf("wrong provider: got %s, want fast", resp.Model)
	}
}
