package usecase

import (
	"fmt"

	"dag-observatory/dag-core/internal/domain/marketdata"
)

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
	if raw, ok := get("high_time", "Hightime", "hightime"); ok {
		t, err := parseUTCTimeValue(raw, "market_ohlcv_bars.high_time")
		if err != nil {
			return marketdata.OHLCV{}, err
		}
		out.Hightime = t
	}
	if raw, ok := get("low", "Low"); ok {
		price, err := parsePriceValue(raw, "market_ohlcv_bars.low")
		if err != nil {
			return marketdata.OHLCV{}, err
		}
		out.Low = price
	}
	if raw, ok := get("low_time", "Lowtime", "lowtime"); ok {
		t, err := parseUTCTimeValue(raw, "market_ohlcv_bars.low_time")
		if err != nil {
			return marketdata.OHLCV{}, err
		}
		out.Lowtime = t
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
