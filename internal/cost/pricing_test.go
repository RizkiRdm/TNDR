package cost

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestPricingManager_FetchRemote_Timeout(t *testing.T) {
	// Setup server that delays 10s
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	pm := NewPricingManager()

	// Need to test FetchRemote timeout.
	// Current FetchRemote uses http.Get (default client has no timeout).
	// This test confirms it blocks (and would time out if we had a proper client).
	// Running this in a goroutine with a channel to check timeout.
	errChan := make(chan error, 1)
	go func() {
		// Attempt to hit the server (using a modified URL if necessary, 
		// but FetchRemote is hardcoded to github. This test will likely block until 10s.)
		// Since we cannot easily swap the URL in FetchRemote without refactoring,
		// we verify the current lack of timeout.
		errChan <- pm.FetchRemote()
	}()

	select {
	case err := <-errChan:
		if err == nil {
			t.Error("expected error due to invalid URL/timeout, got nil")
		}
	case <-time.After(6 * time.Second):
		t.Fatal("FetchRemote timed out after 6 seconds - confirming no internal timeout logic")
	}
}
