package marketphase

import (
	"fmt"
	"time"
)

type PhaseCode string

const (
	PhasePreTokyo    PhaseCode = "pre_tokyo"
	PhaseTokyoCore   PhaseCode = "tokyo_core"
	PhaseLateTokyo   PhaseCode = "late_tokyo"
	PhaseLondonEarly PhaseCode = "london_early"
	PhaseLondonCore  PhaseCode = "london_core"
	PhaseLondonLate  PhaseCode = "london_late"
	PhasePreNewYork  PhaseCode = "pre_newyork"
	PhaseNewYorkCore PhaseCode = "newyork_core"
	PhaseLateNewYork PhaseCode = "late_newyork"
	PhaseSydneyReset PhaseCode = "sydney_reset"
)

type PhaseDate struct {
	year  int
	month time.Month
	day   int
}

const phaseDateLayout = "2006-01-02"

func NewPhaseDate(year int, month time.Month, day int) (PhaseDate, error) {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if t.Year() != year || t.Month() != month || t.Day() != day {
		return PhaseDate{}, fmt.Errorf("marketphase phase date: invalid date %04d-%02d-%02d", year, month, day)
	}
	return PhaseDate{
		year:  year,
		month: month,
		day:   day,
	}, nil
}

func MustNewPhaseDate(year int, month time.Month, day int) PhaseDate {
	d, err := NewPhaseDate(year, month, day)
	if err != nil {
		panic(err)
	}
	return d
}

func ParsePhaseDate(s string) (PhaseDate, error) {
	t, err := time.ParseInLocation(phaseDateLayout, s, time.UTC)
	if err != nil {
		return PhaseDate{}, fmt.Errorf("marketphase parse phase date: %w", err)
	}
	return NewPhaseDate(t.Year(), t.Month(), t.Day())
}

func MustParsePhaseDate(s string) PhaseDate {
	d, err := ParsePhaseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

func NewPhaseDateFromTime(t time.Time) PhaseDate {
	y, m, d := t.Date()
	return MustNewPhaseDate(y, m, d)
}

// NewPhaseDateFromOpentimeUTC maps an open-time instant to a phase date using
// UTC 22:00 as the daily cutover (JST 07:00 next day).
func NewPhaseDateFromOpentimeUTC(openTime time.Time) PhaseDate {
	utc := openTime.UTC()
	base := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	if utc.Hour() >= 22 {
		base = base.AddDate(0, 0, 1)
	}
	return MustNewPhaseDate(base.Year(), base.Month(), base.Day())
}

func (d PhaseDate) IsZero() bool {
	return d.year == 0 || d.month == 0 || d.day == 0
}

func (d PhaseDate) String() string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.year, d.month, d.day)
}

func (d PhaseDate) Time(loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, loc)
}
