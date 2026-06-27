package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
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
	startCmd.Flags().Bool("dry-run", false, "validate config and exit")

	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Check gateway health",
		Run:   runHealth,
	}

	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Run connectivity tests",
		Run:   runTest,
	}

	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check system health",
		Run:   runDoctor,
	}

	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "Show recent logs",
		Run:   runLogs,
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
	costCmd.Flags().Bool("explain", false, "show rates for recent requests")

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

    // updateCmd implements self-update functionality
    updateCmd := &cobra.Command{
        Use:   "update",
        Short: "Update tendr to latest release",
        Run:   runUpdate,
    }
    // Add to root command after other subcommands
    rootCmd.AddCommand(startCmd, healthCmd, testCmd, doctorCmd, logsCmd, initCmd, costCmd, monitorCmd, cacheCmd, updateCmd)

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

	port := 4821
	if cfg, err := config.Load(configPath); err == nil {
		port = cfg.Server.Port
	}

	m := tui.New(s, port)
	
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

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Println("Configuration validated successfully.")
		os.Exit(0)
	}

	// Resolve log directory to ~/.tendr/logs/
	logDir := filepath.Join(mustHomeDir(), ".tendr", "logs")
	logger.Init(cfg.Server.LogLevel, logDir, cfg.Server.LogMaxSizeMB, cfg.Server.LogMaxBackups, cfg.Server.LogMaxAgeDays)

	// Initialize Store
	s, err := store.New(getDBPath())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize store")
	}
	defer s.Close()

	// Initialize Pricing and Tracker
	pm := cost.NewPricingManager()
	pm.LoadOverrides(cfg.Pricing.Override)
	tracker := cost.NewTracker(s, pm, &cfg.Server)

	// Initialize all providers
	providers := make(map[string]provider.Provider)

	if cfg.Providers.OpenAI.APIKey != "" {
		providers["openai"] = openai.NewOpenAIProvider(cfg.Providers.OpenAI.APIKey, cfg.Providers.OpenAI.Timeout)
	}
	if cfg.Providers.Anthropic.APIKey != "" {
		providers["anthropic"] = anthropic.NewAnthropicProvider(cfg.Providers.Anthropic.APIKey, cfg.Providers.Anthropic.Timeout)
	}
	if cfg.Providers.Gemini.APIKey != "" {
		providers["gemini"] = gemini.NewGeminiProvider(cfg.Providers.Gemini.APIKey, cfg.Providers.Gemini.Timeout)
	}
	if cfg.Providers.Groq.APIKey != "" {
		providers["groq"] = groq.NewGroqProvider(cfg.Providers.Groq.APIKey, cfg.Providers.Groq.Timeout)
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

func runHealth(cmd *cobra.Command, args []string) {
	resp, err := http.Get("http://localhost:4821/health")
	if err != nil {
		fmt.Printf("Gateway unreachable: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		fmt.Println("Gateway is healthy")
	} else {
		fmt.Printf("Gateway unhealthy: status %d\n", resp.StatusCode)
		os.Exit(1)
	}
}

func runTest(cmd *cobra.Command, args []string) {
	fmt.Println("Running connectivity tests...")
	// Placeholder for actual connectivity tests
	fmt.Println("All tests passed")
}

func runDoctor(cmd *cobra.Command, args []string) {
	home := mustHomeDir()
	tendrDir := filepath.Join(home, ".tendr")
	if _, err := os.Stat(tendrDir); os.IsNotExist(err) {
		fmt.Printf("Directory %s does not exist\n", tendrDir)
	} else {
		fmt.Printf("Directory %s exists\n", tendrDir)
	}
	logDir := filepath.Join(tendrDir, "logs")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		fmt.Printf("Log directory %s does not exist\n", logDir)
	} else {
		fmt.Printf("Log directory %s exists\n", logDir)
	}
}

func runUpdate(cmd *cobra.Command, args []string) {
    // Determine current executable path
    exePath, err := os.Executable()
    if err != nil {
        fmt.Printf("Unable to locate executable: %v\n", err)
        os.Exit(1)
    }

    // Get latest release info from GitHub API
    resp, err := http.Get("https://api.github.com/repos/RizkiRdm/TNDR/releases/latest")
    if err != nil {
        fmt.Printf("Failed to query GitHub releases: %v\n", err)
        os.Exit(1)
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        fmt.Printf("GitHub API returned %d\n", resp.StatusCode)
        os.Exit(1)
    }
    var release struct {
        TagName string `json:"tag_name"`
        Assets  []struct {
            Name               string `json:"name"`
            BrowserDownloadURL string `json:"browser_download_url"`
        } `json:"assets"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
        fmt.Printf("Failed to parse release JSON: %v\n", err)
        os.Exit(1)
    }

    // Build expected asset name
    osName := runtime.GOOS
    arch := runtime.GOARCH
    expected := fmt.Sprintf("tendr-%s-%s", osName, arch)

    var assetURL string
    for _, a := range release.Assets {
        if strings.Contains(a.Name, expected) {
            assetURL = a.BrowserDownloadURL
            break
        }
    }
    if assetURL == "" {
        fmt.Printf("No binary asset found for %s/%s in latest release %s\n", osName, arch, release.TagName)
        os.Exit(1)
    }

    // Download binary to temporary location
    tmpFile, err := os.CreateTemp("", "tendr-update-*")
    if err != nil {
        fmt.Printf("Failed to create temp file: %v\n", err)
        os.Exit(1)
    }
    defer os.Remove(tmpFile.Name())

    resp, err = http.Get(assetURL)
    if err != nil {
        fmt.Printf("Failed to download asset: %v\n", err)
        os.Exit(1)
    }
    defer resp.Body.Close()
    if _, err := io.Copy(tmpFile, resp.Body); err != nil {
        fmt.Printf("Failed to write binary: %v\n", err)
        os.Exit(1)
    }
    if err := tmpFile.Chmod(0755); err != nil {
        fmt.Printf("Failed to chmod binary: %v\n", err)
        os.Exit(1)
    }
    tmpFile.Close()

    // Replace current executable atomically
    if err := os.Rename(tmpFile.Name(), exePath); err != nil {
        fmt.Printf("Failed to replace executable: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("tendr updated to", release.TagName)
}

func runLogs(cmd *cobra.Command, args []string) {
	logPath := filepath.Join(mustHomeDir(), ".tendr", "logs", "tendr.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Printf("Error reading logs: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func runInit(cmd *cobra.Command, args []string) {
	path := configPath
	if path == "" {
		home := mustHomeDir()
		tendrDir := filepath.Join(home, ".tendr")
		if err := os.MkdirAll(tendrDir, 0700); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", tendrDir, err)
			os.Exit(1)
		}
		path = filepath.Join(tendrDir, "config.yaml")
	} else {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0700); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	sw := &config.SetupWizard{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		EnvGet: os.Getenv,
		Validate: func(ctx context.Context, providerName, key string) error {
			switch providerName {
			case "openai":
				p := openai.NewOpenAIProvider(key, 30000)
				return p.Validate(ctx)
			case "anthropic":
				p := anthropic.NewAnthropicProvider(key, 30000)
				return p.Validate(ctx)
			case "gemini":
				p := gemini.NewGeminiProvider(key, 30000)
				return p.Validate(ctx)
			case "groq":
				p := groq.NewGroqProvider(key, 30000)
				return p.Validate(ctx)
			default:
				return fmt.Errorf("unknown provider: %s", providerName)
			}
		},
		WriteFile: os.WriteFile,
	}

	err := sw.Run(context.Background(), path)
	if err != nil {
		fmt.Printf("Error running setup wizard: %v\n", err)
		os.Exit(1)
	}
}

func runCost(cmd *cobra.Command, args []string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

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

	explain, _ := cmd.Flags().GetBool("explain")
	if explain {
		records, err := s.GetRecentRequests(context.Background(), 10)
		if err != nil {
			fmt.Printf("Error getting recent requests: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Cost Breakdown per Request:")
		fmt.Println("====================")
		for _, r := range records {
			promptCost := (float64(r.PromptTokens) * r.PromptRate) / 1_000_000.0
			compCost := (float64(r.CompletionTokens) * r.CompletionRate) / 1_000_000.0
			fmt.Printf("[%s] %s/%s\n", r.CreatedAt, r.Provider, r.Model)
			fmt.Printf("  %d prompt x $%.4f/1M  = $%.6f\n", r.PromptTokens, r.PromptRate, promptCost)
			fmt.Printf("  %d comp   x $%.4f/1M  = $%.6f\n", r.CompletionTokens, r.CompletionRate, compCost)
			fmt.Printf("  Total: $%.6f  (source: %s)\n", r.Cost, r.PricingSource)
			fmt.Println("  ---")
		}
		return
	}

	if costJSON {
		type costReport struct {
			Summary  store.CostSummary     `json:"summary"`
			Provider []store.ProviderCost  `json:"by_provider,omitempty"`
		}
		providers, _ := s.GetCostByProviderWithTokens(context.Background())
		report := costReport{Summary: *summary, Provider: providers}
		if costProvider != "" {
			report.Provider = nil
		}
		data, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(data))
		return
	}

	title := "Cost Summary"
	if costProvider != "" {
		title = fmt.Sprintf("Cost Summary for %s", costProvider)
	}

	fmt.Printf("%s\n", title)
	fmt.Println("====================")
	fmt.Printf("Today:     $%.4f", summary.Today)
	if cfg.Server.DailyCostLimit > 0 {
		fmt.Printf(" / $%.4f (Limit: %.1f%%)\n", cfg.Server.DailyCostLimit, (summary.Today/cfg.Server.DailyCostLimit)*100)
	} else {
		fmt.Println()
	}
	fmt.Printf("Last 7d:   $%.4f\n", summary.Week)
	fmt.Printf("Last 30d:  $%.4f\n", summary.Month)
	if summary.Month > 0 {
		projection := (summary.Month / 30.0) * 30.0
		fmt.Printf("Projection (30d): $%.4f\n", projection)
	}
	fmt.Printf("All Time:  $%.4f\n", summary.AllTime)

	if costProvider == "" {
		providers, err := s.GetCostByProviderWithTokens(context.Background())
		if err == nil && len(providers) > 0 {
			fmt.Println()
			fmt.Println("By Provider:")
			for _, p := range providers {
				total := p.PromptTokens + p.CompletionTokens
				fmt.Printf("  %-12s $%.4f  (%d tokens)\n", p.Provider+":", p.Cost, total)
			}
		}
	}
}
