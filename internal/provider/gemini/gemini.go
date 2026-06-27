package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

type GeminiProvider struct {
	apiKey  string
	baseURL string
	base    *provider.BaseClient
}

func NewGeminiProvider(apiKey string, timeoutMs int) *GeminiProvider {
	d := time.Duration(timeoutMs) * time.Millisecond
	if d <= 0 {
		d = 30 * time.Second
	}
	return &GeminiProvider{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		base:    provider.NewBaseClient(d),
	}
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (p *GeminiProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	url := fmt.Sprintf("%s/models/%s:generateContent", p.baseURL, req.Model)
	headers := map[string]string{"x-goog-api-key": p.apiKey}

	gemReq := geminiRequest{Contents: mapMessages(req.Messages)}

	var gemResp geminiResponse
	if err := p.base.DoRequest(ctx, url, headers, gemReq, &gemResp); err != nil {
		return nil, err
	}

	if len(gemResp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini: no candidates")
	}

	candidate := gemResp.Candidates[0]
	content := ""
	if len(candidate.Content.Parts) > 0 {
		content = candidate.Content.Parts[0].Text
	}

	role := candidate.Content.Role
	if role == "model" {
		role = "assistant"
	}

	return &provider.CompletionResponse{
		ID:       fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
		Object:   "chat.completion",
		Provider: p.Name(),
		Created:  time.Now().Unix(),
		Model:    req.Model,
		Choices: []provider.Choice{
			{
				Message: provider.Message{
					Role:    role,
					Content: content,
				},
				FinishReason: mapFinishReason(candidate.FinishReason),
			},
		},
		Usage: provider.Usage{
			PromptTokens:     gemResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: gemResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      gemResp.UsageMetadata.TotalTokenCount,
		},
	}, nil}

func (p *GeminiProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	respChan := make(chan *provider.StreamResponse)
	errChan := make(chan error, 1)

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse", p.baseURL, req.Model)
	headers := map[string]string{"x-goog-api-key": p.apiKey}
	gemReq := geminiRequest{Contents: mapMessages(req.Messages)}

	go func() {
		defer close(respChan)
		defer close(errChan)

		err := p.base.StreamSSE(ctx, url, headers, gemReq, func(data []byte) error {
			var gemResp geminiResponse
			if err := json.Unmarshal(data, &gemResp); err != nil {
				return nil
			}

			if len(gemResp.Candidates) > 0 {
				candidate := gemResp.Candidates[0]
				if len(candidate.Content.Parts) > 0 {
					respChan <- &provider.StreamResponse{
						ID:      fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   req.Model,
						Choices: []provider.StreamChoice{
							{
								Delta: provider.MessageDelta{
									Content: candidate.Content.Parts[0].Text,
								},
								FinishReason: mapFinishReason(candidate.FinishReason),
							},
						},
					}
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

func (p *GeminiProvider) Validate(ctx context.Context) error {
	url := fmt.Sprintf("%s/models?key=%s", p.baseURL, p.apiKey)
	return p.base.DoRequest(ctx, url, nil, nil, nil)
}

func (p *GeminiProvider) Health(ctx context.Context) error {
	return p.Validate(ctx)
}

func mapMessages(msgs []provider.Message) []geminiContent {
	var contents []geminiContent
	for _, msg := range msgs {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		if role == "system" {
			role = "user"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: msg.Content}},
		})
	}
	return contents
}

func mapFinishReason(reason string) string {
	switch reason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	default:
		return reason
	}
}
