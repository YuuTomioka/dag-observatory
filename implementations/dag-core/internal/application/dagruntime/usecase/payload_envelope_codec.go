package usecase

import (
	"encoding/json"
	"fmt"

	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/domain/marketdata/ohlc"
)

var inputEnvelopeEncoders = map[artifact.RawKey]func(any) (events.PayloadEnvelope, error){
	InputKeySymbol.Raw():                  encodeSymbolFromInput,
	InputKeyMode.Raw():                    encodeModeFromInput,
	InputKeyMarketBars.Raw():              encodeMarketBarsFromInput,
	InputKeyMarketOHLCVBars.Raw():         encodeMarketOHLCVBarsFromInput,
	InputKeyMarketOHLCVBarsH1.Raw():       encodeMarketOHLCVBarsH1FromInput,
	InputKeyMarketTick.Raw():              encodeMarketTickFromInput,
	InputKeyMarketSpreadBps.Raw():         encodeMarketSpreadBpsFromInput,
	InputKeyExecutionFeeBps.Raw():         encodeExecutionFeeBpsFromInput,
	InputKeyExecutionSlippageBps.Raw():    encodeExecutionSlippageBpsFromInput,
	InputKeyExecutionMinLot.Raw():         encodeExecutionMinLotFromInput,
	InputKeyAccountBalance.Raw():          encodeAccountBalanceFromInput,
	InputKeyMarketdataSymbolID.Raw():      encodeMarketdataSymbolIDFromInput,
	InputKeyMarketdataSymbolCode.Raw():    encodeMarketdataSymbolCodeFromInput,
	InputKeyMarketdataTimeframeCode.Raw(): encodeMarketdataTimeframeCodeFromInput,
	InputKeyMarketdataFrom.Raw():          encodeMarketdataFromFromInput,
	InputKeyMarketdataTo.Raw():            encodeMarketdataToFromInput,
}

var envelopeDecoders = map[events.RawPayloadKey]func(events.PayloadEnvelope, engine.InputMap) error{
	PayloadKeySymbol.Raw():                  decodeSymbolEnvelope,
	PayloadKeyMode.Raw():                    decodeModeEnvelope,
	PayloadKeyMarketBars.Raw():              decodeMarketBarsEnvelope,
	PayloadKeyMarketOHLCVBars.Raw():         decodeMarketOHLCVBarsEnvelope,
	PayloadKeyMarketOHLCVBarsH1.Raw():       decodeMarketOHLCVBarsH1Envelope,
	PayloadKeyMarketTick.Raw():              decodeMarketTickEnvelope,
	PayloadKeyMarketSpreadBps.Raw():         decodeMarketSpreadBpsEnvelope,
	PayloadKeyExecutionFeeBps.Raw():         decodeExecutionFeeBpsEnvelope,
	PayloadKeyExecutionSlippageBps.Raw():    decodeExecutionSlippageBpsEnvelope,
	PayloadKeyExecutionMinLot.Raw():         decodeExecutionMinLotEnvelope,
	PayloadKeyAccountBalance.Raw():          decodeAccountBalanceEnvelope,
	PayloadKeyMarketdataSymbolID.Raw():      decodeMarketdataSymbolIDEnvelope,
	PayloadKeyMarketdataSymbolCode.Raw():    decodeMarketdataSymbolCodeEnvelope,
	PayloadKeyMarketdataTimeframeCode.Raw(): decodeMarketdataTimeframeCodeEnvelope,
	PayloadKeyMarketdataFrom.Raw():          decodeMarketdataFromEnvelope,
	PayloadKeyMarketdataTo.Raw():            decodeMarketdataToEnvelope,
}

func fromInputMap(inputs engine.InputMap) ([]events.PayloadEnvelope, error) {
	envelopes := make([]events.PayloadEnvelope, 0, 8)
	for key, value := range inputs {
		encode, ok := inputEnvelopeEncoders[key.Raw()]
		if !ok {
			continue
		}
		env, err := encode(value)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
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
	if raw, ok := values["execution_fee_bps"]; ok {
		feeBps, err := parseFloat64Value(raw, "execution_fee_bps")
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyExecutionFeeBps, feeBps)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["execution_slippage_bps"]; ok {
		slippageBps, err := parseFloat64Value(raw, "execution_slippage_bps")
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyExecutionSlippageBps, slippageBps)
		if err != nil {
			return nil, err
		}
		envelopes = append(envelopes, env)
	}
	if raw, ok := values["execution_min_lot"]; ok {
		minLot, err := parseFloat64Value(raw, "execution_min_lot")
		if err != nil {
			return nil, err
		}
		env, err := events.EncodePayload(PayloadKeyExecutionMinLot, minLot)
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
		if env.Key == PayloadKeyInput.Raw() || env.Key.Name == PayloadKeyInput.Name {
			value, err := decodeInputEnvelopeCompat(env)
			if err != nil {
				return nil, err
			}
			if err := applyStructuredInputValues(inputs, value); err != nil {
				return nil, err
			}
			continue
		}
		decode, ok := envelopeDecoders[env.Key]
		if !ok {
			continue
		}
		if err := decode(env, inputs); err != nil {
			return nil, err
		}
	}
	return inputs, nil
}

func decodeInputEnvelopeCompat(env events.PayloadEnvelope) (map[string]any, error) {
	if env.Key == PayloadKeyInput.Raw() {
		return events.DecodePayload(PayloadKeyInput, env)
	}
	// Compatibility for producers that emit only key name without stable ID.
	var out map[string]any
	if err := json.Unmarshal(env.Data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func encodeSymbolFromInput(value any) (events.PayloadEnvelope, error) {
	symbol, ok := value.(string)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: symbol must be string")
	}
	return events.EncodePayload(PayloadKeySymbol, symbol)
}

func encodeModeFromInput(value any) (events.PayloadEnvelope, error) {
	mode, ok := value.(string)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: mode must be string")
	}
	return events.EncodePayload(PayloadKeyMode, mode)
}

func encodeMarketBarsFromInput(value any) (events.PayloadEnvelope, error) {
	bars, ok := value.([]float64)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: market_bars must be []float64")
	}
	return events.EncodePayload(PayloadKeyMarketBars, bars)
}

func encodeMarketOHLCVBarsFromInput(value any) (events.PayloadEnvelope, error) {
	bars, ok := value.([]ohlc.OHLCV)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: market_ohlcv_bars must be []ohlc.OHLCV")
	}
	return events.EncodePayload(PayloadKeyMarketOHLCVBars, bars)
}

func encodeMarketOHLCVBarsH1FromInput(value any) (events.PayloadEnvelope, error) {
	bars, ok := value.([]ohlc.OHLCV)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: market_ohlcv_bars_h1 must be []ohlc.OHLCV")
	}
	return events.EncodePayload(PayloadKeyMarketOHLCVBarsH1, bars)
}

func encodeMarketTickFromInput(value any) (events.PayloadEnvelope, error) {
	tick, ok := value.(marketdata.Tick)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: market.tick must be marketdata.Tick")
	}
	return events.EncodePayload(PayloadKeyMarketTick, tick)
}

func encodeMarketSpreadBpsFromInput(value any) (events.PayloadEnvelope, error) {
	spread, ok := value.(float64)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: market.spread_bps must be number")
	}
	return events.EncodePayload(PayloadKeyMarketSpreadBps, spread)
}

func encodeExecutionFeeBpsFromInput(value any) (events.PayloadEnvelope, error) {
	feeBps, ok := value.(float64)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: execution.fee_bps must be number")
	}
	return events.EncodePayload(PayloadKeyExecutionFeeBps, feeBps)
}

func encodeExecutionSlippageBpsFromInput(value any) (events.PayloadEnvelope, error) {
	slippageBps, ok := value.(float64)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: execution.slippage_bps must be number")
	}
	return events.EncodePayload(PayloadKeyExecutionSlippageBps, slippageBps)
}

func encodeExecutionMinLotFromInput(value any) (events.PayloadEnvelope, error) {
	minLot, ok := value.(float64)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: execution.min_lot must be number")
	}
	return events.EncodePayload(PayloadKeyExecutionMinLot, minLot)
}

func encodeAccountBalanceFromInput(value any) (events.PayloadEnvelope, error) {
	balance, ok := value.(float64)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: account.balance must be number")
	}
	return events.EncodePayload(PayloadKeyAccountBalance, balance)
}

func encodeMarketdataSymbolIDFromInput(value any) (events.PayloadEnvelope, error) {
	symbolID, ok := value.(int64)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: marketdata.symbol_id must be int64")
	}
	return events.EncodePayload(PayloadKeyMarketdataSymbolID, symbolID)
}

func encodeMarketdataSymbolCodeFromInput(value any) (events.PayloadEnvelope, error) {
	symbolCode, ok := value.(string)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: marketdata.symbol_code must be string")
	}
	return events.EncodePayload(PayloadKeyMarketdataSymbolCode, symbolCode)
}

func encodeMarketdataTimeframeCodeFromInput(value any) (events.PayloadEnvelope, error) {
	timeframeCode, ok := value.(string)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: marketdata.timeframe_code must be string")
	}
	return events.EncodePayload(PayloadKeyMarketdataTimeframeCode, timeframeCode)
}

func encodeMarketdataFromFromInput(value any) (events.PayloadEnvelope, error) {
	from, ok := value.(marketdata.UTCTime)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: marketdata.from must be UTCTime")
	}
	return events.EncodePayload(PayloadKeyMarketdataFrom, from)
}

func encodeMarketdataToFromInput(value any) (events.PayloadEnvelope, error) {
	to, ok := value.(marketdata.UTCTime)
	if !ok {
		return events.PayloadEnvelope{}, fmt.Errorf("dagruntime: marketdata.to must be UTCTime")
	}
	return events.EncodePayload(PayloadKeyMarketdataTo, to)
}

func decodeSymbolEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeySymbol, env)
	if err != nil {
		return err
	}
	inputs[InputKeySymbol] = value
	return nil
}

func decodeModeEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMode, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMode] = value
	return nil
}

func decodeMarketBarsEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMarketBars, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMarketBars] = value
	return nil
}

func decodeMarketOHLCVBarsEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMarketOHLCVBars, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMarketOHLCVBars] = value
	return nil
}

func decodeMarketOHLCVBarsH1Envelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMarketOHLCVBarsH1, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMarketOHLCVBarsH1] = value
	return nil
}

func decodeMarketTickEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMarketTick, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMarketTick] = value
	return nil
}

func decodeMarketSpreadBpsEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMarketSpreadBps, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMarketSpreadBps] = value
	return nil
}

func decodeExecutionFeeBpsEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyExecutionFeeBps, env)
	if err != nil {
		return err
	}
	inputs[InputKeyExecutionFeeBps] = value
	return nil
}

func decodeExecutionSlippageBpsEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyExecutionSlippageBps, env)
	if err != nil {
		return err
	}
	inputs[InputKeyExecutionSlippageBps] = value
	return nil
}

func decodeExecutionMinLotEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyExecutionMinLot, env)
	if err != nil {
		return err
	}
	inputs[InputKeyExecutionMinLot] = value
	return nil
}

func decodeAccountBalanceEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyAccountBalance, env)
	if err != nil {
		return err
	}
	inputs[InputKeyAccountBalance] = value
	return nil
}

func decodeMarketdataSymbolIDEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMarketdataSymbolID, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMarketdataSymbolID] = value
	return nil
}

func decodeMarketdataSymbolCodeEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMarketdataSymbolCode, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMarketdataSymbolCode] = value
	return nil
}

func decodeMarketdataTimeframeCodeEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMarketdataTimeframeCode, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMarketdataTimeframeCode] = value
	return nil
}

func decodeMarketdataFromEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMarketdataFrom, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMarketdataFrom] = value
	return nil
}

func decodeMarketdataToEnvelope(env events.PayloadEnvelope, inputs engine.InputMap) error {
	value, err := events.DecodePayload(PayloadKeyMarketdataTo, env)
	if err != nil {
		return err
	}
	inputs[InputKeyMarketdataTo] = value
	return nil
}
