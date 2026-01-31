package port

import (
	"context"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

type EventEnqueuer interface {
	Enqueue(ctx context.Context, event events.Event) error
}
