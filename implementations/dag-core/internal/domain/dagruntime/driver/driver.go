package driver

import (
	"context"
	"fmt"

	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type Driver struct {
	Runner   *engine.Runner
	Compiled pipeline.Compiled
	Options  engine.DriverOptions
}

type StreamEvent struct {
	Ctx   context.Context
	Event events.Event
}

func (d *Driver) Run(ctx context.Context, stream <-chan events.Event) error {
	wrapped := make(chan StreamEvent, 1)
	go func() {
		for ev := range stream {
			wrapped <- StreamEvent{Ctx: ctx, Event: ev}
		}
		close(wrapped)
	}()
	return d.RunWithContext(ctx, wrapped)
}

func (d *Driver) RunWithContext(ctx context.Context, stream <-chan StreamEvent) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, ok := <-stream:
			if !ok {
				return nil
			}
			event := item.Event
			runCtx := item.Ctx
			if runCtx == nil {
				runCtx = ctx
			}
			inputs, _ := event.Payload.(engine.InputMap)
			if inputs == nil {
				inputs = engine.InputMap{}
			}
			err := func() (err error) {
				defer func() {
					if recovered := recover(); recovered != nil {
						err = fmt.Errorf("driver: panic while processing event %q: %v", event.EventID, recovered)
					}
				}()
				return d.Runner.RunCycle(runCtx, d.Compiled, inputs, event.Partition, event)
			}()
			if err != nil {
				if d.Options.ContinueOnError {
					continue
				}
				return err
			}
		}
	}
}

func DefaultPartition() state.Partition {
	return state.Partition("default")
}
