package di

import (
	"context"
	"fmt"

	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/dagruntime/workflow"
)

var heavyCalcResultKey = artifact.Key[string]{
	Name:     "heavy_calc_result",
	StableID: "artifact:dagruntime.heavy_calc.result.v1",
}

type heavyCalcNode struct{}

func (n *heavyCalcNode) Name() string { return "dagruntime.heavy_calc" }

func (n *heavyCalcNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{usecase.InputKeySymbol, usecase.InputKeyMode}
}

func (n *heavyCalcNode) Provides() []artifact.AnyKey { return []artifact.AnyKey{heavyCalcResultKey} }
func (n *heavyCalcNode) Reads() []state.AnyKey       { return nil }
func (n *heavyCalcNode) Writes() []state.AnyKey      { return nil }
func (n *heavyCalcNode) Spec() node.ExecutionSpec    { return node.ExecutionSpec{Deterministic: true} }

func (n *heavyCalcNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = txn
	symbol := artifact.MustGet(av, usecase.InputKeySymbol)
	mode := artifact.MustGet(av, usecase.InputKeyMode)
	artifact.Set(aw, heavyCalcResultKey, fmt.Sprintf("%s:%s", symbol, mode))
	return nil
}

func compileDefaultWorkflow() (pipeline.Compiled, error) {
	wf := workflow.Workflow{
		Name:   "dagruntime.default",
		Inputs: []artifact.AnyKey{usecase.InputKeySymbol, usecase.InputKeyMode},
		Nodes: []node.Node{
			&heavyCalcNode{},
		},
	}
	return workflow.Compile(wf)
}
