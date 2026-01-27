package port

import (
	"context"

	"dag-observatory/demo-go/internal/domain/observability/semantics"
)

type IntentLog interface {
	ClockTickReceived(ctx context.Context, e semantics.ClockTickReceived)
	DAGRunStarted(ctx context.Context, e semantics.DAGRunStarted)
	DAGRunFinished(ctx context.Context, e semantics.DAGRunFinished)
	DAGNodeStarted(ctx context.Context, e semantics.DAGNodeStarted)
	DAGNodeFinished(ctx context.Context, e semantics.DAGNodeFinished)
}
