package marketphase

import "dag-observatory/dag-core/internal/domain/marketdata"

const PhaseBarSourceTickBidAskMid = "tick_bid_ask_mid"

type PhaseBar struct {
	PhaseID  PhaseID
	Market   string
	Timezone string
	SymbolID marketdata.SymbolID
	marketdata.OHLCV
	Source string
}

func (b PhaseBar) IsZero() bool {
	return b.SymbolID == 0 || b.PhaseID == "" || b.Opentime.IsZero() || b.Closetime.IsZero()
}
