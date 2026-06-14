package cost

import (
	"testing"
)

func TestPricingManager_LoadEmbedded(t *testing.T) {
	pm := NewPricingManager()
	if len(pm.rates) == 0 {
		t.Error("expected rates map to be populated from embedded data, got empty")
	}

	_, source := pm.GetRate("openai", "gpt-4o")
	if source == "unknown" {
		t.Error("expected gpt-4o pricing to be found in embedded data")
	}
}

func TestPricingManager_GetRate_Override(t *testing.T) {
	pm := NewPricingManager()
	override := Pricing{
		Provider:   "openai",
		Model:      "gpt-4o",
		Prompt:     1.0,
		Completion: 2.0,
	}
	pm.overrides["openai"+"gpt-4o"] = override

	p, source := pm.GetRate("openai", "gpt-4o")
	if source != "override" {
		t.Errorf("expected source 'override', got '%s'", source)
	}
	if p.Prompt != 1.0 {
		t.Errorf("expected prompt rate 1.0, got %f", p.Prompt)
	}
}

func TestPricingManager_GetRate_Unknown(t *testing.T) {
	pm := NewPricingManager()
	_, source := pm.GetRate("fakeprovider", "fakemodel")
	if source != "unknown" {
		t.Errorf("expected source 'unknown', got '%s'", source)
	}
}
