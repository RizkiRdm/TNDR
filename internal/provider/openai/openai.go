package openai

import (
	"context"
	"encoding/json"
	"github.com/RizkiRdm/TNDR/internal/provider"
)

type OpenAIProvider struct {
	apiKey  string
	baseURL string
	base    *provider.BaseClient
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		base:    provider.NewBaseClient(),
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	url := p.baseURL + "/chat/completions"
	headers := map[string]string{
		"Authorization": "Bearer " + p.apiKey,
	}

	var result provider.CompletionResponse
	if err := p.base.DoRequest(ctx, url, headers, req, &result); err != nil {
		return nil, err
	}

	result.Provider = p.Name()
	return &result, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan *provider.StreamResponse, <-chan error) {
	respChan := make(chan *provider.StreamResponse)
	errChan := make(chan error, 1)

	req.Stream = true
	url := p.baseURL + "/chat/completions"
	headers := map[string]string{
		"Authorization": "Bearer " + p.apiKey,
	}

	go func() {
		defer close(respChan)
		defer close(errChan)

		err := p.base.StreamSSE(ctx, url, headers, req, func(data []byte) error {
			var streamResp provider.StreamResponse
			if err := json.Unmarshal(data, &streamResp); err != nil {
				return err
			}
			respChan <- &streamResp
			return nil
		})

		if err != nil {
			errChan <- err
		}
	}()

	return respChan, errChan
}

