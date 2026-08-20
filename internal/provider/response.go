package provider

// Response is the provider-agnostic result of Provider.Complete.
type Response struct {
	Content   string
	ToolCalls []ToolCall
}
