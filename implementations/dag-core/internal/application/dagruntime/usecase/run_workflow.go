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
	if len(req.OHLCVBars) > 0 {
		payload["input"].(map[string]any)["ohlcv_bars"] = req.OHLCVBars
		payload["market_ohlcv_bars"] = req.OHLCVBars
	}
	if req.SpreadBps > 0 {
		payload["input"].(map[string]any)["market.spread_bps"] = req.SpreadBps
		payload["market_spread_bps"] = req.SpreadBps
	}
	if req.AccountBalance > 0 {
		payload["input"].(map[string]any)["account.balance"] = req.AccountBalance
		payload["account_balance"] = req.AccountBalance
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
			if rawOHLCVBars, ok := inputMap["ohlcv_bars"]; ok {
				bars, err := parseOHLCVSlice(rawOHLCVBars)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketOHLCVBars] = bars
			}
			if rawOHLCVBars, ok := inputMap["market.ohlcv_bars"]; ok {
				bars, err := parseOHLCVSlice(rawOHLCVBars)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketOHLCVBars] = bars
			}
			if rawOHLCVBarsH1, ok := inputMap["ohlcv_bars_h1"]; ok {
				bars, err := parseOHLCVSlice(rawOHLCVBarsH1)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketOHLCVBarsH1] = bars
			}
			if rawOHLCVBarsH1, ok := inputMap["market.ohlcv_bars.h1"]; ok {
				bars, err := parseOHLCVSlice(rawOHLCVBarsH1)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketOHLCVBarsH1] = bars
			}
			if raw, ok := inputMap["tick"]; ok {
				tick, err := parseTickValue(raw)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketTick] = tick
			}
			if raw, ok := inputMap["market.tick"]; ok {
				tick, err := parseTickValue(raw)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketTick] = tick
			}
			if raw, ok := inputMap["market.spread_bps"]; ok {
				spread, err := parseFloat64Value(raw, "market.spread_bps")
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketSpreadBps] = spread
			}
			if raw, ok := inputMap["account.balance"]; ok {
				balance, err := parseFloat64Value(raw, "account.balance")
				if err != nil {
					return nil, err
				}
				inputs[InputKeyAccountBalance] = balance
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
	if raw, ok := values["market_ohlcv_bars"]; ok {
		bars, err := parseOHLCVSlice(raw)
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketOHLCVBars] = bars
	}
	if raw, ok := values["market.ohlcv_bars"]; ok {
		bars, err := parseOHLCVSlice(raw)
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketOHLCVBars] = bars
	}
	if raw, ok := values["market_ohlcv_bars_h1"]; ok {
		bars, err := parseOHLCVSlice(raw)
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketOHLCVBarsH1] = bars
	}
	if raw, ok := values["market.ohlcv_bars.h1"]; ok {
		bars, err := parseOHLCVSlice(raw)
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketOHLCVBarsH1] = bars
	}
	if raw, ok := values["market_tick"]; ok {
		tick, err := parseTickValue(raw)
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketTick] = tick
	}
	if raw, ok := values["market.tick"]; ok {
		tick, err := parseTickValue(raw)
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketTick] = tick
	}
	if raw, ok := values["market_spread_bps"]; ok {
		spread, err := parseFloat64Value(raw, "market_spread_bps")
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketSpreadBps] = spread
	}
	if raw, ok := values["market.spread_bps"]; ok {
		spread, err := parseFloat64Value(raw, "market.spread_bps")
		if err != nil {
			return nil, err
		}
		inputs[InputKeyMarketSpreadBps] = spread
	}
	if raw, ok := values["account_balance"]; ok {
		balance, err := parseFloat64Value(raw, "account_balance")
		if err != nil {
			return nil, err
		}
		inputs[InputKeyAccountBalance] = balance
	}
	if raw, ok := values["account.balance"]; ok {
		balance, err := parseFloat64Value(raw, "account.balance")
		if err != nil {
			return nil, err
		}
		inputs[InputKeyAccountBalance] = balance
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
		if key.Raw() == InputKeyMarketOHLCVBars.Raw() {
			bars, ok := value.([]marketdata.OHLCV)
			if !ok {
				return nil, fmt.Errorf("dagruntime: market_ohlcv_bars must be []marketdata.OHLCV")
			}
			env, err := events.EncodePayload(PayloadKeyMarketOHLCVBars, bars)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
		if key.Raw() == InputKeyMarketOHLCVBarsH1.Raw() {
			bars, ok := value.([]marketdata.OHLCV)
			if !ok {
				return nil, fmt.Errorf("dagruntime: market_ohlcv_bars_h1 must be []marketdata.OHLCV")
			}
			env, err := events.EncodePayload(PayloadKeyMarketOHLCVBarsH1, bars)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
		if key.Raw() == InputKeyMarketTick.Raw() {
			tick, ok := value.(marketdata.Tick)
			if !ok {
				return nil, fmt.Errorf("dagruntime: market.tick must be marketdata.Tick")
			}
			env, err := events.EncodePayload(PayloadKeyMarketTick, tick)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
		if key.Raw() == InputKeyMarketSpreadBps.Raw() {
			spread, ok := value.(float64)
			if !ok {
				return nil, fmt.Errorf("dagruntime: market.spread_bps must be number")
			}
			env, err := events.EncodePayload(PayloadKeyMarketSpreadBps, spread)
			if err != nil {
				return nil, err
			}
			envelopes = append(envelopes, env)
			continue
		}
		if key.Raw() == InputKeyAccountBalance.Raw() {
			balance, ok := value.(float64)
			if !ok {
				return nil, fmt.Errorf("dagruntime: account.balance must be number")
			}
			env, err := events.EncodePayload(PayloadKeyAccountBalance, balance)
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
	if raw, ok := values["market_ohlcv_bars"]; ok {
		bars, err := parseOHLCVSlice(raw)
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyMarketOHLCVBars, bars)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["market_ohlcv_bars_h1"]; ok {
		bars, err := parseOHLCVSlice(raw)
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyMarketOHLCVBarsH1, bars)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["market_tick"]; ok {
		tick, err := parseTickValue(raw)
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyMarketTick, tick)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["market_spread_bps"]; ok {
		spread, err := parseFloat64Value(raw, "market_spread_bps")
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyMarketSpreadBps, spread)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["account_balance"]; ok {
		balance, err := parseFloat64Value(raw, "account_balance")
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyAccountBalance, balance)
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
			if raw, ok := value["ohlcv_bars"]; ok {
				bars, err := parseOHLCVSlice(raw)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketOHLCVBars] = bars
			}
			if raw, ok := value["market.ohlcv_bars"]; ok {
				bars, err := parseOHLCVSlice(raw)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketOHLCVBars] = bars
			}
			if raw, ok := value["ohlcv_bars_h1"]; ok {
				bars, err := parseOHLCVSlice(raw)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketOHLCVBarsH1] = bars
			}
			if raw, ok := value["market.ohlcv_bars.h1"]; ok {
				bars, err := parseOHLCVSlice(raw)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketOHLCVBarsH1] = bars
			}
			if raw, ok := value["tick"]; ok {
				tick, err := parseTickValue(raw)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketTick] = tick
			}
			if raw, ok := value["market.tick"]; ok {
				tick, err := parseTickValue(raw)
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketTick] = tick
			}
			if raw, ok := value["market.spread_bps"]; ok {
				spread, err := parseFloat64Value(raw, "market.spread_bps")
				if err != nil {
					return nil, err
				}
				inputs[InputKeyMarketSpreadBps] = spread
			}
			if raw, ok := value["account.balance"]; ok {
				balance, err := parseFloat64Value(raw, "account.balance")
				if err != nil {
					return nil, err
				}
				inputs[InputKeyAccountBalance] = balance
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
		if env.Key == PayloadKeyMarketOHLCVBars.Raw() {
			value, err := events.DecodePayload(PayloadKeyMarketOHLCVBars, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMarketOHLCVBars] = value
			continue
		}
		if env.Key == PayloadKeyMarketOHLCVBarsH1.Raw() {
			value, err := events.DecodePayload(PayloadKeyMarketOHLCVBarsH1, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMarketOHLCVBarsH1] = value
			continue
		}
		if env.Key == PayloadKeyMarketTick.Raw() {
			value, err := events.DecodePayload(PayloadKeyMarketTick, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMarketTick] = value
			continue
		}
		if env.Key == PayloadKeyMarketSpreadBps.Raw() {
			value, err := events.DecodePayload(PayloadKeyMarketSpreadBps, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyMarketSpreadBps] = value
			continue
		}
		if env.Key == PayloadKeyAccountBalance.Raw() {
			value, err := events.DecodePayload(PayloadKeyAccountBalance, env)
			if err != nil {
				return nil, err
			}
			inputs[InputKeyAccountBalance] = value
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

func parseFloat64Value(raw any, name string) (float64, error) {
	switch value := raw.(type) {
	case float64:
		return value, nil
	case float32:
		return float64(value), nil
	case int:
		return float64(value), nil
	case int64:
		return float64(value), nil
	default:
		return 0, fmt.Errorf("dagruntime: %s must be number", name)
	}
}

func parseOHLCVSlice(raw any) ([]marketdata.OHLCV, error) {
	switch value := raw.(type) {
	case []marketdata.OHLCV:
		return value, nil
	case []any:
		bars := make([]marketdata.OHLCV, 0, len(value))
		for _, item := range value {
			bar, err := parseOHLCVItem(item)
			if err != nil {
				return nil, err
			}
			bars = append(bars, bar)
		}
		return bars, nil
	default:
		return nil, fmt.Errorf("dagruntime: market_ohlcv_bars must be []marketdata.OHLCV")
	}
}

func parseOHLCVItem(raw any) (marketdata.OHLCV, error) {
	switch value := raw.(type) {
	case marketdata.OHLCV:
		return value, nil
	case map[string]any:
		return parseOHLCVMap(value)
	default:
		return marketdata.OHLCV{}, fmt.Errorf("dagruntime: market_ohlcv_bars must contain OHLCV objects")
	}
}

func parseTickValue(raw any) (marketdata.Tick, error) {
	switch value := raw.(type) {
	case marketdata.Tick:
		return value, nil
	case map[string]any:
		return parseTickMap(value)
	default:
		return marketdata.Tick{}, fmt.Errorf("dagruntime: market_tick must be marketdata.Tick")
	}
}

func parseTickMap(value map[string]any) (marketdata.Tick, error) {
	get := func(keys ...string) (any, bool) {
		for _, k := range keys {
			if raw, ok := value[k]; ok {
				return raw, true
			}
		}
		return nil, false
	}

	var out marketdata.Tick
	if raw, ok := get("symbol_id", "SymbolID", "symbolID"); ok {
		symbolID, err := parseInt64Value(raw, "market_tick.symbol_id")
		if err != nil {
			return marketdata.Tick{}, err
		}
		out.SymbolID = marketdata.SymbolID(symbolID)
	}
	if raw, ok := get("time", "Time"); ok {
		t, err := parseUTCTimeValue(raw, "market_tick.time")
		if err != nil {
			return marketdata.Tick{}, err
		}
		out.Time = t
	}
	if raw, ok := get("bid", "Bid"); ok {
		bid, err := parsePriceValue(raw, "market_tick.bid")
		if err != nil {
			return marketdata.Tick{}, err
		}
		out.Bid = bid
	}
	if raw, ok := get("ask", "Ask"); ok {
		ask, err := parsePriceValue(raw, "market_tick.ask")
		if err != nil {
			return marketdata.Tick{}, err
		}
		out.Ask = ask
	}
	return out, nil
}

func parseOHLCVMap(value map[string]any) (marketdata.OHLCV, error) {
	get := func(keys ...string) (any, bool) {
		for _, k := range keys {
			if raw, ok := value[k]; ok {
				return raw, true
			}
		}
		return nil, false
	}

	var out marketdata.OHLCV
	if raw, ok := get("open_time", "Opentime", "opentime"); ok {
		t, err := parseUTCTimeValue(raw, "market_ohlcv_bars.open_time")
		if err != nil {
			return marketdata.OHLCV{}, err
		}
		out.Opentime = t
	}
	if raw, ok := get("close_time", "Closetime", "closetime"); ok {
		t, err := parseUTCTimeValue(raw, "market_ohlcv_bars.close_time")
		if err != nil {
			return marketdata.OHLCV{}, err
		}
		out.Closetime = t
	}
	if raw, ok := get("open", "Open"); ok {
		price, err := parsePriceValue(raw, "market_ohlcv_bars.open")
		if err != nil {
			return marketdata.OHLCV{}, err
		}
		out.Open = price
	}
	if raw, ok := get("high", "High"); ok {
		price, err := parsePriceValue(raw, "market_ohlcv_bars.high")
		if err != nil {
			return marketdata.OHLCV{}, err
		}
		out.High = price
	}
	if raw, ok := get("low", "Low"); ok {
		price, err := parsePriceValue(raw, "market_ohlcv_bars.low")
		if err != nil {
			return marketdata.OHLCV{}, err
		}
		out.Low = price
	}
	if raw, ok := get("close", "Close"); ok {
		price, err := parsePriceValue(raw, "market_ohlcv_bars.close")
		if err != nil {
			return marketdata.OHLCV{}, err
		}
		out.Close = price
	}
	if raw, ok := get("volume", "Volume"); ok {
		volume, err := parseVolumeValue(raw, "market_ohlcv_bars.volume")
		if err != nil {
			return marketdata.OHLCV{}, err
		}
		out.Volume = volume
	}
	return out, nil
}

func parsePriceValue(raw any, name string) (marketdata.Price, error) {
	switch value := raw.(type) {
	case marketdata.Price:
		return value, nil
	case int64:
		return marketdata.NewPriceFromRaw(value), nil
	case int:
		return marketdata.NewPriceFromRaw(int64(value)), nil
	case float64:
		return marketdata.NewPriceFromRaw(int64(value)), nil
	default:
		return 0, fmt.Errorf("dagruntime: %s must be price number", name)
	}
}

func parseVolumeValue(raw any, name string) (marketdata.Volume, error) {
	switch value := raw.(type) {
	case marketdata.Volume:
		return value, nil
	case int64:
		return marketdata.Volume(value), nil
	case int:
		return marketdata.Volume(value), nil
	case float64:
		return marketdata.Volume(int64(value)), nil
	default:
		return 0, fmt.Errorf("dagruntime: %s must be volume number", name)
	}
}
