package marketdata

import (
	"testing"
	"time"
)

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
			open, close := TimeframeW1.Window(NewUTCTime(tt.ts))

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
	open, close := TimeframeM5.Window(NewUTCTime(ts))

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
			open, close := TimeframeMN1.Window(NewUTCTime(tt.ts))
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
	from := MustParseUTCTime("2026-03-27T21:59:59Z") // 2026-03-28 06:59:59 JST
	to := MustParseUTCTime("2026-04-03T22:00:00Z")   // next weekly boundary

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
	from := MustParseUTCTime("2026-03-31T21:59:59Z") // 2026-04-01 06:59:59 JST
	to := MustParseUTCTime("2026-04-01T22:00:00Z")   // next monthly window

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
