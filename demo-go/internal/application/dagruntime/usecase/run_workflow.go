package usecase

import (
	"context"

	"dag-observatory/demo-go/internal/domain/dagruntime/driver"
	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
)

type RunWorkflow struct {
	Driver *driver.Driver
}

func (u *RunWorkflow) Handle(ctx context.Context, partition state.Partition, event events.Event) error {
	if u.Driver == nil {
		return nil
	}
	stream := make(chan events.Event, 1)
	stream <- event
	close(stream)
	return u.Driver.Run(ctx, stream)
}
