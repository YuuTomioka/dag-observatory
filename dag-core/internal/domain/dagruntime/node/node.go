package node

import (
	"context"

	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type Node interface {
	Requires() []artifact.AnyKey
	Provides() []artifact.AnyKey
	Reads() []state.AnyKey
	Writes() []state.AnyKey
	Spec() ExecutionSpec
	Run(ctx context.Context, av artifact.View, aw artifact.Writer, txn state.Txn) error
}

type Named interface {
	Name() string
}
