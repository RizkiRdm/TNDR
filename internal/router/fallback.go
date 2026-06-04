package router

import (
	"context"
	"errors"
	"time"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

type FallbackMode string

const (
	ModeReliable FallbackMode = "reliable"
	ModeFast     FallbackMode = "fast"
	ModeSmart    FallbackMode = "smart"
)

// Fallback handles provider chain execution based on mode
type Fallback struct {
	providers []provider.Provider
	mode      FallbackMode
}

func NewFallback(mode FallbackMode, providers []provider.Provider) *Fallback {
	return &Fallback{
		mode:      mode,
		providers: providers,
	}
}

// Execute tries providers sequentially.
func (f *Fallback) Execute(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	var lastErr error

	for _, p := range f.providers {
		start := time.Now()
		resp, err := p.Complete(ctx, req)
		latency := time.Since(start)

		if err == nil {
			if f.mode == ModeFast && latency > 500*time.Millisecond && len(f.providers) > 1 {
				// Fast mode: if too slow, try next if available
				// This logic is simple: if slow, continue loop
				continue
			}
			return resp, nil
		}

		lastErr = err

		if f.mode == ModeReliable {
			if !isHardError(err) {
				return nil, err
			}
		}
	}

	return nil, lastErr
}

func isHardError(err error) bool {
	return errors.Is(err, provider.ErrProviderDown) ||
		errors.Is(err, provider.ErrInvalidKey) ||
		errors.Is(err, provider.ErrRateLimit)
}
