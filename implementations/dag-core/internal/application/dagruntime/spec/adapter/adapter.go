package adapter

import (
	"fmt"

	"dag-observatory/dag-core/internal/application/dagruntime/spec"
	"dag-observatory/dag-core/internal/application/dagruntime/spec/factory"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/node"
	"dag-observatory/dag-core/internal/domain/dagruntime/workflow"
)

func ToWorkflow(
	wfSpec spec.WorkflowSpec,
	inputKeys map[string]artifact.AnyKey,
	registry *factory.Registry,
) (workflow.Workflow, error) {
	inputs := make([]artifact.AnyKey, 0, len(wfSpec.Inputs))
	for _, inputName := range wfSpec.Inputs {
		key, ok := inputKeys[inputName]
		if !ok {
			return workflow.Workflow{}, fmt.Errorf("dagruntime spec adapter: input %q is not mapped", inputName)
		}
		inputs = append(inputs, key)
	}

	nodes := make([]node.Node, 0, len(wfSpec.Nodes))
	for _, nodeSpec := range wfSpec.Nodes {
		n, err := registry.Build(nodeSpec)
		if err != nil {
			return workflow.Workflow{}, fmt.Errorf("dagruntime spec adapter: %w", err)
		}
		nodes = append(nodes, n)
	}

	return workflow.Workflow{
		Name:   wfSpec.Name,
		Inputs: inputs,
		Nodes:  nodes,
	}, nil
}
