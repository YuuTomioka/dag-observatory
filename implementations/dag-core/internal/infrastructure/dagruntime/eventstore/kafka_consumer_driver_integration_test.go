package eventstore

import (
	"context"
	"os"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/driver"
	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/pipeline"
	"dag-observatory/dag-core/internal/domain/dagruntime/policy"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	artifactinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/artifact"
	stateinfra "dag-observatory/dag-core/internal/infrastructure/dagruntime/state"
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
