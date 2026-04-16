package marketdata

type SessionBar struct {
	SessionCode SessionCode
	SessionDate SessionDate
	SymbolID    SymbolID
	OHLCV
	Source string
}

func (b SessionBar) IsZero() bool {
	return b.SymbolID == 0 || b.SessionCode == "" || b.SessionDate.IsZero() || b.Opentime.IsZero() || b.Closetime.IsZero()
}
