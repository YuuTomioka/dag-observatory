package port

import (
	"context"

	"dag-observatory/dag-core/internal/domain/observability/semantics"
)

type IntentLog interface {
	ClockTickReceived(ctx context.Context, e semantics.ClockTickReceived)
	DAGRunStarted(ctx context.Context, e semantics.DAGRunStarted)
	DAGRunFinished(ctx context.Context, e semantics.DAGRunFinished)
	DAGRunFailed(ctx context.Context, e semantics.DAGRunFailed)
	DAGRunStateChanged(ctx context.Context, e semantics.DAGRunStateChanged)
	DAGNodeStarted(ctx context.Context, e semantics.DAGNodeStarted)
	DAGNodeFinished(ctx context.Context, e semantics.DAGNodeFinished)
	DAGNodeFailed(ctx context.Context, e semantics.DAGNodeFailed)
	DAGNodeTimeout(ctx context.Context, e semantics.DAGNodeTimeout)
	DAGNodeSkipped(ctx context.Context, e semantics.DAGNodeSkipped)
	DAGNodeStateChanged(ctx context.Context, e semantics.DAGNodeStateChanged)
}
