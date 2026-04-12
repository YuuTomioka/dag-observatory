package usecase

import (
	"fmt"

	"dag-observatory/dag-core/internal/domain/dagruntime/engine"
)

func fromStringMap(values map[string]any) (engine.InputMap, error) {
	inputs := engine.InputMap{}
	if raw, ok := values["input"]; ok {
		inputMap, ok := raw.(map[string]any)
		if ok {
			if err := applyStructuredInputValues(inputs, inputMap); err != nil {
				return nil, err
			}
		}
	}
	if err := applyTopLevelInputValues(inputs, values); err != nil {
		return nil, err
	}
	return inputs, nil
}

func applyStructuredInputValues(inputs engine.InputMap, values map[string]any) error {
	if raw, ok := values["symbol"]; ok {
		symbol, ok := raw.(string)
		if !ok {
			return fmt.Errorf("dagruntime: symbol must be string")
		}
		inputs[InputKeySymbol] = symbol
	}
	if raw, ok := values["mode"]; ok {
		mode, ok := raw.(string)
		if !ok {
			return fmt.Errorf("dagruntime: mode must be string")
		}
		inputs[InputKeyMode] = mode
	}
	if rawBars, ok := values["bars"]; ok {
		bars, err := parseFloat64Slice(rawBars)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketBars] = bars
	}
	if rawOHLCVBars, ok := values["ohlcv_bars"]; ok {
		bars, err := parseOHLCVSlice(rawOHLCVBars)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketOHLCVBars] = bars
	}
	if rawOHLCVBars, ok := values["market.ohlcv_bars"]; ok {
		bars, err := parseOHLCVSlice(rawOHLCVBars)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketOHLCVBars] = bars
	}
	if rawOHLCVBarsH1, ok := values["ohlcv_bars_h1"]; ok {
		bars, err := parseOHLCVSlice(rawOHLCVBarsH1)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketOHLCVBarsH1] = bars
	}
	if rawOHLCVBarsH1, ok := values["market.ohlcv_bars.h1"]; ok {
		bars, err := parseOHLCVSlice(rawOHLCVBarsH1)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketOHLCVBarsH1] = bars
	}
	if raw, ok := values["tick"]; ok {
		tick, err := parseTickValue(raw)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketTick] = tick
	}
	if raw, ok := values["market.tick"]; ok {
		tick, err := parseTickValue(raw)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketTick] = tick
	}
	if raw, ok := values["market.spread_bps"]; ok {
		spread, err := parseFloat64Value(raw, "market.spread_bps")
		if err != nil {
			return err
		}
		inputs[InputKeyMarketSpreadBps] = spread
	}
	if raw, ok := values["execution.fee_bps"]; ok {
		feeBps, err := parseFloat64Value(raw, "execution.fee_bps")
		if err != nil {
			return err
		}
		inputs[InputKeyExecutionFeeBps] = feeBps
	}
	if raw, ok := values["execution.slippage_bps"]; ok {
		slippageBps, err := parseFloat64Value(raw, "execution.slippage_bps")
		if err != nil {
			return err
		}
		inputs[InputKeyExecutionSlippageBps] = slippageBps
	}
	if raw, ok := values["execution.min_lot"]; ok {
		minLot, err := parseFloat64Value(raw, "execution.min_lot")
		if err != nil {
			return err
		}
		inputs[InputKeyExecutionMinLot] = minLot
	}
	if raw, ok := values["account.balance"]; ok {
		balance, err := parseFloat64Value(raw, "account.balance")
		if err != nil {
			return err
		}
		inputs[InputKeyAccountBalance] = balance
	}
	if raw, ok := values["marketdata.symbol_id"]; ok {
		symbolID, err := parseInt64Value(raw, "marketdata.symbol_id")
		if err != nil {
			return err
		}
		inputs[InputKeyMarketdataSymbolID] = symbolID
	}
	if raw, ok := values["marketdata.symbol_code"]; ok {
		symbolCode, ok := raw.(string)
		if !ok {
			return fmt.Errorf("dagruntime: marketdata.symbol_code must be string")
		}
		inputs[InputKeyMarketdataSymbolCode] = symbolCode
	}
	if raw, ok := values["marketdata.timeframe_code"]; ok {
		timeframeCode, ok := raw.(string)
		if !ok {
			return fmt.Errorf("dagruntime: marketdata.timeframe_code must be string")
		}
		inputs[InputKeyMarketdataTimeframeCode] = timeframeCode
	}
	if raw, ok := values["marketdata.from"]; ok {
		from, err := parseUTCTimeValue(raw, "marketdata.from")
		if err != nil {
			return err
		}
		inputs[InputKeyMarketdataFrom] = from
	}
	if raw, ok := values["marketdata.to"]; ok {
		to, err := parseUTCTimeValue(raw, "marketdata.to")
		if err != nil {
			return err
		}
		inputs[InputKeyMarketdataTo] = to
	}
	return nil
}

func applyTopLevelInputValues(inputs engine.InputMap, values map[string]any) error {
	if raw, ok := values["symbol"]; ok {
		symbol, ok := raw.(string)
		if !ok {
			return fmt.Errorf("dagruntime: symbol must be string")
		}
		inputs[InputKeySymbol] = symbol
	}
	if raw, ok := values["mode"]; ok {
		mode, ok := raw.(string)
		if !ok {
			return fmt.Errorf("dagruntime: mode must be string")
		}
		inputs[InputKeyMode] = mode
	}
	if raw, ok := values["market_bars"]; ok {
		bars, err := parseFloat64Slice(raw)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketBars] = bars
	}
	if raw, ok := values["market_ohlcv_bars"]; ok {
		bars, err := parseOHLCVSlice(raw)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketOHLCVBars] = bars
	}
	if raw, ok := values["market.ohlcv_bars"]; ok {
		bars, err := parseOHLCVSlice(raw)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketOHLCVBars] = bars
	}
	if raw, ok := values["market_ohlcv_bars_h1"]; ok {
		bars, err := parseOHLCVSlice(raw)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketOHLCVBarsH1] = bars
	}
	if raw, ok := values["market.ohlcv_bars.h1"]; ok {
		bars, err := parseOHLCVSlice(raw)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketOHLCVBarsH1] = bars
	}
	if raw, ok := values["market_tick"]; ok {
		tick, err := parseTickValue(raw)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketTick] = tick
	}
	if raw, ok := values["market.tick"]; ok {
		tick, err := parseTickValue(raw)
		if err != nil {
			return err
		}
		inputs[InputKeyMarketTick] = tick
	}
	if raw, ok := values["market_spread_bps"]; ok {
		spread, err := parseFloat64Value(raw, "market_spread_bps")
		if err != nil {
			return err
		}
		inputs[InputKeyMarketSpreadBps] = spread
	}
	if raw, ok := values["market.spread_bps"]; ok {
		spread, err := parseFloat64Value(raw, "market.spread_bps")
		if err != nil {
			return err
		}
		inputs[InputKeyMarketSpreadBps] = spread
	}
	if raw, ok := values["execution_fee_bps"]; ok {
		feeBps, err := parseFloat64Value(raw, "execution_fee_bps")
		if err != nil {
			return err
		}
		inputs[InputKeyExecutionFeeBps] = feeBps
	}
	if raw, ok := values["execution.fee_bps"]; ok {
		feeBps, err := parseFloat64Value(raw, "execution.fee_bps")
		if err != nil {
			return err
		}
		inputs[InputKeyExecutionFeeBps] = feeBps
	}
	if raw, ok := values["execution_slippage_bps"]; ok {
		slippageBps, err := parseFloat64Value(raw, "execution_slippage_bps")
		if err != nil {
			return err
		}
		inputs[InputKeyExecutionSlippageBps] = slippageBps
	}
	if raw, ok := values["execution.slippage_bps"]; ok {
		slippageBps, err := parseFloat64Value(raw, "execution.slippage_bps")
		if err != nil {
			return err
		}
		inputs[InputKeyExecutionSlippageBps] = slippageBps
	}
	if raw, ok := values["execution_min_lot"]; ok {
		minLot, err := parseFloat64Value(raw, "execution_min_lot")
		if err != nil {
			return err
		}
		inputs[InputKeyExecutionMinLot] = minLot
	}
	if raw, ok := values["execution.min_lot"]; ok {
		minLot, err := parseFloat64Value(raw, "execution.min_lot")
		if err != nil {
			return err
		}
		inputs[InputKeyExecutionMinLot] = minLot
	}
	if raw, ok := values["account_balance"]; ok {
		balance, err := parseFloat64Value(raw, "account_balance")
		if err != nil {
			return err
		}
		inputs[InputKeyAccountBalance] = balance
	}
	if raw, ok := values["account.balance"]; ok {
		balance, err := parseFloat64Value(raw, "account.balance")
		if err != nil {
			return err
		}
		inputs[InputKeyAccountBalance] = balance
	}
	if raw, ok := values["marketdata.symbol_id"]; ok {
		symbolID, err := parseInt64Value(raw, "marketdata.symbol_id")
		if err != nil {
			return err
		}
		inputs[InputKeyMarketdataSymbolID] = symbolID
	}
	if raw, ok := values["marketdata.symbol_code"]; ok {
		symbolCode, ok := raw.(string)
		if !ok {
			return fmt.Errorf("dagruntime: marketdata.symbol_code must be string")
		}
		inputs[InputKeyMarketdataSymbolCode] = symbolCode
	}
	if raw, ok := values["marketdata.timeframe_code"]; ok {
		timeframeCode, ok := raw.(string)
		if !ok {
			return fmt.Errorf("dagruntime: marketdata.timeframe_code must be string")
		}
		inputs[InputKeyMarketdataTimeframeCode] = timeframeCode
	}
	if raw, ok := values["marketdata.from"]; ok {
		from, err := parseUTCTimeValue(raw, "marketdata.from")
		if err != nil {
			return err
		}
		inputs[InputKeyMarketdataFrom] = from
	}
	if raw, ok := values["marketdata.to"]; ok {
		to, err := parseUTCTimeValue(raw, "marketdata.to")
		if err != nil {
			return err
		}
		inputs[InputKeyMarketdataTo] = to
	}
	return nil
}
