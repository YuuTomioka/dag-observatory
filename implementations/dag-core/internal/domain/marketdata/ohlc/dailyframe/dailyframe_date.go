package dailyframe

import (
	"fmt"
	"time"
)

type DailyFrameDate struct {
	year  int
	month time.Month
	day   int
}

const dailyFrameDateLayout = "2006-01-02"

func NewDailyFrameDate(year int, month time.Month, day int) (DailyFrameDate, error) {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if t.Year() != year || t.Month() != month || t.Day() != day {
		return DailyFrameDate{}, fmt.Errorf("dailyframe date: invalid date %04d-%02d-%02d", year, month, day)
	}
	return DailyFrameDate{
		year:  year,
		month: month,
		day:   day,
	}, nil
}

func MustNewDailyFrameDate(year int, month time.Month, day int) DailyFrameDate {
	d, err := NewDailyFrameDate(year, month, day)
	if err != nil {
		panic(err)
	}
	return d
}

func ParseDailyFrameDate(s string) (DailyFrameDate, error) {
	t, err := time.ParseInLocation(dailyFrameDateLayout, s, time.UTC)
	if err != nil {
		return DailyFrameDate{}, fmt.Errorf("dailyframe parse date: %w", err)
	}
	return NewDailyFrameDate(t.Year(), t.Month(), t.Day())
}

func MustParseDailyFrameDate(s string) DailyFrameDate {
	d, err := ParseDailyFrameDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

func NewDailyFrameDateFromTime(t time.Time) DailyFrameDate {
	y, m, d := t.Date()
	return MustNewDailyFrameDate(y, m, d)
}

// NewDailyFrameDateFromOpentimeUTC maps an open-time instant to a daily frame date using
// UTC 22:00 as the daily cutover (JST 07:00 next day).
func NewDailyFrameDateFromOpentimeUTC(openTime time.Time) DailyFrameDate {
	utc := openTime.UTC()
	base := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	if utc.Hour() >= 22 {
		base = base.AddDate(0, 0, 1)
	}
	return MustNewDailyFrameDate(base.Year(), base.Month(), base.Day())
}

func (d DailyFrameDate) IsZero() bool {
	return d.year == 0 && d.month == 0 && d.day == 0
}

func (d DailyFrameDate) String() string {
	if d.IsZero() {
		return "0000-00-00"
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.year, d.month, d.day)
}

func (d DailyFrameDate) Time(loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, loc)
}
