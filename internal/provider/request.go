package provider

import "casefile/internal/core/tool"

// Request is the provider-agnostic input to Provider.Complete.
// Each Provider implementation is responsible for translating this into
// its own wire format via its Adapter.
type Request struct {
	Prompt string
	Tools  []tool.Tool
}
