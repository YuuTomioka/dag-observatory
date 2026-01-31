package eventstore

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
)

func TestKafkaRoundTrip(t *testing.T) {
	brokers := splitCSV(os.Getenv("KAFKA_BROKERS"))
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if len(brokers) == 0 || topic == "" {
		t.Skip("KAFKA_BROKERS and KAFKA_TOPIC are required for Kafka integration test")
	}
	if groupID == "" {
		groupID = "dagruntime-test"
	}

	producer, err := NewKafkaProducer(brokers, topic)
	if err != nil {
		t.Fatalf("producer init failed: %v", err)
	}
	defer producer.Close()

	consumer, err := NewKafkaConsumer(brokers, topic, groupID)
	if err != nil {
		t.Fatalf("consumer init failed: %v", err)
	}
	defer consumer.Close()

	envSymbol, err := events.EncodePayload(events.PayloadKey[string]{
		Name:     "symbol",
		StableID: "event:input.symbol.v1",
	}, "EURUSD")
	if err != nil {
		t.Fatalf("encode symbol failed: %v", err)
	}
	envMode, err := events.EncodePayload(events.PayloadKey[string]{
		Name:     "mode",
		StableID: "event:input.mode.v1",
	}, "debug")
	if err != nil {
		t.Fatalf("encode mode failed: %v", err)
	}

	event := events.Event{
		EventID:   "test-event-1",
		EventTime: time.Now().UTC(),
		Partition: state.Partition("EURUSD"),
		Type:      "test.kafka.roundtrip",
		Payload:   []events.PayloadEnvelope{envSymbol, envMode},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out := make(chan events.Event, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- consumer.Run(ctx, out)
	}()

	if err := producer.Enqueue(ctx, event); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	select {
	case got := <-out:
		assertEventEqual(t, event, got)
	case err := <-errCh:
		if err != nil && err != context.DeadlineExceeded {
			t.Fatalf("consumer error: %v", err)
		}
		t.Fatal("no event received before consumer stopped")
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
}

func assertEventEqual(t *testing.T, want events.Event, got events.Event) {
	t.Helper()
	if want.EventID != got.EventID {
		t.Fatalf("event id mismatch: want %s got %s", want.EventID, got.EventID)
	}
	if want.Partition != got.Partition {
		t.Fatalf("partition mismatch: want %s got %s", want.Partition, got.Partition)
	}
	if want.Type != got.Type {
		t.Fatalf("type mismatch: want %s got %s", want.Type, got.Type)
	}
	if got.EventTime.IsZero() {
		t.Fatalf("event time is zero")
	}

	wantPayload, _ := want.Payload.([]events.PayloadEnvelope)
	gotPayload, _ := got.Payload.([]events.PayloadEnvelope)
	if len(wantPayload) != len(gotPayload) {
		t.Fatalf("payload length mismatch: want %d got %d", len(wantPayload), len(gotPayload))
	}
	for i := range wantPayload {
		if wantPayload[i].Key != gotPayload[i].Key {
			t.Fatalf("payload key mismatch at %d: want %v got %v", i, wantPayload[i].Key, gotPayload[i].Key)
		}
		if string(wantPayload[i].Data) != string(gotPayload[i].Data) {
			t.Fatalf("payload data mismatch at %d", i)
		}
	}
}

func splitCSV(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
