package marketdata

import "testing"

func TestAggregateTimeframeBarsSetsHighLowTimeWithLastTouchSemantics(t *testing.T) {
	symbolID := SymbolID(1)
	ticks := []Tick{
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-03-01T00:00:01Z"),
			Bid:      NewPriceFromRaw(1000),
			Ask:      NewPriceFromRaw(1002),
		},
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-03-01T00:00:10Z"),
			Bid:      NewPriceFromRaw(1008),
			Ask:      NewPriceFromRaw(1012),
		},
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-03-01T00:00:20Z"),
			Bid:      NewPriceFromRaw(1007),
			Ask:      NewPriceFromRaw(1012),
		},
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-03-01T00:00:30Z"),
			Bid:      NewPriceFromRaw(990),
			Ask:      NewPriceFromRaw(994),
		},
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-03-01T00:00:40Z"),
			Bid:      NewPriceFromRaw(990),
			Ask:      NewPriceFromRaw(993),
		},
	}

	bars := AggregateTimeframeBars(TimeframeM1, symbolID, ticks)
	if len(bars) != 1 {
		t.Fatalf("expected 1 bar, got %d", len(bars))
	}
	bar := bars[0]
	if got := bar.High.Raw(); got != 1012 {
		t.Fatalf("high mismatch: got=%d want=1012", got)
	}
	if !bar.Hightime.Equal(MustParseUTCTime("2026-03-01T00:00:20Z")) {
		t.Fatalf("high time mismatch: got=%s", bar.Hightime)
	}
	if got := bar.Low.Raw(); got != 990 {
		t.Fatalf("low mismatch: got=%d want=990", got)
	}
	if !bar.Lowtime.Equal(MustParseUTCTime("2026-03-01T00:00:40Z")) {
		t.Fatalf("low time mismatch: got=%s", bar.Lowtime)
	}
	if bar.Source != TimeframeBarSourceTickBidAskMid {
		t.Fatalf("source mismatch: got=%q want=%q", bar.Source, TimeframeBarSourceTickBidAskMid)
	}
}
