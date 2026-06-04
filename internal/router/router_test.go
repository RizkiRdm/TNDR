package router

import (
	"context"
	"testing"

	"github.com/RizkiRdm/TNDR/internal/config"
	"github.com/RizkiRdm/TNDR/internal/provider"
)

type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	return &provider.CompletionResponse{Model: m.name}, nil
}
func (m *mockProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	return nil, nil
}

func TestRouter_Select(t *testing.T) {
	cfg := &config.Config{
		Models: []config.ModelAliasConfig{
			{Alias: "test-alias", Providers: []string{"mock1"}},
		},
	}
	providers := map[string]provider.Provider{
		"mock1": &mockProvider{name: "mock1"},
	}

	r := NewRouter(cfg, providers)
	resp, err := r.Complete(context.Background(), "test-alias", &provider.CompletionRequest{})

	if err != nil {
		t.Fatalf("router fail: %v", err)
	}
	if resp.Model != "mock1" {
		t.Errorf("wrong provider: got %s, want mock1", resp.Model)
	}
}
