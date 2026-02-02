package eventstore

import (
	"context"
	"fmt"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

// KafkaEventStream adapts KafkaConsumer to the application EventStream port.
type KafkaEventStream struct {
	consumer *KafkaConsumer
	buffer   int
}

func NewKafkaEventStream(consumer *KafkaConsumer) *KafkaEventStream {
	return &KafkaEventStream{consumer: consumer, buffer: 32}
}

func (s *KafkaEventStream) Subscribe(ctx context.Context) (<-chan events.Event, error) {
	if s == nil || s.consumer == nil {
		return nil, fmt.Errorf("eventstream: consumer not configured")
	}
	out := make(chan events.Event, s.buffer)
	go func() {
		defer close(out)
		_ = s.consumer.Run(ctx, out)
	}()
	return out, nil
}
