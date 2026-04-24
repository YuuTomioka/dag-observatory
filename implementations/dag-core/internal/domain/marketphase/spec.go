package marketphase

import "time"

type PhaseSpec struct {
	ID          PhaseCode
	Market      string
	Timezone    string
	StartHour   int
	StartMinute int
	EndHour     int
	EndMinute   int
	Tags        []string
	Notes       string
}

type ResolvedPhase struct {
	ID         PhaseCode
	Market     string
	Timezone   string
	Active     bool
	LocalNow   time.Time
	LocalStart time.Time
	LocalEnd   time.Time
	UTCStart   time.Time
	UTCEnd     time.Time
	JSTStart   time.Time
	JSTEnd     time.Time
}

func DefaultPhaseSpecs() []PhaseSpec {
	return []PhaseSpec{
		{
			ID:        PhasePreTokyo,
			Market:    "Tokyo",
			Timezone:  "Asia/Tokyo",
			StartHour: 8,
			EndHour:   9,
			Tags:      []string{"tokyo", "pre_open", "prep"},
			Notes:     "Tokyo preparation and pre-fixing buildup.",
		},
		{
			ID:        PhaseTokyoCore,
			Market:    "Tokyo",
			Timezone:  "Asia/Tokyo",
			StartHour: 9,
			EndHour:   15,
			Tags:      []string{"tokyo", "core", "fixing"},
			Notes:     "Tokyo core flow and real-money activity.",
		},
		{
			ID:        PhaseLateTokyo,
			Market:    "Tokyo",
			Timezone:  "Asia/Tokyo",
			StartHour: 15,
			EndHour:   16,
			Tags:      []string{"tokyo", "late", "handoff"},
			Notes:     "Tokyo late session and Europe handoff.",
		},
		{
			ID:        PhaseLondonEarly,
			Market:    "London",
			Timezone:  "Europe/London",
			StartHour: 8,
			EndHour:   11,
			Tags:      []string{"london", "open", "expansion"},
			Notes:     "London open and early directional expansion.",
		},
		{
			ID:        PhaseLondonCore,
			Market:    "London",
			Timezone:  "Europe/London",
			StartHour: 11,
			EndHour:   13,
			Tags:      []string{"london", "core", "continuation"},
			Notes:     "London core trend continuation.",
		},
		{
			ID:        PhaseLondonLate,
			Market:    "London",
			Timezone:  "Europe/London",
			StartHour: 13,
			EndHour:   16,
			Tags:      []string{"london", "late", "pre_ny"},
			Notes:     "London late session and pre-NY alignment.",
		},
		{
			ID:        PhasePreNewYork,
			Market:    "New York",
			Timezone:  "America/New_York",
			StartHour: 7,
			EndHour:   8,
			Tags:      []string{"newyork", "pre_open", "prep"},
			Notes:     "US preparation before core cash flow.",
		},
		{
			ID:        PhaseNewYorkCore,
			Market:    "New York",
			Timezone:  "America/New_York",
			StartHour: 8,
			EndHour:   11,
			Tags:      []string{"newyork", "core", "macro"},
			Notes:     "US core market flow and macro linkage.",
		},
		{
			ID:        PhaseLateNewYork,
			Market:    "New York",
			Timezone:  "America/New_York",
			StartHour: 11,
			EndHour:   16,
			Tags:      []string{"newyork", "late", "unwind"},
			Notes:     "NY late session and position reduction.",
		},
		{
			ID:        PhaseSydneyReset,
			Market:    "Sydney",
			Timezone:  "Australia/Sydney",
			StartHour: 7,
			EndHour:   9,
			Tags:      []string{"sydney", "reset", "thin_liquidity"},
			Notes:     "Low-liquidity reset before Tokyo buildup.",
		},
	}
}
