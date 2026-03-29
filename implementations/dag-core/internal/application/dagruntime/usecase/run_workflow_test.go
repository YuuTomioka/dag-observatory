package usecase

import (
	"testing"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

func TestToInputMapFromStringMap(t *testing.T) {
	payload := map[string]any{
		"symbol":      "USDJPY",
		"mode":        "normal",
		"market_bars": []any{1.0, 2.0, 3.0},
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
	bars, ok := inputs[InputKeyMarketBars].([]float64)
	if !ok || len(bars) != 3 {
		t.Fatalf("expected market bars to be mapped")
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
	envBars, err := events.EncodePayload(PayloadKeyMarketBars, []float64{1.0, 2.0})
	if err != nil {
		t.Fatalf("encode bars failed: %v", err)
	}

	inputs, err := toInputMap([]events.PayloadEnvelope{envSymbol, envMode, envBars})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inputs[InputKeySymbol] != "EURUSD" {
		t.Fatalf("expected symbol from payload")
	}
	if inputs[InputKeyMode] != "debug" {
		t.Fatalf("expected mode from payload")
	}
	bars, ok := inputs[InputKeyMarketBars].([]float64)
	if !ok || len(bars) != 2 {
		t.Fatalf("expected market bars from payload")
	}
}
