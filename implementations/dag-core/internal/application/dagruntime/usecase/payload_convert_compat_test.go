package usecase

import (
	"encoding/json"
	"testing"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

func TestToInputMapFromWorkerTaskCompletedEnvelope(t *testing.T) {
	inputBody, err := json.Marshal(map[string]any{
		"symbol": "USDJPY",
		"mode":   "normal",
	})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	payload := []events.PayloadEnvelope{
		{
			Key: events.RawPayloadKey{Name: "input", StableID: ""},
			Data: inputBody,
		},
	}
	inputs, err := toInputMap(payload)
	if err != nil {
		t.Fatalf("to input map: %v", err)
	}
	gotSymbol, ok := inputs[InputKeySymbol].(string)
	if !ok || gotSymbol != "USDJPY" {
		t.Fatalf("expected symbol USDJPY, got %#v", inputs[InputKeySymbol])
	}
	gotMode, ok := inputs[InputKeyMode].(string)
	if !ok || gotMode != "normal" {
		t.Fatalf("expected mode normal, got %#v", inputs[InputKeyMode])
	}
}
