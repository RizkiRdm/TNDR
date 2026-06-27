package router

import (
	"context"
	"fmt"
	"time"

	"github.com/RizkiRdm/TNDR/internal/config"
	"github.com/RizkiRdm/TNDR/internal/cost"
	"github.com/RizkiRdm/TNDR/internal/provider"
	"github.com/rs/zerolog/log"
)

type Router struct {
	providers        map[string]provider.Provider
	models           map[string]config.ModelAliasConfig
	tracker          *cost.Tracker
	latencyThreshold time.Duration
}

func NewRouter(cfg *config.Config, providers map[string]provider.Provider, tracker *cost.Tracker) *Router {
	modelMap := make(map[string]config.ModelAliasConfig)
	for _, m := range cfg.Models {
		modelMap[m.Alias] = m
	}

	lt := time.Duration(cfg.Server.LatencyThresholdMs) * time.Millisecond
	if lt <= 0 {
		lt = 500 * time.Millisecond
	}

	return &Router{
		providers:        providers,
		models:           modelMap,
		tracker:          tracker,
		latencyThreshold: lt,
	}
}

func (r *Router) GetModelConfig(alias string) (config.ModelAliasConfig, bool) {
	cfg, ok := r.models[alias]
	return cfg, ok
}

func (r *Router) Complete(ctx context.Context, modelAlias string, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	modelCfg, ok := r.models[modelAlias]
	if !ok {
		return nil, fmt.Errorf("model alias not found: %s", modelAlias)
	}

	var pList []provider.Provider
	for _, name := range modelCfg.Providers {
		if p, ok := r.providers[name]; ok {
			pList = append(pList, p)
		}
	}

	if len(pList) == 0 {
		return nil, fmt.Errorf("no providers configured for alias: %s", modelAlias)
	}

	fb := NewFallback(FallbackMode(modelCfg.FallbackMode), pList, r.latencyThreshold)
	resp, err := fb.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	// Record cost asynchronously
	if r.tracker != nil {
		go func() {
			err := r.tracker.Track(context.Background(), resp.Provider, resp.Model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
			if err != nil {
				log.Error().Err(err).Str("provider", resp.Provider).Str("model", resp.Model).Msg("failed to track cost")
			}
		}()
	}
	return resp, nil
}

func (r *Router) Stream(ctx context.Context, modelAlias string, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	modelCfg, ok := r.models[modelAlias]
	if !ok {
		errChan := make(chan error, 1)
		errChan <- fmt.Errorf("model alias not found: %s", modelAlias)
		respChan := make(chan *provider.StreamResponse)
		close(respChan)
		return respChan, errChan
	}

	// Try providers in order
	for _, name := range modelCfg.Providers {
		p, ok := r.providers[name]
		if !ok {
			continue
		}
		// Stream from first available provider
		return p.Stream(ctx, req)
	}

	errChan := make(chan error, 1)
	errChan <- fmt.Errorf("router: no providers available for alias: %s", modelAlias)
	respChan := make(chan *provider.StreamResponse)
	close(respChan)
	return respChan, errChan
}
