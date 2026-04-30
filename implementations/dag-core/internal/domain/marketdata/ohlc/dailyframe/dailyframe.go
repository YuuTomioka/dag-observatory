package dailyframe

type DailyframeID int64
type DailyframeCode string

type Dailyframe struct {
	ID   DailyframeID
	Code DailyframeCode

	StartTimezone string
	StartHour     int
	StartMinute   int

	EndTimezone string
	EndHour     int
	EndMinute   int
}
