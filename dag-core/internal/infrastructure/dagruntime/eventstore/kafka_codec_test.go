package eventstore

import (
	"reflect"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

func TestEncodeDecodeEvent(t *testing.T) {
	envSymbol, err := events.EncodePayload(events.PayloadKey[string]{
		Name:     "symbol",
		StableID: "event:input.symbol.v1",
	}, "USDJPY")
	if err != nil {
		t.Fatalf("encode symbol failed: %v", err)
	}
	envMode, err := events.EncodePayload(events.PayloadKey[string]{
		Name:     "mode",
		StableID: "event:input.mode.v1",
	}, "normal")
	if err != nil {
		t.Fatalf("encode mode failed: %v", err)
	}

	event := events.Event{
		EventID:   "evt-1",
		EventTime: time.Now().UTC(),
		Partition: state.Partition("USDJPY"),
		Type:      "http.dag.run",
		Payload:   []events.PayloadEnvelope{envSymbol, envMode},
	}

	key, value, err := EncodeEvent(event)
	if err != nil {
		t.Fatalf("encode event failed: %v", err)
	}
	if string(key) != "USDJPY" {
		t.Fatalf("unexpected key: %s", string(key))
	}

	got, err := DecodeEvent(key, value)
	if err != nil {
		t.Fatalf("decode event failed: %v", err)
	}
	if got.EventID != event.EventID {
		t.Fatalf("event id mismatch: want %s got %s", event.EventID, got.EventID)
	}
	if got.Partition != event.Partition {
		t.Fatalf("partition mismatch: want %s got %s", event.Partition, got.Partition)
	}
	if got.Type != event.Type {
		t.Fatalf("type mismatch: want %s got %s", event.Type, got.Type)
	}
	gotPayload, ok := got.Payload.([]events.PayloadEnvelope)
	if !ok {
		t.Fatalf("unexpected payload type %T", got.Payload)
	}
	wantPayload := event.Payload.([]events.PayloadEnvelope)
	if !reflect.DeepEqual(wantPayload, gotPayload) {
		t.Fatalf("payload mismatch")
	}
}
