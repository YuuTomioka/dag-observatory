package dailyframe

import (
	"dag-observatory/dag-core/internal/domain/marketdata"
	"dag-observatory/dag-core/internal/domain/marketdata/ohlc"
)

const DailyFrameBarSourceTickBidAskMid = "tick_bid_ask_mid"

type DailyFrameBar struct {
	SymbolID       marketdata.SymbolID
	DailyframeID   DailyframeID
	DailyFrameDate DailyFrameDate

	ohlc.OHLCV
	Source string
}
