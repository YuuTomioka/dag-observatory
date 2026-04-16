package marketdata

import (
	"testing"
	"time"
)

func TestTokyoSessionWindowForDate(t *testing.T) {
	session := TokyoSession()
	date := MustSessionDate(2026, time.April, 1)

	open, close, err := session.WindowForDate(date)
	if err != nil {
		t.Fatalf("WindowForDate returned error: %v", err)
	}

	wantOpen := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	wantClose := time.Date(2026, time.April, 1, 6, 0, 0, 0, time.UTC)
	if !open.Time().Equal(wantOpen) {
		t.Fatalf("open mismatch: got=%s want=%s", open.Time(), wantOpen)
	}
	if !close.Time().Equal(wantClose) {
		t.Fatalf("close mismatch: got=%s want=%s", close.Time(), wantClose)
	}
}

func TestSessionDateUsesLocalOpenDate(t *testing.T) {
	session := TokyoSession()
	ts := MustParseUTCTime("2026-04-01T00:30:00Z") // 2026-04-01 09:30 JST

	date, open, close, err := session.WindowForTime(ts)
	if err != nil {
		t.Fatalf("WindowForTime returned error: %v", err)
	}
	if date != MustSessionDate(2026, time.April, 1) {
		t.Fatalf("date mismatch: got=%s", date)
	}
	if !open.Equal(MustParseUTCTime("2026-04-01T00:00:00Z")) {
		t.Fatalf("open mismatch: got=%s", open)
	}
	if !close.Equal(MustParseUTCTime("2026-04-01T06:00:00Z")) {
		t.Fatalf("close mismatch: got=%s", close)
	}
}

func TestSessionLocalTimeRejectsInvalidValues(t *testing.T) {
	if _, err := NewSessionLocalTime(24, 0); err == nil {
		t.Fatalf("expected invalid hour error")
	}
	if _, err := NewSessionLocalTime(9, 60); err == nil {
		t.Fatalf("expected invalid minute error")
	}
}
