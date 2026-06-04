package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RizkiRdm/TNDR/internal/config"
	"github.com/RizkiRdm/TNDR/internal/gateway"
	"github.com/RizkiRdm/TNDR/internal/logger"
	"github.com/RizkiRdm/TNDR/internal/provider"
	"github.com/RizkiRdm/TNDR/internal/provider/anthropic"
	"github.com/RizkiRdm/TNDR/internal/provider/gemini"
	"github.com/RizkiRdm/TNDR/internal/provider/groq"
	"github.com/RizkiRdm/TNDR/internal/provider/openai"
	"github.com/RizkiRdm/TNDR/internal/router"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	configPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tendr",
		Short: "TENDR - AI Gateway Binary",
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "path to config file")

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the TENDR gateway",
		Run:   runStart,
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize default configuration",
		Run:   runInit,
	}

	rootCmd.AddCommand(startCmd, initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runStart(cmd *cobra.Command, args []string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.Server.LogLevel, "logs")

	// Initialize all providers
	providers := make(map[string]provider.Provider)
	
	if cfg.Providers.OpenAI.APIKey != "" {
		providers["openai"] = openai.NewOpenAIProvider(cfg.Providers.OpenAI.APIKey)
	}
	if cfg.Providers.Anthropic.APIKey != "" {
		providers["anthropic"] = anthropic.NewAnthropicProvider(cfg.Providers.Anthropic.APIKey)
	}
	if cfg.Providers.Gemini.APIKey != "" {
		providers["gemini"] = gemini.NewGeminiProvider(cfg.Providers.Gemini.APIKey)
	}
	if cfg.Providers.Groq.APIKey != "" {
		providers["groq"] = groq.NewGroqProvider(cfg.Providers.Groq.APIKey)
	}

	// Initialize router
	r := router.NewRouter(cfg, providers)
	
	// Initialize server
	server := gateway.NewServer(cfg.Server.Port, r)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatal().Err(err).Msg("gateway server failed")
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
}

func runInit(cmd *cobra.Command, args []string) {
	exampleConfig := `server:
  port: 4821
  log_level: info

providers:
  openai:
    api_key: ""
  anthropic:
    api_key: ""
  gemini:
    api_key: ""
  groq:
    api_key: ""

models:
  - alias: coding
    fallback_mode: smart
    providers:
      - openai
      - anthropic
`
	err := os.WriteFile("config.yaml", []byte(exampleConfig), 0644)
	if err != nil {
		fmt.Printf("Error writing config.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("config.yaml initialized successfully")
}
