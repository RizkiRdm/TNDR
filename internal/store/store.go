package store

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/001_init.sql
var migrationSQL string

type Store struct {
	db *sql.DB
}

type RequestRecord struct {
	ID               string
	Model            string
	Provider         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	PromptRate       float64
	CompletionRate   float64
	Cost             float64
	PricingSource    string
	CreatedAt        string
}

func (s *Store) RecordRequest(ctx context.Context, r *RequestRecord) error {
	query := `INSERT INTO requests (id, model, provider, prompt_tokens, completion_tokens, total_tokens, prompt_rate, completion_rate, cost, pricing_source, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, r.ID, r.Model, r.Provider, r.PromptTokens, r.CompletionTokens, r.TotalTokens, r.PromptRate, r.CompletionRate, r.Cost, r.PricingSource, r.CreatedAt)
	return err
}

func (s *Store) RecordCacheHit(ctx context.Context, key string) error {
	query := `INSERT INTO cache_entries (key, value, created_at) VALUES (?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, key, "HIT", time.Now().Format(time.RFC3339))
	return err
}

// GetRecentRequests returns last N requests ordered by created_at desc
func (s *Store) GetRecentRequests(ctx context.Context, limit int) ([]RequestRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, model, provider, prompt_tokens, completion_tokens, total_tokens, prompt_rate, completion_rate, cost, pricing_source, created_at
         FROM requests ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RequestRecord
	for rows.Next() {
		var r RequestRecord
		if err := rows.Scan(&r.ID, &r.Model, &r.Provider, &r.PromptTokens,
			&r.CompletionTokens, &r.TotalTokens, &r.PromptRate, &r.CompletionRate, &r.Cost, &r.PricingSource, &r.CreatedAt); err != nil {
			continue
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *Store) GetCacheStats(ctx context.Context) (int, int, error) {
	var totalRequests, cacheHits int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM requests").Scan(&totalRequests)
	if err != nil {
		return 0, 0, err
	}
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cache_entries").Scan(&cacheHits)
	if err != nil {
		return 0, 0, err
	}
	return totalRequests, cacheHits, nil
}

func (s *Store) ClearCache(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM cache_entries")
	return err
}

func New(dbPath string) (*Store, error) {
	// Handle SQLite in-memory database — no directory needed
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("mkdir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// SQLite tuning
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Store{db: db}, nil
}

func runMigrations(db *sql.DB) error {
	_, err := db.Exec(migrationSQL)
	return err
}

func (s *Store) GetCostByProvider(ctx context.Context) (map[string]float64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT provider, COALESCE(SUM(cost), 0) FROM requests GROUP BY provider`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var prov string
		var total float64
		if err := rows.Scan(&prov, &total); err != nil {
			continue
		}
		result[prov] = total
	}
	return result, rows.Err()
}

func (s *Store) Close() error {
	return s.db.Close()
}

type CostSummary struct {
	Today    float64
	Week     float64
	Month    float64
	AllTime  float64
}

func (s *Store) GetCostSummary(ctx context.Context, provider string) (*CostSummary, error) {
	summary := &CostSummary{}

	providerFilter := ""
	args := []interface{}{}
	if provider != "" {
		providerFilter = " AND provider = ?"
		args = append(args, provider)
	}

	// Today
	query := fmt.Sprintf("SELECT COALESCE(SUM(cost), 0) FROM requests WHERE created_at >= ? %s", providerFilter)
	today := time.Now().Truncate(24 * time.Hour).Format(time.RFC3339)
	err := s.db.QueryRowContext(ctx, query, append([]interface{}{today}, args...)...).Scan(&summary.Today)
	if err != nil {
		return nil, err
	}

	// Week (last 7 days)
	week := time.Now().AddDate(0, 0, -7).Format(time.RFC3339)
	err = s.db.QueryRowContext(ctx, query, append([]interface{}{week}, args...)...).Scan(&summary.Week)
	if err != nil {
		return nil, err
	}

	// Month (last 30 days)
	month := time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
	err = s.db.QueryRowContext(ctx, query, append([]interface{}{month}, args...)...).Scan(&summary.Month)
	if err != nil {
		return nil, err
	}

	// All Time
	queryAll := "SELECT COALESCE(SUM(cost), 0) FROM requests"
	if provider != "" {
		queryAll += " WHERE provider = ?"
		err = s.db.QueryRowContext(ctx, queryAll, provider).Scan(&summary.AllTime)
	} else {
		err = s.db.QueryRowContext(ctx, queryAll).Scan(&summary.AllTime)
	}
	if err != nil {
		return nil, err
	}

	return summary, nil
}
