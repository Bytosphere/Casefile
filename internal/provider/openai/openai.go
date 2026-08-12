package openai

import (
	"bytes"
	"casefile/internal/core"
	"casefile/internal/provider"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Provider struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

// New creates a new OpenAI provider loaded from the configuration file.
func New(cfg core.ProviderConfig) *Provider {
	return &Provider{
		Model:   cfg.Model,
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Client:  http.DefaultClient,
	}
}

func (p *Provider) Complete(ctx context.Context, req provider.Request) (string, error) {
	// Prepare the payload for OpenAI.
	payload, err := NewChatRequest(p.Model, req)
	if err != nil {
		return "", fmt.Errorf("create payload: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := p.BaseURL + "/v1/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.APIKey))

	client := p.Client

	res, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var response = MessageResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}
