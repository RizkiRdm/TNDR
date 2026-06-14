package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RizkiRdm/TNDR/internal/config"
)

// nextHandler is a simple handler that writes 200 OK for testing middleware passthrough.
func nextHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func TestGlobalKeyMiddleware_NoKeyConfigured_Passthrough(t *testing.T) {
	// If global_key is empty string, ALL requests pass through without auth check.
	cfg := &config.ServerConfig{GlobalKey: ""}
	mw := GlobalKeyMiddleware(cfg)(http.HandlerFunc(nextHandler))

	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	// No Authorization header set — should still pass
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 when no global_key configured, got %d", w.Code)
	}
}

func TestGlobalKeyMiddleware_ValidBearerToken(t *testing.T) {
	cfg := &config.ServerConfig{GlobalKey: "tndr-secret-key"}
	mw := GlobalKeyMiddleware(cfg)(http.HandlerFunc(nextHandler))

	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer tndr-secret-key")
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with valid Bearer token, got %d", w.Code)
	}
}

func TestGlobalKeyMiddleware_ValidXAPIKey(t *testing.T) {
	cfg := &config.ServerConfig{GlobalKey: "tndr-secret-key"}
	mw := GlobalKeyMiddleware(cfg)(http.HandlerFunc(nextHandler))

	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("X-API-Key", "tndr-secret-key")
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with valid X-API-Key, got %d", w.Code)
	}
}

func TestGlobalKeyMiddleware_WrongKey(t *testing.T) {
	cfg := &config.ServerConfig{GlobalKey: "tndr-secret-key"}
	mw := GlobalKeyMiddleware(cfg)(http.HandlerFunc(nextHandler))

	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong key, got %d", w.Code)
	}
}

func TestGlobalKeyMiddleware_NoKeyProvided(t *testing.T) {
	// global_key is SET but request has NO auth header
	cfg := &config.ServerConfig{GlobalKey: "tndr-secret-key"}
	mw := GlobalKeyMiddleware(cfg)(http.HandlerFunc(nextHandler))

	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	// Intentionally no Authorization or X-API-Key header
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 when no key provided, got %d", w.Code)
	}
}

func TestGlobalKeyMiddleware_ResponseBodyOnUnauthorized(t *testing.T) {
	cfg := &config.ServerConfig{GlobalKey: "tndr-secret-key"}
	mw := GlobalKeyMiddleware(cfg)(http.HandlerFunc(nextHandler))

	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer bad-key")
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	body := w.Body.String()
	if body == "" {
		t.Error("expected non-empty body on 401 response")
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json Content-Type, got %s", w.Header().Get("Content-Type"))
	}
}
