package marketphase

import "time"

type PhaseSpec struct {
	ID                PhaseID
	Category          PhaseCategory
	Market            string
	Timezone          string
	StartHour         int
	StartMinute       int
	EndHour           int
	EndMinute         int
	Liquidity         LiquidityLevel
	BreakoutBias      float64
	MeanReversionBias float64
	Tags              []string
	Notes             string
}

type ResolvedPhase struct {
	ID         PhaseID
	Category   PhaseCategory
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
			ID:                PhasePreTokyo,
			Category:          PhaseCategorySingle,
			Market:            "Tokyo",
			Timezone:          "Asia/Tokyo",
			StartHour:         8,
			EndHour:           9,
			Liquidity:         LiquidityMedium,
			BreakoutBias:      0.50,
			MeanReversionBias: 0.20,
			Tags:              []string{"tokyo", "pre_open", "prep"},
			Notes:             "Tokyo preparation and pre-fixing buildup.",
		},
		{
			ID:                PhaseTokyoCore,
			Category:          PhaseCategorySingle,
			Market:            "Tokyo",
			Timezone:          "Asia/Tokyo",
			StartHour:         9,
			EndHour:           15,
			Liquidity:         LiquidityMedium,
			BreakoutBias:      0.30,
			MeanReversionBias: 0.50,
			Tags:              []string{"tokyo", "core", "fixing"},
			Notes:             "Tokyo core flow and real-money activity.",
		},
		{
			ID:                PhaseLateTokyo,
			Category:          PhaseCategorySingle,
			Market:            "Tokyo",
			Timezone:          "Asia/Tokyo",
			StartHour:         15,
			EndHour:           16,
			Liquidity:         LiquidityMedium,
			BreakoutBias:      0.50,
			MeanReversionBias: 0.20,
			Tags:              []string{"tokyo", "late", "handoff"},
			Notes:             "Tokyo late session and Europe handoff.",
		},
		{
			ID:                PhaseLondonEarly,
			Category:          PhaseCategorySingle,
			Market:            "London",
			Timezone:          "Europe/London",
			StartHour:         8,
			EndHour:           11,
			Liquidity:         LiquidityHigh,
			BreakoutBias:      0.80,
			MeanReversionBias: 0.10,
			Tags:              []string{"london", "open", "expansion"},
			Notes:             "London open and early directional expansion.",
		},
		{
			ID:                PhaseLondonCore,
			Category:          PhaseCategorySingle,
			Market:            "London",
			Timezone:          "Europe/London",
			StartHour:         11,
			EndHour:           13,
			Liquidity:         LiquidityHigh,
			BreakoutBias:      0.70,
			MeanReversionBias: 0.10,
			Tags:              []string{"london", "core", "continuation"},
			Notes:             "London core trend continuation.",
		},
		{
			ID:                PhaseLondonLate,
			Category:          PhaseCategorySingle,
			Market:            "London",
			Timezone:          "Europe/London",
			StartHour:         13,
			EndHour:           16,
			Liquidity:         LiquidityHigh,
			BreakoutBias:      0.60,
			MeanReversionBias: 0.10,
			Tags:              []string{"london", "late", "pre_ny"},
			Notes:             "London late session and pre-NY alignment.",
		},
		{
			ID:                PhasePreNewYork,
			Category:          PhaseCategorySingle,
			Market:            "New York",
			Timezone:          "America/New_York",
			StartHour:         7,
			EndHour:           8,
			Liquidity:         LiquidityMedium,
			BreakoutBias:      0.50,
			MeanReversionBias: 0.10,
			Tags:              []string{"newyork", "pre_open", "prep"},
			Notes:             "US preparation before core cash flow.",
		},
		{
			ID:                PhaseNewYorkCore,
			Category:          PhaseCategorySingle,
			Market:            "New York",
			Timezone:          "America/New_York",
			StartHour:         8,
			EndHour:           11,
			Liquidity:         LiquidityHigh,
			BreakoutBias:      0.80,
			MeanReversionBias: 0.10,
			Tags:              []string{"newyork", "core", "macro"},
			Notes:             "US core market flow and macro linkage.",
		},
		{
			ID:                PhaseLateNewYork,
			Category:          PhaseCategorySingle,
			Market:            "New York",
			Timezone:          "America/New_York",
			StartHour:         11,
			EndHour:           16,
			Liquidity:         LiquidityMedium,
			BreakoutBias:      0.20,
			MeanReversionBias: 0.50,
			Tags:              []string{"newyork", "late", "unwind"},
			Notes:             "NY late session and position reduction.",
		},
		{
			ID:                PhaseSydneyReset,
			Category:          PhaseCategorySingle,
			Market:            "Sydney",
			Timezone:          "Australia/Sydney",
			StartHour:         7,
			EndHour:           9,
			Liquidity:         LiquidityLow,
			BreakoutBias:      0.10,
			MeanReversionBias: 0.10,
			Tags:              []string{"sydney", "reset", "thin_liquidity"},
			Notes:             "Low-liquidity reset before Tokyo buildup.",
		},
	}
}
