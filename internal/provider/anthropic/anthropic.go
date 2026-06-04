package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

type AnthropicProvider struct {
	apiKey string
	client *http.Client
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []provider.Message `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream,omitempty"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (p *AnthropicProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	url := "https://api.anthropic.com/v1/messages"

	anthroReq := anthropicRequest{
		Model:     req.Model,
		Messages:  req.Messages,
		MaxTokens: 4096,
	}

	jsonData, err := json.Marshal(anthroReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("anthropic: create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: %w", provider.ErrProviderDown)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, provider.ErrInvalidKey
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, provider.ErrRateLimit
	}
	if resp.StatusCode != http.StatusOK {
		var errData interface{}
		json.NewDecoder(resp.Body).Decode(&errData)
		return nil, fmt.Errorf("anthropic: provider error (status %d): %v", resp.StatusCode, errData)
	}

	var anthroResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthroResp); err != nil {
		return nil, fmt.Errorf("anthropic: decode response: %w", err)
	}

	content := ""
	if len(anthroResp.Content) > 0 {
		content = anthroResp.Content[0].Text
	}

	normalized := &provider.CompletionResponse{
		ID:      anthroResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   anthroResp.Model,
		Choices: []provider.Choice{
			{
				Index: 0,
				Message: provider.Message{
					Role:    anthroResp.Role,
					Content: content,
				},
				FinishReason: mapStopReason(anthroResp.StopReason),
			},
		},
		Usage: provider.Usage{
			PromptTokens:     anthroResp.Usage.InputTokens,
			CompletionTokens: anthroResp.Usage.OutputTokens,
			TotalTokens:      anthroResp.Usage.InputTokens + anthroResp.Usage.OutputTokens,
		},
	}

	return normalized, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	respChan := make(chan *provider.StreamResponse)
	errChan := make(chan error, 1)

	url := "https://api.anthropic.com/v1/messages"
	anthroReq := anthropicRequest{
		Model:     req.Model,
		Messages:  req.Messages,
		MaxTokens: 4096,
		Stream:    true,
	}

	jsonData, err := json.Marshal(anthroReq)
	if err != nil {
		errChan <- fmt.Errorf("anthropic: marshal request: %w", err)
		close(respChan)
		return respChan, errChan
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		errChan <- fmt.Errorf("anthropic: create request: %w", err)
		close(respChan)
		return respChan, errChan
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	go func() {
		defer close(respChan)
		defer close(errChan)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			errChan <- provider.ErrProviderDown
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("anthropic stream error: status %d", resp.StatusCode)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		var lastID string
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "event: ") {
				// Handle specific events if needed
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			var event struct {
				Type string `json:"type"`
				Message struct {
					ID string `json:"id"`
				} `json:"message"`
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			if event.Type == "message_start" {
				lastID = event.Message.ID
			}

			if event.Type == "content_block_delta" {
				respChan <- &provider.StreamResponse{
					ID:      lastID,
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   req.Model,
					Choices: []provider.StreamChoice{
						{
							Index: 0,
							Delta: provider.MessageDelta{
								Content: event.Delta.Text,
							},
						},
					},
				}
			}

			if event.Type == "message_stop" {
				break
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("scanner error: %w", err)
		}
	}()

	return respChan, errChan
}

func mapStopReason(reason string) string {
	switch reason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "stop_sequence":
		return "stop"
	default:
		return reason
	}
}
