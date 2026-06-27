package config

import (
	"strings"
	"testing"
)

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 4821},
		Models: []ModelAliasConfig{
			{Alias: "test", Providers: []string{"openai"}},
		},
	}
	if err := validate(cfg); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
		want bool
	}{
		{"port 0", 0, true},
		{"port negative", -1, true},
		{"port too high", 99999, true},
		{"port valid", 4821, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Server: ServerConfig{Port: tt.port}}
			err := validate(cfg)
			if (err != nil) != tt.want {
				t.Errorf("got error %v; want error %v", err, tt.want)
			}
		})
	}
}

func TestValidate_EmptyModelAlias(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 4821},
		Models: []ModelAliasConfig{{Alias: "", Providers: []string{"openai"}}},
	}
	err := validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "alias") {
		t.Errorf("expected error containing 'alias', got %v", err)
	}
}

func TestValidate_NoProviders(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 4821},
		Models: []ModelAliasConfig{{Alias: "test", Providers: []string{}}},
	}
	err := validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "providers") {
		t.Errorf("expected error containing 'providers', got %v", err)
	}
}

func TestValidate_InvalidFallbackMode(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 4821},
		Models: []ModelAliasConfig{{Alias: "test", Providers: []string{"openai"}, FallbackMode: "invalid"}},
	}
	err := validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "fallback_mode") {
		t.Errorf("expected error containing 'fallback_mode', got %v", err)
	}
}

func TestValidate_ValidFallbackModes(t *testing.T) {
	modes := []string{"reliable", "fast", "smart"}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{Port: 4821},
				Models: []ModelAliasConfig{{Alias: "test", Providers: []string{"openai"}, FallbackMode: mode}},
			}
			if err := validate(cfg); err != nil {
				t.Errorf("expected no error for mode %s, got %v", mode, err)
			}
		})
	}
}

func TestValidate_PricingURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"empty url", "", false},
		{"valid github url", "https://raw.githubusercontent.com/RizkiRdm/TNDR/main/pricing.json", false},
		{"invalid arbitrary url", "https://example.com/pricing.json", true},
		{"invalid protocol", "http://raw.githubusercontent.com/RizkiRdm/TNDR/main/pricing.json", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{Port: 4821},
				Pricing: PricingConfig{PricingURL: tt.url},
			}
			err := validate(cfg)
			if (err != nil) != tt.want {
				t.Errorf("got error %v; want error %v", err, tt.want)
			}
		})
	}
}
