package tool

// Registry stores Tool entries and offers lookup operations.
type Registry struct {
	tools map[string]*Tool
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]*Tool),
	}
}

// Register registers a new tool by its name.
func (r *Registry) Register(tool *Tool) {
	r.tools[tool.Name] = tool
}

func (r *Registry) Get(name string) (*Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// All returns every registered Tool.
func (r *Registry) All() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, *t)
	}
	return tools
}
