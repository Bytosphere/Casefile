package provider

import "context"

// Provider is an interface that communicates with some AI provider.
type Provider interface {
	// Complete sends a request to the AI provider and returns a response.
	Complete(ctx context.Context, req Request) (Response, error)
}
