package port

import (
	"context"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

type Recorder interface {
	RecordEvent(ctx context.Context, event events.Event)
	RecordNodeResult(ctx context.Context, result NodeResult)
	RecordCycleResult(ctx context.Context, result CycleResult)
}

