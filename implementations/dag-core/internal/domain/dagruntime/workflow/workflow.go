package workflow

import (
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
)

type Workflow struct {
	Name   string
	Inputs []artifact.AnyKey
	Nodes  []node.Node
}
