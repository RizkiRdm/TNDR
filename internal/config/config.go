package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig       `mapstructure:"server"`
	Providers ProvidersConfig    `mapstructure:"providers"`
	Models    []ModelAliasConfig `mapstructure:"models"`
}

type ServerConfig struct {
	Port               int    `mapstructure:"port"`
	LogLevel           string `mapstructure:"log_level"`
	GlobalKey          string `mapstructure:"global_key"`
	LatencyThresholdMs int    `mapstructure:"latency_threshold_ms"`
}

type ProvidersConfig struct {
	OpenAI    ProviderSettings `mapstructure:"openai"`
	Anthropic ProviderSettings `mapstructure:"anthropic"`
	Gemini    ProviderSettings `mapstructure:"gemini"`
	Groq      ProviderSettings `mapstructure:"groq"`
}

type ProviderSettings struct {
	APIKey  string `mapstructure:"api_key"`
	Timeout int    `mapstructure:"timeout_ms"`
}

type ModelAliasConfig struct {
	Alias        string   `mapstructure:"alias"`
	FallbackMode string   `mapstructure:"fallback_mode"`
	Providers    []string `mapstructure:"providers"`
	RateLimit    int      `mapstructure:"rate_limit"`
}

func Load(configPath string) (*Config, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetDefault("server.port", 4821)
	viper.SetDefault("server.log_level", "info")
	viper.SetDefault("server.latency_threshold_ms", 500)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found")
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("validate: invalid server port: %d", cfg.Server.Port)
	}

	validFallbackModes := map[string]bool{"reliable": true, "fast": true, "smart": true}
	for _, m := range cfg.Models {
		if m.Alias == "" {
			return fmt.Errorf("validate: model alias cannot be empty")
		}
		if len(m.Providers) == 0 {
			return fmt.Errorf("validate: model %q has no providers", m.Alias)
		}
		if m.FallbackMode != "" && !validFallbackModes[m.FallbackMode] {
			return fmt.Errorf("validate: model %q has invalid fallback_mode: %q (must be reliable|fast|smart)", m.Alias, m.FallbackMode)
		}
	}
	return nil
}
