package marketcontext

import (
	"time"

	"dag-observatory/dag-core/internal/domain/marketphase"
)

type SignalID string

const (
	SignalTokyoLondonTransition  SignalID = "tokyo_london_transition"
	SignalLondonNewYorkOverlap   SignalID = "london_newyork_overlap"
	SignalNewYorkCloseTransition SignalID = "newyork_close_transition"
)

type Context struct {
	NowUTC          time.Time
	NowJST          time.Time
	ActivePhases    []marketphase.PhaseCode
	ActiveSignals   []SignalID
	DominantMarkets []string
	Tags            []string
}
