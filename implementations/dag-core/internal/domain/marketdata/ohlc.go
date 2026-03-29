package marketdata

type OHLCV struct {
	Opentime  UTCTime
	Closetime UTCTime
	Open      Price
	High      Price
	Low       Price
	Close     Price
	Volume    Volume
}

type CandleTrend int8

const (
	CandleBullish CandleTrend = iota + 1
	CandleBearish
	CandleDoji
)

// Trend は始値と終値を比較してローソク足の方向を返す。
func (b OHLCV) Trend() CandleTrend {
	open := b.Open.Raw()
	close := b.Close.Raw()
	switch {
	case close > open:
		return CandleBullish
	case close < open:
		return CandleBearish
	default:
		return CandleDoji
	}
}

// IsBullish は終値が始値より高い陽線かを返す。
func (b OHLCV) IsBullish() bool {
	return b.Trend() == CandleBullish
}

// IsBearish は終値が始値より低い陰線かを返す。
func (b OHLCV) IsBearish() bool {
	return b.Trend() == CandleBearish
}

// IsDoji は始値と終値が同値の同時線かを返す。
func (b OHLCV) IsDoji() bool {
	return b.Trend() == CandleDoji
}

// CandleDelta は高値と安値の差分（足全体のレンジ）を返す。
func (b OHLCV) CandleDelta() Price {
	return b.High.Sub(b.Low).Abs()
}

// BodyDelta は始値と終値の差分（実体サイズ）を返す。
func (b OHLCV) BodyDelta() Price {
	return b.Close.Sub(b.Open).Abs()
}

// WickDelta は実体上端から高値までの差分（上ヒゲ）を返す。
func (b OHLCV) WickDelta() Price {
	bodyTop := b.Open
	if b.Close.Raw() > b.Open.Raw() {
		bodyTop = b.Close
	}
	return b.High.Sub(bodyTop)
}

// TailDelta は安値から実体下端までの差分（下ヒゲ）を返す。
func (b OHLCV) TailDelta() Price {
	bodyBottom := b.Open
	if b.Close.Raw() < b.Open.Raw() {
		bodyBottom = b.Close
	}
	return bodyBottom.Sub(b.Low)
}
