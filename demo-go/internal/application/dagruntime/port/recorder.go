package port

import (
	"context"

	"dag-observatory/demo-go/internal/domain/dagruntime/events"
)

type Recorder interface {
	RecordEvent(ctx context.Context, event events.Event)
	RecordNodeResult(ctx context.Context, result NodeResult)
	RecordCycleResult(ctx context.Context, result CycleResult)
}

