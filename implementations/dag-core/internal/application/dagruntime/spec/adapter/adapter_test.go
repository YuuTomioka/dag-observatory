package adapter

import (
	"context"
	"testing"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/factory"
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type testFactory struct{}

func (f *testFactory) Kind() string { return "heavy_calc" }

func (f *testFactory) Build(nodeSpec spec.NodeSpec) (node.Node, error) {
	_ = nodeSpec
	return &testNode{}, nil
}

type testNode struct{}

func (n *testNode) Requires() []artifact.AnyKey { return []artifact.AnyKey{usecase.InputKeySymbol} }
func (n *testNode) Provides() []artifact.AnyKey { return nil }
func (n *testNode) Reads() []state.AnyKey       { return nil }
func (n *testNode) Writes() []state.AnyKey      { return nil }
func (n *testNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{} }
func (n *testNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = av
	_ = aw
	_ = txn
	return nil
}

func TestToWorkflow(t *testing.T) {
	t.Parallel()

	registry := factory.NewRegistry()
	if err := registry.Register(&testFactory{}); err != nil {
		t.Fatalf("register factory: %v", err)
	}

	wf, err := ToWorkflow(
		spec.WorkflowSpec{
			Name:    "dagruntime.default",
			Version: "v1",
			Inputs:  []string{"symbol", "mode"},
			Nodes:   []spec.NodeSpec{{ID: "n1", Kind: "heavy_calc"}},
		},
		map[string]artifact.AnyKey{
			"symbol": usecase.InputKeySymbol,
			"mode":   usecase.InputKeyMode,
		},
		registry,
	)
	if err != nil {
		t.Fatalf("to workflow: %v", err)
	}
	if wf.Name != "dagruntime.default" {
		t.Fatalf("expected workflow name, got %q", wf.Name)
	}
	if len(wf.Inputs) != 2 {
		t.Fatalf("expected two inputs, got %d", len(wf.Inputs))
	}
	if len(wf.Nodes) != 1 {
		t.Fatalf("expected one node, got %d", len(wf.Nodes))
	}
}

func TestToWorkflowUnknownInput(t *testing.T) {
	t.Parallel()

	registry := factory.NewRegistry()
	if err := registry.Register(&testFactory{}); err != nil {
		t.Fatalf("register factory: %v", err)
	}

	_, err := ToWorkflow(
		spec.WorkflowSpec{
			Name:    "dagruntime.default",
			Version: "v1",
			Inputs:  []string{"unknown"},
			Nodes:   []spec.NodeSpec{{ID: "n1", Kind: "heavy_calc"}},
		},
		map[string]artifact.AnyKey{},
		registry,
	)
	if err == nil {
		t.Fatal("expected unknown input error")
	}
}
