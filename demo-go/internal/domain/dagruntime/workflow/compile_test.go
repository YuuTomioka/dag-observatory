package workflow

import (
	"context"
	"testing"

	"dag-observatory/demo-go/internal/domain/dagruntime/artifact"
	"dag-observatory/demo-go/internal/domain/dagruntime/node"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
)

type compileNode struct {
	req []artifact.AnyKey
	pro []artifact.AnyKey
	rd  []state.AnyKey
	wr  []state.AnyKey
}

func (n *compileNode) Requires() []artifact.AnyKey { return n.req }
func (n *compileNode) Provides() []artifact.AnyKey { return n.pro }
func (n *compileNode) Reads() []state.AnyKey       { return n.rd }
func (n *compileNode) Writes() []state.AnyKey      { return n.wr }
func (n *compileNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{} }
func (n *compileNode) Run(ctx context.Context, av artifact.View, txn state.Txn) error {
	return nil
}

var (
	keyA = artifact.Key[string]{Name: "a", StableID: "artifact:a"}
	keyB = artifact.Key[string]{Name: "b", StableID: "artifact:b"}

	stateX = state.Key[int]{Name: "x", StableID: "state:x"}
)

func TestCompileDetectsUnresolvedRequire(t *testing.T) {
	wf := Workflow{
		Name:   "unresolved",
		Inputs: []artifact.AnyKey{},
		Nodes: []node.Node{
			&compileNode{req: []artifact.AnyKey{keyA}},
		},
	}
	if _, err := Compile(wf); err == nil {
		t.Fatalf("expected unresolved require error")
	}
}

func TestCompileDetectsDuplicateProvider(t *testing.T) {
	wf := Workflow{
		Name:   "dup_provider",
		Inputs: []artifact.AnyKey{keyA},
		Nodes: []node.Node{
			&compileNode{req: []artifact.AnyKey{keyA}, pro: []artifact.AnyKey{keyB}},
			&compileNode{req: []artifact.AnyKey{keyA}, pro: []artifact.AnyKey{keyB}},
		},
	}
	if _, err := Compile(wf); err == nil {
		t.Fatalf("expected duplicate provider error")
	}
}

func TestCompileDetectsDuplicateWriter(t *testing.T) {
	wf := Workflow{
		Name:   "dup_writer",
		Inputs: []artifact.AnyKey{keyA},
		Nodes: []node.Node{
			&compileNode{req: []artifact.AnyKey{keyA}, wr: []state.AnyKey{stateX}},
			&compileNode{req: []artifact.AnyKey{keyA}, wr: []state.AnyKey{stateX}},
		},
	}
	if _, err := Compile(wf); err == nil {
		t.Fatalf("expected duplicate writer error")
	}
}

func TestCompileDetectsCycle(t *testing.T) {
	wf := Workflow{
		Name:   "cycle",
		Inputs: []artifact.AnyKey{},
		Nodes: []node.Node{
			&compileNode{req: []artifact.AnyKey{keyB}, pro: []artifact.AnyKey{keyA}},
			&compileNode{req: []artifact.AnyKey{keyA}, pro: []artifact.AnyKey{keyB}},
		},
	}
	if _, err := Compile(wf); err == nil {
		t.Fatalf("expected cycle error")
	}
}

func TestCompileDetectsInputProvideCollision(t *testing.T) {
	wf := Workflow{
		Name:   "input_collision",
		Inputs: []artifact.AnyKey{keyA},
		Nodes: []node.Node{
			&compileNode{req: []artifact.AnyKey{keyA}, pro: []artifact.AnyKey{keyA}},
		},
	}
	if _, err := Compile(wf); err == nil {
		t.Fatalf("expected input/provides collision error")
	}
}
