package router

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

func TestFallback_ReliableMode_SkipsOnHardError(t *testing.T) {
	p1 := &mockProvider{err: provider.ErrProviderDown}
	p2 := &mockProvider{resp: &provider.CompletionResponse{Model: "success"}}
	f := NewFallback(ModeReliable, []provider.Provider{p1, p2}, 500*time.Millisecond)

	resp, err := f.Execute(context.Background(), &provider.CompletionRequest{})
	if err != nil || resp.Model != "success" {
		t.Errorf("expected success from p2, got %v, %v", resp, err)
	}
}

func TestFallback_ReliableMode_StopsOnSoftError(t *testing.T) {
	softErr := errors.New("soft error")
	p1 := &mockProvider{err: softErr}
	p2 := &mockProvider{resp: &provider.CompletionResponse{Model: "success"}}
	f := NewFallback(ModeReliable, []provider.Provider{p1, p2}, 500*time.Millisecond)

	_, err := f.Execute(context.Background(), &provider.CompletionRequest{})
	if !errors.Is(err, softErr) {
		t.Errorf("expected soft error, got %v", err)
	}
}

func TestFallback_FastMode_FallsBackOnLatency(t *testing.T) {
	p1 := &mockProvider{resp: &provider.CompletionResponse{Model: "slow"}, delay: 600 * time.Millisecond}
	p2Mock := &mockProvider{resp: &provider.CompletionResponse{Model: "success"}}
	f := NewFallback(ModeFast, []provider.Provider{p1, p2Mock}, 500*time.Millisecond)

	resp, err := f.Execute(context.Background(), &provider.CompletionRequest{})
	if err != nil || resp.Model != "success" {
		t.Errorf("expected success from p2, got %v, %v", resp, err)
	}
}

func TestFallback_FastMode_NoFallbackIfLastProvider(t *testing.T) {
	p1 := &mockProvider{resp: &provider.CompletionResponse{Model: "slow"}, delay: 600 * time.Millisecond}
	f := NewFallback(ModeFast, []provider.Provider{p1}, 500*time.Millisecond)

	resp, err := f.Execute(context.Background(), &provider.CompletionRequest{})
	if err != nil || resp.Model != "slow" {
		t.Errorf("expected success from last p1, got %v, %v", resp, err)
	}
}

func TestFallback_SmartMode_FallsBackOnError(t *testing.T) {
	p1 := &mockProvider{err: provider.ErrRateLimit}
	p2 := &mockProvider{resp: &provider.CompletionResponse{Model: "success"}}
	f := NewFallback(ModeSmart, []provider.Provider{p1, p2}, 500*time.Millisecond)

	resp, err := f.Execute(context.Background(), &provider.CompletionRequest{})
	if err != nil || resp.Model != "success" {
		t.Errorf("expected success from p2, got %v, %v", resp, err)
	}
}

func TestFallback_SmartMode_FallsBackOnLatency(t *testing.T) {
	p1 := &mockProvider{resp: &provider.CompletionResponse{Model: "slow"}, delay: 600 * time.Millisecond}
	p2 := &mockProvider{resp: &provider.CompletionResponse{Model: "success"}}
	f := NewFallback(ModeSmart, []provider.Provider{p1, p2}, 500*time.Millisecond)

	resp, err := f.Execute(context.Background(), &provider.CompletionRequest{})
	if err != nil || resp.Model != "success" {
		t.Errorf("expected success from p2, got %v, %v", resp, err)
	}
}

func TestFallback_AllProvidersFail(t *testing.T) {
	p1 := &mockProvider{err: provider.ErrProviderDown}
	p2 := &mockProvider{err: provider.ErrProviderDown}
	f := NewFallback(ModeReliable, []provider.Provider{p1, p2}, 500*time.Millisecond)

	_, err := f.Execute(context.Background(), &provider.CompletionRequest{})
	var pe *ProviderError
	if !errors.As(err, &pe) {
		t.Errorf("expected *ProviderError, got %T", err)
	}
}
