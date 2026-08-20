package openai

import (
	"encoding/json"
	"fmt"

	"casefile/internal/provider"
)

// MessageRequest is a single message in the conversation sent to the provider.
type MessageRequest struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolCall is a single tool invocation requested by the model, echoed back
// on the assistant message that produced it.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall carries the name and raw JSON-encoded arguments the model
// wants a tool invoked with.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool describes a single callable tool in the shape the chat completions
// endpoint expects.
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function is a tool's name, description, and JSON Schema parameters.
type Function struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  Schema `json:"parameters"`
}

// Schema is the minimal JSON Schema subset used to describe a tool's
// parameters.
type Schema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property describes a single parameter within a Schema.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ChatRequest is the request body for the chat completions endpoint.
type ChatRequest struct {
	Model    string           `json:"model"`
	Messages []MessageRequest `json:"messages"`
	Tools    []Tool           `json:"tools,omitempty"`
}

// NewChatRequest builds a multi-turn request from req.Messages, transforming
// any registered tools into the wire Tool shape.
func NewChatRequest(model string, req provider.Request) (ChatRequest, error) {
	adapter := new(Adapter)
	tools := make([]Tool, 0)

	// Transform all given tools.
	for _, tool := range req.Tools {
		t := Tool{}
		if err := adapter.Transform(&tool, &t); err != nil {
			return ChatRequest{}, err
		}
		tools = append(tools, t)
	}

	messages := make([]MessageRequest, 0, len(req.Messages))
	for _, m := range req.Messages {
		mr, err := toMessageRequest(m)
		if err != nil {
			return ChatRequest{}, err
		}
		messages = append(messages, mr)
	}

	return ChatRequest{
		Model:    model,
		Messages: messages,
		Tools:    tools,
	}, nil
}

// toMessageRequest translates a provider-agnostic Message into the wire
// MessageRequest shape, marshaling any tool call arguments back to JSON.
func toMessageRequest(m provider.Message) (MessageRequest, error) {
	mr := MessageRequest{
		Role:       string(m.Role),
		Content:    m.Content,
		ToolCallID: m.ToolCallID,
	}

	if len(m.ToolCalls) > 0 {
		toolCalls := make([]ToolCall, 0, len(m.ToolCalls))
		for _, tc := range m.ToolCalls {
			args, err := json.Marshal(tc.Arguments)
			if err != nil {
				return MessageRequest{}, fmt.Errorf("marshal tool call arguments: %w", err)
			}
			toolCalls = append(toolCalls, ToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: FunctionCall{
					Name:      tc.Name,
					Arguments: string(args),
				},
			})
		}
		mr.ToolCalls = toolCalls
	}

	return mr, nil
}
