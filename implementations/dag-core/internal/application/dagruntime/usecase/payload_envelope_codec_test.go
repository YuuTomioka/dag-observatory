package usecase

import (
	"testing"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

func TestDecodeEnvelopesAcceptsInputKeyWithoutStableID(t *testing.T) {
	env, err := events.EncodePayload(events.PayloadKey[map[string]any]{Name: "input"}, map[string]any{
		"symbol": "USDJPY",
		"mode":   "normal",
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	inputs, err := decodeEnvelopes([]events.PayloadEnvelope{env})
	if err != nil {
		t.Fatalf("decode envelopes: %v", err)
	}
	if got, ok := inputs[InputKeySymbol].(string); !ok || got != "USDJPY" {
		t.Fatalf("unexpected symbol input: %#v", inputs[InputKeySymbol])
	}
	if got, ok := inputs[InputKeyMode].(string); !ok || got != "normal" {
		t.Fatalf("unexpected mode input: %#v", inputs[InputKeyMode])
	}
}
