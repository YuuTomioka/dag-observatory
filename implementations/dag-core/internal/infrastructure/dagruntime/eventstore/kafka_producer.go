package eventstore

import (
	"context"
	"fmt"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/observability/semantics"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
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
	tracer := otel.Tracer("dag-observatory-dag-core")
	spanCtx, span := tracer.Start(ctx, fmt.Sprintf("kafka.produce %s", event.Type))
	span.SetAttributes(
		attribute.String(semantics.KeyMessagingSystem, "kafka"),
		attribute.String(semantics.KeyMessagingDestination, p.writer.Topic),
		attribute.String(semantics.KeyMessagingOperation, "send"),
		attribute.String(semantics.KeyMessagingMessageID, event.EventID),
	)
	defer span.End()

	key, value, err := EncodeEvent(event)
	if err != nil {
		return err
	}
	headers := kafkaHeadersFromCarrier(spanCtx)
	return p.writer.WriteMessages(spanCtx, kafka.Message{
		Key:   key,
		Value: value,
		Time:  event.EventTime,
		Headers: headers,
	})
}

func (p *KafkaProducer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

func kafkaHeadersFromCarrier(ctx context.Context) []kafka.Header {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if len(carrier) == 0 {
		return nil
	}
	headers := make([]kafka.Header, 0, len(carrier))
	for key, value := range carrier {
		if value == "" {
			continue
		}
		headers = append(headers, kafka.Header{Key: key, Value: []byte(value)})
	}
	return headers
}
