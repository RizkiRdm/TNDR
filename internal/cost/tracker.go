package cost

import (
	"context"
	"fmt"
	"time"

	"github.com/RizkiRdm/TNDR/internal/config"
	"github.com/RizkiRdm/TNDR/internal/store"
	"github.com/google/uuid"
)

type Tracker struct {
	store   *store.Store
	pricing *PricingManager
	cfg     *config.ServerConfig
}

// NewTracker creates a new tracker to monitor request costs using the provided store and pricing manager.
func NewTracker(s *store.Store, pm *PricingManager, cfg *config.ServerConfig) *Tracker {
	return &Tracker{store: s, pricing: pm, cfg: cfg}
}

func (t *Tracker) Config() *config.ServerConfig {
	return t.cfg
}

// Track calculates and records the cost of a request based on prompt and completion token usage.
func (t *Tracker) Track(ctx context.Context, provider, model string, prompt, completion int) error {
	rate, source := t.pricing.GetRate(provider, model)
	if source == "unknown" {
		return fmt.Errorf("no pricing found for %s/%s", provider, model)
	}

	cost := ((float64(prompt) * rate.Prompt) + (float64(completion) * rate.Completion)) / 1000000.0

	return t.store.RecordRequest(ctx, &store.RequestRecord{
		ID:               uuid.New().String(),
		Model:            model,
		Provider:         provider,
		PromptTokens:     prompt,
		CompletionTokens: completion,
		TotalTokens:      prompt + completion,
		PromptRate:       rate.Prompt,
		CompletionRate:   rate.Completion,
		Cost:             cost,
		PricingSource:    source,
		CreatedAt:        time.Now().Format(time.RFC3339),
	})
}
