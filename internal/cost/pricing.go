package cost

import (
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/RizkiRdm/TNDR/internal/config"
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

// LoadOverrides populates the overrides mapping from user configuration overrides.
func (pm *PricingManager) LoadOverrides(override map[string]map[string]config.ModelPricing) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for provider, models := range override {
		for model, p := range models {
			pm.overrides[provider+model] = Pricing{
				Provider:   provider,
				Model:      model,
				Prompt:     p.InputPer1m,
				Completion: p.OutputPer1m,
			}
		}
	}
}
