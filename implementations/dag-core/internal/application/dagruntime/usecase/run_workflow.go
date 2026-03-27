package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/driver"
	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
	"dag-observatory/dag-core/internal/domain/observability/ctxprop"

	"github.com/google/uuid"
)

const (
	defaultSymbol   = "USDJPY"
	defaultMode     = "normal"
	defaultTaskID   = "heavy_calc"
	defaultTaskName = "heavy_calc"
	defaultAttempt  = 1
	defaultEventType = "task.requested"
)

type RunWorkflow struct {
	Driver   *driver.Driver
	Enqueuer port.EventEnqueuer
}

func (u *RunWorkflow) Execute(ctx context.Context, req RunWorkflowRequest) (RunWorkflowResult, error) {
	result, event, partition := buildRunEvent(req)
	ctx = ctxprop.WithRunID(ctx, result.RunID)
	if err := u.Handle(ctx, partition, event); err != nil {
		return result, err
	}
	result.EnqueueMode = u.IsEnqueueMode()
	return result, nil
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

func buildRunEvent(req RunWorkflowRequest) (RunWorkflowResult, events.Event, state.Partition) {
	runID := req.RunID
	if runID == "" {
		runID = uuid.NewString()
	}
	symbol := req.Symbol
	if symbol == "" {
		symbol = defaultSymbol
	}
	mode := req.Mode
	if mode == "" {
		mode = defaultMode
	}

	result := RunWorkflowResult{
		RunID:    runID,
		Symbol:   symbol,
		Mode:     mode,
		TaskID:   defaultTaskID,
		TaskName: defaultTaskName,
		Attempt:  defaultAttempt,
	}

	payload := map[string]any{
		"run_id":    runID,
		"task_id":   defaultTaskID,
		"attempt":   defaultAttempt,
		"task_name": defaultTaskName,
		"input": map[string]any{
			"mode":   mode,
			"symbol": symbol,
		},
		// Keep top-level keys for direct execution compatibility.
		"mode":   mode,
		"symbol": symbol,
	}

	partition := state.Partition(runID)
	event := events.Event{
		EventID:   runID,
		EventTime: time.Now(),
		Partition: partition,
		Type:      defaultEventType,
		Payload:   payload,
	}

	return result, event, partition
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
	if raw, ok := values["input"]; ok {
		inputMap, ok := raw.(map[string]any)
		if ok {
			if rawSymbol, ok := inputMap["symbol"]; ok {
				symbol, ok := rawSymbol.(string)
				if !ok {
					return nil, fmt.Errorf("dagruntime: symbol must be string")
				}
				inputs[InputKeySymbol] = symbol
			}
			if rawMode, ok := inputMap["mode"]; ok {
				mode, ok := rawMode.(string)
				if !ok {
					return nil, fmt.Errorf("dagruntime: mode must be string")
				}
				inputs[InputKeyMode] = mode
			}
		}
	}
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
	envelopes := make([]events.PayloadEnvelope, 0, 8)
	if raw, ok := values["run_id"]; ok {
		if runID, ok := raw.(string); ok {
			env, err := events.EncodePayload(PayloadKeyRunID, runID)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
		}
	}
	if raw, ok := values["task_id"]; ok {
		if taskID, ok := raw.(string); ok {
			env, err := events.EncodePayload(PayloadKeyTaskID, taskID)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
		}
	}
	if raw, ok := values["attempt"]; ok {
		if attempt, ok := raw.(int); ok {
			env, err := events.EncodePayload(PayloadKeyAttempt, attempt)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
		}
	}
	if raw, ok := values["task_name"]; ok {
		if taskName, ok := raw.(string); ok {
			env, err := events.EncodePayload(PayloadKeyTaskName, taskName)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
		}
	}
	if raw, ok := values["input"]; ok {
		if input, ok := raw.(map[string]any); ok {
			env, err := events.EncodePayload(PayloadKeyInput, input)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
		}
	}
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
		if env.Key == PayloadKeyInput.Raw() {
			value, err := events.DecodePayload(PayloadKeyInput, env)
			if err != nil {
				return nil, err
			}
			if raw, ok := value["symbol"]; ok {
				if symbol, ok := raw.(string); ok {
					inputs[InputKeySymbol] = symbol
				} else {
					return nil, fmt.Errorf("dagruntime: symbol must be string")
				}
			}
			if raw, ok := value["mode"]; ok {
				if mode, ok := raw.(string); ok {
					inputs[InputKeyMode] = mode
				} else {
					return nil, fmt.Errorf("dagruntime: mode must be string")
				}
			}
			continue
		}
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
