package agent

import (
	"context"
	"errors"
	"testing"

	"casefile/internal/core/tool"
	"casefile/internal/provider"
)

// fakeProvider replays a fixed sequence of responses, one per Complete
// call, and records the requests it was given.
type fakeProvider struct {
	responses []provider.Response
	calls     int
	requests  []provider.Request
}

func (f *fakeProvider) Complete(_ context.Context, req provider.Request) (provider.Response, error) {
	f.requests = append(f.requests, req)
	if f.calls >= len(f.responses) {
		return provider.Response{}, errors.New("fakeProvider: no more responses")
	}
	res := f.responses[f.calls]
	f.calls++
	return res, nil
}

func newEchoTool(name, result string) *tool.Tool {
	return &tool.Tool{
		Name:        name,
		Description: "echoes a fixed result",
		Handler: func(_ *tool.ExecutorContext, _ tool.Arguments) (string, error) {
			return result, nil
		},
	}
}

func TestLoop_Run_NoToolCalls(t *testing.T) {
	fp := &fakeProvider{
		responses: []provider.Response{
			{Content: "final answer"},
		},
	}
	registry := tool.NewRegistry()
	executor := tool.NewExecutor(t.TempDir())

	l := New(fp, registry, executor)

	result, err := l.Run(context.Background(), "system prompt", "user prompt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "final answer" {
		t.Errorf("expected %q, got %q", "final answer", result)
	}
	if fp.calls != 1 {
		t.Errorf("expected 1 provider call, got %d", fp.calls)
	}
}

func TestLoop_Run_ExecutesToolCallThenReturnsFinalAnswer(t *testing.T) {
	registry := tool.NewRegistry()
	registry.Register(newEchoTool("casefile_search", "match found at file.go:1"))
	executor := tool.NewExecutor(t.TempDir())

	fp := &fakeProvider{
		responses: []provider.Response{
			{
				ToolCalls: []provider.ToolCall{
					{ID: "call_1", Name: "casefile_search", Arguments: tool.Arguments{"pattern": "foo"}},
				},
			},
			{Content: "[]"},
		},
	}

	l := New(fp, registry, executor)

	result, err := l.Run(context.Background(), "system prompt", "user prompt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "[]" {
		t.Errorf("expected %q, got %q", "[]", result)
	}
	if fp.calls != 2 {
		t.Fatalf("expected 2 provider calls, got %d", fp.calls)
	}

	// The second request should carry the tool result fed back as a
	// role: tool message.
	secondReq := fp.requests[1]
	var found bool
	for _, m := range secondReq.Messages {
		if m.Role == provider.RoleTool && m.ToolCallID == "call_1" {
			found = true
			if m.Content != "match found at file.go:1" {
				t.Errorf("expected tool result content, got %q", m.Content)
			}
		}
	}
	if !found {
		t.Error("expected a role:tool message echoing call_1 in the second request")
	}
}

func TestLoop_Run_UnknownToolName(t *testing.T) {
	registry := tool.NewRegistry()
	executor := tool.NewExecutor(t.TempDir())

	fp := &fakeProvider{
		responses: []provider.Response{
			{
				ToolCalls: []provider.ToolCall{
					{ID: "call_1", Name: "does_not_exist", Arguments: tool.Arguments{}},
				},
			},
			{Content: "done"},
		},
	}

	l := New(fp, registry, executor)

	result, err := l.Run(context.Background(), "system prompt", "user prompt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "done" {
		t.Errorf("expected %q, got %q", "done", result)
	}

	secondReq := fp.requests[1]
	var toolMsg *provider.Message
	for i := range secondReq.Messages {
		if secondReq.Messages[i].Role == provider.RoleTool {
			toolMsg = &secondReq.Messages[i]
		}
	}
	if toolMsg == nil {
		t.Fatal("expected a role:tool message for the unknown tool call")
	}
	if toolMsg.Content == "" {
		t.Error("expected a non-empty error message fed back to the model")
	}
}

func TestLoop_Run_ExceedsMaxSteps(t *testing.T) {
	registry := tool.NewRegistry()
	registry.Register(newEchoTool("casefile_search", "result"))
	executor := tool.NewExecutor(t.TempDir())

	responses := make([]provider.Response, 0, maxSteps+1)
	for i := 0; i < maxSteps+1; i++ {
		responses = append(responses, provider.Response{
			ToolCalls: []provider.ToolCall{
				{ID: "call", Name: "casefile_search", Arguments: tool.Arguments{}},
			},
		})
	}

	fp := &fakeProvider{responses: responses}
	l := New(fp, registry, executor)

	_, err := l.Run(context.Background(), "system prompt", "user prompt")
	if err == nil {
		t.Fatal("expected an error for exceeding max steps")
	}
}
