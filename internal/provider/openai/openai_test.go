package openai

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

func TestOpenAIProvider_Complete_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-3.5-turbo",
			"choices": [{
				"message": {"role": "assistant", "content": "Hello!"},
				"finish_reason": "stop",
				"index": 0
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 5,
				"total_tokens": 15
			}
		}`))
	}))
	defer ts.Close()

	p := NewOpenAIProvider("test-key", 30000)
	p.baseURL = ts.URL

	req := &provider.CompletionRequest{Model: "gpt-3.5-turbo"}
	resp, err := p.Complete(context.Background(), req)

	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if len(resp.Choices) == 0 {
		t.Error("expected choices, got none")
	}
	if resp.Usage.TotalTokens <= 0 {
		t.Errorf("expected total tokens > 0, got %d", resp.Usage.TotalTokens)
	}
	if resp.Provider != "openai" {
		t.Errorf("expected provider 'openai', got %s", resp.Provider)
	}
}

func TestOpenAIProvider_Complete_RateLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	p := NewOpenAIProvider("test-key", 30000)
	p.baseURL = ts.URL

	_, err := p.Complete(context.Background(), &provider.CompletionRequest{})
	if !errors.Is(err, provider.ErrRateLimit) {
		t.Errorf("expected ErrRateLimit, got %v", err)
	}
}

func TestOpenAIProvider_Complete_InvalidKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	p := NewOpenAIProvider("test-key", 30000)
	p.baseURL = ts.URL

	_, err := p.Complete(context.Background(), &provider.CompletionRequest{})
	if !errors.Is(err, provider.ErrInvalidKey) {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}

func TestOpenAIProvider_Complete_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	p := NewOpenAIProvider("test-key", 30000)
	p.baseURL = ts.URL

	_, err := p.Complete(context.Background(), &provider.CompletionRequest{})
	if err == nil {
		t.Error("expected error for 500, got nil")
	}
}

func TestOpenAIProvider_Stream_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"id\":\"1\",\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\ndata: [DONE]\n\n"))
	}))
	defer ts.Close()

	p := NewOpenAIProvider("test-key", 30000)
	p.baseURL = ts.URL

	respChan, errChan := p.Stream(context.Background(), &provider.CompletionRequest{})

	select {
	case resp := <-respChan:
		if resp == nil {
			t.Error("expected stream response, got nil")
		}
	case err := <-errChan:
		t.Fatalf("unexpected error: %v", err)
	}
}
