package gemini

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

func TestGeminiProvider_Complete_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"candidates": [{
				"content": {
					"parts": [{"text": "Hello!"}],
					"role": "model"
				},
				"finishReason": "STOP"
			}],
			"usageMetadata": {
				"promptTokenCount": 10,
				"candidatesTokenCount": 5,
				"totalTokenCount": 15
			}
		}`))
	}))
	defer ts.Close()

	p := NewGeminiProvider("test-key")
	p.baseURL = ts.URL

	req := &provider.CompletionRequest{Model: "gemini-pro"}
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
