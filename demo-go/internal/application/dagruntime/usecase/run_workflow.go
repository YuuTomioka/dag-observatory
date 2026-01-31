package usecase

import (
	"context"
	"errors"
	"fmt"

	"dag-observatory/demo-go/internal/application/dagruntime/port"
	"dag-observatory/demo-go/internal/domain/dagruntime/driver"
	"dag-observatory/demo-go/internal/domain/dagruntime/engine"
	"dag-observatory/demo-go/internal/domain/dagruntime/events"
	"dag-observatory/demo-go/internal/domain/dagruntime/state"
)

type RunWorkflow struct {
	Driver   *driver.Driver
	Enqueuer port.EventEnqueuer
}

func (u *RunWorkflow) Handle(ctx context.Context, partition state.Partition, event events.Event) error {
	if u.Enqueuer != nil {
		payload, err := toPayloadEnvelopes(event.Payload)
		if err != nil {
			return err
		}
		event.Payload = payload
		event.Partition = partition
		return u.Enqueuer.Enqueue(ctx, event)
	}
	if u.Driver == nil {
		return errors.New("dagruntime: driver not configured")
	}
	inputs, err := toInputMap(event.Payload)
	if err != nil {
		return err
	}
	event.Payload = inputs
	event.Partition = partition
	stream := make(chan events.Event, 1)
	stream <- event
	close(stream)
	return u.Driver.Run(ctx, stream)
}

func (u *RunWorkflow) IsEnqueueMode() bool {
	return u != nil && u.Enqueuer != nil
}

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

func fromStringMap(values map[string]any) (engine.InputMap, error) {
	inputs := engine.InputMap{}
	if raw, ok := values["symbol"]; ok {
		symbol, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("dagruntime: symbol must be string")
		}
		inputs[InputKeySymbol] = symbol
	}
	if raw, ok := values["mode"]; ok {
		mode, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("dagruntime: mode must be string")
		}
		inputs[InputKeyMode] = mode
	}
	return inputs, nil
}

func fromInputMap(inputs engine.InputMap) ([]events.PayloadEnvelope, error) {
	envelopes := make([]events.PayloadEnvelope, 0, 2)
	for key, value := range inputs {
		if key.Raw() == InputKeySymbol.Raw() {
			symbol, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("dagruntime: symbol must be string")
			}
			env, err := events.EncodePayload(PayloadKeySymbol, symbol)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
		if key.Raw() == InputKeyMode.Raw() {
			mode, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("dagruntime: mode must be string")
			}
			env, err := events.EncodePayload(PayloadKeyMode, mode)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
	}
	return envelopes, nil
}

func encodeEnvelopes(values map[string]any) ([]events.PayloadEnvelope, error) {
	envelopes := make([]events.PayloadEnvelope, 0, 2)
	if raw, ok := values["symbol"]; ok {
		symbol, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("dagruntime: symbol must be string")
		}
		env, err := events.EncodePayload(PayloadKeySymbol, symbol)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["mode"]; ok {
		mode, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("dagruntime: mode must be string")
		}
		env, err := events.EncodePayload(PayloadKeyMode, mode)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	return envelopes, nil
}

func decodeEnvelopes(envelopes []events.PayloadEnvelope) (engine.InputMap, error) {
	inputs := engine.InputMap{}
	for _, env := range envelopes {
		if env.Key == PayloadKeySymbol.Raw() {
			value, err := events.DecodePayload(PayloadKeySymbol, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeySymbol] = value
			continue
		}
		if env.Key == PayloadKeyMode.Raw() {
			value, err := events.DecodePayload(PayloadKeyMode, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMode] = value
			continue
		}
	}
	return inputs, nil
}
