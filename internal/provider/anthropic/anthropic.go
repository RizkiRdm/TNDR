package anthropic

import (
	"context"
	"encoding/json"
	"time"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

type AnthropicProvider struct {
	apiKey  string
	baseURL string
	base    *provider.BaseClient
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1",
		base:    provider.NewBaseClient(),
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
	Role    string `json:"role"`
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (p *AnthropicProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	url := p.baseURL + "/messages"
	headers := map[string]string{
		"x-api-key":         p.apiKey,
		"anthropic-version": "2023-06-01",
	}

	anthroReq := anthropicRequest{
		Model:     req.Model,
		Messages:  req.Messages,
		MaxTokens: 4096,
	}

	var anthroResp anthropicResponse
	if err := p.base.DoRequest(ctx, url, headers, anthroReq, &anthroResp); err != nil {
		return nil, err
	}

	content := ""
	if len(anthroResp.Content) > 0 {
		content = anthroResp.Content[0].Text
	}

	return &provider.CompletionResponse{
		ID:       anthroResp.ID,
		Object:   "chat.completion",
		Provider: p.Name(),
		Created:  time.Now().Unix(),
		Model:    anthroResp.Model,
		Choices: []provider.Choice{
			{
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
	}, nil}

func (p *AnthropicProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	respChan := make(chan *provider.StreamResponse)
	errChan := make(chan error, 1)

	anthroReq := anthropicRequest{
		Model:     req.Model,
		Messages:  req.Messages,
		MaxTokens: 4096,
		Stream:    true,
	}

	url := p.baseURL + "/messages"
	headers := map[string]string{
		"x-api-key":         p.apiKey,
		"anthropic-version": "2023-06-01",
	}

	go func() {
		defer close(respChan)
		defer close(errChan)

		var lastID string
		err := p.base.StreamSSE(ctx, url, headers, anthroReq, func(data []byte) error {
			var event struct {
				Type    string `json:"type"`
				Message struct {
					ID string `json:"id"`
				} `json:"message"`
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}

			if err := json.Unmarshal(data, &event); err != nil {
				return nil // skip malformed
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
							Delta: provider.MessageDelta{
								Content: event.Delta.Text,
							},
						},
					},
				}
			}
			return nil
		})

		if err != nil {
			errChan <- err
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
	default:
		return reason
	}
}
