package eventstore

import (
	"context"
	"fmt"

	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) (*KafkaProducer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("eventstore: kafka brokers are required")
	}
	if topic == "" {
		return nil, fmt.Errorf("eventstore: kafka topic is required")
	}
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.Hash{},
	}
	return &KafkaProducer{writer: writer}, nil
}

func (p *KafkaProducer) Enqueue(ctx context.Context, event events.Event) error {
	if p == nil || p.writer == nil {
		return fmt.Errorf("eventstore: producer not configured")
	}
	key, value, err := EncodeEvent(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
		Time:  event.EventTime,
	})
}

func (p *KafkaProducer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
