package workflow

import (
	"fmt"

	"dag-observatory/demo-go/internal/domain/dagruntime/artifact"
	dagerrors "dag-observatory/demo-go/internal/domain/dagruntime/errors"
	"dag-observatory/demo-go/internal/domain/dagruntime/node"
	"dag-observatory/demo-go/internal/domain/dagruntime/pipeline"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
)

func Compile(wf Workflow) (pipeline.Compiled, error) {
	inputs := map[artifact.RawKey]struct{}{}
	for _, key := range wf.Inputs {
		inputs[key.Raw()] = struct{}{}
	}

	providers := map[artifact.RawKey]int{}
	writers := map[state.RawKey]int{}
	for idx, n := range wf.Nodes {
		for _, key := range n.Provides() {
			raw := key.Raw()
			if _, ok := inputs[raw]; ok {
				return pipeline.Compiled{}, dagerrors.CompileError{
					Kind:    dagerrors.CompileErrInputProvidesCollision,
					Message: fmt.Sprintf("artifact key %s provided by node %d also listed in inputs", key.String(), idx),
				}
			}
			if prev, ok := providers[raw]; ok {
				return pipeline.Compiled{}, dagerrors.CompileError{
					Kind:    dagerrors.CompileErrDuplicateProvider,
					Message: fmt.Sprintf("artifact key %s provided by both node %d and node %d", key.String(), prev, idx),
				}
			}
			providers[raw] = idx
		}
		for _, key := range n.Writes() {
			raw := key.Raw()
			if prev, ok := writers[raw]; ok {
				return pipeline.Compiled{}, dagerrors.CompileError{
					Kind:    dagerrors.CompileErrDuplicateWriter,
					Message: fmt.Sprintf("state key %s written by both node %d and node %d", key.String(), prev, idx),
				}
			}
			writers[raw] = idx
		}
	}

	for idx, n := range wf.Nodes {
		for _, key := range n.Requires() {
			raw := key.Raw()
			if _, ok := inputs[raw]; ok {
				continue
			}
			if _, ok := providers[raw]; ok {
				continue
			}
			return pipeline.Compiled{}, dagerrors.CompileError{
				Kind:    dagerrors.CompileErrUnresolvedRequire,
				Message: fmt.Sprintf("node %d requires unresolved artifact key %s", idx, key.String()),
			}
		}
	}

	order, err := topoSort(wf, providers, inputs)
	if err != nil {
		return pipeline.Compiled{}, err
	}

	return pipeline.Compiled{
		Name:   wf.Name,
		Order:  order,
		Nodes:  wf.Nodes,
		Inputs: wf.Inputs,
	}, nil
}

func topoSort(wf Workflow, providers map[artifact.RawKey]int, inputs map[artifact.RawKey]struct{}) ([]node.Node, error) {
	nodeCount := len(wf.Nodes)
	inDegree := make([]int, nodeCount)
	deps := make([][]int, nodeCount)

	for idx, n := range wf.Nodes {
		for _, key := range n.Requires() {
			raw := key.Raw()
			if _, ok := inputs[raw]; ok {
				continue
			}
			if providerIdx, ok := providers[raw]; ok {
				deps[providerIdx] = append(deps[providerIdx], idx)
				inDegree[idx]++
			}
		}
	}

	queue := make([]int, 0, nodeCount)
	for i := 0; i < nodeCount; i++ {
		if inDegree[i] == 0 {
			queue = append(queue, i)
		}
	}

	order := make([]node.Node, 0, nodeCount)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, wf.Nodes[current])

		for _, next := range deps[current] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(order) != nodeCount {
		return nil, dagerrors.CompileError{
			Kind:    dagerrors.CompileErrCycleDetected,
			Message: "cycle detected in workflow",
		}
	}

	return order, nil
}
