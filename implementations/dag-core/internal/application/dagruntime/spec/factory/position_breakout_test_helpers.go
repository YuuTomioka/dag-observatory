package factory

import (
	"context"

	"dag-observatory/dag-core/internal/domain/algotrade"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type writePositionStateNode struct {
	inputKey artifact.Key[algotrade.PositionSnapshot]
	stateKey state.Key[algotrade.PositionSnapshot]
}

func (n *writePositionStateNode) Name() string { return "write.position.state" }
func (n *writePositionStateNode) Requires() []artifact.AnyKey {
	return []artifact.AnyKey{n.inputKey}
}
func (n *writePositionStateNode) Provides() []artifact.AnyKey { return nil }
func (n *writePositionStateNode) Reads() []state.AnyKey       { return nil }
func (n *writePositionStateNode) Writes() []state.AnyKey      { return []state.AnyKey{n.stateKey} }
func (n *writePositionStateNode) Spec() node.ExecutionSpec {
	return node.ExecutionSpec{Deterministic: true}
}
func (n *writePositionStateNode) Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error {
	_ = ctx
	_ = aw
	snapshot := artifact.MustGet(av, n.inputKey)
	state.StageWrite(txn, n.stateKey, snapshot)
	return nil
}
