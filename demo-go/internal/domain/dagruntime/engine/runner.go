package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dag-observatory/demo-go/internal/application/dagruntime/port"
	"dag-observatory/demo-go/internal/domain/dagruntime/artifact"
	dagerrors "dag-observatory/demo-go/internal/domain/dagruntime/errors"
	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"dag-observatory/demo-go/internal/domain/dagruntime/node"
	"dag-observatory/demo-go/internal/domain/dagruntime/pipeline"
	"dag-observatory/demo-go/internal/domain/dagruntime/policy"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
)

type Runner struct {
	ArtifactStore artifact.Store
	StateStore    state.Store
	StateHasher   state.Hasher
	Observer      port.Observer
	Policy        policy.Policy
	Recorder      port.Recorder
	ContinueOnError bool
}

type InputMap map[artifact.AnyKey]any

func (r *Runner) RunCycle(ctx context.Context, compiled pipeline.Compiled, inputs InputMap, partition state.Partition, event events.Event) error {
	ctx = events.WithEvent(ctx, event)
	if r.Observer != nil {
		r.Observer.OnCycleStart(ctx, port.CycleInfo{
			WorkflowName: compiled.Name,
			Partition:    partition,
			Event:        event,
		})
	}
	if r.Recorder != nil {
		r.Recorder.RecordEvent(ctx, event)
	}

	start := time.Now()
	var lastErr error
	var failedNode node.Node
	var failedDuration time.Duration
	lastAttempt := 1

	for attempt := 1; ; attempt++ {
		lastAttempt = attempt
		r.ArtifactStore.Clear()
		for key, value := range inputs {
			r.ArtifactStore.Set(key, value)
		}

		txn := r.StateStore.BeginTxn(partition)
		nodeErr := r.runPipeline(ctx, compiled, txn, event.EventID, int64(attempt-1))
		if nodeErr == nil {
			if err := txn.Commit(); err != nil {
				lastErr = dagerrors.RuntimeError{
					Kind:    dagerrors.RuntimeErrTxnCommitFailure,
					Message: err.Error(),
					Err:     err,
				}
				break
			}
			lastErr = nil
			break
		}

		txn.Rollback()
		lastErr = nodeErr.err
		failedNode = nodeErr.node
		failedDuration = nodeErr.duration

		if !canRetry(nodeErr.spec) {
			break
		}
		if attempt >= maxAttempts(r.Policy, nodeErr.spec) {
			lastErr = dagerrors.RuntimeError{
				Kind:    dagerrors.RuntimeErrRetryExhausted,
				Message: lastErr.Error(),
				Err:     lastErr,
			}
			break
		}
	}

	duration := time.Since(start)
	stateHash := ""
	if r.StateHasher != nil {
		stateHash = r.StateHasher.Hash(partition)
	} else if hasher, ok := r.StateStore.(state.HashableStore); ok {
		stateHash = hasher.Hash(partition)
	}
	if r.Observer != nil {
		r.Observer.OnCycleEnd(ctx, port.CycleResult{
			WorkflowName: compiled.Name,
			Partition:    partition,
			Event:        event,
			Duration:     duration,
			Err:          lastErr,
			StateHash:    stateHash,
			RetryCount:   int64(maxInt(0, lastAttempt-1)),
		})
	}
	if r.Recorder != nil {
		r.Recorder.RecordCycleResult(ctx, port.CycleResult{
			WorkflowName: compiled.Name,
			Partition:    partition,
			Event:        event,
			Duration:     duration,
			Err:          lastErr,
			StateHash:    stateHash,
			RetryCount:   int64(maxInt(0, lastAttempt-1)),
		})
	}

	if lastErr != nil && failedNode != nil && r.Observer != nil {
		r.Observer.OnError(ctx, port.ErrorInfo{
			Stage:   "node",
			Node:    nodeName(failedNode),
			Err:     lastErr,
			Details: fmt.Sprintf("duration=%s", failedDuration),
		})
	}

	return lastErr
}

type DriverOptions struct {
	ContinueOnError bool
}

type nodeFailure struct {
	node     node.Node
	spec     node.ExecutionSpec
	err      error
	duration time.Duration
}

func (r *Runner) runPipeline(ctx context.Context, compiled pipeline.Compiled, txn state.Txn, runID string, retryCount int64) *nodeFailure {
	for _, n := range compiled.Order {
		if r.Observer != nil {
			r.Observer.OnNodeStart(ctx, port.NodeInfo{
				RunID:    runID,
				NodeName: nodeName(n),
				QueueWaitMS: 0,
			})
		}

		start := time.Now()
		runCtx := ctx
		timeout := effectiveTimeout(r.Policy, n.Spec())
		cancel := func() {}
		if timeout > 0 {
			runCtx, cancel = context.WithTimeout(ctx, timeout)
		}

		err := n.Run(runCtx, r.ArtifactStore.View(), r.ArtifactStore, txn)
		cancel()
		duration := time.Since(start)

		if err != nil && errors.Is(err, context.DeadlineExceeded) {
			err = dagerrors.TimeoutError{
				Node: nodeName(n),
				Err:  err,
			}
		}

		if r.Observer != nil {
			r.Observer.OnNodeEnd(ctx, port.NodeResult{
				RunID:    runID,
				NodeName: nodeName(n),
				Duration: duration,
				Err:      err,
				RetryCount: retryCount,
			})
		}
		if r.Recorder != nil {
			r.Recorder.RecordNodeResult(ctx, port.NodeResult{
				RunID:    runID,
				NodeName: nodeName(n),
				Duration: duration,
				Err:      err,
				RetryCount: retryCount,
			})
		}

		if err != nil {
			return &nodeFailure{
				node:     n,
				spec:     n.Spec(),
				err:      err,
				duration: duration,
			}
		}
	}
	return nil
}

func effectiveTimeout(policy policy.Policy, spec node.ExecutionSpec) time.Duration {
	if spec.Timeout > 0 {
		return spec.Timeout
	}
	return policy.DefaultTimeout
}

func maxAttempts(policy policy.Policy, spec node.ExecutionSpec) int {
	if spec.RetryPolicy != nil && spec.RetryPolicy.MaxAttempts > 0 {
		return spec.RetryPolicy.MaxAttempts
	}
	if policy.DefaultRetry.MaxAttempts > 0 {
		return policy.DefaultRetry.MaxAttempts
	}
	return 1
}

func canRetry(spec node.ExecutionSpec) bool {
	if spec.SideEffect && !spec.Idempotent {
		return false
	}
	return true
}

func nodeName(n node.Node) string {
	if named, ok := n.(node.Named); ok {
		if name := named.Name(); name != "" {
			return name
		}
	}
	return fmt.Sprintf("%T", n)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
