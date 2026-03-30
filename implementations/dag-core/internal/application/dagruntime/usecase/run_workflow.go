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
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/domain/observability/ctxprop"

	"github.com/google/uuid"
)

const (
	defaultSymbol                = "USDJPY"
	defaultMode                  = "normal"
	defaultTaskID                = "heavy_calc"
	defaultTaskName              = "heavy_calc"
	defaultAttempt               = 1
	defaultEventType             = "task.requested"
	defaultDriverDestinationName = "dagruntime-driver"
)

type RunWorkflow struct {
	Driver        *driver.Driver
	Enqueuer      port.EventEnqueuer
	Clock         port.Clock
	ProducerTopic string
}

func (u *RunWorkflow) Execute(ctx context.Context, req RunWorkflowRequest) (RunWorkflowResult, error) {
	result, event, partition := buildRunEvent(req, u.now())
	ctx = ctxprop.WithRunID(ctx, result.RunID)
	if err := u.Handle(ctx, partition, event); err != nil {
		return result, err
	}
	result.EnqueueMode = u.IsEnqueueMode()
	result.MessagingDestination = u.messagingDestination(result.EnqueueMode)
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

func (u *RunWorkflow) messagingDestination(enqueueMode bool) string {
	if enqueueMode {
		if u != nil && u.ProducerTopic != "" {
			return u.ProducerTopic
		}
		return "unknown"
	}
	return defaultDriverDestinationName
}

func (u *RunWorkflow) now() time.Time {
	if u != nil && u.Clock != nil {
		return u.Clock.Now()
	}
	return time.Now()
}

func buildRunEvent(req RunWorkflowRequest, now time.Time) (RunWorkflowResult, events.Event, state.Partition) {
	runID := req.RunID
	if runID == "" {
		runID = uuid.NewString()
	}
	symbol := req.Symbol
	if symbol == "" && req.Marketdata != nil && req.Marketdata.SymbolCode != "" {
		symbol = req.Marketdata.SymbolCode
	}
	if symbol == "" && req.Marketdata == nil {
		symbol = defaultSymbol
	}
	mode := req.Mode
	if mode == "" {
		mode = defaultMode
	}

	result := RunWorkflowResult{
		RunID:      runID,
		Symbol:     symbol,
		Mode:       mode,
		Marketdata: cloneMarketdataInput(req.Marketdata),
		TaskID:     defaultTaskID,
		TaskName:   defaultTaskName,
		Attempt:    defaultAttempt,
	}

	payload := map[string]any{
		"run_id":    runID,
		"task_id":   defaultTaskID,
		"attempt":   defaultAttempt,
		"task_name": defaultTaskName,
		"input":     map[string]any{},
		// Keep top-level keys for direct execution compatibility.
		"mode": mode,
	}
	payload["input"].(map[string]any)["mode"] = mode
	if symbol != "" {
		payload["input"].(map[string]any)["symbol"] = symbol
		payload["symbol"] = symbol
	}
	if len(req.Bars) > 0 {
		payload["input"].(map[string]any)["bars"] = req.Bars
		payload["market_bars"] = req.Bars
	}
	if req.Marketdata != nil {
		if req.Marketdata.SymbolID > 0 {
			payload["input"].(map[string]any)["marketdata.symbol_id"] = req.Marketdata.SymbolID
			payload["marketdata.symbol_id"] = req.Marketdata.SymbolID
		}
		if req.Marketdata.SymbolCode != "" {
			payload["input"].(map[string]any)["marketdata.symbol_code"] = req.Marketdata.SymbolCode
			payload["marketdata.symbol_code"] = req.Marketdata.SymbolCode
		}
		if req.Marketdata.TimeframeCode != "" {
			payload["input"].(map[string]any)["marketdata.timeframe_code"] = req.Marketdata.TimeframeCode
			payload["marketdata.timeframe_code"] = req.Marketdata.TimeframeCode
		}
		if !req.Marketdata.From.IsZero() {
			payload["input"].(map[string]any)["marketdata.from"] = req.Marketdata.From
			payload["marketdata.from"] = req.Marketdata.From
		}
		if !req.Marketdata.To.IsZero() {
			payload["input"].(map[string]any)["marketdata.to"] = req.Marketdata.To
			payload["marketdata.to"] = req.Marketdata.To
		}
	}

	partition := state.Partition(runID)
	event := events.Event{
		EventID:   runID,
		EventTime: now,
		Partition: partition,
		Type:      defaultEventType,
		Payload:   payload,
	}

	return result, event, partition
}

func cloneMarketdataInput(input *MarketdataRunInput) *MarketdataRunInput {
	if input == nil {
		return nil
	}
	cloned := *input
	return &cloned
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
			if rawBars, ok := inputMap["bars"]; ok {
				bars, err := parseFloat64Slice(rawBars)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketBars] = bars
			}
			if raw, ok := inputMap["marketdata.symbol_id"]; ok {
				symbolID, err := parseInt64Value(raw, "marketdata.symbol_id")
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketdataSymbolID] = symbolID
			}
			if raw, ok := inputMap["marketdata.symbol_code"]; ok {
				symbolCode, ok := raw.(string)
				if !ok {
					return nil, fmt.Errorf("dagruntime: marketdata.symbol_code must be string")
				}
				inputs[InputKeyMarketdataSymbolCode] = symbolCode
			}
			if raw, ok := inputMap["marketdata.timeframe_code"]; ok {
				timeframeCode, ok := raw.(string)
				if !ok {
					return nil, fmt.Errorf("dagruntime: marketdata.timeframe_code must be string")
				}
				inputs[InputKeyMarketdataTimeframeCode] = timeframeCode
			}
			if raw, ok := inputMap["marketdata.from"]; ok {
				from, err := parseUTCTimeValue(raw, "marketdata.from")
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketdataFrom] = from
			}
			if raw, ok := inputMap["marketdata.to"]; ok {
				to, err := parseUTCTimeValue(raw, "marketdata.to")
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketdataTo] = to
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
	if raw, ok := values["market_bars"]; ok {
		bars, err := parseFloat64Slice(raw)
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketBars] = bars
	}
	if raw, ok := values["marketdata.symbol_id"]; ok {
		symbolID, err := parseInt64Value(raw, "marketdata.symbol_id")
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketdataSymbolID] = symbolID
	}
	if raw, ok := values["marketdata.symbol_code"]; ok {
		symbolCode, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("dagruntime: marketdata.symbol_code must be string")
		}
		inputs[InputKeyMarketdataSymbolCode] = symbolCode
	}
	if raw, ok := values["marketdata.timeframe_code"]; ok {
		timeframeCode, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("dagruntime: marketdata.timeframe_code must be string")
		}
		inputs[InputKeyMarketdataTimeframeCode] = timeframeCode
	}
	if raw, ok := values["marketdata.from"]; ok {
		from, err := parseUTCTimeValue(raw, "marketdata.from")
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketdataFrom] = from
	}
	if raw, ok := values["marketdata.to"]; ok {
		to, err := parseUTCTimeValue(raw, "marketdata.to")
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketdataTo] = to
	}
	return inputs, nil
}

func fromInputMap(inputs engine.InputMap) ([]events.PayloadEnvelope, error) {
	envelopes := make([]events.PayloadEnvelope, 0, 8)
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
		if key.Raw() == InputKeyMarketBars.Raw() {
			bars, ok := value.([]float64)
			if !ok {
				return nil, fmt.Errorf("dagruntime: market_bars must be []float64")
			}
			env, err := events.EncodePayload(PayloadKeyMarketBars, bars)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
		if key.Raw() == InputKeyMarketdataSymbolID.Raw() {
			symbolID, ok := value.(int64)
			if !ok {
				return nil, fmt.Errorf("dagruntime: marketdata.symbol_id must be int64")
			}
			env, err := events.EncodePayload(PayloadKeyMarketdataSymbolID, symbolID)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
		if key.Raw() == InputKeyMarketdataSymbolCode.Raw() {
			symbolCode, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("dagruntime: marketdata.symbol_code must be string")
			}
			env, err := events.EncodePayload(PayloadKeyMarketdataSymbolCode, symbolCode)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
		if key.Raw() == InputKeyMarketdataTimeframeCode.Raw() {
			timeframeCode, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("dagruntime: marketdata.timeframe_code must be string")
			}
			env, err := events.EncodePayload(PayloadKeyMarketdataTimeframeCode, timeframeCode)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
		if key.Raw() == InputKeyMarketdataFrom.Raw() {
			from, ok := value.(marketdata.UTCTime)
			if !ok {
				return nil, fmt.Errorf("dagruntime: marketdata.from must be UTCTime")
			}
			env, err := events.EncodePayload(PayloadKeyMarketdataFrom, from)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
		if key.Raw() == InputKeyMarketdataTo.Raw() {
			to, ok := value.(marketdata.UTCTime)
			if !ok {
				return nil, fmt.Errorf("dagruntime: marketdata.to must be UTCTime")
			}
			env, err := events.EncodePayload(PayloadKeyMarketdataTo, to)
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
	if raw, ok := values["market_bars"]; ok {
		bars, err := parseFloat64Slice(raw)
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyMarketBars, bars)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["marketdata.symbol_id"]; ok {
		symbolID, err := parseInt64Value(raw, "marketdata.symbol_id")
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyMarketdataSymbolID, symbolID)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["marketdata.symbol_code"]; ok {
		symbolCode, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("dagruntime: marketdata.symbol_code must be string")
		}
		env, err := events.EncodePayload(PayloadKeyMarketdataSymbolCode, symbolCode)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["marketdata.timeframe_code"]; ok {
		timeframeCode, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("dagruntime: marketdata.timeframe_code must be string")
		}
		env, err := events.EncodePayload(PayloadKeyMarketdataTimeframeCode, timeframeCode)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["marketdata.from"]; ok {
		from, err := parseUTCTimeValue(raw, "marketdata.from")
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyMarketdataFrom, from)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["marketdata.to"]; ok {
		to, err := parseUTCTimeValue(raw, "marketdata.to")
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyMarketdataTo, to)
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
			if raw, ok := value["bars"]; ok {
				bars, err := parseFloat64Slice(raw)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketBars] = bars
			}
			if raw, ok := value["marketdata.symbol_id"]; ok {
				symbolID, err := parseInt64Value(raw, "marketdata.symbol_id")
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketdataSymbolID] = symbolID
			}
			if raw, ok := value["marketdata.symbol_code"]; ok {
				symbolCode, ok := raw.(string)
				if !ok {
					return nil, fmt.Errorf("dagruntime: marketdata.symbol_code must be string")
				}
				inputs[InputKeyMarketdataSymbolCode] = symbolCode
			}
			if raw, ok := value["marketdata.timeframe_code"]; ok {
				timeframeCode, ok := raw.(string)
				if !ok {
					return nil, fmt.Errorf("dagruntime: marketdata.timeframe_code must be string")
				}
				inputs[InputKeyMarketdataTimeframeCode] = timeframeCode
			}
			if raw, ok := value["marketdata.from"]; ok {
				from, err := parseUTCTimeValue(raw, "marketdata.from")
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketdataFrom] = from
			}
			if raw, ok := value["marketdata.to"]; ok {
				to, err := parseUTCTimeValue(raw, "marketdata.to")
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketdataTo] = to
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
		if env.Key == PayloadKeyMarketBars.Raw() {
			value, err := events.DecodePayload(PayloadKeyMarketBars, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMarketBars] = value
			continue
		}
		if env.Key == PayloadKeyMarketdataSymbolID.Raw() {
			value, err := events.DecodePayload(PayloadKeyMarketdataSymbolID, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMarketdataSymbolID] = value
			continue
		}
		if env.Key == PayloadKeyMarketdataSymbolCode.Raw() {
			value, err := events.DecodePayload(PayloadKeyMarketdataSymbolCode, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMarketdataSymbolCode] = value
			continue
		}
		if env.Key == PayloadKeyMarketdataTimeframeCode.Raw() {
			value, err := events.DecodePayload(PayloadKeyMarketdataTimeframeCode, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMarketdataTimeframeCode] = value
			continue
		}
		if env.Key == PayloadKeyMarketdataFrom.Raw() {
			value, err := events.DecodePayload(PayloadKeyMarketdataFrom, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMarketdataFrom] = value
			continue
		}
		if env.Key == PayloadKeyMarketdataTo.Raw() {
			value, err := events.DecodePayload(PayloadKeyMarketdataTo, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMarketdataTo] = value
			continue
		}
	}
	return inputs, nil
}

func parseInt64Value(raw any, name string) (int64, error) {
	switch value := raw.(type) {
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	case float64:
		return int64(value), nil
	default:
		return 0, fmt.Errorf("dagruntime: %s must be int64", name)
	}
}

func parseUTCTimeValue(raw any, name string) (marketdata.UTCTime, error) {
	switch value := raw.(type) {
	case marketdata.UTCTime:
		return value, nil
	case string:
		var t marketdata.UTCTime
		if err := t.UnmarshalJSON([]byte(`"` + value + `"`)); err != nil {
			return marketdata.UTCTime{}, fmt.Errorf("dagruntime: %s must be RFC3339Nano", name)
		}
		return t, nil
	default:
		return marketdata.UTCTime{}, fmt.Errorf("dagruntime: %s must be UTCTime", name)
	}
}

func parseFloat64Slice(raw any) ([]float64, error) {
	switch value := raw.(type) {
	case []float64:
		return value, nil
	case []any:
		bars := make([]float64, 0, len(value))
		for _, v := range value {
			f, ok := v.(float64)
			if !ok {
				return nil, fmt.Errorf("dagruntime: market_bars must contain numbers")
			}
			bars = append(bars, f)
		}
		return bars, nil
	default:
		return nil, fmt.Errorf("dagruntime: market_bars must be []float64")
	}
}
