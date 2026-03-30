package marketdata

const TimeframeBarSourceTickMid = "tick_mid"

type TimeframeBar struct {
	TimeframeCode TimeframeCode
	SymbolID      SymbolID
	OHLCV
	Source string
}

func (b TimeframeBar) IsZero() bool {
	return b.SymbolID == 0 || b.TimeframeCode == "" || b.Opentime.IsZero() || b.Closetime.IsZero()
}
