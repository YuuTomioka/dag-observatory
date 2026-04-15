package usecase

import (
	"testing"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

func TestToInputMapFromStringMap(t *testing.T) {
	payload := map[string]any{
		"symbol":      "USDJPY",
		"mode":        "normal",
		"market_bars": []any{1.0, 2.0, 3.0},
	}

	inputs, err := toInputMap(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inputs[InputKeySymbol] != "USDJPY" {
		t.Fatalf("expected symbol to be mapped")
	}
	if inputs[InputKeyMode] != "normal" {
		t.Fatalf("expected mode to be mapped")
	}
	bars, ok := inputs[InputKeyMarketBars].([]float64)
	if !ok || len(bars) != 3 {
		t.Fatalf("expected market bars to be mapped")
	}
}

func TestToInputMapFromPayloadEnvelopes(t *testing.T) {
	envSymbol, err := events.EncodePayload(PayloadKeySymbol, "EURUSD")
	if err != nil {
		t.Fatalf("encode symbol failed: %v", err)
	}
	envMode, err := events.EncodePayload(PayloadKeyMode, "debug")
	if err != nil {
		t.Fatalf("encode mode failed: %v", err)
	}
	envBars, err := events.EncodePayload(PayloadKeyMarketBars, []float64{1.0, 2.0})
	if err != nil {
		t.Fatalf("encode bars failed: %v", err)
	}

	inputs, err := toInputMap([]events.PayloadEnvelope{envSymbol, envMode, envBars})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inputs[InputKeySymbol] != "EURUSD" {
		t.Fatalf("expected symbol from payload")
	}
	if inputs[InputKeyMode] != "debug" {
		t.Fatalf("expected mode from payload")
	}
	bars, ok := inputs[InputKeyMarketBars].([]float64)
	if !ok || len(bars) != 2 {
		t.Fatalf("expected market bars from payload")
	}
}

func TestToInputMapFromStringMapWithOHLCVBars(t *testing.T) {
	payload := map[string]any{
		"market_ohlcv_bars": []any{
			map[string]any{
				"open_time":  "2026-04-01T00:00:00Z",
				"close_time": "2026-04-01T00:01:00Z",
				"open":       1000.0,
				"high":       1010.0,
				"high_time":  "2026-04-01T00:00:30Z",
				"low":        995.0,
				"low_time":   "2026-04-01T00:00:10Z",
				"close":      1005.0,
				"volume":     12.0,
			},
		},
		"market_spread_bps": 2.5,
		"account_balance":   10000.0,
	}

	inputs, err := toInputMap(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bars, ok := inputs[InputKeyMarketOHLCVBars].([]marketdata.OHLCV)
	if !ok {
		t.Fatalf("expected market ohlcv bars to be mapped")
	}
	if len(bars) != 1 {
		t.Fatalf("expected 1 ohlcv bar, got %d", len(bars))
	}
	if got := bars[0].Close.Raw(); got != 1005 {
		t.Fatalf("expected close=1005, got %d", got)
	}
	if !bars[0].Hightime.Equal(marketdata.MustParseUTCTime("2026-04-01T00:00:30Z")) {
		t.Fatalf("expected high_time to be mapped, got %s", bars[0].Hightime)
	}
	if !bars[0].Lowtime.Equal(marketdata.MustParseUTCTime("2026-04-01T00:00:10Z")) {
		t.Fatalf("expected low_time to be mapped, got %s", bars[0].Lowtime)
	}
	if got := inputs[InputKeyMarketSpreadBps]; got != 2.5 {
		t.Fatalf("expected spread_bps=2.5, got %v", got)
	}
	if got := inputs[InputKeyAccountBalance]; got != 10000.0 {
		t.Fatalf("expected account_balance=10000, got %v", got)
	}
}

func TestToInputMapFromPayloadEnvelopesWithOHLCVBars(t *testing.T) {
	ohlcv := []marketdata.OHLCV{
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-01T00:00:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-01T00:01:00Z"),
			Open:      marketdata.NewPriceFromRaw(1000),
			High:      marketdata.NewPriceFromRaw(1010),
			Low:       marketdata.NewPriceFromRaw(995),
			Close:     marketdata.NewPriceFromRaw(1005),
			Volume:    marketdata.Volume(12),
		},
	}
	envBars, err := events.EncodePayload(PayloadKeyMarketOHLCVBars, ohlcv)
	if err != nil {
		t.Fatalf("encode ohlcv bars failed: %v", err)
	}
	envSpread, err := events.EncodePayload(PayloadKeyMarketSpreadBps, 1.2)
	if err != nil {
		t.Fatalf("encode spread failed: %v", err)
	}
	envBalance, err := events.EncodePayload(PayloadKeyAccountBalance, 20000.0)
	if err != nil {
		t.Fatalf("encode balance failed: %v", err)
	}

	inputs, err := toInputMap([]events.PayloadEnvelope{envBars, envSpread, envBalance})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bars, ok := inputs[InputKeyMarketOHLCVBars].([]marketdata.OHLCV)
	if !ok || len(bars) != 1 {
		t.Fatalf("expected market ohlcv bars from payload")
	}
	if got := bars[0].Open.Raw(); got != 1000 {
		t.Fatalf("expected open=1000, got %d", got)
	}
	if got := inputs[InputKeyMarketSpreadBps]; got != 1.2 {
		t.Fatalf("expected spread_bps=1.2, got %v", got)
	}
	if got := inputs[InputKeyAccountBalance]; got != 20000.0 {
		t.Fatalf("expected account_balance=20000, got %v", got)
	}
}

func TestToInputMapBridgeFromStringMapWithTickAndH1Bars(t *testing.T) {
	payload := map[string]any{
		"market_tick": map[string]any{
			"symbol_id": 1,
			"time":      "2026-04-02T00:00:00Z",
			"bid":       1000.0,
			"ask":       1002.0,
		},
		"market_ohlcv_bars_h1": []any{
			map[string]any{
				"open_time":  "2026-04-02T00:00:00Z",
				"close_time": "2026-04-02T01:00:00Z",
				"open":       990.0,
				"high":       1010.0,
				"low":        980.0,
				"close":      1008.0,
				"volume":     100.0,
			},
		},
	}

	inputs, err := toInputMap(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tick, ok := inputs[InputKeyMarketTick].(marketdata.Tick)
	if !ok {
		t.Fatalf("expected market tick to be mapped")
	}
	if tick.Ask.Raw() != 1002 || tick.Bid.Raw() != 1000 {
		t.Fatalf("unexpected tick mapped: %#v", tick)
	}
	barsH1, ok := inputs[InputKeyMarketOHLCVBarsH1].([]marketdata.OHLCV)
	if !ok || len(barsH1) != 1 {
		t.Fatalf("expected market h1 bars to be mapped")
	}
	if barsH1[0].Close.Raw() != 1008 {
		t.Fatalf("expected h1 close=1008, got %d", barsH1[0].Close.Raw())
	}
}

func TestToInputMapBridgeFromPayloadEnvelopesWithTickAndH1Bars(t *testing.T) {
	tick := marketdata.Tick{
		SymbolID: 2,
		Time:     marketdata.MustParseUTCTime("2026-04-02T00:00:00Z"),
		Bid:      marketdata.NewPriceFromRaw(2000),
		Ask:      marketdata.NewPriceFromRaw(2003),
	}
	h1 := []marketdata.OHLCV{
		{
			Opentime:  marketdata.MustParseUTCTime("2026-04-02T00:00:00Z"),
			Closetime: marketdata.MustParseUTCTime("2026-04-02T01:00:00Z"),
			Open:      marketdata.NewPriceFromRaw(1990),
			High:      marketdata.NewPriceFromRaw(2010),
			Low:       marketdata.NewPriceFromRaw(1980),
			Close:     marketdata.NewPriceFromRaw(2008),
		},
	}
	envTick, err := events.EncodePayload(PayloadKeyMarketTick, tick)
	if err != nil {
		t.Fatalf("encode market tick failed: %v", err)
	}
	envH1, err := events.EncodePayload(PayloadKeyMarketOHLCVBarsH1, h1)
	if err != nil {
		t.Fatalf("encode market h1 bars failed: %v", err)
	}

	inputs, err := toInputMap([]events.PayloadEnvelope{envTick, envH1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotTick, ok := inputs[InputKeyMarketTick].(marketdata.Tick)
	if !ok {
		t.Fatalf("expected market tick from payload")
	}
	if gotTick.SymbolID != 2 || gotTick.Ask.Raw() != 2003 {
		t.Fatalf("unexpected tick from payload: %#v", gotTick)
	}
	gotH1, ok := inputs[InputKeyMarketOHLCVBarsH1].([]marketdata.OHLCV)
	if !ok || len(gotH1) != 1 {
		t.Fatalf("expected market h1 bars from payload")
	}
	if gotH1[0].Open.Raw() != 1990 {
		t.Fatalf("expected h1 open=1990, got %d", gotH1[0].Open.Raw())
	}
}
