package marketdata

import (
	"time"
)

type TimeframeCode string

const (
	TimeframeM1  TimeframeCode = "M1"
	TimeframeM5  TimeframeCode = "M5"
	TimeframeM15 TimeframeCode = "M15"
	TimeframeM30 TimeframeCode = "M30"
	TimeframeH1  TimeframeCode = "H1"
	TimeframeH4  TimeframeCode = "H4"
	TimeframeD1  TimeframeCode = "D1"
	TimeframeW1  TimeframeCode = "W1"  // Week1
	TimeframeMN1 TimeframeCode = "MN1" // Month1
)

type TimeframeDef struct {
	Code     TimeframeCode
	Duration time.Duration
}

var TimeframeDefs = map[TimeframeCode]TimeframeDef{
	TimeframeM1:  {Code: TimeframeM1, Duration: time.Minute * 1},
	TimeframeM5:  {Code: TimeframeM5, Duration: time.Minute * 5},
	TimeframeM15: {Code: TimeframeM15, Duration: time.Minute * 15},
	TimeframeM30: {Code: TimeframeM30, Duration: time.Minute * 30},
	TimeframeH1:  {Code: TimeframeH1, Duration: time.Hour * 1},
	TimeframeH4:  {Code: TimeframeH4, Duration: time.Hour * 4},
	TimeframeD1:  {Code: TimeframeD1, Duration: time.Hour * 24},
	// W1 / MN1 では Duration は dummy 実装
	TimeframeW1:  {Code: TimeframeW1, Duration: 7 * 24 * time.Hour},
	TimeframeMN1: {Code: TimeframeMN1, Duration: 30 * 24 * time.Hour},
}

// Def は TimeframeCode に対応する定義を返す。未定義コードは panic する。
func (c TimeframeCode) Def() TimeframeDef {
	def, ok := TimeframeDefs[c]
	if !ok {
		panic("unknown TimeframeCode: " + string(c))
	}
	return def
}

// Duration は TimeframeCode に対応する時間幅を返す。
func (c TimeframeCode) Duration() time.Duration {
	return c.Def().Duration
}

// --- Window 計算 -------------------------------------------------------------
// 要件：
// - すべて UTC ベースで計算する
// - intraday / D1: UTC+0 で 22:00 起点（＝UTC+9で翌日 07:00）
// - W1: UTC+9 で「土曜 07:00」を起点とする週足
// - MN1: UTC+9 で「各月 1日の 07:00」を起点とする月足

// 22:00 UTC 起点
var sessionOffset = 22 * time.Hour

// JST (UTC+9) ロケーション
var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// Window は指定時刻 t が属する足の [open, close) を返す。
func (c TimeframeCode) Window(t UTCTime) (open UTCTime, close UTCTime) {
	base := t.Time().UTC()

	switch c {
	case TimeframeW1:
		openUTC, closeUTC := week1Window(base)
		return NewUTCTime(openUTC), NewUTCTime(closeUTC)

	case TimeframeMN1:
		openUTC, closeUTC := month1Window(base)
		return NewUTCTime(openUTC), NewUTCTime(closeUTC)

	default:
		openUTC, closeUTC := intradayWindow(c, base)
		return NewUTCTime(openUTC), NewUTCTime(closeUTC)
	}
}

// intradayWindow は M1〜H4/D1 向けに、22:00 UTC 起点で時刻をバケット化する。
func intradayWindow(c TimeframeCode, ts time.Time) (openUTC, closeUTC time.Time) {
	def := c.Def()
	dur := def.Duration

	shifted := ts.Add(-sessionOffset)
	bucketStart := shifted.Truncate(dur)
	openUTC = bucketStart.Add(sessionOffset)
	closeUTC = openUTC.Add(dur)
	return
}

// week1Window は JST 土曜 07:00 を始点として、週足の [open, close) を計算する。
func week1Window(ts time.Time) (openUTC, closeUTC time.Time) {
	tJST := ts.In(jst)

	weekday := int(tJST.Weekday())
	target := int(time.Saturday)
	deltaDays := (weekday - target + 7) % 7
	weekStartDateJST := tJST.AddDate(0, 0, -deltaDays)

	y, m, d := weekStartDateJST.Date()
	openJST := time.Date(y, m, d, 7, 0, 0, 0, jst)
	if tJST.Before(openJST) {
		openJST = openJST.AddDate(0, 0, -7)
	}
	openUTC = openJST.UTC()
	closeUTC = openUTC.Add(7 * 24 * time.Hour)
	return
}

// month1Window は JST 毎月 1 日 07:00 を始点として、月足の [open, close) を計算する。
func month1Window(ts time.Time) (openUTC, closeUTC time.Time) {
	tJST := ts.In(jst)
	y, m, _ := tJST.Date()

	thisMonthStartJST := time.Date(y, m, 1, 7, 0, 0, 0, jst)

	if tJST.Before(thisMonthStartJST) {
		thisMonthStartJST = thisMonthStartJST.AddDate(0, -1, 0)
	}

	openJST := thisMonthStartJST
	nextMonthStartJST := openJST.AddDate(0, 1, 0)

	openUTC = openJST.UTC()
	closeUTC = nextMonthStartJST.UTC()
	return
}
