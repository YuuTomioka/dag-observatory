package pipeline

import (
	"dag-observatory/demo-go/internal/domain/dagruntime/artifact"
	"dag-observatory/demo-go/internal/domain/dagruntime/node"
)

type Compiled struct {
	Name   string
	Order  []node.Node
	Nodes  []node.Node
	Inputs []artifact.AnyKey
}
