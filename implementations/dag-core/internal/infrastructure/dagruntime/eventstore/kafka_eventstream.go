package eventstore

import (
	"context"
	"fmt"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/driver"
)

// KafkaEventStream adapts KafkaConsumer to the application EventStream port.
type KafkaEventStream struct {
	consumer *KafkaConsumer
	buffer   int
}

func NewKafkaEventStream(consumer *KafkaConsumer) *KafkaEventStream {
	return &KafkaEventStream{consumer: consumer, buffer: 32}
}

func (s *KafkaEventStream) Subscribe(ctx context.Context) (<-chan port.StreamEvent, error) {
	if s == nil || s.consumer == nil {
		return nil, fmt.Errorf("eventstream: consumer not configured")
	}
	raw := make(chan driver.StreamEvent, s.buffer)
	out := make(chan port.StreamEvent, s.buffer)
	go func() {
		defer close(out)
		for item := range raw {
			out <- port.StreamEvent{Ctx: item.Ctx, Event: item.Event}
		}
	}()
	go func() {
		defer close(raw)
		_ = s.consumer.RunWithContext(ctx, raw)
	}()
	return out, nil
}
