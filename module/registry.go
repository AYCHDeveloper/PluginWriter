package module

import "fmt"

type (
	// Creator is a Job builder
	Creator struct {
		UpdateEvery       int
		DisabledByDefault bool
		Create            func() Module
	}
	// Registry is a collection of Creators
	Registry map[string]Creator
)

// DefaultRegistry DefaultRegistry
var DefaultRegistry = Registry{}

// Register registers a module in the DefaultRegistry
func Register(name string, creator Creator) {
	DefaultRegistry.Register(name, creator)
}

// Register registers a module
func (r *Registry) Register(name string, creator Creator) {
	if _, ok := (*r)[name]; ok {
		panic(fmt.Sprintf("%s is already in registry", name))
	}
	(*r)[name] = creator
}
