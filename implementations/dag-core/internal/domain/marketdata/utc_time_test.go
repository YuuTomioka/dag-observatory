package marketdata

import (
	"encoding/json"
	"testing"
	"time"
)

// TestUTCTimeNormalizationAndComparison は UTC 正規化と基本比較/演算を検証する。
func TestUTCTimeNormalizationAndComparison(t *testing.T) {
	tokyo := time.FixedZone("Asia/Tokyo", 9*60*60)
	local := time.Date(2026, 3, 30, 12, 34, 56, 123456789, tokyo)
	u := NewUTCTime(local)

	if u.Time().Location() != time.UTC {
		t.Fatalf("NewUTCTime must normalize to UTC")
	}

	a := MustParseUTCTime("2026-03-30T03:34:56Z")
	b := MustParseUTCTime("2026-03-30T03:34:57Z")
	if got := a.Compare(b); got != -1 {
		t.Fatalf("Compare mismatch: got=%d want=-1", got)
	}
	if !a.Before(b) || !b.After(a) || a.Equal(b) {
		t.Fatalf("Before/After/Equal mismatch")
	}
	if !b.InRangeInclusive(a, b) {
		t.Fatalf("InRangeInclusive mismatch")
	}
	if got := a.Add(2 * time.Second); !got.Equal(MustParseUTCTime("2026-03-30T03:34:58Z")) {
		t.Fatalf("Add mismatch: got=%s", got.String())
	}
	if got := b.Sub(a); got != time.Second {
		t.Fatalf("Sub mismatch: got=%v want=1s", got)
	}
}

// TestUTCTimeJSONRoundTripAndNullHandling は JSON 往復変換と null 取り扱いを検証する。
func TestUTCTimeJSONRoundTripAndNullHandling(t *testing.T) {
	original := MustParseUTCTime("2026-03-30T03:34:56.123456789Z")
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var decoded UTCTime
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("UnmarshalJSON round-trip error: %v", err)
	}
	if !decoded.Equal(original) {
		t.Fatalf("round-trip mismatch: got=%s want=%s", decoded.String(), original.String())
	}

	var zero UTCTime
	if err := json.Unmarshal([]byte("null"), &zero); err != nil {
		t.Fatalf("UnmarshalJSON null error: %v", err)
	}
	if !zero.IsZero() {
		t.Fatalf("null should decode to zero time")
	}
}

// TestUTCTimeUnmarshalJSONRejectsInvalidFormat は不正フォーマット入力がエラーになることを検証する。
func TestUTCTimeUnmarshalJSONRejectsInvalidFormat(t *testing.T) {
	var u UTCTime
	if err := json.Unmarshal([]byte(`"not-a-time"`), &u); err == nil {
		t.Fatalf("expected error for invalid time format")
	}
}

// TestMustParseUTCTimePanicsOnInvalidInput は不正文字列で panic することを検証する。
func TestMustParseUTCTimePanicsOnInvalidInput(t *testing.T) {
	assertPanics(t, func() {
		_ = MustParseUTCTime("invalid")
	})
}
