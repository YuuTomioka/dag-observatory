package usecase

import (
	"fmt"

	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

func toInputMap(payload any) (engine.InputMap, error) {
	if payload == nil {
		return engine.InputMap{}, nil
	}
	switch value := payload.(type) {
	case engine.InputMap:
		return value, nil
	case map[string]any:
		return fromStringMap(value)
	case []events.PayloadEnvelope:
		return decodeEnvelopes(value)
	default:
		return nil, fmt.Errorf("dagruntime: unsupported payload type %T", payload)
	}
}

// ToInputMap normalizes transport payloads into runtime input map.
func ToInputMap(payload any) (engine.InputMap, error) {
	return toInputMap(payload)
}

func toPayloadEnvelopes(payload any) ([]events.PayloadEnvelope, error) {
	if payload == nil {
		return nil, nil
	}
	switch value := payload.(type) {
	case []events.PayloadEnvelope:
		return value, nil
	case engine.InputMap:
		return fromInputMap(value)
	case map[string]any:
		return encodeEnvelopes(value)
	default:
		return nil, fmt.Errorf("dagruntime: unsupported payload type %T", payload)
	}
}
