package eventstore

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"dag-observatory/demo-go/internal/domain/dagruntime/engine"
	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"dag-observatory/demo-go/internal/domain/dagruntime/pipeline"
	"dag-observatory/demo-go/internal/domain/dagruntime/policy"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
	stateinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/state"
	artifactinfra "dag-observatory/demo-go/internal/infrastructure/dagruntime/artifact"
	"dag-observatory/demo-go/internal/domain/dagruntime/driver"
)

func TestKafkaConsumerToDriver(t *testing.T) {
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
	}, "USDJPY")
	if err != nil {
		t.Fatalf("encode symbol failed: %v", err)
	}

	event := events.Event{
		EventID:   "evt-1",
		EventTime: time.Now().UTC(),
		Partition: state.Partition("USDJPY"),
		Type:      "http.dag.run",
		Payload:   []events.PayloadEnvelope{envSymbol},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out := make(chan events.Event, 1)
	go func() {
		_ = consumer.Run(ctx, out)
	}()

	if err := producer.Enqueue(ctx, event); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	var got events.Event
	select {
	case got = <-out:
	case <-ctx.Done():
		t.Fatal("timeout waiting for event")
	}
	cancel()

	stream := make(chan events.Event, 1)
	stream <- got
	close(stream)

	driver := &driver.Driver{
		Runner: &engine.Runner{
			ArtifactStore: artifactinfra.NewMemoryStore(),
			StateStore:    stateinfra.NewMemoryStore(),
			Policy: policy.Policy{
				DefaultRetry: policy.RetryPolicy{MaxAttempts: 1},
			},
		},
		Compiled: pipeline.Compiled{Name: "test"},
	}

	if err := driver.Run(context.Background(), stream); err != nil {
		t.Fatalf("driver run failed: %v", err)
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
