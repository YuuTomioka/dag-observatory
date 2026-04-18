package marketphase

type phaseScore struct {
	Liquidity     float64
	Breakout      float64
	MeanReversion float64
	Tags          []string
}

var derivedScores = map[PhaseID]phaseScore{
	PhaseTokyoLondonTransition: {
		Liquidity:     0.75,
		Breakout:      0.70,
		MeanReversion: 0.10,
		Tags:          []string{"transition", "tokyo", "london"},
	},
	PhaseLondonNewYorkOverlap: {
		Liquidity:     1.00,
		Breakout:      1.00,
		MeanReversion: 0.10,
		Tags:          []string{"overlap", "london", "newyork", "high_liquidity"},
	},
	PhaseNewYorkCloseTransition: {
		Liquidity:     0.20,
		Breakout:      0.20,
		MeanReversion: 0.60,
		Tags:          []string{"transition", "newyork", "close", "unwind"},
	},
}

func scoreForSpec(spec PhaseSpec) phaseScore {
	return phaseScore{
		Liquidity:     liquidityScore(spec.Liquidity),
		Breakout:      clamp01(spec.BreakoutBias),
		MeanReversion: clamp01(spec.MeanReversionBias),
		Tags:          append([]string(nil), spec.Tags...),
	}
}

func liquidityScore(level LiquidityLevel) float64 {
	switch level {
	case LiquidityLow:
		return 0.25
	case LiquidityMedium:
		return 0.50
	case LiquidityHigh:
		return 0.75
	case LiquidityVeryHigh:
		return 1.00
	default:
		return 0
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
