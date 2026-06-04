package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/RizkiRdm/TNDR/internal/provider"
)

type OpenAIProvider struct {
	apiKey string
	client *http.Client
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	url := "https://api.openai.com/v1/chat/completions"

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: %w", provider.ErrProviderDown)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, provider.ErrInvalidKey
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, provider.ErrRateLimit
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai: provider error: status code %d", resp.StatusCode)
	}

	var result provider.CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	respChan := make(chan *provider.StreamResponse)
	errChan := make(chan error, 1)

	req.Stream = true
	url := "https://api.openai.com/v1/chat/completions"

	jsonData, err := json.Marshal(req)
	if err != nil {
		errChan <- fmt.Errorf("marshal request: %w", err)
		close(respChan)
		return respChan, errChan
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		errChan <- fmt.Errorf("create request: %w", err)
		close(respChan)
		return respChan, errChan
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

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
			errChan <- fmt.Errorf("openai stream error: status %d", resp.StatusCode)
			return
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

			var streamResp provider.StreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				errChan <- fmt.Errorf("unmarshal stream: %w", err)
				return
			}
			respChan <- &streamResp
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("scanner error: %w", err)
		}
	}()

	return respChan, errChan
}
