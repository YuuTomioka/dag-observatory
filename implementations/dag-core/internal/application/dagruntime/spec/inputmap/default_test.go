package inputmap

import (
	"testing"

	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
)

func TestDefaultContainsRequiredKeys(t *testing.T) {
	t.Parallel()

	m := Default()
	cases := map[string]string{
		"symbol":                    usecase.InputKeySymbol.String(),
		"mode":                      usecase.InputKeyMode.String(),
		"market.symbol":             usecase.InputKeySymbol.String(),
		"market.ohlcv_bars":         usecase.InputKeyMarketOHLCVBars.String(),
		"market.ohlcv_bars.h1":      usecase.InputKeyMarketOHLCVBarsH1.String(),
		"marketdata.symbol_code":    usecase.InputKeyMarketdataSymbolCode.String(),
		"marketdata.timeframe_code": usecase.InputKeyMarketdataTimeframeCode.String(),
		"marketdata.from":           usecase.InputKeyMarketdataFrom.String(),
		"marketdata.to":             usecase.InputKeyMarketdataTo.String(),
	}
	for name, want := range cases {
		got, ok := m[name]
		if !ok {
			t.Fatalf("expected key %q to exist", name)
		}
		if got.String() != want {
			t.Fatalf("key %q: want %q, got %q", name, want, got.String())
		}
	}
}
