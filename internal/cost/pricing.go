package cost

import (
	_ "embed"
	"encoding/json"
	"sync"
)

//go:embed pricing.json
var embeddedPricing []byte

type Pricing struct {
	Provider   string  `json:"provider"`
	Model      string  `json:"model"`
	Prompt     float64 `json:"prompt_rate"`
	Completion float64 `json:"completion_rate"`
}

type PricingManager struct {
	mu        sync.RWMutex
	rates     map[string]Pricing
	overrides map[string]Pricing
}

func NewPricingManager() *PricingManager {
	pm := &PricingManager{
		rates:     make(map[string]Pricing),
		overrides: make(map[string]Pricing),
	}
	pm.loadEmbedded()
	return pm
}

func (pm *PricingManager) loadEmbedded() {
	var data []Pricing
	if err := json.Unmarshal(embeddedPricing, &data); err == nil {
		for _, p := range data {
			pm.rates[p.Provider+p.Model] = p
		}
	}
}

func (pm *PricingManager) GetRate(provider, model string) (Pricing, string) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if p, ok := pm.overrides[provider+model]; ok {
		return p, "override"
	}
	if p, ok := pm.rates[provider+model]; ok {
		return p, "remote"
	}
	return Pricing{}, "unknown"
}
