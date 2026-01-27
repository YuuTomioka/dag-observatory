package port

import (
	"context"

	"dag-observatory/demo-go/internal/infrastructure/observability/intentlog"
)

type IntentLog interface {
	DAGRunStarted(ctx context.Context, e intentlog.DAGRunEvent)
	DAGRunFinished(ctx context.Context, e intentlog.DAGRunEvent)
}
