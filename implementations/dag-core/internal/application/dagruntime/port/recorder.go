package port

import (
	"context"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

type Recorder interface {
	RecordEvent(ctx context.Context, event events.Event)
	RecordNodeExecution(ctx context.Context, event events.NodeExecutionEvent)
	RecordNodeResult(ctx context.Context, result NodeResult)
	RecordCycleResult(ctx context.Context, result CycleResult)
}
