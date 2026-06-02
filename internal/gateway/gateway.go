package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RizkiRdm/TNDR/internal/provider"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type Server struct {
	httpServer *http.Server
	router     *chi.Mux
	provider   provider.Provider
}

func NewServer(port int, p provider.Provider) *Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s := &Server{
		router:   r,
		provider: p,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: r,
		},
	}

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	s.router.Route("/v1", func(r chi.Router) {
		r.Post("/chat/completions", s.handleChatCompletions)
	})
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req provider.CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := s.provider.Complete(r.Context(), &req)
	if err != nil {
		log.Error().Err(err).Msg("provider completion failed")
		http.Error(w, "provider failure", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) Start() error {
	log.Info().Str("addr", s.httpServer.Addr).Msg("starting gateway server")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info().Msg("stopping gateway server")
	return s.httpServer.Shutdown(ctx)
}
