package marketdata

import (
	"math"
)

type SymbolID int64

type Symbol struct {
	ID    SymbolID
	Code  string
	Base  string
	Quote string

	// Price = Raw / 10^PriceScale
	PriceScale uint8

	// 1Tick が PriceRaw 何カウントか
	TickSizeRaw int64

	// 1Pip が PriceRaw 何カウントか
	PipSizeRaw int64
}

// PriceDeltaFromPips は pips 値を Price の差分に変換する。
func (s Symbol) PriceDeltaFromPips(pips int64) Price {
	return NewPriceFromRaw(pips * s.PipSizeRaw)
}

// PriceDeltaFromTicks は ticks 値を Price の差分に変換する。
func (s Symbol) PriceDeltaFromTicks(ticks int64) Price {
	return NewPriceFromRaw(ticks * s.TickSizeRaw)
}

// PipsOfPriceDelta は Price 差分を pips に変換する。PipSizeRaw が 0 の場合は panic する。
func (s Symbol) PipsOfPriceDelta(delta Price) int64 {
	if s.PipSizeRaw == 0 {
		panic("Symbol.PipSizeRaw is 0")
	}
	return delta.Raw() / s.PipSizeRaw
}

// TicksOfPriceDelta は Price 差分を ticks に変換する。TickSizeRaw が 0 の場合は panic する。
func (s Symbol) TicksOfPriceDelta(delta Price) int64 {
	if s.TickSizeRaw == 0 {
		panic("Symbol.TickSizeRaw is 0")
	}
	return delta.Raw() / s.TickSizeRaw
}

// PipsBetween は 2 つの Price の絶対差分を pips に変換する。PipSizeRaw が 0 の場合は panic する。
func (s Symbol) PipsBetween(rawA, rawB Price) int64 {
	diff := rawB.Sub(rawA).Abs().Raw()
	if s.PipSizeRaw == 0 {
		panic("Symbol.PipSizeRaw is 0")
	}
	return diff / s.PipSizeRaw
}

// TicksBetween は 2 つの Price の絶対差分を ticks に変換する。TickSizeRaw が 0 の場合は panic する。
func (s Symbol) TicksBetween(rawA, rawB Price) int64 {
	diff := rawB.Sub(rawA).Abs().Raw()
	if s.TickSizeRaw == 0 {
		panic("Symbol.TickSizeRaw is 0")
	}
	return diff / s.TickSizeRaw
}

// PriceFromFloat64 は UI/外部境界向けに float64 の価格を内部 Price へ丸め変換する。
// PriceScale に従い 10^scale を掛け、四捨五入して整数化する。
func (s Symbol) PriceFromFloat64(f float64) Price {
	scale := int(s.PriceScale)
	raw := math.Round(f * math.Pow10(scale))
	return NewPriceFromRaw(int64(raw))
}

// Float64FromPrice は UI/外部境界向けに内部 Price を float64 表現へ変換する。
func (s Symbol) Float64FromPrice(p Price) float64 {
	scale := int(s.PriceScale)
	return float64(p.Raw()) / math.Pow10(scale)
}
