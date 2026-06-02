package provider

import (
	"context"
	"errors"
)

var (
	ErrRateLimit    = errors.New("rate_limit")
	ErrTimeout      = errors.New("timeout")
	ErrProviderDown = errors.New("provider_down")
	ErrInvalidKey   = errors.New("invalid_key")
)

type Provider interface {
	Name() string
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
}

type CompletionRequest struct {
	Model    string           `json:"model"`
	Messages []Message        `json:"messages"`
	Stream   bool             `json:"stream,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage  `json:"usage"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
