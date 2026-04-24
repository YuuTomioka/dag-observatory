package marketcontext

import (
	"time"

	"dag-observatory/dag-core/internal/domain/marketphase"
)

func Build(at time.Time, singles []marketphase.ResolvedPhase, signals []SignalID) Context {
	at = at.UTC()
	phaseCodes := make([]marketphase.PhaseCode, 0, len(singles))
	dominantMarkets := make([]string, 0, len(singles))
	seenMarkets := make(map[string]struct{}, len(singles))
	tags := make([]string, 0)
	seenTags := make(map[string]struct{})

	specByID := make(map[marketphase.PhaseCode]marketphase.PhaseSpec, len(marketphase.DefaultPhaseSpecs()))
	for _, spec := range marketphase.DefaultPhaseSpecs() {
		specByID[spec.ID] = spec
	}

	for _, single := range singles {
		phaseCodes = append(phaseCodes, single.ID)
		if _, ok := seenMarkets[single.Market]; !ok && single.Market != "" {
			seenMarkets[single.Market] = struct{}{}
			dominantMarkets = append(dominantMarkets, single.Market)
		}
		if spec, ok := specByID[single.ID]; ok {
			tags = appendUniqueStrings(tags, seenTags, spec.Tags)
		}
	}

	for _, id := range signals {
		src, ok := signalTags[id]
		if !ok {
			continue
		}
		tags = appendUniqueStrings(tags, seenTags, src)
	}

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	return Context{
		NowUTC:          at,
		NowJST:          at.In(jst),
		ActivePhases:    phaseCodes,
		ActiveSignals:   append([]SignalID(nil), signals...),
		DominantMarkets: dominantMarkets,
		Tags:            tags,
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
