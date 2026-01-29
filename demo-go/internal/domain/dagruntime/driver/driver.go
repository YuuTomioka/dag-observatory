package driver

import (
	"context"

	"dag-observatory/demo-go/internal/domain/dagruntime/engine"
	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"dag-observatory/demo-go/internal/domain/dagruntime/pipeline"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
)

type Driver struct {
	Runner   *engine.Runner
	Compiled pipeline.Compiled
}

func (d *Driver) Run(ctx context.Context, stream <-chan events.Event) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-stream:
			if !ok {
				return nil
			}
			inputs, _ := event.Payload.(engine.InputMap)
			if inputs == nil {
				inputs = engine.InputMap{}
			}
			err := d.Runner.RunCycle(ctx, d.Compiled, inputs, event.Partition, event)
			if err != nil {
				return err
			}
		}
	}
}

func DefaultPartition() state.Partition {
	return state.Partition("default")
}

