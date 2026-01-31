package port

import (
	"context"

	"dag-observatory/demo-go/internal/domain/dagruntime/events"
)

type EventEnqueuer interface {
	Enqueue(ctx context.Context, event events.Event) error
}
