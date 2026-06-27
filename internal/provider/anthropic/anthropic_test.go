package anthropic

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

func TestAnthropicProvider_Complete_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "msg-123",
			"role": "assistant",
			"content": [{"text": "Hello!"}],
			"model": "claude-3-opus",
			"stop_reason": "end_turn",
			"usage": {
				"input_tokens": 10,
				"output_tokens": 5
			}
		}`))
	}))
	defer ts.Close()

	p := NewAnthropicProvider("test-key", 30000)
	p.baseURL = ts.URL

	req := &provider.CompletionRequest{Model: "claude-3-opus"}
	resp, err := p.Complete(context.Background(), req)

	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if resp.Choices[0].Message.Content != "Hello!" {
		t.Errorf("expected Hello!, got %s", resp.Choices[0].Message.Content)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("expected 15 tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestAnthropicProvider_Complete_RateLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	p := NewAnthropicProvider("test-key", 30000)
	p.baseURL = ts.URL

	_, err := p.Complete(context.Background(), &provider.CompletionRequest{})
	if !errors.Is(err, provider.ErrRateLimit) {
		t.Errorf("expected ErrRateLimit, got %v", err)
	}
}
