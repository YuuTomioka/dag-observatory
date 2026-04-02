package usecase

import (
	"dag-observatory/dag-core/internal/domain/dagruntime/artifact"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/marketdata"
)

var (
	PayloadKeyRunID = events.PayloadKey[string]{
		Name:     "run_id",
		StableID: "event:run_id.v1",
	}
	PayloadKeyTaskID = events.PayloadKey[string]{
		Name:     "task_id",
		StableID: "event:task_id.v1",
	}
	PayloadKeyAttempt = events.PayloadKey[int]{
		Name:     "attempt",
		StableID: "event:attempt.v1",
	}
	PayloadKeyTaskName = events.PayloadKey[string]{
		Name:     "task_name",
		StableID: "event:task_name.v1",
	}
	PayloadKeyInput = events.PayloadKey[map[string]any]{
		Name:     "input",
		StableID: "event:input.v1",
	}
	PayloadKeyOutput = events.PayloadKey[map[string]any]{
		Name:     "output",
		StableID: "event:output.v1",
	}
	PayloadKeyError = events.PayloadKey[map[string]any]{
		Name:     "error",
		StableID: "event:error.v1",
	}

	PayloadKeySymbol = events.PayloadKey[string]{
		Name:     "symbol",
		StableID: "event:input.symbol.v1",
	}
	PayloadKeyMode = events.PayloadKey[string]{
		Name:     "mode",
		StableID: "event:input.mode.v1",
	}
	PayloadKeyMarketBars = events.PayloadKey[[]float64]{
		Name:     "market_bars",
		StableID: "event:input.market.bars.v1",
	}
	PayloadKeyMarketOHLCVBars = events.PayloadKey[[]marketdata.OHLCV]{
		Name:     "market_ohlcv_bars",
		StableID: "event:input.market.ohlcv_bars.v1",
	}
	PayloadKeyMarketSpreadBps = events.PayloadKey[float64]{
		Name:     "market_spread_bps",
		StableID: "event:input.market.spread_bps.v1",
	}
	PayloadKeyAccountBalance = events.PayloadKey[float64]{
		Name:     "account_balance",
		StableID: "event:input.account.balance.v1",
	}
	PayloadKeyMarketdataSymbolID = events.PayloadKey[int64]{
		Name:     "marketdata.symbol_id",
		StableID: "event:input.marketdata.symbol_id.v1",
	}
	PayloadKeyMarketdataSymbolCode = events.PayloadKey[string]{
		Name:     "marketdata.symbol_code",
		StableID: "event:input.marketdata.symbol_code.v1",
	}
	PayloadKeyMarketdataTimeframeCode = events.PayloadKey[string]{
		Name:     "marketdata.timeframe_code",
		StableID: "event:input.marketdata.timeframe_code.v1",
	}
	PayloadKeyMarketdataFrom = events.PayloadKey[marketdata.UTCTime]{
		Name:     "marketdata.from",
		StableID: "event:input.marketdata.from.v1",
	}
	PayloadKeyMarketdataTo = events.PayloadKey[marketdata.UTCTime]{
		Name:     "marketdata.to",
		StableID: "event:input.marketdata.to.v1",
	}

	InputKeySymbol             = artifact.Key[string]{Name: "symbol", StableID: "artifact:input.symbol.v1"}
	InputKeyMode               = artifact.Key[string]{Name: "mode", StableID: "artifact:input.mode.v1"}
	InputKeyMarketBars         = artifact.Key[[]float64]{Name: "market.bars", StableID: "artifact:input.market.bars.v1"}
	InputKeyMarketOHLCVBars    = artifact.Key[[]marketdata.OHLCV]{Name: "market.ohlcv_bars", StableID: "artifact:input.market.ohlcv_bars.v1"}
	InputKeyMarketSpreadBps    = artifact.Key[float64]{Name: "market.spread_bps", StableID: "artifact:input.market.spread_bps.v1"}
	InputKeyAccountBalance     = artifact.Key[float64]{Name: "account.balance", StableID: "artifact:input.account.balance.v1"}
	InputKeyMarketdataSymbolID = artifact.Key[int64]{
		Name:     "marketdata.symbol_id",
		StableID: "artifact:input.marketdata.symbol_id.v1",
	}
	InputKeyMarketdataSymbolCode = artifact.Key[string]{
		Name:     "marketdata.symbol_code",
		StableID: "artifact:input.marketdata.symbol_code.v1",
	}
	InputKeyMarketdataTimeframeCode = artifact.Key[string]{
		Name:     "marketdata.timeframe_code",
		StableID: "artifact:input.marketdata.timeframe_code.v1",
	}
	InputKeyMarketdataFrom = artifact.Key[marketdata.UTCTime]{
		Name:     "marketdata.from",
		StableID: "artifact:input.marketdata.from.v1",
	}
	InputKeyMarketdataTo = artifact.Key[marketdata.UTCTime]{
		Name:     "marketdata.to",
		StableID: "artifact:input.marketdata.to.v1",
	}
)
