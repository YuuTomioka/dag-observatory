package marketphase

import "time"

type MarketContext struct {
	NowUTC              time.Time
	NowJST              time.Time
	ActiveSinglePhases  []PhaseID
	ActiveDerivedPhases []PhaseID
	DominantMarkets     []string
	LiquidityScore      float64
	BreakoutScore       float64
	MeanReversionScore  float64
	Tags                []string
}
