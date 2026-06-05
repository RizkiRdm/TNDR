package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

type FallbackMode string

const (
	ModeReliable FallbackMode = "reliable"
	ModeFast     FallbackMode = "fast"
	ModeSmart    FallbackMode = "smart"
)

type ProviderError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider error: %s: %v", e.Message, e.Err)
}

// Fallback handles provider chain execution based on mode
type Fallback struct {
	providers        []provider.Provider
	mode             FallbackMode
	latencyThreshold time.Duration
}

func NewFallback(mode FallbackMode, providers []provider.Provider, latencyThreshold time.Duration) *Fallback {
	return &Fallback{
		mode:             mode,
		providers:        providers,
		latencyThreshold: latencyThreshold,
	}
}

// Execute tries providers sequentially.
func (f *Fallback) Execute(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	var lastErr error

	for i, p := range f.providers {
		start := time.Now()
		resp, err := p.Complete(ctx, req)
		latency := time.Since(start)

		if err == nil {
			isLast := i == len(f.providers)-1
			
			// Fast mode: trigger fallback if slow, unless it's the last provider
			if f.mode == ModeFast && latency > f.latencyThreshold && !isLast {
				continue
			}
			
			// Smart mode: trigger fallback if slow OR hard error, unless last
			if f.mode == ModeSmart && latency > f.latencyThreshold && !isLast {
				continue
			}

			return resp, nil
		}

		lastErr = err

		// Reliable and Smart modes: trigger fallback on hard errors
		if f.mode == ModeReliable || f.mode == ModeSmart {
			if !isHardError(err) {
				return nil, err
			}
		}
	}

	return nil, &ProviderError{
		StatusCode: http.StatusBadGateway,
		Message:    "all providers exhausted",
		Err:        lastErr,
	}
}

func isHardError(err error) bool {
	return errors.Is(err, provider.ErrProviderDown) ||
		errors.Is(err, provider.ErrInvalidKey) ||
		errors.Is(err, provider.ErrRateLimit)
}
