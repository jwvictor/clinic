package registry

// Registry holds all known tool and stack definitions.
// In the MVP this is embedded; later it pulls from a remote registry.
type Registry struct {
	Tools  map[string]ToolDef
	Stacks map[string]StackDef
}

// Load returns the built-in registry.
func Load() *Registry {
	r := &Registry{
		Tools:  make(map[string]ToolDef),
		Stacks: make(map[string]StackDef),
	}

	for _, t := range builtinTools() {
		r.Tools[t.Name] = t
	}
	for _, s := range builtinStacks() {
		r.Stacks[s.Name] = s
	}

	return r
}

// GetTool returns a tool definition by name.
func (r *Registry) GetTool(name string) (ToolDef, bool) {
	t, ok := r.Tools[name]
	return t, ok
}

// GetStack returns a stack definition by name.
func (r *Registry) GetStack(name string) (StackDef, bool) {
	s, ok := r.Stacks[name]
	return s, ok
}
