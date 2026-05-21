package tools

import (
	"fmt"
	"sort"
)

type Registry struct {
	tools map[string]Tool
}

func NewRegistry(tools ...Tool) (*Registry, error) {
	registry := &Registry{tools: make(map[string]Tool, len(tools))}
	for _, tool := range tools {
		if tool == nil {
			return nil, fmt.Errorf("tool registry contains nil tool")
		}
		name := tool.Name()
		if name == "" {
			return nil, fmt.Errorf("tool registry contains unnamed tool")
		}
		if _, exists := registry.tools[name]; exists {
			return nil, fmt.Errorf("duplicate tool name %q", name)
		}
		registry.tools[name] = tool
	}
	return registry, nil
}

func (r *Registry) Get(name string) (Tool, bool) {
	if r == nil {
		return nil, false
	}
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *Registry) Schemas() []ToolSchema {
	if r == nil {
		return nil
	}

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)

	schemas := make([]ToolSchema, 0, len(names))
	for _, name := range names {
		schemas = append(schemas, r.tools[name].Schema())
	}
	return schemas
}
