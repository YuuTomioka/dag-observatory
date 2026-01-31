package eventstore

import (
	"context"
	"fmt"

	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(brokers []string, topic, groupID string) (*KafkaConsumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("eventstore: kafka brokers are required")
	}
	if topic == "" {
		return nil, fmt.Errorf("eventstore: kafka topic is required")
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	return &KafkaConsumer{reader: reader}, nil
}

func (c *KafkaConsumer) Run(ctx context.Context, out chan<- events.Event) error {
	if c == nil || c.reader == nil {
		return fmt.Errorf("eventstore: consumer not configured")
	}
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		event, err := DecodeEvent(msg.Key, msg.Value)
		if err != nil {
			return err
		}
		select {
		case out <- event:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *KafkaConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
