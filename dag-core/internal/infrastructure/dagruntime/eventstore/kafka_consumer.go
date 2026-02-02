package eventstore

import (
	"context"
	"fmt"

	"dag-observatory/dag-core/internal/domain/dagruntime/driver"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/observability/semantics"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
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

func (c *KafkaConsumer) RunWithContext(ctx context.Context, out chan<- driver.StreamEvent) error {
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
		carrier := headerCarrierFromKafka(msg.Headers)
		msgCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)
		tracer := otel.Tracer("dag-observatory-dag-core")
		spanCtx, span := tracer.Start(msgCtx, fmt.Sprintf("kafka.consume %s", event.Type))
		span.SetAttributes(
			attribute.String(semantics.KeyMessagingSystem, "kafka"),
			attribute.String(semantics.KeyMessagingDestination, c.reader.Config().Topic),
			attribute.String(semantics.KeyMessagingOperation, "receive"),
			attribute.String(semantics.KeyMessagingMessageID, event.EventID),
		)
		select {
		case out <- driver.StreamEvent{Ctx: spanCtx, Event: event}:
			span.End()
		case <-ctx.Done():
			span.End()
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

type headerCarrier map[string]string

func (c headerCarrier) Get(key string) string {
	return c[key]
}

func (c headerCarrier) Set(key string, value string) {
	c[key] = value
}

func (c headerCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

func headerCarrierFromKafka(headers []kafka.Header) propagation.TextMapCarrier {
	carrier := headerCarrier{}
	for _, h := range headers {
		if len(h.Value) == 0 {
			continue
		}
		carrier[h.Key] = string(h.Value)
	}
	return carrier
}
