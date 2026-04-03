package inputmap

import (
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
)

func Default() map[string]artifact.AnyKey {
	return map[string]artifact.AnyKey{
		"symbol":                    usecase.InputKeySymbol,
		"mode":                      usecase.InputKeyMode,
		"market.symbol":             usecase.InputKeySymbol,
		"market.bars":               usecase.InputKeyMarketBars,
		"market.ohlcv_bars":         usecase.InputKeyMarketOHLCVBars,
		"market.ohlcv_bars.h1":      usecase.InputKeyMarketOHLCVBarsH1,
		"market.tick":               usecase.InputKeyMarketTick,
		"market.spread_bps":         usecase.InputKeyMarketSpreadBps,
		"account.balance":           usecase.InputKeyAccountBalance,
		"marketdata.symbol_id":      usecase.InputKeyMarketdataSymbolID,
		"marketdata.symbol_code":    usecase.InputKeyMarketdataSymbolCode,
		"marketdata.timeframe_code": usecase.InputKeyMarketdataTimeframeCode,
		"marketdata.from":           usecase.InputKeyMarketdataFrom,
		"marketdata.to":             usecase.InputKeyMarketdataTo,
	}
}
