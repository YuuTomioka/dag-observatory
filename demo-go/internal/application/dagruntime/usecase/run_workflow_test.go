package usecase

import (
	"testing"

	"dag-observatory/demo-go/internal/domain/dagruntime/events"
)

func TestToInputMapFromStringMap(t *testing.T) {
	payload := map[string]any{
		"symbol": "USDJPY",
		"mode":   "normal",
	}

	inputs, err := toInputMap(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inputs[InputKeySymbol] != "USDJPY" {
		t.Fatalf("expected symbol to be mapped")
	}
	if inputs[InputKeyMode] != "normal" {
		t.Fatalf("expected mode to be mapped")
	}
}

func TestToInputMapFromPayloadEnvelopes(t *testing.T) {
	envSymbol, err := events.EncodePayload(PayloadKeySymbol, "EURUSD")
	if err != nil {
		t.Fatalf("encode symbol failed: %v", err)
	}
	envMode, err := events.EncodePayload(PayloadKeyMode, "debug")
	if err != nil {
		t.Fatalf("encode mode failed: %v", err)
	}

	inputs, err := toInputMap([]events.PayloadEnvelope{envSymbol, envMode})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inputs[InputKeySymbol] != "EURUSD" {
		t.Fatalf("expected symbol from payload")
	}
	if inputs[InputKeyMode] != "debug" {
		t.Fatalf("expected mode from payload")
	}
}
