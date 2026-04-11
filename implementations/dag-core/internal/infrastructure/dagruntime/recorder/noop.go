package recorder

import (
	"context"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

type NoopRecorder struct{}

func NewNoopRecorder() *NoopRecorder {
	return &NoopRecorder{}
}

func (r *NoopRecorder) RecordEvent(ctx context.Context, event events.Event)            {}
func (r *NoopRecorder) RecordNodeExecution(ctx context.Context, event events.NodeExecutionEvent) {
}
func (r *NoopRecorder) RecordNodeResult(ctx context.Context, result port.NodeResult)   {}
func (r *NoopRecorder) RecordCycleResult(ctx context.Context, result port.CycleResult) {}
