// Package agent drives the tool-calling loop between a provider.Provider
// and a tool.Registry until the model produces a final answer.
package agent

import (
	"context"
	"fmt"
	"log/slog"

	"casefile/internal/core/tool"
	"casefile/internal/provider"
)

// maxSteps bounds the number of round-trips to the provider so a
// misbehaving model can't loop forever.
const maxSteps = 200

// Loop drives a system+user conversation through a Provider, executing any
// requested tool calls via an Executor and feeding results back, until the
// model responds with no tool calls.
type Loop struct {
	provider provider.Provider
	registry *tool.Registry
	executor *tool.Executor
}

// New creates a new Loop.
func New(p provider.Provider, registry *tool.Registry, executor *tool.Executor) *Loop {
	return &Loop{
		provider: p,
		registry: registry,
		executor: executor,
	}
}

// Run sends system and user messages through the provider, executing any
// requested tool calls until a response carries no tool calls, and returns
// that response's content.
func (l *Loop) Run(ctx context.Context, system, user string) (string, error) {
	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: system},
		{Role: provider.RoleUser, Content: user},
	}

	tools := l.registry.All()

	slog.Info("agent: loop starting", "max_steps", maxSteps, "tools", len(tools))

	for step := 0; step < maxSteps; step++ {
		slog.Debug("agent: step starting", "step", step, "messages", len(messages))

		res, err := l.provider.Complete(ctx, provider.Request{
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			slog.Error("agent: provider completion failed", "step", step, "error", err)
			return "", fmt.Errorf("provider completion: %w", err)
		}

		slog.Debug("agent: provider responded", "step", step, "content_length", len(res.Content), "tool_calls", len(res.ToolCalls))

		if len(res.ToolCalls) == 0 {
			slog.Info("agent: loop finished, final answer received", "step", step)
			return res.Content, nil
		}

		messages = append(messages, provider.Message{
			Role:      provider.RoleAssistant,
			Content:   res.Content,
			ToolCalls: res.ToolCalls,
		})

		for _, call := range res.ToolCalls {
			slog.Info("agent: executing tool call", "step", step, "tool", call.Name, "id", call.ID, "arguments", call.Arguments)

			result := l.executeToolCall(ctx, call)

			slog.Debug("agent: tool call finished", "step", step, "tool", call.Name, "id", call.ID, "result_length", len(result))

			messages = append(messages, provider.Message{
				Role:       provider.RoleTool,
				Content:    result,
				ToolCallID: call.ID,
			})
		}
	}

	slog.Error("agent: loop exceeded max steps", "max_steps", maxSteps)
	return "", fmt.Errorf("agent loop: exceeded max steps (%d)", maxSteps)
}

// executeToolCall resolves and runs a single requested tool call, returning
// its result as a string. Unknown tools and execution failures are turned
// into a message fed back to the model rather than aborting the loop.
func (l *Loop) executeToolCall(ctx context.Context, call provider.ToolCall) string {
	t, ok := l.registry.Get(call.Name)
	if !ok {
		slog.Warn("agent: unknown tool requested", "tool", call.Name, "id", call.ID)
		return fmt.Sprintf("error: unknown tool %q", call.Name)
	}

	result, err := l.executor.Run(ctx, t, call.Arguments)
	if err != nil {
		slog.Warn("agent: tool execution failed", "tool", call.Name, "id", call.ID, "error", err)
		return fmt.Sprintf("error: %s", err)
	}

	return result
}
