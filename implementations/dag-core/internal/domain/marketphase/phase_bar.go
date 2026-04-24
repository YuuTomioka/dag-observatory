package marketphase

import (
	"dag-observatory/dag-core/internal/domain/marketdata"
)

const PhaseBarSourceTickBidAskMid = "tick_bid_ask_mid"

type PhaseBar struct {
	PhaseCode PhaseCode
	PhaseDate PhaseDate
	Market    string
	Timezone  string
	SymbolID  marketdata.SymbolID
	marketdata.OHLCV
	Source string
}

func (b PhaseBar) IsZero() bool {
	return b.SymbolID == 0 || b.PhaseCode == "" || b.PhaseDate.IsZero() || b.Opentime.IsZero() || b.Closetime.IsZero()
}
