package ohlc

import (
	"testing"

	marketdata "dag-observatory/dag-core/internal/domain/marketdata"
)

// TestOHLCVTrendAndFlags はローソク足方向判定と補助フラグの整合性を検証する。
func TestOHLCVTrendAndFlags(t *testing.T) {
	bull := OHLCV{Open: marketdata.NewPriceFromRaw(100), Close: marketdata.NewPriceFromRaw(110)}
	bear := OHLCV{Open: marketdata.NewPriceFromRaw(110), Close: marketdata.NewPriceFromRaw(100)}
	doji := OHLCV{Open: marketdata.NewPriceFromRaw(100), Close: marketdata.NewPriceFromRaw(100)}

	if got := bull.Trend(); got != CandleBullish {
		t.Fatalf("bull trend mismatch: got=%v want=%v", got, CandleBullish)
	}
	if !bull.IsBullish() || bull.IsBearish() || bull.IsDoji() {
		t.Fatalf("bull flags mismatch")
	}

	if got := bear.Trend(); got != CandleBearish {
		t.Fatalf("bear trend mismatch: got=%v want=%v", got, CandleBearish)
	}
	if !bear.IsBearish() || bear.IsBullish() || bear.IsDoji() {
		t.Fatalf("bear flags mismatch")
	}

	if got := doji.Trend(); got != CandleDoji {
		t.Fatalf("doji trend mismatch: got=%v want=%v", got, CandleDoji)
	}
	if !doji.IsDoji() || doji.IsBullish() || doji.IsBearish() {
		t.Fatalf("doji flags mismatch")
	}
}

// TestOHLCVDeltas はレンジ・実体・上下ヒゲの差分計算を検証する。
func TestOHLCVDeltas(t *testing.T) {
	b := OHLCV{
		Open:  marketdata.NewPriceFromRaw(100),
		High:  marketdata.NewPriceFromRaw(125),
		Low:   marketdata.NewPriceFromRaw(90),
		Close: marketdata.NewPriceFromRaw(115),
	}
	if got := b.CandleDelta().Raw(); got != 35 {
		t.Fatalf("CandleDelta mismatch: got=%d want=35", got)
	}
	if got := b.BodyDelta().Raw(); got != 15 {
		t.Fatalf("BodyDelta mismatch: got=%d want=15", got)
	}
	if got := b.WickDelta().Raw(); got != 10 {
		t.Fatalf("WickDelta mismatch: got=%d want=10", got)
	}
	if got := b.TailDelta().Raw(); got != 10 {
		t.Fatalf("TailDelta mismatch: got=%d want=10", got)
	}
}
