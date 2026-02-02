package port

import (
	"context"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

// EventStream provides a unified event source for streaming transports.
type EventStream interface {
	Subscribe(ctx context.Context) (<-chan events.Event, error)
}
