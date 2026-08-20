package provider

import "casefile/internal/core/tool"

// Role identifies who authored a Message in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ToolCall is a single tool invocation requested by the model.
type ToolCall struct {
	ID        string
	Name      string
	Arguments tool.Arguments
}

// Message is a single provider-agnostic turn in a conversation.
type Message struct {
	Role    Role
	Content string

	// ToolCalls is set on assistant messages that requested tool calls.
	ToolCalls []ToolCall

	// ToolCallID is set on tool-role messages, echoing which call this
	// message answers.
	ToolCallID string
}
