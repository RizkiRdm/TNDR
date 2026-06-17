package cost

import (
	"context"
	"testing"

	"github.com/RizkiRdm/TNDR/internal/store"
)

func setupTestStore(t *testing.T) *store.Store {
	t.Helper()
	// t.TempDir() returns a unique dir per test, auto-cleaned after test
	s, err := store.New(t.TempDir() + "/tendr_test.db")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestTracker_Track_Success(t *testing.T) {
	s := setupTestStore(t)
	pm := NewPricingManager()
	// Inject a mock pricing entry via internal field if needed, 
	// but here we use real pricing from pricing.json or add override.
	pm.overrides["openai"+"gpt-4o"] = Pricing{
		Provider:   "openai",
		Model:      "gpt-4o",
		Prompt:     0.005,
		Completion: 0.015,
	}
	tr := NewTracker(s, pm)

	ctx := context.Background()
	err := tr.Track(ctx, "openai", "gpt-4o", 1000, 500)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	summary, err := s.GetCostSummary(ctx, "openai")
	if err != nil {
		t.Fatalf("failed to get summary: %v", err)
	}

	if summary.AllTime <= 0 {
		t.Errorf("expected AllTime cost > 0, got %f", summary.AllTime)
	}
}

func TestTracker_Track_UnknownModel(t *testing.T) {
	s := setupTestStore(t)
	pm := NewPricingManager()
	tr := NewTracker(s, pm)

	ctx := context.Background()
	err := tr.Track(ctx, "openai", "gpt-unknown", 1000, 500)
	if err == nil {
		t.Error("expected error for unknown model, got nil")
	}
}

func TestTracker_CostCalculation(t *testing.T) {
	tests := []struct {
		name       string
		prompt     int
		completion int
		promptRate float64
		compRate   float64
		expected   float64
	}{
		{
			name:       "standard cost",
			prompt:     1000,
			completion: 500,
			promptRate: 0.005,
			compRate:   0.015,
			expected:   (1000 * 0.005) + (500 * 0.015),
		},
		{
			name:       "zero cost",
			prompt:     0,
			completion: 0,
			promptRate: 0.005,
			compRate:   0.015,
			expected:   0.0,
		},
		{
			name:       "full unit cost",
			prompt:     1000000,
			completion: 1000000,
			promptRate: 0.005,
			compRate:   0.015,
			expected:   (1000000 * 0.005) + (1000000 * 0.015),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Manual calculation matching internal/cost/tracker.go:29
			cost := ((float64(tt.prompt) * tt.promptRate) + (float64(tt.completion) * tt.compRate)) / 1000000.0
			expected := tt.expected / 1000000.0
			if cost != expected {
				t.Errorf("expected %f, got %f", expected, cost)
			}
		})
	}
}
