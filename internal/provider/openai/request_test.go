package openai

import (
	"encoding/json"
	"testing"

	"casefile/internal/core/tool"
	"casefile/internal/provider"
)

func TestNewChatRequest_MultiMessage(t *testing.T) {
	req := provider.Request{
		Messages: []provider.Message{
			{Role: provider.RoleSystem, Content: "system prompt"},
			{Role: provider.RoleUser, Content: "user prompt"},
		},
	}

	chatReq, err := NewChatRequest("gpt-test", req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(chatReq.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(chatReq.Messages))
	}
	if chatReq.Messages[0].Role != "system" || chatReq.Messages[0].Content != "system prompt" {
		t.Errorf("unexpected first message: %+v", chatReq.Messages[0])
	}
	if chatReq.Messages[1].Role != "user" || chatReq.Messages[1].Content != "user prompt" {
		t.Errorf("unexpected second message: %+v", chatReq.Messages[1])
	}
}

func TestNewChatRequest_ToolCallRoundTrip(t *testing.T) {
	req := provider.Request{
		Messages: []provider.Message{
			{
				Role: provider.RoleAssistant,
				ToolCalls: []provider.ToolCall{
					{ID: "call_1", Name: "casefile_search", Arguments: tool.Arguments{"pattern": "foo"}},
				},
			},
			{Role: provider.RoleTool, Content: "result", ToolCallID: "call_1"},
		},
	}

	chatReq, err := NewChatRequest("gpt-test", req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assistantMsg := chatReq.Messages[0]
	if len(assistantMsg.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(assistantMsg.ToolCalls))
	}
	tc := assistantMsg.ToolCalls[0]
	if tc.ID != "call_1" || tc.Type != "function" || tc.Function.Name != "casefile_search" {
		t.Errorf("unexpected tool call: %+v", tc)
	}

	var args tool.Arguments
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		t.Fatalf("expected valid JSON arguments, got error: %v", err)
	}
	if args["pattern"] != "foo" {
		t.Errorf("expected pattern=foo, got %v", args)
	}

	toolMsg := chatReq.Messages[1]
	if toolMsg.Role != "tool" || toolMsg.Content != "result" || toolMsg.ToolCallID != "call_1" {
		t.Errorf("unexpected tool message: %+v", toolMsg)
	}
}

func TestNewChatRequest_TransformsTools(t *testing.T) {
	req := provider.Request{
		Tools: []tool.Tool{
			{
				Name:        "casefile_search",
				Description: "search the repo",
				Parameters: tool.Schema{
					Type:     "object",
					Required: []string{"pattern"},
				},
			},
		},
	}

	chatReq, err := NewChatRequest("gpt-test", req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(chatReq.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(chatReq.Tools))
	}
	if chatReq.Tools[0].Function.Name != "casefile_search" {
		t.Errorf("expected casefile_search, got %q", chatReq.Tools[0].Function.Name)
	}
}
