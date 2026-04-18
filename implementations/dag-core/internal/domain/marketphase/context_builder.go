package marketphase

import "time"

func BuildMarketContext(at time.Time, singles []ResolvedPhase, derived []PhaseID) MarketContext {
	at = at.UTC()
	singleIDs := make([]PhaseID, 0, len(singles))
	dominantMarkets := make([]string, 0, len(singles))
	seenMarkets := make(map[string]struct{}, len(singles))
	tags := make([]string, 0)
	seenTags := make(map[string]struct{})

	specByID := make(map[PhaseID]PhaseSpec, len(DefaultPhaseSpecs()))
	for _, spec := range DefaultPhaseSpecs() {
		specByID[spec.ID] = spec
	}

	var liquidityScore float64
	var breakoutScore float64
	var meanReversionScore float64

	for _, single := range singles {
		singleIDs = append(singleIDs, single.ID)
		if _, ok := seenMarkets[single.Market]; !ok && single.Market != "" {
			seenMarkets[single.Market] = struct{}{}
			dominantMarkets = append(dominantMarkets, single.Market)
		}
		if spec, ok := specByID[single.ID]; ok {
			score := scoreForSpec(spec)
			liquidityScore = maxScore(liquidityScore, score.Liquidity)
			breakoutScore = maxScore(breakoutScore, score.Breakout)
			meanReversionScore = maxScore(meanReversionScore, score.MeanReversion)
			tags = appendUniqueStrings(tags, seenTags, score.Tags)
		}
	}

	for _, id := range derived {
		score, ok := derivedScores[id]
		if !ok {
			continue
		}
		liquidityScore = maxScore(liquidityScore, score.Liquidity)
		breakoutScore = maxScore(breakoutScore, score.Breakout)
		meanReversionScore = maxScore(meanReversionScore, score.MeanReversion)
		tags = appendUniqueStrings(tags, seenTags, score.Tags)
	}

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	return MarketContext{
		NowUTC:              at,
		NowJST:              at.In(jst),
		ActiveSinglePhases:  singleIDs,
		ActiveDerivedPhases: append([]PhaseID(nil), derived...),
		DominantMarkets:     dominantMarkets,
		LiquidityScore:      liquidityScore,
		BreakoutScore:       breakoutScore,
		MeanReversionScore:  meanReversionScore,
		Tags:                tags,
	}
}

func appendUniqueStrings(dst []string, seen map[string]struct{}, src []string) []string {
	for _, v := range src {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		dst = append(dst, v)
	}
	return dst
}

func maxScore(a, b float64) float64 {
	if b > a {
		return b
	}
	return a
}
