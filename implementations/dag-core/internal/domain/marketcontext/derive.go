package marketcontext

import (
	"fmt"
	"sync"
	"time"

	"dag-observatory/dag-core/internal/domain/marketphase"
)

var locationCache sync.Map

func ResolveSignals(activeSingles []marketphase.ResolvedPhase) []SignalID {
	byID := make(map[marketphase.PhaseCode]marketphase.ResolvedPhase, len(activeSingles))
	for _, phase := range activeSingles {
		byID[phase.ID] = phase
	}

	signals := make([]SignalID, 0, 3)
	if lateTokyo, ok := byID[marketphase.PhaseLateTokyo]; ok && londonIsEnteringOrActive(lateTokyo) {
		signals = append(signals, SignalTokyoLondonTransition)
	}
	if (hasPhase(byID, marketphase.PhaseLondonCore) || hasPhase(byID, marketphase.PhaseLondonLate)) &&
		(hasPhase(byID, marketphase.PhasePreNewYork) || hasPhase(byID, marketphase.PhaseNewYorkCore)) {
		signals = append(signals, SignalLondonNewYorkOverlap)
	}
	if hasPhase(byID, marketphase.PhaseLateNewYork) &&
		!hasPhase(byID, marketphase.PhaseSydneyReset) &&
		!hasPhase(byID, marketphase.PhasePreTokyo) &&
		!hasPhase(byID, marketphase.PhaseTokyoCore) &&
		!hasPhase(byID, marketphase.PhaseLateTokyo) &&
		!hasPhase(byID, marketphase.PhaseLondonEarly) &&
		!hasPhase(byID, marketphase.PhaseLondonCore) &&
		!hasPhase(byID, marketphase.PhaseLondonLate) &&
		lastLocalHour(byID[marketphase.PhaseLateNewYork]) >= 15 {
		signals = append(signals, SignalNewYorkCloseTransition)
	}
	return signals
}

func hasPhase(active map[marketphase.PhaseCode]marketphase.ResolvedPhase, id marketphase.PhaseCode) bool {
	_, ok := active[id]
	return ok
}

func lastLocalHour(phase marketphase.ResolvedPhase) int {
	if phase.LocalNow.IsZero() {
		return 0
	}
	return phase.LocalNow.Hour()
}

func londonIsEnteringOrActive(phase marketphase.ResolvedPhase) bool {
	if phase.LocalNow.IsZero() {
		return false
	}
	loc, err := loadLocation("Europe/London")
	if err != nil {
		return false
	}
	londonNow := phase.LocalNow.In(loc)
	start := time.Date(londonNow.Year(), londonNow.Month(), londonNow.Day(), 7, 0, 0, 0, loc)
	end := time.Date(londonNow.Year(), londonNow.Month(), londonNow.Day(), 13, 0, 0, 0, loc)
	return !londonNow.Before(start) && londonNow.Before(end)
}

func loadLocation(name string) (*time.Location, error) {
	if v, ok := locationCache.Load(name); ok {
		return v.(*time.Location), nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", name, err)
	}
	actual, _ := locationCache.LoadOrStore(name, loc)
	return actual.(*time.Location), nil
}

func Resolve(at time.Time, specs []marketphase.PhaseSpec) (Context, error) {
	singles, err := marketphase.ResolveSinglePhases(at, specs)
	if err != nil {
		return Context{}, err
	}
	signals := ResolveSignals(singles)
	return Build(at, singles, signals), nil
}
