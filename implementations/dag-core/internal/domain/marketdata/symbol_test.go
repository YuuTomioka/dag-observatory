package marketdata

import "testing"

// TestSymbolPriceDeltaConversions は pips/ticks と Price 差分の相互変換を検証する。
func TestSymbolPriceDeltaConversions(t *testing.T) {
	s := Symbol{
		PriceScale:  5,
		TickSizeRaw: 1,
		PipSizeRaw:  10,
	}

	if got := s.PriceDeltaFromPips(3).Raw(); got != 30 {
		t.Fatalf("PriceDeltaFromPips mismatch: got=%d want=30", got)
	}
	if got := s.PriceDeltaFromTicks(7).Raw(); got != 7 {
		t.Fatalf("PriceDeltaFromTicks mismatch: got=%d want=7", got)
	}
	if got := s.PipsOfPriceDelta(NewPriceFromRaw(49)); got != 4 {
		t.Fatalf("PipsOfPriceDelta mismatch: got=%d want=4", got)
	}
	if got := s.TicksOfPriceDelta(NewPriceFromRaw(49)); got != 49 {
		t.Fatalf("TicksOfPriceDelta mismatch: got=%d want=49", got)
	}
	if got := s.PipsBetween(NewPriceFromRaw(100), NewPriceFromRaw(130)); got != 3 {
		t.Fatalf("PipsBetween mismatch: got=%d want=3", got)
	}
	if got := s.PipsBetween(NewPriceFromRaw(130), NewPriceFromRaw(100)); got != 3 {
		t.Fatalf("PipsBetween(abs) mismatch: got=%d want=3", got)
	}
	if got := s.TicksBetween(NewPriceFromRaw(100), NewPriceFromRaw(97)); got != 3 {
		t.Fatalf("TicksBetween(abs) mismatch: got=%d want=3", got)
	}
}

// TestSymbolFloatConversions は UI 境界向けの float64 変換と丸めを検証する。
func TestSymbolFloatConversions(t *testing.T) {
	s := Symbol{PriceScale: 5}

	if got := s.PriceFromFloat64(1.23456).Raw(); got != 123456 {
		t.Fatalf("PriceFromFloat64 mismatch: got=%d want=123456", got)
	}
	if got := s.PriceFromFloat64(1.234567).Raw(); got != 123457 {
		t.Fatalf("PriceFromFloat64 rounding mismatch: got=%d want=123457", got)
	}
	if got := s.Float64FromPrice(NewPriceFromRaw(123456)); got != 1.23456 {
		t.Fatalf("Float64FromPrice mismatch: got=%v want=1.23456", got)
	}
}

// TestSymbolPanicsWhenTickOrPipSizeIsZero はサイズ 0 の設定で panic することを検証する。
func TestSymbolPanicsWhenTickOrPipSizeIsZero(t *testing.T) {
	withZeroPip := Symbol{PipSizeRaw: 0, TickSizeRaw: 1}
	withZeroTick := Symbol{PipSizeRaw: 1, TickSizeRaw: 0}

	assertPanics(t, func() {
		_ = withZeroPip.PipsOfPriceDelta(NewPriceFromRaw(10))
	})
	assertPanics(t, func() {
		_ = withZeroTick.TicksOfPriceDelta(NewPriceFromRaw(10))
	})
	assertPanics(t, func() {
		_ = withZeroPip.PipsBetween(NewPriceFromRaw(1), NewPriceFromRaw(2))
	})
	assertPanics(t, func() {
		_ = withZeroTick.TicksBetween(NewPriceFromRaw(1), NewPriceFromRaw(2))
	})
}
