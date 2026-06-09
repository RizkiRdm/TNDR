package tabs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/RizkiRdm/TNDR/internal/store"
	"github.com/spf13/viper"
)

func DashboardView(s *store.Store) string {
	var sb strings.Builder
	sb.WriteString("┌─ GATEWAY STATUS ──────────────────────────────────────────────┐\n")
	sb.WriteString("│  Status    ● RUNNING       Port    4821                       │\n")
	sb.WriteString("│  Uptime    " + time.Since(time.Now().Add(-2*time.Hour)).Truncate(time.Minute).String() + "          Requests  1,847 total              │\n")
	sb.WriteString("└───────────────────────────────────────────────────────────────┘\n\n")

	sb.WriteString("┌─ PROVIDER HEALTH ─────────────────────────────────────────────┐\n")
	sb.WriteString("│  OpenAI      ● ok    142ms avg    2,341 req    $0.0421        │\n")
	sb.WriteString("│  Anthropic   ● ok    891ms avg      412 req    $0.0211        │\n")
	sb.WriteString("│  Gemini      ● ok     88ms avg      103 req    $0.0012        │\n")
	sb.WriteString("│  Groq        ✗ down    —             —           —            │\n")
	sb.WriteString("└───────────────────────────────────────────────────────────────┘\n\n")

	sb.WriteString("┌─ LAST 10 REQUESTS ────────────────────────────────────────────┐\n")
	sb.WriteString("│  TIME        MODEL         PROVIDER    TOKENS   COST    STATUS│\n")
	sb.WriteString("│  11:42:01    coding        openai       1,204   $0.0021  ✓    │\n")
	sb.WriteString("│  11:41:58    fast          groq           341   $0.0002  ✗ fb │\n")
	sb.WriteString("│  11:41:44    default       anthropic      892   $0.0091  ✓    │\n")
	sb.WriteString("└───────────────────────────────────────────────────────────────┘\n")

	return sb.String()
}

func CostView(s *store.Store) string {
	summary, err := s.GetCostSummary(context.Background(), "")
	if err != nil {
		return fmt.Sprintf("Error loading cost: %v", err)
	}

	var sb strings.Builder
	sb.WriteString("┌─ COST SUMMARY ────────────────────────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│  Today         $%8.6f     This Week    $%8.6f             │\n", summary.Today, summary.Week))
	sb.WriteString(fmt.Sprintf("│  This Month    $%8.6f     All Time     $%8.6f             │\n", summary.Month, summary.AllTime))
	sb.WriteString("└───────────────────────────────────────────────────────────────┘\n\n")

	sb.WriteString("┌─ BY PROVIDER ─────────────────────────────────────────────────┐\n")
	sb.WriteString("│  OpenAI        $0.0282   ████████████░░░░  67%               │\n")
	sb.WriteString("│  Anthropic     $0.0112   █████░░░░░░░░░░░  27%               │\n")
	sb.WriteString("│  Gemini        $0.0027   █░░░░░░░░░░░░░░░   6%               │\n")
	sb.WriteString("└───────────────────────────────────────────────────────────────┘\n")

	return sb.String()
}

func CacheView(s *store.Store) string {
	total, hits, err := s.GetCacheStats(context.Background())
	if err != nil {
		return fmt.Sprintf("Error loading cache stats: %v", err)
	}

	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	var sb strings.Builder
	sb.WriteString("┌─ CACHE STATUS ────────────────────────────────────────────────┐\n")
	sb.WriteString("│  Type       In-Memory (Exact)                                 │\n")
	sb.WriteString(fmt.Sprintf("│  Entries    %d total                                          │\n", total))
	sb.WriteString(fmt.Sprintf("│  Hit Rate   %.1f%%                                             │\n", hitRate))
	sb.WriteString("└───────────────────────────────────────────────────────────────┘\n\n")

	sb.WriteString(" [c] clear all   [d] clear selected   [enter] view entry\n")
	return sb.String()
}

func ConfigView(s *store.Store) string {
	cfgFile := viper.ConfigFileUsed()
	content, err := os.ReadFile(cfgFile)
	if err != nil {
		return fmt.Sprintf("Error reading config: %v", err)
	}

	// Mask API Keys
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "api_key:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[1])
				if len(key) > 6 {
					lines[i] = fmt.Sprintf("%s: %s••••••••", parts[0], key[:6])
				}
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("┌─ ACTIVE CONFIG ───────────────────────────────────────────────┐\n"))
	sb.WriteString(fmt.Sprintf("│  File: %s      [e] open in $EDITOR\n", cfgFile))
	sb.WriteString(strings.Join(lines, "\n"))
	sb.WriteString("\n└───────────────────────────────────────────────────────────────┘\n")
	sb.WriteString(" [e] edit in $EDITOR   [r] reload config   [v] validate\n")

	return sb.String()
}

func OpenConfigInEditor(cfgFile string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cmd := exec.Command(editor, cfgFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func LogsView(s *store.Store) string {
	content, err := os.ReadFile("logs/tendr.log")
	if err != nil {
		return fmt.Sprintf("Error reading logs: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 20 {
		lines = lines[len(lines)-20:]
	}

	var sb strings.Builder
	sb.WriteString("┌─ RECENT LOGS (LIVE) ──────────────────────────────────────────┐\n")
	sb.WriteString(strings.Join(lines, "\n"))
	sb.WriteString("\n└───────────────────────────────────────────────────────────────┘\n")
	sb.WriteString(" [p] pause/resume   [f] filter level (all/info/warn/err)\n")
	return sb.String()
}
