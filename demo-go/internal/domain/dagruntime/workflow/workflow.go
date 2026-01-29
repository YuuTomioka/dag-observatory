package workflow

import (
	"dag-observatory/demo-go/internal/domain/dagruntime/artifact"
	"dag-observatory/demo-go/internal/domain/dagruntime/node"
)

type Workflow struct {
	Name   string
	Inputs []artifact.AnyKey
	Nodes  []node.Node
}

