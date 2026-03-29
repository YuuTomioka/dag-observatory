package factory

import (
	"fmt"
	"strings"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
)

type NodeFactory interface {
	Kind() string
	Build(nodeSpec spec.NodeSpec) (node.Node, error)
}

type Registry struct {
	factories map[string]NodeFactory
}

func NewRegistry() *Registry {
	return &Registry{factories: map[string]NodeFactory{}}
}

func (r *Registry) Register(factory NodeFactory) error {
	if factory == nil {
		return fmt.Errorf("dagruntime node factory registry: factory is nil")
	}
	kind := strings.TrimSpace(factory.Kind())
	if kind == "" {
		return fmt.Errorf("dagruntime node factory registry: factory kind is empty")
	}
	if _, exists := r.factories[kind]; exists {
		return fmt.Errorf("dagruntime node factory registry: kind %q already registered", kind)
	}
	r.factories[kind] = factory
	return nil
}

func (r *Registry) Has(kind string) bool {
	if r == nil {
		return false
	}
	_, ok := r.factories[strings.TrimSpace(kind)]
	return ok
}

func (r *Registry) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	if r == nil {
		return nil, fmt.Errorf("dagruntime node factory registry: registry is nil")
	}
	kind := strings.TrimSpace(nodeSpec.Kind)
	factory, ok := r.factories[kind]
	if !ok {
		return nil, fmt.Errorf("dagruntime node factory registry: kind %q is not registered", kind)
	}
	n, err := factory.Build(nodeSpec)
	if err != nil {
		return nil, fmt.Errorf("dagruntime node factory registry: build node id %q kind %q: %w", nodeSpec.ID, kind, err)
	}
	return n, nil
}

func (r *Registry) ValidateNodeSpec(nodeSpec spec.NodeSpec) error {
	_, err := r.Build(nodeSpec)
	return err
}
