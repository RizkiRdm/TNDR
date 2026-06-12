package gateway

import (
	"crypto/subtle"
	"net/http"

	"github.com/RizkiRdm/TNDR/internal/config"
)

// GlobalKeyMiddleware validates the X-API-Key or Authorization Bearer header
// against the configured global_key. If global_key is empty, skips auth.
func GlobalKeyMiddleware(cfg *config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.GlobalKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Support both OpenAI-style Bearer and direct key
			key := r.Header.Get("Authorization")
			if len(key) > 7 && key[:7] == "Bearer " {
				key = key[7:]
			}
			if key == "" {
				key = r.Header.Get("X-API-Key")
			}

			if key == "" || subtle.ConstantTimeCompare([]byte(key), []byte(cfg.GlobalKey)) != 1 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":{"code":"invalid_key","message":"Invalid API key"}}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
