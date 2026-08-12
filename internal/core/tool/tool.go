// Package tool provides shared logic for defining and executing agent tools.
package tool

// Arguments is the parsed JSON arguments from a tool call.
type Arguments = map[string]string

// HandlerFunc executes the tool and produces a result.
type HandlerFunc = func(ctx *ExecutorContext, args Arguments) (string, error)

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// Schema is a minimal JSON Schema subset — enough for flat tool parameters.
// Extend with Items/Enum/etc. if a tool ever needs them.
type Schema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Tool is an agent tool usable from an MCP server.
type Tool struct {
	Name        string
	Description string
	Parameters  Schema
	Handler     HandlerFunc
}
