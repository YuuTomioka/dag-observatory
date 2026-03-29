package marketdata

type Tick struct {
	SymbolID SymbolID
	Time     UTCTime
	Bid      Price
	Ask      Price
}

// Spread は Ask と Bid の差分（Ask - Bid）を返す。
// 負のスプレッドもそのまま返す。
func (t Tick) Spread() Price {
	return t.Ask.Sub(t.Bid)
}
