package _example

import (
	"context"
	"fmt"

	"dag-observatory/demo-go/internal/domain/dagruntime/artifact"
	"dag-observatory/demo-go/internal/domain/dagruntime/node"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
	"dag-observatory/demo-go/internal/domain/dagruntime/workflow"
)

var (
	KeyInput  = artifact.Key[string]{Name: "input", StableID: "artifact:example.input.v1"}
	KeyUpper  = artifact.Key[string]{Name: "upper", StableID: "artifact:example.upper.v1"}
	KeyCount  = artifact.Key[int]{Name: "count", StableID: "artifact:example.count.v1"}
	KeyResult = artifact.Key[string]{Name: "result", StableID: "artifact:example.result.v1"}

	StateCount = state.Key[int]{Name: "count", StableID: "state:example.count.v1"}
)

type NodeUpper struct{}

func (n *NodeUpper) Requires() []artifact.AnyKey { return []artifact.AnyKey{KeyInput} }
func (n *NodeUpper) Provides() []artifact.AnyKey { return []artifact.AnyKey{KeyUpper} }
func (n *NodeUpper) Reads() []state.AnyKey       { return nil }
func (n *NodeUpper) Writes() []state.AnyKey      { return nil }
func (n *NodeUpper) Spec() node.ExecutionSpec    { return node.ExecutionSpec{Deterministic: true} }

func (n *NodeUpper) Run(ctx context.Context, av artifact.View, txn state.Txn) error {
	_ = ctx
	input := artifact.MustGet(av, KeyInput)
	if store, ok := av.(artifact.Store); ok {
		artifact.Set(store, KeyUpper, fmt.Sprintf("%s", toUpperASCII(input)))
	}
	return nil
}

type NodeCount struct {
	Fail bool
}

func (n *NodeCount) Requires() []artifact.AnyKey { return []artifact.AnyKey{KeyUpper} }
func (n *NodeCount) Provides() []artifact.AnyKey { return []artifact.AnyKey{KeyCount} }
func (n *NodeCount) Reads() []state.AnyKey       { return []state.AnyKey{StateCount} }
func (n *NodeCount) Writes() []state.AnyKey      { return []state.AnyKey{StateCount} }
func (n *NodeCount) Spec() node.ExecutionSpec    { return node.ExecutionSpec{Deterministic: true} }

func (n *NodeCount) Run(ctx context.Context, av artifact.View, txn state.Txn) error {
	_ = ctx
	current, ok := state.Get(txn, StateCount)
	if !ok {
		current = 0
	}
	next := current + 1
	state.StageWrite(txn, StateCount, next)
	if store, ok := av.(artifact.Store); ok {
		artifact.Set(store, KeyCount, next)
	}
	if n.Fail {
		return fmt.Errorf("node count failed")
	}
	return nil
}

type NodeResult struct{}

func (n *NodeResult) Requires() []artifact.AnyKey { return []artifact.AnyKey{KeyUpper, KeyCount} }
func (n *NodeResult) Provides() []artifact.AnyKey { return []artifact.AnyKey{KeyResult} }
func (n *NodeResult) Reads() []state.AnyKey       { return nil }
func (n *NodeResult) Writes() []state.AnyKey      { return nil }
func (n *NodeResult) Spec() node.ExecutionSpec    { return node.ExecutionSpec{Deterministic: true} }

func (n *NodeResult) Run(ctx context.Context, av artifact.View, txn state.Txn) error {
	_ = ctx
	_ = txn
	upper := artifact.MustGet(av, KeyUpper)
	count := artifact.MustGet(av, KeyCount)
	if store, ok := av.(artifact.Store); ok {
		artifact.Set(store, KeyResult, fmt.Sprintf("%s:%d", upper, count))
	}
	return nil
}

func BuildWorkflow(failNode2 bool) workflow.Workflow {
	return workflow.Workflow{
		Name:   "example",
		Inputs: []artifact.AnyKey{KeyInput},
		Nodes: []node.Node{
			&NodeUpper{},
			&NodeCount{Fail: failNode2},
			&NodeResult{},
		},
	}
}

func toUpperASCII(value string) string {
	buf := []byte(value)
	for i, b := range buf {
		if b >= 'a' && b <= 'z' {
			buf[i] = b - ('a' - 'A')
		}
	}
	return string(buf)
}

