package port

import (
	"context"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

type StreamEvent struct {
	Ctx   context.Context
	Event events.Event
}

// EventStream provides a unified event source for streaming transports.
type EventStream interface {
	Subscribe(ctx context.Context) (<-chan StreamEvent, error)
}
