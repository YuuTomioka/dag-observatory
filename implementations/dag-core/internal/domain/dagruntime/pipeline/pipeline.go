package pipeline

import (
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
)

type Compiled struct {
	Name   string
	Order  []node.Node
	Nodes  []node.Node
	Inputs []artifact.AnyKey
}
