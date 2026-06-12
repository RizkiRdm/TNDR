package groq

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

func TestGroqProvider_Complete_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "chatcmpl-123",
			"choices": [{
				"message": {"role": "assistant", "content": "Hello!"},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 5,
				"total_tokens": 15
			}
		}`))
	}))
	defer ts.Close()

	p := NewGroqProvider("test-key")
	p.baseURL = ts.URL

	req := &provider.CompletionRequest{Model: "llama3"}
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

func TestGroqProvider_Complete_RateLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	p := NewGroqProvider("test-key")
	p.baseURL = ts.URL

	_, err := p.Complete(context.Background(), &provider.CompletionRequest{})
	if !errors.Is(err, provider.ErrRateLimit) {
		t.Errorf("expected ErrRateLimit, got %v", err)
	}
}
