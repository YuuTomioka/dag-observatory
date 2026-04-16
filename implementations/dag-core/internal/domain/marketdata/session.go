package marketdata

import (
	"fmt"
	"time"
)

type SessionCode string

const SessionTokyo SessionCode = "TOKYO"

type SessionLocalTime struct {
	duration time.Duration
}

func NewSessionLocalTime(hour, minute int) (SessionLocalTime, error) {
	if hour < 0 || hour > 23 {
		return SessionLocalTime{}, fmt.Errorf("marketdata session local time: hour must be 0..23")
	}
	if minute < 0 || minute > 59 {
		return SessionLocalTime{}, fmt.Errorf("marketdata session local time: minute must be 0..59")
	}
	return SessionLocalTime{duration: time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute}, nil
}

func MustSessionLocalTime(hour, minute int) SessionLocalTime {
	t, err := NewSessionLocalTime(hour, minute)
	if err != nil {
		panic(err)
	}
	return t
}

func (t SessionLocalTime) Duration() time.Duration {
	return t.duration
}

func (t SessionLocalTime) IsZero() bool {
	return t.duration == 0
}

type SessionDate struct {
	year  int
	month time.Month
	day   int
}

func NewSessionDate(year int, month time.Month, day int) (SessionDate, error) {
	d := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if d.Year() != year || d.Month() != month || d.Day() != day {
		return SessionDate{}, fmt.Errorf("marketdata session date: invalid date")
	}
	return SessionDate{year: year, month: month, day: day}, nil
}

func MustSessionDate(year int, month time.Month, day int) SessionDate {
	d, err := NewSessionDate(year, month, day)
	if err != nil {
		panic(err)
	}
	return d
}

func SessionDateFromLocalTime(t time.Time) SessionDate {
	year, month, day := t.Date()
	return SessionDate{year: year, month: month, day: day}
}

func (d SessionDate) IsZero() bool {
	return d.year == 0
}

func (d SessionDate) String() string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.year, d.month, d.day)
}

func (d SessionDate) Time(loc *time.Location) time.Time {
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, loc)
}

type Session struct {
	Code       SessionCode
	Location   *time.Location
	OpenLocal  SessionLocalTime
	CloseLocal SessionLocalTime
}

func TokyoSession() Session {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		loc = time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	return Session{
		Code:       SessionTokyo,
		Location:   loc,
		OpenLocal:  MustSessionLocalTime(9, 0),
		CloseLocal: MustSessionLocalTime(15, 0),
	}
}

func (s Session) WindowForDate(date SessionDate) (open UTCTime, close UTCTime, err error) {
	if s.Code == "" {
		return UTCTime{}, UTCTime{}, fmt.Errorf("marketdata session window: code is required")
	}
	if s.Location == nil {
		return UTCTime{}, UTCTime{}, fmt.Errorf("marketdata session window: location is required")
	}
	if date.IsZero() {
		return UTCTime{}, UTCTime{}, fmt.Errorf("marketdata session window: session date is required")
	}
	if s.CloseLocal.Duration() <= s.OpenLocal.Duration() {
		return UTCTime{}, UTCTime{}, fmt.Errorf("marketdata session window: close must be after open")
	}

	localMidnight := date.Time(s.Location)
	openLocal := localMidnight.Add(s.OpenLocal.Duration())
	closeLocal := localMidnight.Add(s.CloseLocal.Duration())
	return NewUTCTime(openLocal), NewUTCTime(closeLocal), nil
}

func (s Session) WindowForTime(t UTCTime) (SessionDate, UTCTime, UTCTime, error) {
	if s.Location == nil {
		return SessionDate{}, UTCTime{}, UTCTime{}, fmt.Errorf("marketdata session window: location is required")
	}
	date := SessionDateFromLocalTime(t.Time().In(s.Location))
	open, close, err := s.WindowForDate(date)
	return date, open, close, err
}
