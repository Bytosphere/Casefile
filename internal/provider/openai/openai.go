package openai

import (
	"bytes"
	"casefile/internal/core"
	"casefile/internal/core/tool"
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

func (p *Provider) Complete(ctx context.Context, req provider.Request) (provider.Response, error) {
	// Prepare the payload for OpenAI.
	payload, err := NewChatRequest(p.Model, req)
	if err != nil {
		return provider.Response{}, fmt.Errorf("create payload: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return provider.Response{}, fmt.Errorf("marshal request: %w", err)
	}

	url := p.BaseURL + "/v1/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return provider.Response{}, fmt.Errorf("build request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.APIKey))

	client := p.Client

	res, err := client.Do(httpReq)
	if err != nil {
		return provider.Response{}, fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return provider.Response{}, fmt.Errorf("read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return provider.Response{}, fmt.Errorf("provider returned status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var response = MessageResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		return provider.Response{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		return provider.Response{}, fmt.Errorf("provider returned no choices")
	}

	message := response.Choices[0].Message

	toolCalls := make([]provider.ToolCall, 0, len(message.ToolCalls))
	for _, tc := range message.ToolCalls {
		var args tool.Arguments
		if tc.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				return provider.Response{}, fmt.Errorf("unmarshal tool call arguments: %w", err)
			}
		}
		toolCalls = append(toolCalls, provider.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		})
	}

	return provider.Response{
		Content:   strings.TrimSpace(message.Content),
		ToolCalls: toolCalls,
	}, nil
}
