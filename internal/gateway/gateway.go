package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/RizkiRdm/TNDR/internal/cache"
	"github.com/RizkiRdm/TNDR/internal/config"
	"github.com/RizkiRdm/TNDR/internal/provider"
	"github.com/RizkiRdm/TNDR/internal/ratelimit"
	"github.com/RizkiRdm/TNDR/internal/router"
	"github.com/RizkiRdm/TNDR/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type Server struct {
	httpServer *http.Server
	router     *chi.Mux
	gwRouter   *router.Router
	cache      *cache.Exact
	store      *store.Store
	limiters   map[string]*ratelimit.Limiter
	config     *config.ServerConfig
	sem        chan struct{}
	globalRL   *ratelimit.Limiter
}

func NewServer(port int, r *router.Router, c *cache.Exact, st *store.Store, l map[string]*ratelimit.Limiter, cfg *config.ServerConfig) *Server {
	mux := chi.NewRouter()

	mux.Use(middleware.RequestID)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(GlobalKeyMiddleware(cfg))

	var sem chan struct{}
	if cfg.MaxConcurrentRequests > 0 {
		sem = make(chan struct{}, cfg.MaxConcurrentRequests)
	}

	var globalRL *ratelimit.Limiter
	if cfg.MaxRequestsPerSecond > 0 {
		globalRL = ratelimit.NewLimiter(float64(cfg.MaxRequestsPerSecond), float64(cfg.MaxRequestsPerSecond))
	}

	s := &Server{
		router:    mux,
		gwRouter:  r,
		cache:     c,
		store:     st,
		limiters:  l,
		config:    cfg,
		sem:       sem,
		globalRL:  globalRL,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	s.router.Route("/v1", func(r chi.Router) {
		r.Post("/chat/completions", s.handleChatCompletions)
	})
	s.router.Get("/health", s.handleHealth)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	reqID := middleware.GetReqID(r.Context())

	// Limit request body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	var req provider.CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Str("request_id", reqID).Err(err).Msg("gateway.request.invalid_body")
		http.Error(w, "invalid request body or body too large", http.StatusBadRequest)
		return
	}

	// GW3: concurrent request semaphore
	if s.sem != nil {
		select {
		case s.sem <- struct{}{}:
			defer func() { <-s.sem }()
		default:
			log.Warn().Str("request_id", reqID).Str("model", req.Model).Msg("gateway.request.concurrency_limit")
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate_limit_exceeded", http.StatusTooManyRequests)
			return
		}
	}

	// GW3: global rate limiter
	if s.globalRL != nil && !s.globalRL.Allow() {
		log.Warn().Str("request_id", reqID).Str("model", req.Model).Msg("gateway.request.rate_limited_global")
		w.Header().Set("Retry-After", "1")
		http.Error(w, "rate_limit_exceeded", http.StatusTooManyRequests)
		return
	}

	// GW4: structured log - request start
	log.Info().Str("request_id", reqID).Str("model", req.Model).Bool("stream", req.Stream).Msg("gateway.request.start")

	// Per-model rate limit check
	if limiter, ok := s.limiters[req.Model]; ok {
		if !limiter.Allow() {
			log.Warn().Str("request_id", reqID).Str("model", req.Model).Msg("gateway.request.rate_limited_model")
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate_limit_exceeded", http.StatusTooManyRequests)
			return
		}
	}

	modelCfg, ok := s.gwRouter.GetModelConfig(req.Model)
	if ok {
		for _, p := range modelCfg.Providers {
			if limiter, ok := s.limiters[p]; ok {
				if !limiter.Allow() {
					log.Warn().Str("request_id", reqID).Str("provider", p).Msg("gateway.request.rate_limited_provider")
					w.Header().Set("Retry-After", "1")
					http.Error(w, "rate_limit_exceeded", http.StatusTooManyRequests)
					return
				}
			}
		}
	}

	if req.Stream {
		s.handleStreaming(w, r, &req)
		return
	}

	key, err := cache.HashKey(req)
	if err == nil {
		if val, ok := s.cache.Get(key); ok {
			log.Info().Str("request_id", reqID).Str("model", req.Model).Msg("gateway.request.cache_hit")
			go s.store.RecordCacheHit(context.Background(), key)
			go s.store.RecordRequest(context.Background(), &store.RequestRecord{
				ID:            reqID,
				Model:         req.Model,
				Cost:          0.0,
				PricingSource: "cache",
				CreatedAt:     time.Now().Format(time.RFC3339),
			})
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Tendr-Cache", "HIT")
			w.Write([]byte(val))
			return
		}
	}
	log.Info().Str("request_id", reqID).Str("model", req.Model).Msg("gateway.request.cache_miss")

	// In TENDR, we use the 'model' field as the alias
	resp, err := s.gwRouter.Complete(r.Context(), req.Model, &req)
	latency := time.Since(start)
	if err != nil {
		writeErrorResponse(w, reqID, req.Model, start, err)
		return
	}
	if err == nil {
		b, err := json.Marshal(resp)
		if err == nil {
			s.cache.Set(key, string(b))
		}
	}

	log.Info().
		Str("request_id", reqID).
		Str("model", req.Model).
		Str("provider", resp.Provider).
		Int64("latency_ms", latency.Milliseconds()).
		Int("prompt_tokens", resp.Usage.PromptTokens).
		Int("completion_tokens", resp.Usage.CompletionTokens).
		Msg("gateway.request.complete")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Tendr-Cache", "MISS")
	w.Header().Set("X-Tendr-Provider", resp.Provider)
	w.Header().Set("X-Tendr-Latency-Ms", fmt.Sprintf("%d", latency.Milliseconds()))
	json.NewEncoder(w).Encode(resp)
}

func writeErrorResponse(w http.ResponseWriter, reqID, model string, start time.Time, err error) {
	latency := time.Since(start)
	var statusCode int
	var errorCode string
	var attempts []router.Attempt

	switch {
	case errors.Is(err, provider.ErrRateLimit):
		statusCode = http.StatusTooManyRequests
		errorCode = "rate_limit_exceeded"
	case errors.Is(err, provider.ErrInvalidKey):
		statusCode = http.StatusUnauthorized
		errorCode = "invalid_key"
	case errors.Is(err, provider.ErrTimeout):
		statusCode = http.StatusGatewayTimeout
		errorCode = "provider_timeout"
	case errors.Is(err, provider.ErrProviderDown):
		statusCode = http.StatusBadGateway
		errorCode = "provider_unavailable"
	default:
		statusCode = http.StatusBadGateway
		errorCode = "provider_unavailable"
	}

	var pe *router.ProviderError
	if errors.As(err, &pe) {
		attempts = pe.Attempts
		// Fallback exhaustion with ≥2 attempts → infrastructure failure
		if len(attempts) >= 2 {
			statusCode = http.StatusBadGateway
			errorCode = "provider_unavailable"
		}
	}

	log.Error().
		Str("request_id", reqID).
		Str("model", model).
		Err(err).
		Int("status", statusCode).
		Str("error_code", errorCode).
		Int64("latency_ms", latency.Milliseconds()).
		Msg("gateway.request.error")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Tendr-Error-Code", errorCode)
	w.Header().Set("X-Tendr-Latency-Ms", fmt.Sprintf("%d", latency.Milliseconds()))
	w.WriteHeader(statusCode)

	resp := map[string]interface{}{
		"error":   errorCode,
		"message": "The requested AI provider is currently unavailable or returned an error.",
	}
	if len(attempts) > 0 {
		resp["attempts"] = attempts
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleStreaming(w http.ResponseWriter, r *http.Request, req *provider.CompletionRequest) {
	respChan, errChan := s.gwRouter.Stream(r.Context(), req.Model, req)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error().Msg("gateway.stream.flusher_unsupported")
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-r.Context().Done():
			log.Info().Str("request_id", middleware.GetReqID(r.Context())).Msg("gateway.stream.cancelled")
			return
		case err, ok := <-errChan:
			if ok && err != nil {
				errorCode := "provider_unavailable"
				switch {
				case errors.Is(err, provider.ErrRateLimit):
					errorCode = "rate_limit_exceeded"
				case errors.Is(err, provider.ErrInvalidKey):
					errorCode = "invalid_key"
				case errors.Is(err, provider.ErrTimeout):
					errorCode = "provider_timeout"
				}
				log.Error().Str("request_id", middleware.GetReqID(r.Context())).Err(err).Str("error_code", errorCode).Msg("gateway.stream.error")
				fmt.Fprintf(w, "event: error\ndata: {\"error\":%q}\n\n", errorCode)
				flusher.Flush()
			}
			return
		case resp, ok := <-respChan:
			if !ok {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				return
			}
			jsonData, err := json.Marshal(resp)
			if err != nil {
				log.Error().Err(err).Msg("gateway.stream.marshal_failed")
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()
		}
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) Start() error {
	log.Info().Str("addr", s.httpServer.Addr).Msg("starting gateway server")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info().Msg("stopping gateway server")
	return s.httpServer.Shutdown(ctx)
}
