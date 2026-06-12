package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultProviderTimeout = 30 * time.Second

// BaseClient menyediakan fungsionalitas umum untuk semua provider AI.
type BaseClient struct {
	HTTPClient *http.Client
}

// NewBaseClient creates a new instance of BaseClient.
func NewBaseClient() *BaseClient {
	return &BaseClient{
		HTTPClient: &http.Client{
			Timeout: defaultProviderTimeout,
		},
	}
}

// DoRequest sends a JSON POST request and decodes the response into the target interface.
func (c *BaseClient) DoRequest(ctx context.Context, url string, headers map[string]string, body interface{}, target interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("provider: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("provider: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return ErrProviderDown
	}
	defer resp.Body.Close()

	if err := MapHTTPError(resp.StatusCode); err != nil {
		return err
	}

	if target == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// StreamSSE handles streaming Server-Sent Events (SSE) responses from providers.
func (c *BaseClient) StreamSSE(ctx context.Context, url string, headers map[string]string, body interface{}, handler func(data []byte) error) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("provider: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("provider: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return ErrProviderDown
	}
	defer resp.Body.Close()

	if err := MapHTTPError(resp.StatusCode); err != nil {
		return err
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		if err := handler([]byte(data)); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// MapHTTPError maps HTTP status codes to standardized provider error types.
func MapHTTPError(statusCode int) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrInvalidKey
	case http.StatusTooManyRequests:
		return ErrRateLimit
	default:
		return fmt.Errorf("provider error: status %d", statusCode)
	}
}
