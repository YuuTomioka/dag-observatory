package marketphase

import (
	"fmt"
	"sync"
	"time"
)

var locationCache sync.Map

func ResolveSinglePhases(at time.Time, specs []PhaseSpec) ([]ResolvedPhase, error) {
	if at.IsZero() {
		return nil, fmt.Errorf("marketphase resolve singles: timestamp is required")
	}
	at = at.UTC()
	jst, err := loadLocation("Asia/Tokyo")
	if err != nil {
		return nil, err
	}

	out := make([]ResolvedPhase, 0)
	for _, spec := range specs {
		resolved, ok, err := resolveSinglePhase(at, spec, jst)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, resolved)
		}
	}
	return out, nil
}

func resolveSinglePhase(at time.Time, spec PhaseSpec, jst *time.Location) (ResolvedPhase, bool, error) {
	if spec.ID == "" {
		return ResolvedPhase{}, false, fmt.Errorf("marketphase resolve single: phase id is required")
	}
	if spec.Timezone == "" {
		return ResolvedPhase{}, false, fmt.Errorf("marketphase resolve single %q: timezone is required", spec.ID)
	}
	if spec.EndHour < spec.StartHour || (spec.EndHour == spec.StartHour && spec.EndMinute <= spec.StartMinute) {
		return ResolvedPhase{}, false, fmt.Errorf("marketphase resolve single %q: end must be after start", spec.ID)
	}

	loc, err := loadLocation(spec.Timezone)
	if err != nil {
		return ResolvedPhase{}, false, fmt.Errorf("marketphase resolve single %q: %w", spec.ID, err)
	}
	localNow := at.In(loc)
	localStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), spec.StartHour, spec.StartMinute, 0, 0, loc)
	localEnd := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), spec.EndHour, spec.EndMinute, 0, 0, loc)
	if localNow.Before(localStart) || !localNow.Before(localEnd) {
		return ResolvedPhase{}, false, nil
	}

	utcStart := localStart.UTC()
	utcEnd := localEnd.UTC()
	return ResolvedPhase{
		ID:         spec.ID,
		Market:     spec.Market,
		Timezone:   spec.Timezone,
		Active:     true,
		LocalNow:   localNow,
		LocalStart: localStart,
		LocalEnd:   localEnd,
		UTCStart:   utcStart,
		UTCEnd:     utcEnd,
		JSTStart:   utcStart.In(jst),
		JSTEnd:     utcEnd.In(jst),
	}, true, nil
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
