package marketdata

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type UTCTime time.Time

const utcLayout = time.RFC3339Nano

// NewUTCTime は任意の time.Time を UTC 正規化した UTCTime に変換する。
func NewUTCTime(t time.Time) UTCTime {
	return UTCTime(t.UTC())
}

// NowUTCTime は現在時刻を UTC の UTCTime で返す。
func NowUTCTime() UTCTime {
	return NewUTCTime(time.Now().UTC())
}

// Time は UTCTime を time.Time として返す。
func (u UTCTime) Time() time.Time {
	return time.Time(u)
}

// IsZero はゼロ値時刻かを返す。
func (u UTCTime) IsZero() bool {
	return time.Time(u).IsZero()
}

// String は RFC3339Nano 形式の UTC 文字列を返す。ゼロ値の場合は空文字を返す。
func (u UTCTime) String() string {
	if u.IsZero() {
		return ""
	}
	return time.Time(u).Format(utcLayout)
}

// Compare は u と other を比較し、u<other なら -1、u>other なら 1、等しい場合は 0 を返す。
func (u UTCTime) Compare(other UTCTime) int {
	ut := u.Time()
	ot := other.Time()
	if ut.Before(ot) {
		return -1
	}
	if ut.After(ot) {
		return 1
	}
	return 0
}

// Before は u が other より前かを返す。
func (u UTCTime) Before(other UTCTime) bool { return u.Time().Before(other.Time()) }

// After は u が other より後かを返す。
func (u UTCTime) After(other UTCTime) bool  { return u.Time().After(other.Time()) }

// Equal は同一インスタントかを返す（time.Time の == ではなく Equal 判定）。
func (u UTCTime) Equal(other UTCTime) bool { return u.Time().Equal(other.Time()) }

// InRangeInclusive は u が [from, to] の範囲内（端を含む）かを返す。
func (u UTCTime) InRangeInclusive(from, to UTCTime) bool {
	if u.Before(from) {
		return false
	}
	if u.After(to) {
		return false
	}
	return true
}

// Add は duration を加算した新しい UTCTime を返す。
func (u UTCTime) Add(d time.Duration) UTCTime {
	return NewUTCTime(u.Time().Add(d))
}

// Sub は u - other の時間差を返す。
func (u UTCTime) Sub(other UTCTime) time.Duration {
	return u.Time().Sub(other.Time())
}

// MarshalJSON は UTCTime を JSON にシリアライズする。ゼロ値は null として出力する。
func (u UTCTime) MarshalJSON() ([]byte, error) {
	if u.IsZero() {
		return []byte(`null`), nil
	}
	return json.Marshal(u.String())
}

// UnmarshalJSON は JSON 文字列（または null/空文字）から UTCTime を復元する。
func (u *UTCTime) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" || s == `""` || s == "" {
		*u = UTCTime{}
		return nil
	}

	s = strings.Trim(s, `"`)

	t, err := time.Parse(utcLayout, s)
	if err != nil {
		return fmt.Errorf("UTCTime.UnmarshalJSON: %w", err)
	}
	*u = NewUTCTime(t)
	return nil
}

// UnixNano は Unix epoch からの経過ナノ秒を返す。
func (u UTCTime) UnixNano() int64 { return u.Time().UnixNano() }

// MustParseUTCTime は RFC3339Nano 文字列を UTCTime に変換する。失敗時は panic する。
func MustParseUTCTime(s string) UTCTime {
	t, err := time.Parse(utcLayout, s)
	if err != nil {
		panic(err)
	}
	return NewUTCTime(t)
}
