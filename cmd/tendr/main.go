package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/RizkiRdm/TNDR/internal/cache"
	"github.com/RizkiRdm/TNDR/internal/config"
	"github.com/RizkiRdm/TNDR/internal/cost"
	"github.com/RizkiRdm/TNDR/internal/gateway"
	"github.com/RizkiRdm/TNDR/internal/logger"
	"github.com/RizkiRdm/TNDR/internal/provider"
	"github.com/RizkiRdm/TNDR/internal/provider/anthropic"
	"github.com/RizkiRdm/TNDR/internal/provider/gemini"
	"github.com/RizkiRdm/TNDR/internal/provider/groq"
	"github.com/RizkiRdm/TNDR/internal/provider/openai"
	"github.com/RizkiRdm/TNDR/internal/ratelimit"
	"github.com/RizkiRdm/TNDR/internal/router"
	"github.com/RizkiRdm/TNDR/internal/store"
	"github.com/RizkiRdm/TNDR/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	configPath   string
	costProvider string
	costJSON     bool
	Version      = "0.1.0-dev"
	Commit       = "none"
	Date         = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "tendr",
		Short:   "TENDR - AI Gateway Binary",
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", Version, Commit, Date),
		Run:     runTUI,
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

	costCmd := &cobra.Command{
		Use:   "cost",
		Short: "Show cost breakdown",
		Run:   runCost,
	}
	costCmd.Flags().StringVar(&costProvider, "provider", "", "filter by provider")
	costCmd.Flags().BoolVar(&costJSON, "json", false, "output in JSON format")

	monitorCmd := &cobra.Command{
		Use:   "monitor",
		Short: "Launch TUI dashboard",
		Run:   runTUI,
	}

	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage cache",
		Run:   runCache,
	}
	cacheClearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear cache",
		Run:   runCacheClear,
	}
	cacheCmd.AddCommand(cacheClearCmd)

	rootCmd.AddCommand(startCmd, initCmd, costCmd, monitorCmd, cacheCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "tendr.db"
	}
	return filepath.Join(home, ".tendr", "tendr.db")
}

// mustHomeDir returns user home dir or panics (only called at startup)
func mustHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot determine home directory: %v\n", err)
		os.Exit(1)
	}
	return home
}

func runTUI(cmd *cobra.Command, args []string) {
	s, err := store.New(getDBPath())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	m := tui.New(s)
	
	if cmd.Use == "cost" {
		// Set active tab to Cost (index 1)
		// Assuming we can set activeTab directly in the model via a method if needed,
		// but since TUI pkg New sets to 0, for now we will rely on default launch.
		// The requirement asks to launch on Dashboard or Cost tab.
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func runCache(cmd *cobra.Command, args []string) {
	s, err := store.New(getDBPath())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	total, hits, err := s.GetCacheStats(context.Background())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	rate := 0.0
	if total > 0 {
		rate = float64(hits) / float64(total) * 100
	}

	fmt.Printf("Cache Hits:     %d\n", hits)
	fmt.Printf("Total Requests: %d\n", total)
	fmt.Printf("Hit Rate:       %.2f%%\n", rate)
}

func runCacheClear(cmd *cobra.Command, args []string) {
	s, err := store.New(getDBPath())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	if err := s.ClearCache(context.Background()); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Cache cleared")
}

func runStart(cmd *cobra.Command, args []string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Resolve log directory to ~/.tendr/logs/
	logDir := filepath.Join(mustHomeDir(), ".tendr", "logs")
	logger.Init(cfg.Server.LogLevel, logDir)

	// Initialize Store
	s, err := store.New(getDBPath())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize store")
	}
	defer s.Close()

	// Initialize Pricing and Tracker
	pm := cost.NewPricingManager()
	tracker := cost.NewTracker(s, pm, &cfg.Server)

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
	r := router.NewRouter(cfg, providers, tracker)

	// Initialize cache
	c := cache.NewExact(5 * time.Minute)

	// Initialize limiters
	limiters := make(map[string]*ratelimit.Limiter)
	for _, m := range cfg.Models {
		limit := m.RateLimit
		if limit <= 0 {
			limit = 10
		}
		limiters[m.Alias] = ratelimit.NewLimiter(float64(limit), float64(limit*2))
	}

	// Initialize server
	server := gateway.NewServer(cfg.Server.Port, r, c, s, limiters, &cfg.Server)

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

func runCost(cmd *cobra.Command, args []string) {
	s, err := store.New(getDBPath())
	if err != nil {
		fmt.Printf("Error initializing store: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	summary, err := s.GetCostSummary(context.Background(), costProvider)
	if err != nil {
		fmt.Printf("Error getting cost summary: %v\n", err)
		os.Exit(1)
	}

	if costJSON {
		data, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Println(string(data))
		return
	}

	title := "Cost Summary"
	if costProvider != "" {
		title = fmt.Sprintf("Cost Summary for %s", costProvider)
	}

	fmt.Printf("%s\n", title)
	fmt.Println("====================")
	fmt.Printf("Today:     $%.4f\n", summary.Today)
	fmt.Printf("Last 7d:   $%.4f\n", summary.Week)
	fmt.Printf("Last 30d:  $%.4f\n", summary.Month)
	fmt.Printf("All Time:  $%.4f\n", summary.AllTime)
}
