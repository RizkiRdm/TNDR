package store

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestStore_RecordRequest_Success(t *testing.T) {
	s := setupTestStore(t)
	req := &RequestRecord{
		ID:        "1",
		Provider:  "openai",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	err := s.RecordRequest(context.Background(), req)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestStore_RecordRequest_DuplicateID(t *testing.T) {
	s := setupTestStore(t)
	req := &RequestRecord{
		ID:        "1",
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	_ = s.RecordRequest(context.Background(), req)
	err := s.RecordRequest(context.Background(), req)
	if err == nil {
		t.Error("expected primary key error for duplicate ID, got nil")
	}
}

func TestStore_GetCostSummary_Empty(t *testing.T) {
	s := setupTestStore(t)
	sum, err := s.GetCostSummary(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Today != 0.0 || sum.Week != 0.0 || sum.Month != 0.0 || sum.AllTime != 0.0 {
		t.Errorf("expected all zeros, got %+v", sum)
	}
}

func TestStore_GetCostSummary_WithData(t *testing.T) {
	s := setupTestStore(t)
	costs := []float64{0.01, 0.02, 0.03}
	for i, c := range costs {
		_ = s.RecordRequest(context.Background(), &RequestRecord{
			ID:        fmt.Sprintf("%d", i),
			Cost:      c,
			CreatedAt: time.Now().Format(time.RFC3339),
		})
	}
	sum, _ := s.GetCostSummary(context.Background(), "")
	const tolerance = 1e-9
	if diff := sum.AllTime - 0.06; diff > tolerance || diff < -tolerance {
		t.Errorf("expected ~0.06, got %f", sum.AllTime)
	}
}

func TestStore_GetCostSummary_ProviderFilter(t *testing.T) {
	s := setupTestStore(t)
	_ = s.RecordRequest(context.Background(), &RequestRecord{ID: "1", Provider: "openai", Cost: 0.1, CreatedAt: time.Now().Format(time.RFC3339)})
	_ = s.RecordRequest(context.Background(), &RequestRecord{ID: "2", Provider: "anthropic", Cost: 0.2, CreatedAt: time.Now().Format(time.RFC3339)})
	
	sum, _ := s.GetCostSummary(context.Background(), "openai")
	if sum.AllTime != 0.1 {
		t.Errorf("expected 0.1, got %f", sum.AllTime)
	}
}

func TestStore_GetCacheStats(t *testing.T) {
	s := setupTestStore(t)
	_ = s.RecordRequest(context.Background(), &RequestRecord{ID: "1", CreatedAt: time.Now().Format(time.RFC3339)})
	_ = s.RecordRequest(context.Background(), &RequestRecord{ID: "2", CreatedAt: time.Now().Format(time.RFC3339)})
	_ = s.RecordCacheHit(context.Background(), "key")
	
	total, hits, _ := s.GetCacheStats(context.Background())
	if total != 2 || hits != 1 {
		t.Errorf("expected 2 total and 1 hit, got %d and %d", total, hits)
	}
}

func TestStore_ClearCache(t *testing.T) {
	s := setupTestStore(t)
	_ = s.RecordCacheHit(context.Background(), "key")
	_ = s.ClearCache(context.Background())
	_, hits, _ := s.GetCacheStats(context.Background())
	if hits != 0 {
		t.Errorf("expected 0 hits after clear, got %d", hits)
	}
}

func TestStore_GetRecentRequests(t *testing.T) {
	s := setupTestStore(t)
	for i := 0; i < 15; i++ {
		_ = s.RecordRequest(context.Background(), &RequestRecord{
			ID:        fmt.Sprintf("%d", i),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second).Format(time.RFC3339),
		})
	}
	recs, _ := s.GetRecentRequests(context.Background(), 10)
	if len(recs) != 10 {
		t.Errorf("expected 10 records, got %d", len(recs))
	}
	// Check sorting: descending by created_at. Since created_at was added sequentially,
	// newest should be first.
	// Last ID inserted was "14" (newest)
	if recs[0].ID != "14" {
		t.Errorf("expected newest ID 14, got %s", recs[0].ID)
	}
}

func TestStore_GetCostByProvider(t *testing.T) {
	s := setupTestStore(t)

	// Insert records for two different providers
	_ = s.RecordRequest(context.Background(), &RequestRecord{
		ID: "1", Provider: "openai", Cost: 0.10, CreatedAt: time.Now().Format(time.RFC3339),
	})
	_ = s.RecordRequest(context.Background(), &RequestRecord{
		ID: "2", Provider: "openai", Cost: 0.05, CreatedAt: time.Now().Format(time.RFC3339),
	})
	_ = s.RecordRequest(context.Background(), &RequestRecord{
		ID: "3", Provider: "anthropic", Cost: 0.20, CreatedAt: time.Now().Format(time.RFC3339),
	})

	result, err := s.GetCostByProvider(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const tol = 1e-9
	openaiCost := result["openai"]
	if diff := openaiCost - 0.15; diff > tol || diff < -tol {
		t.Errorf("expected openai cost ~0.15, got %f", openaiCost)
	}

	anthropicCost := result["anthropic"]
	if diff := anthropicCost - 0.20; diff > tol || diff < -tol {
		t.Errorf("expected anthropic cost ~0.20, got %f", anthropicCost)
	}

	// Provider not in DB should not exist in result map
	if _, ok := result["gemini"]; ok {
		t.Error("expected gemini to not be in result, got entry")
	}
}

func TestStore_GetCostByProviderWithTokens(t *testing.T) {
	s := setupTestStore(t)

	_ = s.RecordRequest(context.Background(), &RequestRecord{
		ID: "1", Provider: "openai", Model: "gpt-4o",
		Cost: 0.10, PromptTokens: 1000, CompletionTokens: 500, TotalTokens: 1500,
		CreatedAt: time.Now().Format(time.RFC3339),
	})
	_ = s.RecordRequest(context.Background(), &RequestRecord{
		ID: "2", Provider: "openai", Model: "gpt-4o-mini",
		Cost: 0.05, PromptTokens: 2000, CompletionTokens: 1000, TotalTokens: 3000,
		CreatedAt: time.Now().Format(time.RFC3339),
	})
	_ = s.RecordRequest(context.Background(), &RequestRecord{
		ID: "3", Provider: "anthropic", Model: "claude-3-5-sonnet-20240620",
		Cost: 0.20, PromptTokens: 500, CompletionTokens: 250, TotalTokens: 750,
		CreatedAt: time.Now().Format(time.RFC3339),
	})

	results, err := s.GetCostByProviderWithTokens(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(results))
	}

	const tol = 1e-9
	for _, r := range results {
		switch r.Provider {
		case "openai":
			if diff := r.Cost - 0.15; diff > tol || diff < -tol {
				t.Errorf("expected openai cost 0.15, got %f", r.Cost)
			}
			if r.PromptTokens != 3000 {
				t.Errorf("expected openai prompt_tokens 3000, got %d", r.PromptTokens)
			}
			if r.CompletionTokens != 1500 {
				t.Errorf("expected openai completion_tokens 1500, got %d", r.CompletionTokens)
			}
		case "anthropic":
			if diff := r.Cost - 0.20; diff > tol || diff < -tol {
				t.Errorf("expected anthropic cost 0.20, got %f", r.Cost)
			}
			if r.PromptTokens != 500 {
				t.Errorf("expected anthropic prompt_tokens 500, got %d", r.PromptTokens)
			}
			if r.CompletionTokens != 250 {
				t.Errorf("expected anthropic completion_tokens 250, got %d", r.CompletionTokens)
			}
		default:
			t.Errorf("unexpected provider: %s", r.Provider)
		}
	}
}

