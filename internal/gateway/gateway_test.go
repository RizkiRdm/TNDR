package gateway

import (
	"bytes"
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

func TestRateLimit(t *testing.T) {
	// Setup
	c := cache.NewExact(5 * time.Minute)
	s, _ := store.New(":memory:")
	limiters := map[string]*ratelimit.Limiter{
		"test-model": ratelimit.NewLimiter(0.1, 1), // 0.1 rps, 1 burst
	}
	r := router.NewRouter(&config.Config{}, nil, nil)
	srv := NewServer(0, r, c, s, limiters)

	// Trigger limit
	reqBody, _ := json.Marshal(provider.CompletionRequest{Model: "test-model"})
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(reqBody))
		srv.router.ServeHTTP(rec, req)
		
		if i > 0 && rec.Code == http.StatusTooManyRequests {
			return // Success
		}
	}
	t.Error("expected rate limit")
}
