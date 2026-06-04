package router

import (
	"context"
	"fmt"

	"github.com/RizkiRdm/TNDR/internal/config"
	"github.com/RizkiRdm/TNDR/internal/provider"
)

type Router struct {
	providers map[string]provider.Provider
	models    map[string]config.ModelAliasConfig
}

func NewRouter(cfg *config.Config, providers map[string]provider.Provider) *Router {
	modelMap := make(map[string]config.ModelAliasConfig)
	for _, m := range cfg.Models {
		modelMap[m.Alias] = m
	}

	return &Router{
		providers: providers,
		models:    modelMap,
	}
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

	fb := NewFallback(FallbackMode(modelCfg.FallbackMode), pList)
	return fb.Execute(ctx, req)
}

func (r *Router) Stream(ctx context.Context, modelAlias string, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	// Stream fallback not yet implemented
	modelCfg, ok := r.models[modelAlias]
	if !ok {
		errChan := make(chan error, 1)
		errChan <- fmt.Errorf("model alias not found: %s", modelAlias)
		respChan := make(chan *provider.StreamResponse)
		close(respChan)
		return respChan, errChan
	}

	providerName := modelCfg.Providers[0]
	p, ok := r.providers[providerName]
	if !ok {
		errChan := make(chan error, 1)
		errChan <- fmt.Errorf("provider not found: %s", providerName)
		respChan := make(chan *provider.StreamResponse)
		close(respChan)
		return respChan, errChan
	}

	return p.Stream(ctx, req)
}
