package marketphase

import "time"

func ResolveDerivedPhases(activeSingles []ResolvedPhase) []PhaseID {
	byID := make(map[PhaseID]ResolvedPhase, len(activeSingles))
	for _, phase := range activeSingles {
		byID[phase.ID] = phase
	}

	derived := make([]PhaseID, 0, 3)
	if lateTokyo, ok := byID[PhaseLateTokyo]; ok && londonIsEnteringOrEarly(lateTokyo) {
		derived = append(derived, PhaseTokyoLondonTransition)
	}
	if (hasPhase(byID, PhaseLondonCore) || hasPhase(byID, PhaseLondonLate)) &&
		(hasPhase(byID, PhasePreNewYork) || hasPhase(byID, PhaseNewYorkCore)) {
		derived = append(derived, PhaseLondonNewYorkOverlap)
	}
	if hasPhase(byID, PhaseLateNewYork) &&
		!hasPhase(byID, PhaseSydneyReset) &&
		!hasPhase(byID, PhasePreTokyo) &&
		!hasPhase(byID, PhaseTokyoCore) &&
		!hasPhase(byID, PhaseLateTokyo) &&
		!hasPhase(byID, PhaseLondonEarly) &&
		!hasPhase(byID, PhaseLondonCore) &&
		!hasPhase(byID, PhaseLondonLate) &&
		lastLocalHour(byID[PhaseLateNewYork]) >= 15 {
		derived = append(derived, PhaseNewYorkCloseTransition)
	}
	return derived
}

func hasPhase(active map[PhaseID]ResolvedPhase, id PhaseID) bool {
	_, ok := active[id]
	return ok
}

func lastLocalHour(phase ResolvedPhase) int {
	if phase.LocalNow.IsZero() {
		return 0
	}
	return phase.LocalNow.Hour()
}

func londonIsEnteringOrEarly(phase ResolvedPhase) bool {
	if phase.LocalNow.IsZero() {
		return false
	}
	loc, err := loadLocation("Europe/London")
	if err != nil {
		return false
	}
	londonNow := phase.LocalNow.In(loc)
	start := time.Date(londonNow.Year(), londonNow.Month(), londonNow.Day(), 7, 0, 0, 0, loc)
	end := time.Date(londonNow.Year(), londonNow.Month(), londonNow.Day(), 11, 0, 0, 0, loc)
	return !londonNow.Before(start) && londonNow.Before(end)
}
