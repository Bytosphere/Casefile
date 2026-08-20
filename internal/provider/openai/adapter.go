package openai

import (
	"fmt"

	"casefile/internal/core/tool"
	"casefile/internal/provider"
)

// Adapter implements provider.Adapter for OpenAI-compatible backends.
type Adapter struct{}

var _ provider.Adapter = (*Adapter)(nil)

func (a *Adapter) Transform(t *tool.Tool, target any) error {
	if t == nil {
		return fmt.Errorf("openai adapter: nil tool")
	}

	dst, ok := target.(*Tool)
	if !ok {
		return fmt.Errorf("openai adapter: target must be *openai.Tool, got %T", target)
	}

	properties := make(map[string]Property, len(t.Parameters.Properties))
	for name, prop := range t.Parameters.Properties {
		properties[name] = Property{
			Type:        prop.Type,
			Description: prop.Description,
		}
	}

	dst.Type = "function"
	dst.Function = Function{
		Name:        t.Name,
		Description: t.Description,
		Parameters: Schema{
			Type:       t.Parameters.Type,
			Properties: properties,
			Required:   t.Parameters.Required,
		},
	}

	return nil
}
