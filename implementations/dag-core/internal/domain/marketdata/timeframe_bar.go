package marketdata

const TimeframeBarSourceTickBidAskMid = "tick_bid_ask_mid"

type TimeframeBar struct {
	TimeframeCode TimeframeCode
	SymbolID      SymbolID
	OHLCV
	Source string
}

func (b TimeframeBar) IsZero() bool {
	return b.SymbolID == 0 || b.TimeframeCode == "" || b.Opentime.IsZero() || b.Closetime.IsZero()
}
