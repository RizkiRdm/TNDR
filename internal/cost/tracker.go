package cost

import (
	"context"
	"fmt"
	"time"

	"github.com/RizkiRdm/TNDR/internal/store"
	"github.com/google/uuid"
)

type Tracker struct {
	store   *store.Store
	pricing *PricingManager
}

func NewTracker(s *store.Store, pm *PricingManager) *Tracker {
	return &Tracker{store: s, pricing: pm}
}

func (t *Tracker) Track(ctx context.Context, provider, model string, prompt, completion int) error {
	rate, source := t.pricing.GetRate(provider, model)
	if source == "unknown" {
		return fmt.Errorf("no pricing found for %s/%s", provider, model)
	}

	cost := (float64(prompt) * rate.Prompt) + (float64(completion) * rate.Completion)

	return t.store.RecordRequest(ctx, &store.RequestRecord{
		ID:               uuid.New().String(),
		Model:            model,
		Provider:         provider,
		PromptTokens:     prompt,
		CompletionTokens: completion,
		TotalTokens:      prompt + completion,
		Cost:             cost,
		PricingSource:    source,
		CreatedAt:        time.Now().Format(time.RFC3339),
	})
}
