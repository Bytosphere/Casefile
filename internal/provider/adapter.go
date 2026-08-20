package provider

import "casefile/internal/core/tool"

// Adapter is the translation layer that transforms tool definitions or calls
// into the appropriate shapes for a Provider.
type Adapter interface {
	// Transform takes a tool and transforms into the desired shape that the provider
	// expects.
	Transform(t *tool.Tool, target any) error
}
