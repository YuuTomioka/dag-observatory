package port

import (
	"context"
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type CompileInfo struct {
	WorkflowName string
	NodeCount    int
}

type CycleInfo struct {
	WorkflowName string
	Partition    state.Partition
	Event        events.Event
}

type CycleResult struct {
	WorkflowName string
	Partition    state.Partition
	Event        events.Event
	Duration     time.Duration
	Err          error
	StateHash    string
	RetryCount   int64
}

type NodeInfo struct {
	RunID       string
	NodeName    string
	QueueWaitMS int64
}

type NodeResult struct {
	RunID       string
	NodeName    string
	Duration    time.Duration
	Err         error
	RetryCount  int64
	QueueWaitMS int64
}

type ErrorInfo struct {
	Stage   string
	Node    string
	Err     error
	Details string
}

type Observer interface {
	OnCompile(ctx context.Context, info CompileInfo)
	OnCycleStart(ctx context.Context, info CycleInfo)
	OnCycleEnd(ctx context.Context, info CycleResult)
	OnNodeStart(ctx context.Context, info NodeInfo)
	OnNodeEnd(ctx context.Context, info NodeResult)
	OnError(ctx context.Context, info ErrorInfo)
}
