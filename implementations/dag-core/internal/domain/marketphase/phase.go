package marketphase

type PhaseID string
type PhaseCategory string
type LiquidityLevel string

const (
	PhaseCategorySingle  PhaseCategory = "single"
	PhaseCategoryDerived PhaseCategory = "derived"
)

const (
	LiquidityLow      LiquidityLevel = "low"
	LiquidityMedium   LiquidityLevel = "medium"
	LiquidityHigh     LiquidityLevel = "high"
	LiquidityVeryHigh LiquidityLevel = "very_high"
)

const (
	PhasePreTokyo               PhaseID = "pre_tokyo"
	PhaseTokyoCore              PhaseID = "tokyo_core"
	PhaseLateTokyo              PhaseID = "late_tokyo"
	PhaseLondonEarly            PhaseID = "london_early"
	PhaseLondonCore             PhaseID = "london_core"
	PhaseLondonLate             PhaseID = "london_late"
	PhasePreNewYork             PhaseID = "pre_newyork"
	PhaseNewYorkCore            PhaseID = "newyork_core"
	PhaseLateNewYork            PhaseID = "late_newyork"
	PhaseSydneyReset            PhaseID = "sydney_reset"
	PhaseTokyoLondonTransition  PhaseID = "tokyo_london_transition"
	PhaseLondonNewYorkOverlap   PhaseID = "london_newyork_overlap"
	PhaseNewYorkCloseTransition PhaseID = "newyork_close_transition"
)
