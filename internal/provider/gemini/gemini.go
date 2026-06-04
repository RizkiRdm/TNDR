package gemini

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

type GeminiProvider struct {
	apiKey string
	client *http.Client
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
		client: &http.Client{},
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
		Index        int    `json:"index"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (p *GeminiProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", req.Model, p.apiKey)

	var contents []geminiContent
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		if role == "system" {
			role = "user"
		}

		contents = append(contents, geminiContent{
			Role: role,
			Parts: []geminiPart{
				{Text: msg.Content},
			},
		})
	}

	gemReq := geminiRequest{
		Contents: contents,
	}

	jsonData, err := json.Marshal(gemReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("gemini: create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", provider.ErrProviderDown)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return nil, provider.ErrInvalidKey
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, provider.ErrRateLimit
	}
	if resp.StatusCode != http.StatusOK {
		var errData interface{}
		json.NewDecoder(resp.Body).Decode(&errData)
		return nil, fmt.Errorf("gemini: provider error (status %d): %v", resp.StatusCode, errData)
	}

	var gemResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gemResp); err != nil {
		return nil, fmt.Errorf("gemini: decode response: %w", err)
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

	normalized := &provider.CompletionResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []provider.Choice{
			{
				Index: 0,
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
	}

	return normalized, nil
}

func (p *GeminiProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	respChan := make(chan *provider.StreamResponse)
	errChan := make(chan error, 1)

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", req.Model, p.apiKey)

	var contents []geminiContent
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		if role == "system" {
			role = "user"
		}
		contents = append(contents, geminiContent{
			Role: role,
			Parts: []geminiPart{
				{Text: msg.Content},
			},
		})
	}

	gemReq := geminiRequest{
		Contents: contents,
	}

	jsonData, err := json.Marshal(gemReq)
	if err != nil {
		errChan <- fmt.Errorf("gemini: marshal request: %w", err)
		close(respChan)
		return respChan, errChan
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		errChan <- fmt.Errorf("gemini: create request: %w", err)
		close(respChan)
		return respChan, errChan
	}

	httpReq.Header.Set("Content-Type", "application/json")

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
			errChan <- fmt.Errorf("gemini stream error: status %d", resp.StatusCode)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			var gemResp geminiResponse
			if err := json.Unmarshal([]byte(data), &gemResp); err != nil {
				continue
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
								Index: 0,
								Delta: provider.MessageDelta{
									Content: candidate.Content.Parts[0].Text,
								},
								FinishReason: mapFinishReason(candidate.FinishReason),
							},
						},
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("scanner error: %w", err)
		}
	}()

	return respChan, errChan
}

func mapFinishReason(reason string) string {
	switch reason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY":
		return "content_filter"
	default:
		return reason
	}
}
