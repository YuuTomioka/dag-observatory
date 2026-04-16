package marketdata

import (
	"testing"
	"time"
)

func TestAggregateSessionBarUsesHalfOpenWindowAndBidAskMidSemantics(t *testing.T) {
	session := TokyoSession()
	date := MustSessionDate(2026, time.April, 1)
	symbolID := SymbolID(1)
	ticks := []Tick{
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-03-31T23:59:59Z"),
			Bid:      NewPriceFromRaw(900),
			Ask:      NewPriceFromRaw(902),
		},
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-04-01T00:00:00Z"),
			Bid:      NewPriceFromRaw(1000),
			Ask:      NewPriceFromRaw(1002),
		},
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-04-01T01:00:00Z"),
			Bid:      NewPriceFromRaw(1008),
			Ask:      NewPriceFromRaw(1012),
		},
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-04-01T02:00:00Z"),
			Bid:      NewPriceFromRaw(1007),
			Ask:      NewPriceFromRaw(1012),
		},
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-04-01T03:00:00Z"),
			Bid:      NewPriceFromRaw(990),
			Ask:      NewPriceFromRaw(994),
		},
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-04-01T04:00:00Z"),
			Bid:      NewPriceFromRaw(990),
			Ask:      NewPriceFromRaw(993),
		},
		{
			SymbolID: symbolID,
			Time:     MustParseUTCTime("2026-04-01T06:00:00Z"),
			Bid:      NewPriceFromRaw(1200),
			Ask:      NewPriceFromRaw(1202),
		},
	}

	bar, ok, err := AggregateSessionBar(session, date, symbolID, ticks)
	if err != nil {
		t.Fatalf("AggregateSessionBar returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected session bar")
	}
	if bar.SessionCode != SessionTokyo {
		t.Fatalf("session code mismatch: got=%q", bar.SessionCode)
	}
	if bar.SessionDate != date {
		t.Fatalf("session date mismatch: got=%s want=%s", bar.SessionDate, date)
	}
	if !bar.Opentime.Equal(MustParseUTCTime("2026-04-01T00:00:00Z")) {
		t.Fatalf("open time mismatch: got=%s", bar.Opentime)
	}
	if !bar.Closetime.Equal(MustParseUTCTime("2026-04-01T06:00:00Z")) {
		t.Fatalf("close time mismatch: got=%s", bar.Closetime)
	}
	if got := bar.Open.Raw(); got != 1001 {
		t.Fatalf("open mismatch: got=%d want=1001", got)
	}
	if got := bar.High.Raw(); got != 1012 {
		t.Fatalf("high mismatch: got=%d want=1012", got)
	}
	if !bar.Hightime.Equal(MustParseUTCTime("2026-04-01T02:00:00Z")) {
		t.Fatalf("high time mismatch: got=%s", bar.Hightime)
	}
	if got := bar.Low.Raw(); got != 990 {
		t.Fatalf("low mismatch: got=%d want=990", got)
	}
	if !bar.Lowtime.Equal(MustParseUTCTime("2026-04-01T04:00:00Z")) {
		t.Fatalf("low time mismatch: got=%s", bar.Lowtime)
	}
	if got := bar.Close.Raw(); got != 991 {
		t.Fatalf("close mismatch: got=%d want=991", got)
	}
	if got := int64(bar.Volume); got != 5 {
		t.Fatalf("volume mismatch: got=%d want=5", got)
	}
	if bar.Source != TimeframeBarSourceTickBidAskMid {
		t.Fatalf("source mismatch: got=%q want=%q", bar.Source, TimeframeBarSourceTickBidAskMid)
	}
}

func TestAggregateSessionBarReturnsFalseWhenNoTicksInWindow(t *testing.T) {
	session := TokyoSession()
	date := MustSessionDate(2026, time.April, 1)
	ticks := []Tick{
		{
			SymbolID: 1,
			Time:     MustParseUTCTime("2026-03-31T23:59:59Z"),
			Bid:      NewPriceFromRaw(1000),
			Ask:      NewPriceFromRaw(1002),
		},
		{
			SymbolID: 1,
			Time:     MustParseUTCTime("2026-04-01T06:00:00Z"),
			Bid:      NewPriceFromRaw(1000),
			Ask:      NewPriceFromRaw(1002),
		},
	}

	_, ok, err := AggregateSessionBar(session, date, 1, ticks)
	if err != nil {
		t.Fatalf("AggregateSessionBar returned error: %v", err)
	}
	if ok {
		t.Fatalf("expected no session bar")
	}
}
