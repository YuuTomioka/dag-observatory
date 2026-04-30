package timeframe

import (
	"testing"
	"time"

	marketdata "dag-observatory/dag-core/internal/domain/marketdata"
)

func assertPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	f()
}

// TestTimeframeW1Window_BoundaryAtSaturday0700JST は週足の土曜07:00(JST)境界を検証する。
func TestTimeframeW1Window_BoundaryAtSaturday0700JST(t *testing.T) {
	tests := []struct {
		name      string
		ts        time.Time
		wantOpen  time.Time
		wantClose time.Time
	}{
		{
			name:      "just_before_boundary_belongs_to_previous_week",
			ts:        time.Date(2026, 3, 28, 6, 59, 59, 0, jst),
			wantOpen:  time.Date(2026, 3, 21, 7, 0, 0, 0, jst),
			wantClose: time.Date(2026, 3, 28, 7, 0, 0, 0, jst),
		},
		{
			name:      "at_boundary_belongs_to_new_week",
			ts:        time.Date(2026, 3, 28, 7, 0, 0, 0, jst),
			wantOpen:  time.Date(2026, 3, 28, 7, 0, 0, 0, jst),
			wantClose: time.Date(2026, 4, 4, 7, 0, 0, 0, jst),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			open, close := TimeframeW1.Window(marketdata.NewUTCTime(tt.ts))

			if !open.Time().Equal(tt.wantOpen.UTC()) {
				t.Fatalf("open mismatch: got=%s want=%s", open.Time().UTC(), tt.wantOpen.UTC())
			}
			if !close.Time().Equal(tt.wantClose.UTC()) {
				t.Fatalf("close mismatch: got=%s want=%s", close.Time().UTC(), tt.wantClose.UTC())
			}
			if open.Time().After(tt.ts.UTC()) {
				t.Fatalf("window open must not be after ts: open=%s ts=%s", open.Time().UTC(), tt.ts.UTC())
			}
			if !tt.ts.UTC().Before(close.Time()) {
				t.Fatalf("ts must be strictly before close: ts=%s close=%s", tt.ts.UTC(), close.Time().UTC())
			}
		})
	}
}

// TestTimeframeM5Window_Uses2200UTCSessionOffset は intraday 足が 22:00 UTC 起点で切られることを検証する。
func TestTimeframeM5Window_Uses2200UTCSessionOffset(t *testing.T) {
	ts := time.Date(2026, 3, 30, 22, 3, 12, 0, time.UTC)
	open, close := TimeframeM5.Window(marketdata.NewUTCTime(ts))

	wantOpen := time.Date(2026, 3, 30, 22, 0, 0, 0, time.UTC)
	wantClose := time.Date(2026, 3, 30, 22, 5, 0, 0, time.UTC)

	if !open.Time().Equal(wantOpen) {
		t.Fatalf("open mismatch: got=%s want=%s", open.Time().UTC(), wantOpen)
	}
	if !close.Time().Equal(wantClose) {
		t.Fatalf("close mismatch: got=%s want=%s", close.Time().UTC(), wantClose)
	}
}

// TestTimeframeMN1Window_BoundaryAtFirstDay0700JST は月足の毎月1日07:00(JST)境界を検証する。
func TestTimeframeMN1Window_BoundaryAtFirstDay0700JST(t *testing.T) {
	tests := []struct {
		name      string
		ts        time.Time
		wantOpen  time.Time
		wantClose time.Time
	}{
		{
			name:      "just_before_boundary_belongs_to_previous_month",
			ts:        time.Date(2026, 4, 1, 6, 59, 59, 0, jst),
			wantOpen:  time.Date(2026, 3, 1, 7, 0, 0, 0, jst),
			wantClose: time.Date(2026, 4, 1, 7, 0, 0, 0, jst),
		},
		{
			name:      "at_boundary_belongs_to_new_month",
			ts:        time.Date(2026, 4, 1, 7, 0, 0, 0, jst),
			wantOpen:  time.Date(2026, 4, 1, 7, 0, 0, 0, jst),
			wantClose: time.Date(2026, 5, 1, 7, 0, 0, 0, jst),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			open, close := TimeframeMN1.Window(marketdata.NewUTCTime(tt.ts))
			if !open.Time().Equal(tt.wantOpen.UTC()) {
				t.Fatalf("open mismatch: got=%s want=%s", open.Time().UTC(), tt.wantOpen.UTC())
			}
			if !close.Time().Equal(tt.wantClose.UTC()) {
				t.Fatalf("close mismatch: got=%s want=%s", close.Time().UTC(), tt.wantClose.UTC())
			}
		})
	}
}

func TestTimeframeW1BackfillRange_ExpandsToWeekWindow(t *testing.T) {
	from := marketdata.MustParseUTCTime("2026-03-27T21:59:59Z") // 2026-03-28 06:59:59 JST
	to := marketdata.MustParseUTCTime("2026-04-03T22:00:00Z")   // next weekly boundary

	open, close, err := TimeframeW1.BackfillRange(from, to)
	if err != nil {
		t.Fatalf("BackfillRange returned error: %v", err)
	}

	wantOpen := time.Date(2026, 3, 20, 22, 0, 0, 0, time.UTC)
	wantClose := time.Date(2026, 4, 3, 22, 0, 0, 0, time.UTC)

	if !open.Time().Equal(wantOpen) {
		t.Fatalf("open mismatch: got=%s want=%s", open.Time().UTC(), wantOpen)
	}
	if !close.Time().Equal(wantClose) {
		t.Fatalf("close mismatch: got=%s want=%s", close.Time().UTC(), wantClose)
	}
}

func TestTimeframeMN1BackfillRange_ExpandsAcrossMonthBoundary(t *testing.T) {
	from := marketdata.MustParseUTCTime("2026-03-31T21:59:59Z") // 2026-04-01 06:59:59 JST
	to := marketdata.MustParseUTCTime("2026-04-01T22:00:00Z")   // next monthly window

	open, close, err := TimeframeMN1.BackfillRange(from, to)
	if err != nil {
		t.Fatalf("BackfillRange returned error: %v", err)
	}

	wantOpen := time.Date(2026, 2, 28, 22, 0, 0, 0, time.UTC)
	wantClose := time.Date(2026, 4, 30, 22, 0, 0, 0, time.UTC)

	if !open.Time().Equal(wantOpen) {
		t.Fatalf("open mismatch: got=%s want=%s", open.Time().UTC(), wantOpen)
	}
	if !close.Time().Equal(wantClose) {
		t.Fatalf("close mismatch: got=%s want=%s", close.Time().UTC(), wantClose)
	}
}

// TestTimeframeCodeDefPanicsOnUnknownCode は未定義コード参照時に panic することを検証する。
func TestTimeframeCodeDefPanicsOnUnknownCode(t *testing.T) {
	assertPanics(t, func() {
		_ = TimeframeCode("UNKNOWN").Def()
	})
}

func TestAggregateTimeframeBarsSetsHighLowTimeWithLastTouchSemantics(t *testing.T) {
	symbolID := marketdata.SymbolID(1)
	ticks := []marketdata.Tick{
		{
			SymbolID: symbolID,
			Time:     marketdata.MustParseUTCTime("2026-03-01T00:00:01Z"),
			Bid:      marketdata.NewPriceFromRaw(1000),
			Ask:      marketdata.NewPriceFromRaw(1002),
		},
		{
			SymbolID: symbolID,
			Time:     marketdata.MustParseUTCTime("2026-03-01T00:00:10Z"),
			Bid:      marketdata.NewPriceFromRaw(1008),
			Ask:      marketdata.NewPriceFromRaw(1012),
		},
		{
			SymbolID: symbolID,
			Time:     marketdata.MustParseUTCTime("2026-03-01T00:00:20Z"),
			Bid:      marketdata.NewPriceFromRaw(1007),
			Ask:      marketdata.NewPriceFromRaw(1012),
		},
		{
			SymbolID: symbolID,
			Time:     marketdata.MustParseUTCTime("2026-03-01T00:00:30Z"),
			Bid:      marketdata.NewPriceFromRaw(990),
			Ask:      marketdata.NewPriceFromRaw(994),
		},
		{
			SymbolID: symbolID,
			Time:     marketdata.MustParseUTCTime("2026-03-01T00:00:40Z"),
			Bid:      marketdata.NewPriceFromRaw(990),
			Ask:      marketdata.NewPriceFromRaw(993),
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
	if !bar.Hightime.Equal(marketdata.MustParseUTCTime("2026-03-01T00:00:20Z")) {
		t.Fatalf("high time mismatch: got=%s", bar.Hightime)
	}
	if got := bar.Low.Raw(); got != 990 {
		t.Fatalf("low mismatch: got=%d want=990", got)
	}
	if !bar.Lowtime.Equal(marketdata.MustParseUTCTime("2026-03-01T00:00:40Z")) {
		t.Fatalf("low time mismatch: got=%s", bar.Lowtime)
	}
	if bar.Source != TimeframeBarSourceTickBidAskMid {
		t.Fatalf("source mismatch: got=%q want=%q", bar.Source, TimeframeBarSourceTickBidAskMid)
	}
}
