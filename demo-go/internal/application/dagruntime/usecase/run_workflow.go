package usecase

import (
	"context"

	"dag-observatory/demo-go/internal/domain/dagruntime/engine"
	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"dag-observatory/demo-go/internal/domain/dagruntime/pipeline"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
)

type RunWorkflow struct {
	Runner   *engine.Runner
	Compiled pipeline.Compiled
}

func (u *RunWorkflow) Handle(ctx context.Context, inputs engine.InputMap, partition state.Partition, event events.Event) error {
	return u.Runner.RunCycle(ctx, u.Compiled, inputs, partition, event)
}

