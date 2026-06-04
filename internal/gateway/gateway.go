package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RizkiRdm/TNDR/internal/provider"
	"github.com/RizkiRdm/TNDR/internal/router"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type Server struct {
	httpServer *http.Server
	router     *chi.Mux
	gwRouter   *router.Router
}

func NewServer(port int, r *router.Router) *Server {
	mux := chi.NewRouter()

	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	s := &Server{
		router:   mux,
		gwRouter: r,
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
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req provider.CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Stream {
		s.handleStreaming(w, r, &req)
		return
	}

	// In TENDR, we use the 'model' field as the alias
	resp, err := s.gwRouter.Complete(r.Context(), req.Model, &req)
	if err != nil {
		log.Error().Err(err).Str("model", req.Model).Msg("routing failed")
		http.Error(w, fmt.Sprintf("gateway error: %v", err), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case err, ok := <-errChan:
			if ok && err != nil {
				log.Error().Err(err).Msg("stream error")
				fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
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
				log.Error().Err(err).Msg("marshal stream resp failed")
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()
		}
	}
}

func (s *Server) Start() error {
	log.Info().Str("addr", s.httpServer.Addr).Msg("starting gateway server")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info().Msg("stopping gateway server")
	return s.httpServer.Shutdown(ctx)
}
